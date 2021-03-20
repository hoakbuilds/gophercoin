package gcd

import (
	"github.com/murlokito/gophercoin/blockchain"
	log "github.com/murlokito/gophercoin/log"
	"github.com/murlokito/gophercoin/mining"
	"github.com/murlokito/gophercoin/peer"
	"github.com/murlokito/gophercoin/wallet"
	"sync"
)

// gcdMain is the real main entrypoint for gophercoind.  It is necessary to work around
// the fact that deferred functions do not run when os.Exit() is called.  The
// optional serverChan parameter is mainly used by the service code to be
// notified with the server once it is setup so it can gracefully stop it when
// requested from the service control manager.
func gcdMain(serverChan chan<- *Server, cfg *Config) error {
	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.
	logger := log.NewLogger(log.InfoLevel)
	logger.Info("Preparing to launch.")
	var (
		wg         sync.WaitGroup
		gcd        *Server
		w          *wallet.Wallet
		peerServer *peer.PeerServer
		miner      *mining.MinerServer
	)

	// attempt to load the wallet from file
	w, err := wallet.NewWallet(cfg.walletPath)
	if err != nil {
		return err
	}
	logger.WithDetails(
		log.NewDetail("wallet", cfg.walletPath),
	).Info("Successfully loaded wallet")

	// attempt to load the database from file
	chain, err := blockchain.NewBlockchain(cfg.dbPath)
	if err != nil {
		newChain, err := blockchain.CreateBlockchain(w.CreateAddress())
		if err != nil {
			return err
		}
		chain = newChain
	}
	logger.WithDetails(
		log.NewDetail("database", cfg.dbPath),
	).Info("Successfully loaded database")

	// attempt to load and reindex utxo set
	utxoSet := &blockchain.UTXOSet{
		Chain: chain,
		Mutex: &sync.RWMutex{},
	}

	logger.WithDetails(
		log.NewDetail("database", cfg.dbPath),
	).Info("Successfully loaded utxo set")

	// initialize the chain manager to pass onto other components
	chainMgr := blockchain.NewChainManager(chain, utxoSet)

	peerConfig := peer.Config{
		Port:     cfg.peerPort,
		LogLevel: log.InfoLevel,
	}
	// initialize the peer server for network communication
	peerServer = peer.NewPeerServer(peerConfig, &wg, chainMgr)
	peerServer.Start()
	logger.Info("Successfully started peer server")

	// initialize the mining server
	if cfg.miningNode {
		miner = mining.NewMinerServer(chainMgr, &wg, w.GetInitialAddress(), peerServer)
		miner.StartMiner()
		peerServer.MinerChan = miner.MinerChan
	}
	logger.Info("Successfully started mining server")

	// initialize the server that exposes the REST API
	gcd = NewServer(cfg, chainMgr, w, miner, &wg)
	gcd.StartServer()
	logger.Info("Successfully started api server")

	go func() {
		logger.Info("Reindexing UTXO Set.")

		utxoSet.Reindex()
		ctx := utxoSet.CountTransactions()
		logger.WithDetails(
			log.NewDetail("txcount", ctx),
		).Info("Finished reindexing UTXO Set.")
	}()

	if serverChan != nil {
		serverChan <- gcd
	}
	// Enter an infinite select, the daemon is still stoppable
	// due to the goroutine checking for a sigterm
	select {}

}

// Main is the entrypoint for the Gophercoin Daemon
func Main() error {

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Work around defer not working after os.Exit()
	if err := gcdMain(nil, cfg); err != nil {
		return err
	}

	return nil
}
