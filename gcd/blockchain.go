package gcd

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const (
	noExistingBlockchainFound = "No existing blockchain found"
)

// Blockchain is an array of blocks.
// Arrays in Go are ordered by default,
// which helps with some minor issues
type Blockchain struct {
	tip []byte
	db  *bolt.DB
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
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
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
		bc.tip = newBlock.Hash

		return nil
	})
	return newBlock
}

// AddBlock saves the block into the blockchain
func (bc *Blockchain) AddBlock(block *Block) {
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
			bc.tip = block.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// AddGenesis saves the block into the blockchain
func (bc *Blockchain) AddGenesis(block *Block) {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Printf("[GCDB] err creating blockchain bucket: %+v\n", err)
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
		bc.tip = block.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// FindTransaction is used to get a Transaction by the given transaction hash
// passed as the ID
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
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

	return Transaction{}, errors.New("Transaction is not found")
}

// SignTransaction is used by the blockchain to sign the given transaction with the
// given private key
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Printf("Error finding for transaction")
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

//VerifyTransaction is used to verify the given
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

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
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
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
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address string) (Blockchain, error) {
	dbFile := fmt.Sprintf("%s%s", blocksBucket, bucketExtension)

	if fileExists(dbFile) {
		return Blockchain{}, fmt.Errorf("Blockchain already exists")
	}

	var tip []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Printf("[GCDB] err opening db: %+v\n", err)
		return Blockchain{}, err
	}

	if address != "" {

		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := genesisBlock(cbtx)

		err = db.Update(func(tx *bolt.Tx) error {

			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Printf("[GCDB] err creating blockchain bucket: %+v\n", err)
			}
			ser, err := genesis.SerializeBlock()
			if err != nil {
				log.Printf("[GCDB] err serializing genesis block: %+v\n", err)
			}
			err = b.Put(genesis.Hash, ser)
			if err != nil {
				log.Printf("[GCDB] err updating genesis hash: %+v\n", err)
			}

			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Printf("[GCDB] err updating last block hash: %+v\n", err)
			}
			tip = genesis.Hash

			return nil
		})
	}

	if err != nil {
		log.Printf("[GCDB] err in blockchain creation db method: %+v\n", err)
		return Blockchain{}, err
	}

	bc := Blockchain{
		tip: tip,
		db:  db,
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
		return nil, fmt.Errorf(noExistingBlockchainFound)
	}

	var tip []byte
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		log.Printf("[GCDB] err opening db: %+v\n", err)
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
		tip: tip,
		db:  db,
	}

	return &bc, nil
}
