package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

// Block is the unit structure of
// a blockchain. It will store information
// which will be hashed, the hash of the
// previous block and the time of its creation
type Block struct {
	timestamp     int64
	info          []byte
	prevBlockHash []byte
	hash          []byte
	counter       int
}

// Serialize is used to encode the Block before
// insertion in BoltDB
func (b *Block) Serialize() ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		fmt.Printf("Error serializing block")
		return nil, err
	}

	return result.Bytes(), nil
}

//	func to create a new block
func newBlock(prevBlockHash []byte, info string) *Block {

	b := &Block{time.Now().Unix(), []byte(info), prevBlockHash, []byte{}, 0}
	pow := newProofOfWork(b)
	counter, hash := pow.run()
	b.hash = hash[:]
	b.counter = counter

	return b
}

//	func that creates the Blockchain with
//	the Genesis Block as its first block
func genesisBlock() *Block {

	return newBlock([]byte{}, "Genesis Block")
}
