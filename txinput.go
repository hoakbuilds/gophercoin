package main

import "bytes"

// TXInput represents a transaction input
type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

// UsesKey checks whether the address initiated the transaction
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

// CanUnlockOutputWith checks whether the address initiated the transaction
func (in *TXInput) CanUnlockOutputWith(unlockingData []byte) bool {
	return bytes.Compare(in.PubKey, unlockingData) == 0
}
