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
		sync sync.WaitGroup
		gcd  *Server
	)

	externalAddress := getExternalAddress()

	if externalAddress == "" {
		return fmt.Errorf("[GCD] Could not fetch external address")
	} else {
		log.Printf("[GCD] External address: %s", externalAddress)
	}

	// base Server structure, after declaring it we try to initiate
	// some of the components from a possible config structure
	gcd = &Server{
		knownNodes:      []Peer{},
		nodeAddress:     externalAddress + ":" + string(defaultProtocolPort),
		blocksInTransit: [][]byte{},
		memPool:         map[string]Transaction{},
		wg:              &sync,
		cfg:             cfg,
		quitChan:        make(chan int),
		minerChan:       make(chan interface{}),
		nodeServChan:    make(chan interface{}),
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
				gcd.wallet = wallet
			}

		}
		// In case flags provide a wallet path
		if cfg.dbPath != "" {
			// initialize blockchain
			db, err := NewBlockchain(cfg.dbPath)
			if err != nil {
				log.Printf("[GCD] Failed to create/load Wallet: %+v", err)
			} else {
				log.Printf("[GCD] Database successfully opened.")
				gcd.db = db
				// perform utxo reindexing task
				UTXOSet := UTXOSet{
					chain: gcd.db,
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
	if gcd.cfg.restPort != "" {
		go gcd.BuildAndServeAPI()
		sync.Add(1)
	}

	go gcd.StartServer()
	sync.Add(1)

	if gcd.cfg.miningNode == true {
		go gcd.StartMiner()
		sync.Add(1)
	}

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
