package gcd

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// Wallet stores a collection of Wallet
type Wallet struct {
	Wallet map[string]*Address
}

// NewWallet creates Wallet and fills it from a file if it exists
func NewWallet() (*Wallet, error) {
	Wallet := Wallet{}
	Wallet.Wallet = make(map[string]*Address)

	err := Wallet.LoadFromFile()
	if os.IsNotExist(err) {
		log.Printf("[WLLT] Wallet did not exist, creating.")
		Wallet.SaveToFile()
	}

	return &Wallet, nil
}

// CreateAddress adds an Address to Wallet
func (ws *Wallet) CreateAddress() string {
	wallet := NewAddress()
	address := fmt.Sprintf("%s", wallet.GetAddress())
	log.Printf("[WLLT] New address created: %s", address)
	ws.Wallet[address] = wallet

	return address
}

// GetAddresses returns an array of addresses stored in the wallet file
func (ws *Wallet) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallet {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetInitialAddress returns the first address in the wallet
func (ws *Wallet) GetInitialAddress() string {
	var address string

	for addr := range ws.Wallet {
		address = addr
		break
	}

	return address
}

// GetAddress returns an Address by its address
func (ws Wallet) GetAddress(address string) Address {
	return *ws.Wallet[address]
}

// LoadFromFile loads Wallet from the file
func (ws *Wallet) LoadFromFile() error {
	walletFile := fmt.Sprintf("%s%s", walletBucket, walletExtension)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var Wallet Wallet
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&Wallet)
	if err != nil {
		log.Panic(err)
	}

	ws.Wallet = Wallet.Wallet

	return nil
}

// SaveToFile saves Wallet to a file
func (ws Wallet) SaveToFile() {
	var content bytes.Buffer
	walletFile := fmt.Sprintf("%s%s", walletBucket, walletExtension)

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
