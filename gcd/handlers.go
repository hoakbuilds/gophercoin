package gcd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// ResponseError defined to be used for serialization purposes
type ResponseError struct {
	Status      int    `json:"Status"`
	Description string `json:"Description"`
}

// ResponseAddress defined to be used for serialization purposes
type ResponseAddress struct {
	Address string `json:"Address,omitempty"`
}

// ResponseCreateBlockchain defined to be used for serialization purposes
type ResponseCreateBlockchain struct {
	Status      int    `json:"Status"`
	Description string `json:"Description"`
}

// ResponseCreateWallet defined to be used for serialization purposes
type ResponseCreateWallet struct {
	Status      int    `json:"Status"`
	Description string `json:"Description"`
	Address     string `json:"Address,omitempty"`
}

// ResponseBalance defined to be used for serialization purposes
type ResponseBalance struct {
	Address string `json:"Address,omitempty"`
	Balance int    `json:"Balance,omitempty"`
}

// Index is the handler for the '/' endpoint, which is to be used for
// tests only
func (s *GcdServer) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.wallet != nil {
		json.NewEncoder(w).Encode([]ResponseAddress{
			ResponseAddress{
				Address: s.wallet.CreateAddress(),
			},
			ResponseAddress{
				Address: s.wallet.CreateAddress(),
			},
			ResponseAddress{
				Address: s.wallet.CreateAddress(),
			},
		})
	}

	json.NewEncoder(w).Encode(
		ResponseError{
			Status:      401,
			Description: fmt.Errorf("Wallet uninitialized").Error(),
		},
	)

}

// CreateWallet is the handler for the '/create_wallet' endpoint, which is
// responsible for asking the wallet for a new address.
func (s *GcdServer) CreateWallet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	wallet, err := NewWallet()
	if err != nil {
		json.NewEncoder(w).Encode(
			ResponseError{
				Status:      404,
				Description: fmt.Errorf("Failed to create new Wallet: %+v", err).Error(),
			},
		)
	}
	s.wallet = wallet
	json.NewEncoder(w).Encode(
		ResponseCreateWallet{
			Status:      200,
			Description: "Successfully created wallet!",
			Address:     wallet.CreateAddress(),
		},
	)

}

// NewAddress is the handler for the '/new_address' endpoint, which is
// responsible for asking the wallet for a new address.
func (s *GcdServer) NewAddress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(
		ResponseAddress{
			Address: s.wallet.CreateAddress(),
		},
	)
}

// CreateBlockchain is the handler for the '/create_blockchain' endpoint
func (s *GcdServer) CreateBlockchain(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	addr := s.wallet.GetInitialAddress()

	db, err := CreateBlockchain(addr)
	if err != nil {
		json.NewEncoder(w).Encode(
			ResponseError{
				Status:      404,
				Description: fmt.Errorf("Failed to create new Wallet: %+v", err).Error(),
			},
		)
	}
	s.db = db
	json.NewEncoder(w).Encode(
		ResponseCreateBlockchain{
			Status:      200,
			Description: "Successfully created the blockchain!",
		},
	)

	//todo, lançar o daemon

}

// GetBalance is the handler for the '/get_balance' endpoint
func (s *GcdServer) GetBalance(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var addr ResponseAddress
	var balance int
	balance = 0

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body",
			http.StatusInternalServerError)
	}

	addr.Address = string(body)
	log.Printf("Get balance for: %v", string(addr.Address))

	if s.db == nil {
		json.NewEncoder(w).Encode(
			ResponseError{
				Status:      404,
				Description: fmt.Errorf("Blockhain not found").Error(),
			},
		)

	}

	if s.utxoSet == nil {

		log.Printf("utxoSet not found, creating it.")
		// perform task
		UTXOSet := UTXOSet{
			chain: s.db,
		}
		s.utxoSet = &UTXOSet
	}

	log.Printf("Finding all utxos for address: %v.", string(addr.Address))
	pubKeyHash := Base58Decode([]byte(addr.Address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := s.utxoSet.FindUTXO(pubKeyHash)

	if len(UTXOs) > 1 {
		for _, out := range UTXOs {

			log.Printf("utxo: %v.", out)
			balance += out.Value
		}
	}
	json.NewEncoder(w).Encode(
		ResponseBalance{
			Address: addr.Address,
			Balance: balance,
		},
	)

}

// GenerateBlocks is the handler for the '/get_balance' endpoint
func (s *GcdServer) GenerateBlocks(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body",
			http.StatusInternalServerError)
	}
	log.Printf("Get balance for: %v", string(body))

	if s.db == nil {
		json.NewEncoder(w).Encode(
			ResponseError{
				Status:      404,
				Description: fmt.Errorf("Blockhain not found").Error(),
			},
		)

	}

}

// SubmitTx is the handler for the '/submit_tx' endpoint
func (s *GcdServer) SubmitTx(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

}
