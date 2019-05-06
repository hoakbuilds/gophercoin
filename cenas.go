package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"
)

var (
	maxNonce = math.MaxInt64
)

const targetBits = 24

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

// Blockchain is an array of blocks.
// Arrays in Go are ordered by default,
// which helps with some minor issues
type Blockchain struct {
	blocks []*Block
}

// ProofOfWork is a mechanism used in blockchains.
// The main ideia is that some hard work has to be
// done to add a block to the blockchain.
// This helps maintain the stability of the
// blockchain database.
type ProofOfWork struct {
	block  *Block
	target *big.Int
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

//	func to add blocks to the blockchain
func (bc *Blockchain) addBlock(info string) {

	prevblock := bc.blocks[len(bc.blocks)-1]
	newBlock := newBlock(prevblock.hash, info)
	bc.blocks = append(bc.blocks, newBlock)
}

//	func that creates the Blockchain with
//	the Genesis Block as its first block
func genesisBlock() *Block {

	return newBlock([]byte{}, "Genesis Block")
}

// func that creates a new Blockchain with
// the the genesis block in the first position
func newBlockchain() *Blockchain {
	return &Blockchain{[]*Block{genesisBlock()}}
}

// func that creates a new ProofOfWork struct.
// We use a big int because later we'll convert
// a hash into a big int and compare if it's
// less than the target. The target is like
// the upper boundary of a range. If a hash
// is lower than the boundary it's valid.
// Lowering the boundary makes if more
// difficult to find a valid hash.
func newProofOfWork(b *Block) *ProofOfWork {

	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	POW := &ProofOfWork{b, target}
	return POW
}

// IntToHex converst an int into a hexadecimal.
// This will be used in the function prepareData
// coded below
func IntToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n, 16))
}

// func that merges the block fields with the target and counter
// this "prepared" the data to be hashed
func (pow *ProofOfWork) prepareData(counter int) []byte {

	data := bytes.Join(
		[][]byte{
			pow.block.prevBlockHash,
			pow.block.info,
			IntToHex(pow.block.timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(counter)),
		},
		[]byte{},
	)
	return data
}

// basic proof of work function which will
// enable us to mine blocks
func (pow *ProofOfWork) run() (int, []byte) {

	var hashInt big.Int // hash represented in integer form
	var hash [32]byte
	counter := 0

	fmt.Printf("Mining block containg \"%s\"\n", pow.block.info)

	for counter < maxNonce {
		data := pow.prepareData(counter)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			counter++
		}
	}
	fmt.Print("\n\n")

	return counter, hash[:]
}

// func that decides wheather the proof of worl
// is valid of not
func (pow *ProofOfWork) validate() bool {

	var hashInt big.Int

	data := pow.prepareData(pow.block.counter)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

func main() {

	bc := newBlockchain()
	bc.addBlock("Send 1 WeedCoin to Murloc")
	bc.addBlock("Send 2 WeedCoin to Pasta Bro")

	for _, block := range bc.blocks {
		fmt.Printf("Previous block hash: %x\n", block.prevBlockHash)
		fmt.Printf("Info: %s\n", block.info)
		fmt.Printf("Hash: %x\n", block.hash)

		pow := newProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.validate()))
		fmt.Println()
	}
}
