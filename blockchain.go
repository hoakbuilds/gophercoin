package main

import (
	"encoding/hex"
	"fmt"

	"github.com/boltdb/bolt"
)

// Blockchain is an array of blocks.
// Arrays in Go are ordered by default,
// which helps with some minor issues
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// AddBlock is the method used to add a block
// to the blockchain. The string given with the
// parameter data is used to be hashed in the block
func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("1"))

		return nil
	})
	if err != nil {
		fmt.Printf("Error getting last block")
	}

	newBlock := NewBlock([]byte(data), string(lastHash))

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		block, err := newBlock.SerializeBlock()
		if err != nil {
			fmt.Printf("Error serializing new block")
			return nil
		}
		err = b.Put(newBlock.Hash, block)
		if err != nil {
			fmt.Printf("Error updating bucket with new block")
			return nil
		}
		err = b.Put([]byte("1"), newBlock.Hash)
		if err != nil {
			fmt.Printf("Error serializing genesis block")
			return nil
		}
		bc.tip = newBlock.Hash

		return nil
	})
}

// BlockchainIterator is the struct defining
// the iterator used to iterate over all the keys
// in a boltdb bucket
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

//Iterator is the method used to create an iterator,
// it will be linked to the blockchain tip
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

// Next is the method used to get the next block
// while iterating the blockchain
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		nblock, err := DeserializeBlock(encodedBlock)
		if err != nil {
			fmt.Printf("Error deserializing block")
			return nil
		}
		block = nblock
		return nil
	})
	if err != nil {
		fmt.Printf("Error serializing new block")

	}

	i.currentHash = block.PrevBlockHash

	return block
}

// FindUnspentTransactions returns a list of transactions containing unspent outputs
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// FindUTXO finds and returns all unspent transaction outputs
func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

// NewBlockchain is used to open a db file,
// check if a Blockchain already existed,
// if so gets the current blockchain tip,
// else generates the genesis block and
// sets it as the tip
func NewBlockchain(data []byte) *Blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		fmt.Printf("Error opening db file while creating blockchain")
		return nil
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			genesis := genesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			serializedGenesis, err := genesis.SerializeBlock()
			if err != nil {
				fmt.Printf("Error serializing genesis block")
				return nil
			}
			err = b.Put(genesis.Hash, serializedGenesis)
			err = b.Put([]byte("l"), genesis.Hash)
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	bc := Blockchain{
		tip: tip,
		db:  db,
	}

	return &bc
}
