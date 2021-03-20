package gcd

import (
	"github.com/murlokito/gophercoin/api"
	"github.com/murlokito/gophercoin/blockchain"
	"github.com/murlokito/gophercoin/log"
	"github.com/murlokito/gophercoin/mining"
	"github.com/murlokito/gophercoin/peer"
	"github.com/murlokito/gophercoin/wallet"
	"sync"
)

// Server is the structure which defines the Gophercoin
type Server struct {
	cfg          *Config
	chainMgr     *blockchain.ChainManager
	peerServer   *peer.PeerServer
	miner        *mining.MinerServer
	wallet       *wallet.Wallet
	wg           *sync.WaitGroup
	api          *api.APIServer
	nodeServChan chan interface{}
}

// StartServer is the function used to start the gophercoind Server
func (s *Server) StartServer() {
	apiConfig := api.Config{
		Port:      s.cfg.restPort,
		Protected: s.cfg.restProtected,
		Password:  s.cfg.restPassword,
		LogLevel:  log.InfoLevel,
		Routes:    s.Routes(),
	}
	api := api.NewAPIServer(s.wg, apiConfig)

	s.api = api
}

// Routes returns the API routes exposed to interact with the node
func (s *Server) Routes() api.Routes {
	return api.Routes{
		api.Route{
			Name:        "Index",
			Method:      "GET",
			Pattern:     "/",
			HandlerFunc: s.Index,
		},
		api.Route{
			Name:        "NewAddress",
			Method:      "GET",
			Pattern:     "/new_address",
			HandlerFunc: s.NewAddress,
		},
		api.Route{
			Name:        "CreateWallet",
			Method:      "POST",
			Pattern:     "/create_wallet",
			HandlerFunc: s.CreateWallet,
		},
		api.Route{
			Name:        "CreateBlockchain",
			Method:      "POST",
			Pattern:     "/create_blockchain",
			HandlerFunc: s.CreateBlockchain,
		},
		api.Route{
			Name:        "GenerateBlocks",
			Method:      "POST",
			Pattern:     "/generate_blocks/{Amount}",
			HandlerFunc: s.GenerateBlocks,
		},
		api.Route{
			Name:        "GetBalance",
			Method:      "GET",
			Pattern:     "/get_balance/{Address}",
			HandlerFunc: s.GetBalance,
		},
		api.Route{
			Name:        "ListAddresses",
			Method:      "GET",
			Pattern:     "/list_addresses",
			HandlerFunc: s.ListAddresses,
		},
		api.Route{
			Name:        "ListMempool",
			Method:      "GET",
			Pattern:     "/list_mempool",
			HandlerFunc: s.ListMempool,
		},
		api.Route{
			Name:        "ListBlocks",
			Method:      "GET",
			Pattern:     "/list_blocks",
			HandlerFunc: s.ListBlocks,
		},
		api.Route{
			Name:        "NodeInfo",
			Method:      "GET",
			Pattern:     "/node_info",
			HandlerFunc: s.NodeInfo,
		},
		api.Route{
			Name:        "SubmitTx",
			Method:      "POST",
			Pattern:     "/submit_tx/{From}/{To}/{Amount}",
			HandlerFunc: s.SubmitTx,
		},
		api.Route{
			Name:        "AddNode",
			Method:      "POST",
			Pattern:     "/add_node/{Address}",
			HandlerFunc: s.AddNode,
		},
	}
}

// NewServer creates a new server with all the needed components
func NewServer(config *Config, chainMgr *blockchain.ChainManager, wallet *wallet.Wallet, miner *mining.MinerServer, wg *sync.WaitGroup) *Server {
	return &Server{
		cfg:          config,
		chainMgr:     chainMgr,
		wallet:       wallet,
		miner:        miner,
		wg:           wg,
		nodeServChan: make(chan interface{}),
	}
}
