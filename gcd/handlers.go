package gcd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// ResponseError defined to be used for serialization purposes
type ResponseError struct {
	Status      int    `json:"Status"`
	Description string `json:"Description"`
}

// ResponseAddress defined to be used for serialization purposes
type ResponseAddress struct {
	Address string `json:"Address,omitempty"`
}

// ResponseListAddresses defined to be used for serialization purposes
type ResponseListAddresses struct {
	Addresses []ResponseAddress `json:"Addresses,omitempty"`
}

// ResponseBlock defined to be used for serialization purposes
type ResponseBlock struct {
	Timestamp     int64
	PrevBlockHash []byte
	Transactions  []*Transaction
	Hash          []byte
	Nonce         int
	Height        int
	ProofOfWork   string
}

// ResponseListBlocks defined to be used for serialization purposes
type ResponseListBlocks struct {
	Blocks []ResponseBlock `json:"Blocks,omitempty"`
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

	if s.wallet == nil {
		respondWithError(w, http.StatusBadRequest, "Wallet uninitialized")
	}

	p := []ResponseAddress{
		ResponseAddress{
			Address: s.wallet.CreateAddress(),
		},
		ResponseAddress{
			Address: s.wallet.CreateAddress(),
		},
		ResponseAddress{
			Address: s.wallet.CreateAddress(),
		},
	}

	respondWithJSON(w, http.StatusOK, p)
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

	if s.wallet == nil {
		respondWithError(w, http.StatusBadRequest, "Wallet uninitialized")
	}
	addr := ResponseAddress{
		Address: s.wallet.CreateAddress(),
	}

	respondWithJSON(w, http.StatusOK, addr)
}

// ListAddresses is the handler for the '/list_addresses' endpoint, which is
// responsible for asking the wallet for a new address.
func (s *GcdServer) ListAddresses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.wallet == nil {
		respondWithError(w, http.StatusBadRequest, "Wallet uninitialized")
	}

	list := s.wallet.GetAddresses()
	var responseList ResponseListAddresses
	for _, addr := range list {
		a := ResponseAddress{
			Address: addr,
		}

		responseList.Addresses = append(responseList.Addresses, a)
	}

	respondWithJSON(w, http.StatusOK, responseList)
}

// ListBlocks is the handler for the '/list_blocks' endpoint, which is
// responsible for asking the wallet for a new address.
func (s *GcdServer) ListBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.db == nil {
		respondWithError(w, http.StatusBadRequest, "Blockchain uninitialized")
	}

	var responseList ResponseListBlocks
	log.Printf("Chain tip: %v", s.db.tip)
	bci := s.db.Iterator()
	for {
		block := bci.Next()

		pow := NewProofOfWork(block)
		b := ResponseBlock{
			Height:        block.Height,
			PrevBlockHash: block.PrevBlockHash,
			Transactions:  block.Transactions,
			Hash:          block.Hash,
			ProofOfWork:   strconv.FormatBool(pow.Validate()),
		}
		responseList.Blocks = append(responseList.Blocks, b)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	respondWithJSON(w, http.StatusOK, responseList)
}

// CreateBlockchain is the handler for the '/create_blockchain' endpoint
func (s *GcdServer) CreateBlockchain(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	if s.wallet == nil {
		wallet, err := NewWallet()
		if err != nil {
			json.NewEncoder(w).Encode(
				ResponseError{
					Status:      404,
					Description: fmt.Errorf("Wallet uninitialized: %+v", err).Error(),
				},
			)
		}
		s.wallet = wallet
	}

	addr := s.wallet.GetInitialAddress()

	db, err := CreateBlockchain(addr)
	if err != nil {
		json.NewEncoder(w).Encode(
			ResponseError{
				Status:      404,
				Description: fmt.Errorf("Failed to create new db: %+v", err).Error(),
			},
		)
	} else {
		s.db = &db
		json.NewEncoder(w).Encode(
			ResponseCreateBlockchain{
				Status:      200,
				Description: "Successfully created the blockchain!",
			},
		)
	}

}

// GetBalance is the handler for the '/get_balance/{Address}' endpoint
func (s *GcdServer) GetBalance(w http.ResponseWriter, r *http.Request) {
	data := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	var balance int
	balance = 0

	if data["Address"] != "" {

		log.Printf("Get balance for: %v", string(data["Address"]))

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

			go func() {
				s.utxoSet.Reindex()
				ctx := s.utxoSet.CountTransactions()
				log.Printf("Finished reindexing utxoSet, there are %d transactions in it.", ctx)
			}()

			json.NewEncoder(w).Encode(
				ResponseError{
					Status:      201,
					Description: fmt.Errorf("Reindexing UTXOs").Error(),
				},
			)
		} else {
			log.Printf("Finding all utxos for address: %v.", string(data["Address"]))
			pubKeyHash := Base58Decode([]byte(data["Address"]))
			pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
			UTXOs := s.utxoSet.FindUTXO(pubKeyHash)

			if len(UTXOs) >= 1 {
				for _, out := range UTXOs {

					log.Printf("utxo: %v.", out)
					balance += out.Value
				}
			}
			json.NewEncoder(w).Encode(
				ResponseBalance{
					Address: data["Address"],
					Balance: balance,
				},
			)
		}
	} else {
		json.NewEncoder(w).Encode(
			ResponseError{
				Status:      415,
				Description: fmt.Errorf("Error validating input").Error(),
			},
		)
	}

}

// GenerateBlocks is the handler for the '/generate_blocks/{Amount}' endpoint
func (s *GcdServer) GenerateBlocks(w http.ResponseWriter, r *http.Request) {

	data := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

	if data["Amount"] != "" {
		log.Printf("Generating %v blocks.", data["Amount"])

		if s.db == nil {
			json.NewEncoder(w).Encode(
				ResponseError{
					Status:      404,
					Description: fmt.Errorf("Blockhain not found").Error(),
				},
			)

		} else {
			addr := s.wallet.CreateAddress()
			amt, err := strconv.Atoi(data["Amount"])
			if err != nil {
				json.NewEncoder(w).Encode(
					ResponseError{
						Status:      415,
						Description: fmt.Errorf("Error validating input").Error(),
					},
				)
			} else {
				for i := 0; i < amt; i++ {
					cbTx := NewCoinbaseTX(addr, "")
					txs := []*Transaction{cbTx}

					newBlock := s.db.MineBlock(txs)
					s.utxoSet.Update(newBlock)
				}
			}
		}
	}

}

// SubmitTx is the handler for the '/submit_tx/{Transaction}' endpoint
func (s *GcdServer) SubmitTx(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

	if vars["Transaction"] == "" {
		respondWithError(w, http.StatusBadRequest, "Empty transaction")
	}

	respondWithJSON(w, http.StatusOK, p)

}

// AddNode is the handler for the '/add_node/{Address}' endpoint
func (s *GcdServer) AddNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

}
