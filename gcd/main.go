package gcd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// gcdMain is the real main function for gcd.  It is necessary to work around
// the fact that deferred functions do not run when os.Exit() is called.  The
// optional serverChan parameter is mainly used by the service code to be
// notified with the server once it is setup so it can gracefully stop it when
// requested from the service control manager.
func gcdMain(serverChan chan<- *gcdServer) error {
	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.

	var sync sync.WaitGroup

	w, err := NewWallet()
	if err != nil {
		return fmt.Errorf("Failed to create new Wallet: %+v", err)
	}

	grpcChan := make(chan string)
	s := &GrpcServer{
		grpcChan: grpcChan,
		wg:       &sync,
	}

	go s.StartGRPCServer(w)
	sync.Add(1)

	db, err := NewBlockchain()
	if err != nil {
		log.Printf("[GCD] %v", err)
	}

	var address string

	if len(w.Wallet) > 0 {
		log.Printf("[GCD] Wallet with %d address(es) found", len(w.Wallet))
		address = w.GetInitialAddress()
	} else {
		address = w.CreateAddress()
		w.SaveToFile()
	}

	if err != nil && err.Error() == noExistingBlockchainFound {
		db = CreateBlockchain(address)
	}

	utxoset := &UTXOSet{
		chain: db,
	}

	externalAddress := getExternalAddress()

	if externalAddress == "" {
		log.Print("[GCD] Could not fetch external address, exiting.")
	} else {
		log.Printf("[GCD] External address: %s", externalAddress)
	}

	gcd := &gcdServer{
		db:              db,
		wallet:          w,
		utxoSet:         utxoset,
		rpcServer:       s,
		knownNodes:      []Peer{},
		nodeAddress:     externalAddress + ":" + string(defaultProtocolPort),
		blocksInTransit: [][]byte{},
		memPool:         map[string]Transaction{},
		wg:              &sync,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Printf("[GCD] Catching signal, terminating gracefully.")
		os.Exit(1)
	}()

	go gcd.StartServer()
	sync.Add(1)

	if serverChan != nil {
		serverChan <- gcd
	}
	select {
	case <-grpcChan:

	}
	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	//<-interrupt
	return nil
}

// Main is the entrypoint for the Gophercoin Daemon
func Main() error {

	// Work around defer not working after os.Exit()
	if err := gcdMain(nil); err != nil {
		return err
	}

	return nil
}
