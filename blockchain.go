package main

// Blockchain is an array of blocks.
// Arrays in Go are ordered by default,
// which helps with some minor issues
type Blockchain struct {
	blocks []*Block
}

//	func to add blocks to the blockchain
func (bc *Blockchain) addBlock(info string) {

	prevblock := bc.blocks[len(bc.blocks)-1]
	newBlock := newBlock(prevblock.hash, info)
	bc.blocks = append(bc.blocks, newBlock)
}

// func that creates a new Blockchain with
// the the genesis block in the first position
func newBlockchain() *Blockchain {
	return &Blockchain{[]*Block{genesisBlock()}}
}
