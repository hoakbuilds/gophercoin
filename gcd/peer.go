package gcd

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

// Peer is the structure that defines a peer
// in the peer to peer network
type Peer struct {
	// address defines the peer's IP address
	address string
	// version defines the peer's best block height
	version int64
}

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

// Version structure specifies the structure used
// to represent the Version message passed between
// nodes
type Version struct {
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

func extractCommand(request []byte) []byte {
	return request[:commandLength]
}

func (s *gcdServer) requestBlocks() {
	for _, node := range s.knownNodes {
		s.sendGetBlocks(node.address)
	}
}

func (s *gcdServer) sendAddr(address string) {
	nodes := addr{}

	for _, node := range s.knownNodes {
		nodes.AddrList = append(nodes.AddrList, node.address)
	}

	payload := gobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)

	s.sendData(address, request)
}

func (s *gcdServer) sendBlock(addr string, b *Block) {
	serBlock, err := b.SerializeBlock()
	if err != nil {
		log.Panicf("err: %v", err)
	}
	data := block{s.nodeAddress, serBlock}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	s.sendData(addr, request)
}

func (s *gcdServer) sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		log.Printf("%s is not available\n", addr)
		var updatedNodes []Peer

		for _, node := range s.knownNodes {
			if node.address != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		s.knownNodes = updatedNodes

		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func (s *gcdServer) sendInv(address, kind string, items [][]byte) {
	inventory := inv{s.nodeAddress, kind, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	s.sendData(address, request)
}

func (s *gcdServer) sendGetBlocks(address string) {
	payload := gobEncode(getblocks{s.nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)

	s.sendData(address, request)
}

func (s *gcdServer) sendGetData(address, kind string, id []byte) {
	payload := gobEncode(getdata{s.nodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	s.sendData(address, request)
}

func (s *gcdServer) sendTx(addr string, tnx *Transaction) {
	data := tx{s.nodeAddress, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	s.sendData(addr, request)
}

func (s *gcdServer) sendVersion(addr string, bc *Blockchain) {
	bestHeight := bc.GetBestHeight()
	log.Printf("best height: %d \n", bestHeight)
	version := Version{nodeVersion, bestHeight, s.nodeAddress}
	payload := gobEncode(version)
	log.Printf("sending payload: %+v\n", bytesToCommand(payload))
	request := append(commandToBytes("version"), payload...)

	s.sendData(addr, request)
}

func (s *gcdServer) handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	var updatedNodes []Peer
	updatedNodes = s.knownNodes
	for _, node := range s.knownNodes {

		for _, addr := range payload.AddrList {
			if node.address != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		s.knownNodes = updatedNodes
	}

	s.knownNodes = updatedNodes
	log.Printf("There are %d known nodes now!\n", len(s.knownNodes))

	s.requestBlocks()
}

func (s *gcdServer) handleBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block, err := DeserializeBlock(blockData)
	if err != nil {
		log.Print(err)
	}
	fmt.Println("Recevied a new block!")
	bc.AddBlock(block)

	log.Printf("Added block %x\n", block.Hash)

	if len(s.blocksInTransit) > 0 {
		blockHash := s.blocksInTransit[0]
		s.sendGetData(payload.AddrFrom, "block", blockHash)

		s.blocksInTransit = s.blocksInTransit[1:]
	} else {
		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()
	}
}

func (s *gcdServer) handleInv(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

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

		if s.memPool[hex.EncodeToString(txID)].ID == nil {
			s.sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

func (s *gcdServer) handleGetBlocks(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := bc.GetBlockHashes()
	s.sendInv(payload.AddrFrom, "block", blocks)
}

func (s *gcdServer) handleGetData(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		s.sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := s.memPool[txID]

		s.sendTx(payload.AddrFrom, &tx)
		// delete(memPool, txID)
	}
}

func (s *gcdServer) handleTx(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := DeserializeTransaction(txData)
	s.memPool[hex.EncodeToString(tx.ID)] = tx

	if s.nodeAddress == s.knownNodes[0].address {
		for _, node := range s.knownNodes {
			if node.address != s.nodeAddress && node.address != payload.AddFrom {
				s.sendInv(node.address, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(s.memPool) >= 2 && len(s.miningAddress) > 0 {
		MineTransactions:
			var txs []*Transaction

			for id := range s.memPool {
				tx := s.memPool[id]
				log.Printf("verifying transaction: %s\n", id)
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return
			}

			cbTx := NewCoinbaseTX(s.miningAddress, "")
			txs = append(txs, cbTx)

			newBlock := bc.MineBlock(txs)
			UTXOSet := UTXOSet{bc}
			UTXOSet.Reindex()

			fmt.Println("New block is mined!")

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(s.memPool, txID)
			}

			for _, node := range s.knownNodes {
				if node.address != s.nodeAddress {
					s.sendInv(node.address, "block", [][]byte{newBlock.Hash})
				}
			}

			if len(s.memPool) > 0 {
				goto MineTransactions
			}
		}
	}
}

func (s *gcdServer) handleVersion(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload Version

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	log.Printf("best height: %d \tpeer %s best height: %d\n", myBestHeight, payload.AddrFrom, foreignerBestHeight)
	if myBestHeight < foreignerBestHeight {

		log.Printf("sending getblocks message to %s\n", s.knownNodes[0])
		s.sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {

		log.Printf("sending version message to %s\n", s.knownNodes[0])
		s.sendVersion(payload.AddrFrom, bc)
	}

	// sendAddr(payload.AddrFrom)
	if !s.nodeIsKnown(payload.AddrFrom) {
		log.Printf("node %s is unknown, adding to peer list\n", payload.AddrFrom)
		s.knownNodes = append(s.knownNodes, Peer{address: payload.AddrFrom})
	}
}

func (s *gcdServer) handleConnection(conn net.Conn, bc *Blockchain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])
	log.Printf("Received %s command\n", command)

	switch command {
	case "addr":
		s.handleAddr(request)
	case "block":
		s.handleBlock(request, bc)
	case "inv":
		s.handleInv(request, bc)
	case "getblocks":
		s.handleGetBlocks(request, bc)
	case "getdata":
		s.handleGetData(request, bc)
	case "tx":
		s.handleTx(request, bc)
	case "version":
		s.handleVersion(request, bc)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()

	s.wg.Done()
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func (s *gcdServer) nodeIsKnown(addr string) bool {
	for _, node := range s.knownNodes {
		if node.address == addr {
			return true
		}
	}

	return false
}
