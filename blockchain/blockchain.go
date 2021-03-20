package blockchain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/murlokito/gophercoin/transaction"
	"log"
	"os"
	"sync"

	"github.com/boltdb/bolt"
)

const (
	noExistingBlockchainFound = "No existing blockchain found"
)

// Blockchain is an array of blocks.
// Arrays in Go are ordered by default,
// which helps with some minor issues
type Blockchain struct {
	Tip   []byte
	db    *bolt.DB
	mutex *sync.RWMutex
}

// fileExists is used to check if the database
// already exists locally or not
func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

// GetBestHeight returns the height of the latest block
func (bc *Blockchain) GetBestHeight() int {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	var block *Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		block, err = DeserializeBlock(blockData)
		if err != nil {
			log.Panicf("err: %v", err)

			return nil
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return block.Height
}

// GetBlock finds a block by its hash and returns it
func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	var block *Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("block not found")
		}

		block, err = DeserializeBlock(blockData)
		if err != nil {
			return errors.New(err.Error())
		}
		return nil
	})

	if err != nil {
		return *block, err
	}

	return *block, nil
}

// GetBlockHashes returns a list of hashes of all the blocks in the chain
func (bc *Blockchain) GetBlockHashes() [][]byte {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}

// MineBlock is the method used to mine a block
// with the provided transactions. The parameter `transactions`
// passed as a pointer to a slice of transactions
func (bc *Blockchain) MineBlock(transactions []*transaction.Transaction) *Block {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	var lastHash []byte
	var lastHeight int
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		block, err := DeserializeBlock(blockData)
		if err != nil {
			log.Printf("Error deserializing blockchain tip")
			return nil
		}
		lastHeight = block.Height
		return nil
	})
	if err != nil {
		log.Printf("Error getting last block")
	}

	log.Printf("Previous Height: %d Previous Hash: %v", lastHeight, lastHash)

	newBlock := NewBlock([]byte(lastHash), transactions, lastHeight+1)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		block, err := newBlock.SerializeBlock()
		if err != nil {
			log.Printf("Error serializing new block")
			return nil
		}
		err = b.Put(newBlock.Hash, block)
		if err != nil {
			log.Printf("Error updating bucket with new block")
			return nil
		}
		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Printf("Error serializing genesis block")
			return nil
		}
		bc.Tip = newBlock.Hash

		return nil
	})

	log.Printf("Update Tip: %d Latest Hash: %v", newBlock.Height, newBlock.Hash)

	return newBlock
}

// AddBlock saves the block into the blockchain
func (bc *Blockchain) AddBlock(block *Block) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDb := b.Get(block.Hash)

		if blockInDb != nil {
			return nil
		}

		blockData, err := block.SerializeBlock()
		if err != nil {
			log.Panic(err)
		}
		err = b.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock, err := DeserializeBlock(lastBlockData)
		if err != nil {
			log.Panic(err)
		}
		if block.Height > lastBlock.Height {
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			bc.Tip = block.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// AddGenesis saves the block into the blockchain
func (bc *Blockchain) AddGenesis(block *Block) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Printf("err creating blockchain bucket: %+v\n", err)
		}

		blockData, err := block.SerializeBlock()
		if err != nil {
			log.Panic(err)
		}
		err = b.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), block.Hash)
		if err != nil {
			log.Panic(err)
		}
		bc.Tip = block.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// FindTransaction is used to get a Transaction by the given transaction hash
// passed as the ID
func (bc *Blockchain) FindTransaction(ID []byte) (transaction.Transaction, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return transaction.Transaction{}, errors.New("transaction was not found")
}

// FindPreviousTransactions is used to get the previous transactions associated with the passed
// transaction's Vins
func (bc *Blockchain) FindPreviousTransactions(tx *transaction.Transaction) (map[string]transaction.Transaction, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	prevTXs := make(map[string]transaction.Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Printf("Error finding for transaction")
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return prevTXs, nil
}

//VerifyTransaction is used to verify the given transaction
func (bc *Blockchain) VerifyTransaction(tx *transaction.Transaction) bool {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]transaction.Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Printf("Error finding for transaction")
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

// BlockchainIterator is the struct defining
// the iterator used to iterate over all the keys
// in a boltdb bucket
type BlockchainIterator struct {
	currentHash []byte
	db          *Blockchain
}

//Iterator is the method used to create an iterator,
// it will be linked to the blockchain tip
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	bci := &BlockchainIterator{bc.Tip, bc}

	return bci
}

// Next is the method used to get the next block
// while iterating the blockchain
func (i *BlockchainIterator) Next() *Block {
	i.db.mutex.RLock()
	defer i.db.mutex.RUnlock()
	var block *Block

	err := i.db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		nblock, err := DeserializeBlock(encodedBlock)
		if err != nil {
			log.Printf("Error deserializing block")
			return nil
		}
		block = nblock
		return nil
	})
	if err != nil {
		log.Printf("Error serializing new block")

	}

	i.currentHash = block.PrevBlockHash

	return block
}

// FindUTXO finds and returns all unspent transaction outputs
func (bc *Blockchain) FindUTXO() map[string]transaction.TXOutputs {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	unspentOutputs := make(map[string]transaction.TXOutputs)
	spentOutputs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentOutputs[txID] != nil {
					for _, spentOutIdx := range spentOutputs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := unspentOutputs[txID]
				outs.Outputs = append(outs.Outputs, out)
				unspentOutputs[txID] = outs
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentOutputs[inTxID] = append(spentOutputs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentOutputs
}

// CreateBlockchain creates a new blockchain
func CreateBlockchain(address string) (*Blockchain, error) {
	dbFile := fmt.Sprintf("%s%s", blocksBucket, bucketExtension)

	if fileExists(dbFile) {
		return &Blockchain{}, errors.New("blockchain already exists")
	}

	var tip []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Printf("err opening db: %+v\n", err)
		return &Blockchain{}, err
	}

	if address != "" {

		coinbaseTx := transaction.NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := genesisBlock(coinbaseTx)

		err = db.Update(func(tx *bolt.Tx) error {

			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Printf("err creating blockchain bucket: %+v\n", err)
			}
			ser, err := genesis.SerializeBlock()
			if err != nil {
				log.Printf("err serializing genesis block: %+v\n", err)
			}
			err = b.Put(genesis.Hash, ser)
			if err != nil {
				log.Printf("err updating genesis hash: %+v\n", err)
			}

			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Printf("err updating last block hash: %+v\n", err)
			}
			tip = genesis.Hash

			return nil
		})
	}

	if err != nil {
		log.Printf("err in blockchain creation db method: %+v\n", err)
		return &Blockchain{}, err
	}

	bc := &Blockchain{
		Tip:   tip,
		db:    db,
		mutex: &sync.RWMutex{},
	}

	return bc, nil
}

// NewBlockchain is used to open a db file,
// check if a Blockchain already existed,
// if so gets the current blockchain tip,
// else generates the genesis block and
// sets it as the tip
func NewBlockchain(path string) (*Blockchain, error) {
	if fileExists(path) == false {
		return nil, errors.New(noExistingBlockchainFound)
	}

	var tip []byte
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		log.Printf("err opening db: %+v\n", err)
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{
		Tip:   tip,
		db:    db,
		mutex: &sync.RWMutex{},
	}

	return &bc, nil
}
