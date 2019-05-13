package gcd

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
	Timestamp     int64
	PrevBlockHash []byte
	Transactions  []*Transaction
	Hash          []byte
	Nonce         int
	Height        int
}

// HashTransactions returns a hash of the transactions in the block
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Serialize())
	}

	mTree := NewMerkleTree(txHashes)

	return mTree.RootNode.Data
}

// DeserializeBlock is used to decode the Block before
// insertion in BoltDB
func DeserializeBlock(d []byte) (*Block, error) {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		fmt.Printf("Error deserializing block")
		return nil, err
	}
	return &block, nil
}

// SerializeBlock is used to encode the Block before
// insertion in BoltDB
func (b *Block) SerializeBlock() ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		fmt.Printf("Error serializing block")
		return nil, err
	}

	return result.Bytes(), nil
}

// NewBlock is the func to create a new block
func NewBlock(prevBlockHash []byte, transactions []*Transaction,
	height int) *Block {

	//Initialize the block structure with the given data
	b := &Block{
		Timestamp:     time.Now().Unix(),
		PrevBlockHash: prevBlockHash,
		Transactions:  transactions,
		Hash:          []byte{},
		Nonce:         0,
		Height:        height,
	}
	pow := NewProofOfWork(b)
	nonce, hash := pow.run()
	b.Hash = hash[:]
	b.Nonce = nonce

	return b
}

//	func that creates the Blockchain with
//	the Genesis Block as its first block
func genesisBlock(coinbasetx *Transaction) *Block {
	return NewBlock([]byte{}, []*Transaction{coinbasetx}, 0)
}
