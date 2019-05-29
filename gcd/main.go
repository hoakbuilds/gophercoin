package gcd

import (
	"fmt"
	"log"
	"sync"
)

// gcdMain is the real main function for gcd.  It is necessary to work around
// the fact that deferred functions do not run when os.Exit() is called.  The
// optional serverChan parameter is mainly used by the service code to be
// notified with the server once it is setup so it can gracefully stop it when
// requested from the service control manager.
func gcdMain(serverChan chan<- *Server, cfg Config) error {
	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.
	log.Printf("[GCD] Preparing to launch.")
	var (
		wg    sync.WaitGroup
		gcd   *Server
		miner *MinerServer
	)

	externalAddress := getExternalAddress()

	if externalAddress == "" {
		return fmt.Errorf("[GCD] Could not fetch external address")
	}

	log.Printf("[GCD] External address: %s", externalAddress)

	// base Server structure, after declaring it we try to initiate
	// some of the components from a possible config structure
	gcd = &Server{
		knownNodes:      []Peer{},
		nodeAddress:     externalAddress + ":" + string(defaultProtocolPort),
		blocksInTransit: [][]byte{},
		memPool:         map[string]Transaction{},
		wg:              &wg,
		cfg:             cfg,
		nodeServChan:    make(chan interface{}),
	}
	// base MinerServer structure, after declaring it we try to initiate
	// some of the components from a possible config structure
	miner = &MinerServer{
		server:    gcd,
		quitChan:  make(chan int),
		minerChan: make(chan []byte, 5),
		timeChan:  make(chan int64, 5),
		miningTxs: false,
	}

	// In case a config structure was able to be built from flags
	if (cfg != Config{}) {
		// In case flags provide a peer port
		if cfg.peerPort != "" {
			gcd.nodeAddress = ":" + string(cfg.peerPort)
		}
		// In case flags provide a peer port
		if cfg.restPort != "" {
			gcd.nodeAddress = ":" + string(cfg.peerPort)
		}
		// In case flags provide a wallet path
		if cfg.walletPath != "" {
			wallet, err := NewWallet()
			if err != nil {
				log.Printf("[GCD] Failed to create/load Wallet: %+v", err)

			} else {
				log.Printf("[GCD] Wallet successfully opened.")
				gcd.wallet = wallet
			}

		}
		// In case flags provide a wallet path
		if cfg.dbPath != "" {
			// initialize blockchain
			db, err := NewBlockchain(cfg.dbPath)
			if err != nil {
				log.Printf("[GCD] Failed to create/load Wallet: %+v", err)
				gcd.db = db
				miner.db = db
			} else {
				log.Printf("[GCD] Database successfully opened.")
				log.Printf("[GCD] Chain Tip: %v ", db.tip)
				gcd.db = db
				miner.db = db
				// perform utxo reindexing task
				UTXOSet := UTXOSet{
					chain: gcd.db,
					mutex: &sync.RWMutex{},
				}
				gcd.utxoSet = &UTXOSet
				go func() {
					log.Printf("[GCDB] Reindexing UTXO Set.")

					gcd.utxoSet.Reindex()
					ctx := gcd.utxoSet.CountTransactions()
					log.Printf("[GCDB] Finished reindexing UTXO Set, there are %d transactions in it.", ctx)

				}()

			}
		}
	}

	if gcd.cfg.miningAddr != "" {
		miner.miningAddress = gcd.cfg.miningAddr
	}

	gcd.miner = miner

	if gcd.cfg.restPort != "" {
		log.Printf("[GCD] Starting API Server.")
		go gcd.BuildAndServeAPI()
		gcd.wg.Add(1)
	}

	if gcd.cfg.miningNode == true {
		log.Printf("[GCD] Starting Mining Server.")
		go gcd.miner.StartMiner()
		gcd.wg.Add(1)
	}

	go gcd.StartServer()
	gcd.wg.Add(1)

	if serverChan != nil {
		serverChan <- gcd
	}
	// Enter an infinite select, the daemon is still stoppable
	// due to the goroutine checking for a sigterm
	select {}

}

// Main is the entrypoint for the Gophercoin Daemon
func Main() error {

	cfg := loadConfig()

	// Work around defer not working after os.Exit()
	if err := gcdMain(nil, cfg); err != nil {
		return err
	}

	return nil
}
