package gcd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	address2 "github.com/murlokito/gophercoin/address"
	"github.com/murlokito/gophercoin/transaction"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/murlokito/gophercoin/blockchain"
	"github.com/murlokito/gophercoin/peer"
	"github.com/murlokito/gophercoin/wallet"

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

// ResponseMessage defined to be used for serialization purposes
type ResponseMessage struct {
	Description string `json:"Description"`
}

// ResponseInfo defined to be used for serialization purposes
type ResponseInfo struct {
	Protocol string `json:"Protocol"`
	Peers    int    `json:"Peers"`
	Mining   string `json:"Mining"`
	Synced   string `json:"Synced"`
}

// ResponseTx defined to be used for serialization purposes
type ResponseTx struct {
	ID   []byte                 `json:"ID"`
	Vin  []transaction.TXInput  `json:"Vin"`
	Vout []transaction.TXOutput `json:"Vout"`
}

// ResponseListTx defined to be used for serialization purposes
type ResponseListTx struct {
	Transactions []ResponseTx `json:"Transactions,omitempty"`
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
	Transactions  []*transaction.Transaction
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
	Balance int64  `json:"Balance,omitempty"`
}

// ResponseSubmitTx defined to be used for serialization purposes
type ResponseSubmitTx struct {
	Status   string                  `json:"Status"`
	Tx       transaction.Transaction `json:"Transaction"`
	NewBlock blockchain.Block        `json:"NewBlock"`
}

// Index is the handler for the '/' endpoint, which is to be used for
// tests only
func (s *Server) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.wallet == nil {
		respondWithError(w, http.StatusBadRequest, "Wallet uninitialized")
		return
	}

	p := []ResponseAddress{
		ResponseAddress{
			Address: s.wallet.CreateAddress(),
		},
	}

	respondWithJSON(w, http.StatusOK, p)
	return
}

// CreateWallet is the handler for the '/create_wallet' endpoint, which is
// responsible for asking the wallet for a new address.
func (s *Server) CreateWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

	if vars["WalletFile"] == "" {
		respondWithError(w, http.StatusBadRequest, "Empty transaction")
		return
	}

	wallet, err := wallet.NewWallet(vars["WalletFile"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest,
			fmt.Errorf("Failed to create new Wallet: %+v", err).Error())
		return

	}
	s.wallet = wallet
	respondWithJSON(w, http.StatusOK, ResponseAddress{
		Address: wallet.CreateAddress(),
	})
	return
}

// NewAddress is the handler for the '/new_address' endpoint, which is
// responsible for asking the wallet for a new address.
func (s *Server) NewAddress(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	if s.wallet == nil {
		respondWithError(w, http.StatusBadRequest, "Wallet uninitialized")
		return
	}
	addr := ResponseAddress{
		Address: s.wallet.CreateAddress(),
	}

	respondWithJSON(w, http.StatusOK, addr)
	return
}

// NodeInfo is the handler for the '/list_addresses' endpoint, which is
// responsible for asking the wallet for a new address.
func (s *Server) NodeInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.wallet == nil {
		respondWithError(w, http.StatusBadRequest, "Wallet uninitialized")
		return
	}

	var response ResponseInfo

	if s.cfg.miningNode == true {
		response.Mining = "true"
	} else {
		response.Mining = "false"
	}

	if s.peerServer.Config.Port != "" {
		response.Protocol = s.peerServer.Config.Port
	} else {
		response.Protocol = peer.DefaultProtocolPort
	}

	response.Peers = len(s.peerServer.KnownNodes)
	response.Synced = "Consensus can't tell yet. Bullish."

	respondWithJSON(w, http.StatusOK, response)
	return
}

// ListAddresses is the handler for the '/list_addresses' endpoint, which is
// responsible for asking the wallet for a new address.
func (s *Server) ListAddresses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.wallet == nil {
		respondWithError(w, http.StatusBadRequest, "Wallet uninitialized")
		return
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
	return
}

// ListBlocks is the handler for the '/list_blocks' endpoint, which is
// responsible for asking the wallet for a new address.
func (s *Server) ListBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.chainMgr == nil {
		respondWithError(w, http.StatusBadRequest, "Blockchain uninitialized")
		return
	}

	var responseList ResponseListBlocks
	log.Printf("Chain tip: %v", s.chainMgr.Chain.Tip)
	bci := s.chainMgr.Chain.Iterator()
	for {
		block := bci.Next()

		pow := blockchain.NewProofOfWork(block)
		b := ResponseBlock{
			Timestamp:     block.Timestamp,
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
	return
}

// ListMempool is the handler for the '/list_mempool' endpoint, which is
// responsible for asking the wallet for a new address.
func (s *Server) ListMempool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if len(s.chainMgr.MemPool) < 1 {
		respondWithError(w, http.StatusBadRequest, "No transactions in mempool")
		return
	}

	var responseListTxs ResponseListTx

	for _, tx := range s.chainMgr.MemPool {
		newtx := ResponseTx{
			ID:   tx.ID,
			Vin:  tx.Vin,
			Vout: tx.Vout,
		}
		responseListTxs.Transactions = append(responseListTxs.Transactions, newtx)
	}

	respondWithJSON(w, http.StatusOK, responseListTxs)
	return
}

// CreateBlockchain is the handler for the '/create_blockchain' endpoint
func (s *Server) CreateBlockchain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

	if vars["WalletFile"] == "" {
		respondWithError(w, http.StatusBadRequest, "Empty transaction")
		return
	}

	if s.wallet == nil {
		wallet, err := wallet.NewWallet(vars["WalletFile"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, fmt.Errorf("Wallet uninitialized: %+v", err).Error())
			return
		}
		s.wallet = wallet
	}

	addr := s.wallet.CreateAddress()
	log.Printf("Mining genesis block to address: %v", addr)
	db, err := blockchain.CreateBlockchain(addr)
	var msg string
	if err != nil {
		msg = fmt.Errorf("Failed to create db: %+v", err).Error()
	}

	s.chainMgr.Chain = db

	if msg != "" {
		respondWithJSON(w, http.StatusOK, ResponseMessage{
			Description: msg,
		})
	} else {
		respondWithJSON(w, http.StatusOK, ResponseMessage{
			Description: "Successfully created blockchain",
		})
	}

	return

}

// GetBalance is the handler for the '/get_balance/{Address}' endpoint
func (s *Server) GetBalance(w http.ResponseWriter, r *http.Request) {
	data := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")
	var balance int
	balance = 0

	if data["Address"] != "" {

		if s.chainMgr.Chain == nil {
			respondWithError(w, http.StatusBadRequest,
				fmt.Errorf("Blockhain not found").Error())
			return

		}

		if s.chainMgr.UTXOSet == nil {
			log.Printf("utxoSet not found, creating it.")
			// perform task
			UTXOSet := &blockchain.UTXOSet{
				Chain: s.chainMgr.Chain,
				Mutex: &sync.RWMutex{},
			}
			s.chainMgr.UTXOSet = UTXOSet

			go func() {
				s.chainMgr.UTXOSet.Reindex()
				ctx := s.chainMgr.UTXOSet.CountTransactions()
				log.Printf("Finished reindexing utxoSet, there are %d transactions in it.", ctx)
			}()

			respondWithJSON(w, http.StatusOK, ResponseMessage{
				Description: "Reindexing UTXOs.",
			})
			return

		}
		pubKeyHash := address2.Base58Decode([]byte(data["Address"]))
		pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
		UTXOs := s.chainMgr.UTXOSet.FindUTXO(pubKeyHash)

		if len(UTXOs) >= 1 {
			for _, out := range UTXOs {

				balance += out.Value
			}
		}
		log.Printf("Address: %v Balance: %v", string(data["Address"]), balance)
		respondWithJSON(w, http.StatusOK, ResponseBalance{
			Address: data["Address"],
			Balance: int64(balance),
		})
		return
	}

	respondWithError(w, http.StatusBadRequest,
		fmt.Errorf("Error validating input").Error())
	return

}

// GenerateBlocks is the handler for the '/generate_blocks/{Amount}' endpoint
func (s *Server) GenerateBlocks(w http.ResponseWriter, r *http.Request) {

	data := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

	if data["Amount"] != "" {
		log.Printf("Generating %v blocks.", data["Amount"])

		if s.chainMgr.Chain == nil {
			respondWithError(w, http.StatusBadRequest,
				fmt.Errorf("Blockhain not found").Error())
			return

		}
		addr := s.wallet.CreateAddress()
		amt, err := strconv.Atoi(data["Amount"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest,
				fmt.Errorf("Error validating input").Error())
			return

		}
		var responseList ResponseListBlocks
		for i := 0; i < amt; i++ {
			cbTx := transaction.NewCoinbaseTX(addr, "")
			txs := []*transaction.Transaction{cbTx}

			newBlock := s.chainMgr.Chain.MineBlock(txs)
			s.chainMgr.UTXOSet.Update(newBlock)

			pow := blockchain.NewProofOfWork(newBlock)
			b := ResponseBlock{
				Height:        newBlock.Height,
				PrevBlockHash: newBlock.PrevBlockHash,
				Transactions:  newBlock.Transactions,
				Hash:          newBlock.Hash,
				ProofOfWork:   strconv.FormatBool(pow.Validate()),
			}
			responseList.Blocks = append(responseList.Blocks, b)
		}
		respondWithJSON(w, http.StatusOK, responseList)

	}
	respondWithError(w, http.StatusBadRequest,
		fmt.Errorf("Error validating input").Error())
	return

}

// SubmitTx is the handler for the '/submit_tx/{Transaction}' endpoint
func (s *Server) SubmitTx(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

	if vars["From"] == "" || vars["To"] == "" ||
		vars["Amount"] == "" {
		respondWithError(w, http.StatusBadRequest, "Empty transaction")
		return
	}

	if !address2.ValidateAddress(vars["From"]) {
		respondWithError(w, http.StatusBadRequest, "Invalid wallet address")
		return
	}

	if !address2.ValidateAddress(vars["To"]) {
		respondWithError(w, http.StatusBadRequest, "Invalid destiny address")
		return
	}

	address := s.wallet.GetAddress(vars["From"])
	amount, err := strconv.Atoi(vars["Amount"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid amount")
		return
	}
	pubKeyHash := address2.HashPubKey(address.PublicKey)
	acc, validOutputs := s.chainMgr.UTXOSet.FindSpendableOutputs(pubKeyHash, amount)

	tx, err := transaction.NewUTXOTransaction(acc, validOutputs, vars["To"], amount, address.GetAddress())
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	prevTxs, err := s.chainMgr.Chain.FindPreviousTransactions(tx)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = tx.Sign(address.PrivateKey, prevTxs)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	p := ResponseSubmitTx{
		Status: "OK",
		Tx:     *tx,
	}

	if s.peerServer.KnownNodes != nil {
		s.chainMgr.MemPool[hex.EncodeToString(tx.ID)] = *tx
		p.Status = "No peers available, added to mempool."
		if (s.miner.MinerChan) != nil {
			s.miner.MinerChan <- tx.ID
		}
	} else {
		for _, node := range s.peerServer.KnownNodes {
			if node.Address != s.peerServer.NodeAddress {
				s.peerServer.SendInv(node.Address, "tx", [][]byte{tx.ID})
			}
		}
		s.chainMgr.MemPool[hex.EncodeToString(tx.ID)] = *tx
	}

	respondWithJSON(w, http.StatusOK, p)
	return
}

// AddNode is the handler for the '/add_node/{Address}' endpoint
func (s *Server) AddNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

	if vars["Address"] == "" {
		respondWithError(w, http.StatusBadRequest, "No peer address given")
		return
	}
	splitInput := strings.Split(vars["Address"], ":")
	ip, port := splitInput[0], splitInput[1]

	if ip == "localhost" || ip == "127.0.0.1" || ip == "" {
		if port == peer.DefaultProtocolPort {
			respondWithError(w, http.StatusBadRequest, "Cannot add own node")
			return
		}
	}
	log.Printf("Trying to add new peer: %v", vars["Address"])

	for _, peer := range s.peerServer.KnownNodes {
		if vars["Address"] == peer.Address {
			s.peerServer.SendVersion(peer.Address)
			respondWithError(w, http.StatusBadRequest, "Peer is known")
			return
		}

		split := strings.Split(peer.Address, ":")

		if split[1] == port {
			s.peerServer.SendVersion(peer.Address)
			respondWithError(w, http.StatusBadRequest, "Peer is known")
			return
		}

	}

	s.peerServer.KnownNodes = append(s.peerServer.KnownNodes, peer.Peer{
		Address: vars["Address"],
	})

	s.peerServer.SendVersion(vars["Address"])

	log.Printf("Successfully added new peer: %v", vars["Address"])

	respondWithJSON(w, http.StatusOK, s.peerServer.KnownNodes)
	return
}
