package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"github.com/murlokito/gophercoin/address"
	"io/ioutil"
	"log"
	"os"
)

// Wallet stores a collection of Wallet
type Wallet struct {
	Wallet map[string]*address.Address
}

// NewWallet creates Wallet and fills it from a file if it exists
func NewWallet(fileName string) (*Wallet, error) {
	Wallet := Wallet{}
	Wallet.Wallet = make(map[string]*address.Address)

	err := Wallet.LoadFromFile(fileName)
	if os.IsNotExist(err) {
		log.Printf("[gcw] Wallet did not exist, creating.")
		Wallet.SaveToFile("")
	}

	return &Wallet, nil
}

// CreateAddress adds an Address to Wallet
func (ws *Wallet) CreateAddress() string {
	wallet := address.NewAddress()
	address := fmt.Sprintf("%s", wallet.GetAddress())
	log.Printf("New address created: %s", address)
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
func (ws Wallet) GetAddress(address string) address.Address {
	return *ws.Wallet[address]
}

// LoadFromFile loads Wallet from a file
func (ws *Wallet) LoadFromFile(fileName string) error {
	var (
		walletFile string
	)

	// In case no wallet file name is passed we'll use the default file name
	if fileName == "" {
		fileName = Bucket
	}

	walletFile = fmt.Sprintf("%s%s", fileName, Extension)

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
func (ws Wallet) SaveToFile(fileName string) {
	var (
		content    bytes.Buffer
		walletFile string
	)

	// In case no wallet file name is passed we'll use the default file name
	if fileName == "" {
		fileName = Bucket
	}
	walletFile = fmt.Sprintf("%s%s", fileName, Extension)

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
