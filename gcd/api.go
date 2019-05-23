package gcd

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// BuildAndServeAPI is the function used to serve the API endpoints
func (s *Server) BuildAndServeAPI() {
	log.Println("[GCDAPI] Building API endpoints.")

	s.Router = mux.NewRouter()
	s.Router.HandleFunc("/", s.Index).Methods("GET")
	s.Router.HandleFunc("/new_address", s.NewAddress).Methods("GET")

	//todo
	s.Router.HandleFunc("/create_wallet", s.CreateWallet).Methods("POST")
	s.Router.HandleFunc("/create_blockchain", s.CreateBlockchain).Methods("POST")
	s.Router.HandleFunc("/generate_blocks/{Amount}", s.GenerateBlocks).Methods("POST")

	s.Router.HandleFunc("/get_balance/{Address}", s.GetBalance).Methods("GET")
	s.Router.HandleFunc("/list_addresses", s.ListAddresses).Methods("GET")
	s.Router.HandleFunc("/list_blocks", s.ListBlocks).Methods("GET")

	s.Router.HandleFunc("/submit_tx/{Transaction}", s.SubmitTx).Methods("POST")
	s.Router.HandleFunc("/add_node/{Address}", s.AddNode).Methods("POST")

	log.Println("[GCDAPI] Listening and Serving API. Port: 9000")
	log.Fatal(http.ListenAndServe(":9000", s.Router))
}
