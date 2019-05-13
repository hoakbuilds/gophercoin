package gcd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

// ProofOfWork is a mechanism used in blockchains.
// The main ideia is that some hard work has to be
// done to add a block to the blockchain.
// This helps maintain the stability of the
// blockchain database.
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// func that merges the block fields with the target and counter
// this "prepared" the data to be hashed
func (pow *ProofOfWork) prepareData(counter int) []byte {

	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
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

	fmt.Printf("Mining block containing \n\"%s\"\n", pow.block.Transactions)

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

// Validate is the func that decides whether the proof of work
// is valid of not
func (pow *ProofOfWork) Validate() bool {

	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

// NewProofOfWork is the
// func that creates a new ProofOfWork struct.
// We use a big int because later we'll convert
// a hash into a big int and compare if it's
// less than the target. The target is like
// the upper boundary of a range. If a hash
// is lower than the boundary it's valid.
// Lowering the boundary makes if more
// difficult to find a valid hash.
func NewProofOfWork(b *Block) *ProofOfWork {

	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	POW := &ProofOfWork{b, target}
	return POW
}
