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
	Address string `json:"Address"`
	// version defines the peer's best block height
	Version int64 `json:"Version"`
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

func (s *Server) requestBlocks() {
	for _, node := range s.knownNodes {
		s.sendGetBlocks(node.Address)
	}
}

func (s *Server) sendAddr(address string) {
	nodes := addr{}

	for _, node := range s.knownNodes {
		nodes.AddrList = append(nodes.AddrList, node.Address)
	}

	payload := gobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)

	s.sendData(address, request)
}

func (s *Server) sendBlock(addr string, b *Block) {
	serBlock, err := b.SerializeBlock()
	if err != nil {
		log.Panicf("err: %v", err)
	}
	data := block{s.nodeAddress, serBlock}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	s.sendData(addr, request)
}

func (s *Server) sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		log.Printf("[PRSRV] %s is not available\n", addr)
		var updatedNodes []Peer

		for _, node := range s.knownNodes {
			if node.Address != addr {
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

func (s *Server) sendInv(address, kind string, items [][]byte) {
	inventory := inv{s.nodeAddress, kind, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	s.sendData(address, request)
}

func (s *Server) sendGetBlocks(address string) {
	payload := gobEncode(getblocks{s.nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)

	s.sendData(address, request)
}

func (s *Server) sendGetData(address, kind string, id []byte) {
	payload := gobEncode(getdata{s.nodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	s.sendData(address, request)
}

func (s *Server) sendTx(addr string, tnx *Transaction) {
	data := tx{s.nodeAddress, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	s.sendData(addr, request)
}

func (s *Server) sendVersion(addr string) {
	var bestHeight int
	if s.db != nil {
		bestHeight = s.db.GetBestHeight()
	} else {
		bestHeight = 0
	}
	log.Printf("[PRSRV] Best height: %d \n", bestHeight)
	version := Version{nodeVersion, bestHeight, s.nodeAddress}
	payload := gobEncode(version)
	log.Printf("[PRSRV] Sending payload:\n%+v\n-------------\n",
		bytesToCommand(payload))
	request := append(commandToBytes("version"), payload...)

	s.sendData(addr, request)
}

func (s *Server) handleAddr(request []byte) {
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
			if node.Address != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		s.knownNodes = updatedNodes
	}

	s.knownNodes = updatedNodes
	log.Printf("[PRSRV] There are %d known nodes now!\n", len(s.knownNodes))

	s.requestBlocks()
}

func (s *Server) handleBlock(request []byte) {

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
	log.Printf("[PRSRV] Received a new block!\n%+v\v", block)

	if s.db != nil {
		s.db.AddBlock(block)
	} else {
		db, err := CreateBlockchain("")
		if err != nil {
			log.Printf("[PRSRV] Failed to create db: %v", err)
			return
		}
		s.db = &db
		s.db.AddGenesis(block)
	}

	log.Printf("[PRSRV] Added block %x\n", block.Hash)

	if len(s.blocksInTransit) > 0 {
		blockHash := s.blocksInTransit[0]
		s.sendGetData(payload.AddrFrom, "block", blockHash)

		s.blocksInTransit = s.blocksInTransit[1:]
	} else {
		UTXOSet := UTXOSet{s.db}
		UTXOSet.Reindex()
	}
}

func (s *Server) handleInv(request []byte) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("[PRSRV] Received inventory with %d %s\n", len(payload.Items), payload.Type)

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

func (s *Server) handleGetBlocks(request []byte) {

	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := s.db.GetBlockHashes()
	s.sendInv(payload.AddrFrom, "block", blocks)
}

func (s *Server) handleGetData(request []byte) {

	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := s.db.GetBlock([]byte(payload.ID))
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

func (s *Server) handleTx(request []byte) {

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

	s.minerChan <- txData

	if len(s.knownNodes) > 0 {
		for _, node := range s.knownNodes {
			if node.Address != s.nodeAddress && node.Address != payload.AddFrom {
				s.sendInv(node.Address, "tx", [][]byte{tx.ID})
			}
		}
	}

}

func (s *Server) handleVersion(request []byte) {
	var buff bytes.Buffer
	var payload Version

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	// sendAddr(payload.AddrFrom)
	if !s.nodeIsKnown(payload.AddrFrom) {
		log.Printf("[PRSRV] Node %s is unknown, adding to peer list\n", payload.AddrFrom)
		s.knownNodes = append(s.knownNodes, Peer{Address: payload.AddrFrom})
	}

	if s.db == nil && payload.BestHeight != 0 {
		log.Printf("[PRSRV] Sending getblocks message to %v\n", payload.AddrFrom)
		s.sendGetBlocks(payload.AddrFrom)
		return
	}
	var myBestHeight int
	if s.db != nil {
		myBestHeight = s.db.GetBestHeight()
	} else {
		myBestHeight = 0
	}
	foreignerBestHeight := payload.BestHeight

	log.Printf("[PRSRV] My best height: %d \tPeer %s best height: %d\n", myBestHeight, payload.AddrFrom, foreignerBestHeight)

	if myBestHeight < foreignerBestHeight {
		log.Printf("[PRSRV] Sending getblocks message to %v\n", payload.AddrFrom)
		s.sendGetBlocks(payload.AddrFrom)
		return
	} else if myBestHeight > foreignerBestHeight {
		log.Printf("[PRSRV] Sending version message to %v\n", payload.AddrFrom)
		s.sendVersion(payload.AddrFrom)
		return
	}

}

func (s *Server) handleConnection(conn net.Conn) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])
	log.Printf("[PRSRV] Received %s command\n", command)

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
		log.Println("[PRSRV] Unknown command!")
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

func (s *Server) nodeIsKnown(addr string) bool {
	for _, node := range s.knownNodes {
		if node.Address == addr {
			return true
		}
	}

	return false
}
