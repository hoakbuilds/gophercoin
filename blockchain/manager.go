package blockchain

import "github.com/murlokito/gophercoin/transaction"

type TransactionPool map[string]transaction.Transaction

type ChainManager struct {
	Chain   *Blockchain
	MemPool TransactionPool
	UTXOSet *UTXOSet
}

func NewChainManager(chain *Blockchain, set *UTXOSet) *ChainManager {
	return &ChainManager{
		Chain:   chain,
		UTXOSet: set,
		MemPool: make(TransactionPool, 0),
	}
}
