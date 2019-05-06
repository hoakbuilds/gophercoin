package main

import (
	"fmt"
	"strconv"
)

func main() {

	bc := newBlockchain()
	bc.addBlock("Send 1 gophercoin to Murloc")
	bc.addBlock("Send 2 gophercoin to Pasta Bro")

	for _, block := range bc.blocks {
		fmt.Printf("Previous block hash: %x\n", block.prevBlockHash)
		fmt.Printf("Info: %s\n", block.info)
		fmt.Printf("Hash: %x\n", block.hash)

		pow := newProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.validate()))
		fmt.Println()
	}
}
