package peer

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/murlokito/gophercoin/blockchain"
	"github.com/murlokito/gophercoin/transaction"
	"io"
	"io/ioutil"
	"net"
	"sync"
)

type addr struct {
	AddrList []string
}

type block struct {
	AddrFrom string
	Block    []byte
}

type getblocks struct {
	AddrFrom string
}

type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type tx struct {
	AddFrom     string
	Transaction []byte
}

type version struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func (s PeerServer) handleConnection(conn net.Conn) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		s.logger.WithError(err)
		return
	}
	command := bytesToCommand(request[:commandLength])
	s.logger.Info("Received %s command\n", command)

	switch command {
	case "addr":
		s.handleAddr(request)
	case "block":
		s.handleBlock(request)
	case "inv":
		s.handleInv(request)
	case "getblocks":
		s.handleGetBlocks(request)
	case "getdata":
		s.handleGetData(request)
	case "tx":
		s.handleTx(request)
	case "version":
		s.handleVersion(request)
	default:
		s.logger.Info("Unknown command received, ignoring.")
	}

	conn.Close()

	s.wg.Done()
}

func gobEncode(data interface{}) ([]byte, error) {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		return []byte{}, err
	}

	return buff.Bytes(), nil
}

func extractCommand(request []byte) []byte {
	return request[:commandLength]
}

func (s PeerServer) requestBlocks() {
	for _, node := range s.KnownNodes {
		s.sendGetBlocks(node.Address)
	}
}

func (s PeerServer) sendAddr(address string) {
	nodes := addr{}

	for _, node := range s.KnownNodes {
		nodes.AddrList = append(nodes.AddrList, node.Address)
	}

	payload, err := gobEncode(nodes)
	if err != nil {
		s.logger.WithError(err)
		return
	}
	request := append(commandToBytes("addr"), payload...)

	s.sendData(address, request)
}

func (s PeerServer) sendBlock(addr string, b *blockchain.Block) {
	serBlock, err := b.SerializeBlock()
	if err != nil {
		s.logger.WithError(err)
	}
	data := block{s.NodeAddress, serBlock}
	payload, err := gobEncode(data)
	if err != nil {
		s.logger.WithError(err)
		return
	}
	request := append(commandToBytes("block"), payload...)

	s.sendData(addr, request)
}

func (s PeerServer) sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	defer conn.Close()
	if err != nil {
		s.logger.Info("%s is not available\n", addr)
		var updatedNodes []Peer

		for _, node := range s.KnownNodes {
			if node.Address != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		s.KnownNodes = updatedNodes

		return
	}

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		s.logger.WithError(err)
		return
	}
}

func (s PeerServer) SendInv(address, kind string, items [][]byte) {
	inventory := inv{s.NodeAddress, kind, items}
	payload, err := gobEncode(inventory)
	if err != nil {
		s.logger.WithError(err)
		return
	}
	request := append(commandToBytes("inv"), payload...)

	s.sendData(address, request)
}

func (s PeerServer) sendGetBlocks(address string) {
	payload, err := gobEncode(getblocks{s.NodeAddress})
	if err != nil {
		s.logger.WithError(err)
		return
	}
	request := append(commandToBytes("getblocks"), payload...)

	s.sendData(address, request)
}

func (s PeerServer) sendGetData(address, kind string, id []byte) {
	payload, err := gobEncode(getdata{s.NodeAddress, kind, id})
	if err != nil {
		s.logger.WithError(err)
		return
	}
	request := append(commandToBytes("getdata"), payload...)

	s.sendData(address, request)
}

func (s PeerServer) sendTx(addr string, tnx *transaction.Transaction) {
	data := tx{s.NodeAddress, tnx.Serialize()}
	payload, err := gobEncode(data)
	if err != nil {
		s.logger.WithError(err)
		return
	}
	request := append(commandToBytes("tx"), payload...)

	s.sendData(addr, request)
}

func (s PeerServer) SendVersion(addr string) {
	var bestHeight int
	if s.chainMgr.Chain != nil {
		bestHeight = s.chainMgr.Chain.GetBestHeight()
	} else {
		bestHeight = -1
	}
	s.logger.Info("Best height: %d \n", bestHeight)
	version := version{nodeVersion, bestHeight, s.NodeAddress}
	payload, err := gobEncode(version)
	if err != nil {
		s.logger.WithError(err)
		return
	}
	s.logger.Info("Sending payload:\n%+v\n-------------\n",
		bytesToCommand(payload))
	request := append(commandToBytes("version"), payload...)

	s.sendData(addr, request)
}

func (s PeerServer) handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		s.logger.WithError(err)
		return
	}
	var updatedNodes []Peer
	updatedNodes = s.KnownNodes
	for _, node := range s.KnownNodes {

		for _, addr := range payload.AddrList {
			if node.Address != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		s.KnownNodes = updatedNodes
	}

	s.KnownNodes = updatedNodes
	s.logger.Info("There are %d known nodes now!\n", len(s.KnownNodes))

	s.requestBlocks()
}

func (s PeerServer) handleBlock(request []byte) {

	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		s.logger.WithError(err)
		return
	}

	blockData := payload.Block
	block, err := blockchain.DeserializeBlock(blockData)
	if err != nil {
		s.logger.WithError(err)
		return
	}
	s.logger.Info("Received a new block!\n%+v\v", block)

	if s.chainMgr.Chain != nil {
		s.chainMgr.Chain.AddBlock(block)
	} else {
		db, err := blockchain.CreateBlockchain("")
		if err != nil {
			s.logger.Info("Failed to create db: %v", err)
			return
		}
		s.chainMgr.Chain = db
		s.chainMgr.Chain.AddGenesis(block)
	}

	s.logger.Info("Added block %x\n", block.Hash)

	if len(s.blocksInTransit) > 0 {
		blockHash := s.blocksInTransit[0]
		s.sendGetData(payload.AddrFrom, "block", blockHash)

		s.blocksInTransit = s.blocksInTransit[1:]
	} else {
		UTXOSet := blockchain.UTXOSet{s.chainMgr.Chain, &sync.RWMutex{}}
		UTXOSet.Reindex()
	}
}

func (s PeerServer) handleInv(request []byte) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		s.logger.WithError(err)
		return
	}

	s.logger.Info("Received inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit := payload.Items

		blockHash := payload.Items[0]
		s.sendGetData(payload.AddrFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		s.blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if s.chainMgr.MemPool[hex.EncodeToString(txID)].ID == nil {
			s.sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

func (s PeerServer) handleGetBlocks(request []byte) {

	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		s.logger.WithError(err)
		return
	}

	blocks := s.chainMgr.Chain.GetBlockHashes()
	s.SendInv(payload.AddrFrom, "block", blocks)
}

func (s PeerServer) handleGetData(request []byte) {

	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		s.logger.WithError(err)
		return
	}

	if payload.Type == "block" {
		block, err := s.chainMgr.Chain.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		s.sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := s.chainMgr.MemPool[txID]

		s.sendTx(payload.AddrFrom, &tx)
		// delete(memPool, txID)
	}
}

func (s PeerServer) handleTx(request []byte) {

	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		s.logger.WithError(err)
		return
	}

	txData := payload.Transaction
	tx := transaction.DeserializeTransaction(txData)
	s.chainMgr.MemPool[hex.EncodeToString(tx.ID)] = tx

	if s.MinerChan != nil {
		s.MinerChan <- txData
	}

	if len(s.KnownNodes) > 0 {
		for _, node := range s.KnownNodes {
			if node.Address != s.NodeAddress && node.Address != payload.AddFrom {
				s.SendInv(node.Address, "tx", [][]byte{tx.ID})
			}
		}
	}

}

func (s PeerServer) handleVersion(request []byte) {
	var buff bytes.Buffer
	var payload version

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		s.logger.WithError(err)
		return
	}

	// sendAddr(payload.AddrFrom)
	if !s.nodeIsKnown(payload.AddrFrom) {
		s.logger.Info("Node %s is unknown, adding to peer list\n", payload.AddrFrom)
		s.KnownNodes = append(s.KnownNodes, Peer{Address: payload.AddrFrom})
	}

	if s.chainMgr.Chain == nil && payload.BestHeight != 0 {
		s.logger.Info("Sending getblocks message to %v\n", payload.AddrFrom)
		s.sendGetBlocks(payload.AddrFrom)
		return
	}
	var myBestHeight int
	if s.chainMgr.Chain != nil {
		s.logger.Info("Getting best block height from db.")
		myBestHeight = s.chainMgr.Chain.GetBestHeight()
	} else {
		s.logger.Info("Database not found, best height is 0.")
		myBestHeight = -1
	}
	foreignerBestHeight := payload.BestHeight

	s.logger.Info("My best height: %d \tPeer %s best height: %d\n", myBestHeight, payload.AddrFrom, foreignerBestHeight)

	if myBestHeight < foreignerBestHeight {
		s.logger.Info("Sending getblocks message to %v\n", payload.AddrFrom)
		s.sendGetBlocks(payload.AddrFrom)
		return
	} else if myBestHeight > foreignerBestHeight {
		s.logger.Info("Sending version message to %v\n", payload.AddrFrom)
		s.SendVersion(payload.AddrFrom)
		return
	}

}

func (s PeerServer) nodeIsKnown(addr string) bool {
	for _, node := range s.KnownNodes {
		if node.Address == addr {
			return true
		}
	}

	return false
}
