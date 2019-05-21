package gcd

import (
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
func gcdMain(serverChan chan<- *GcdServer) error {
	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.

	var sync sync.WaitGroup

	externalAddress := getExternalAddress()

	if externalAddress == "" {
		log.Print("[GCD] Could not fetch external address, exiting.")
	} else {
		log.Printf("[GCD] External address: %s", externalAddress)
	}

	gcd := &GcdServer{
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

	go gcd.BuildAndServeAPI()
	sync.Add(1)
	go gcd.StartServer()
	sync.Add(1)

	if serverChan != nil {
		serverChan <- gcd
	}
	// Enter an infinite select, the daemon is still stoppable
	// due to the goroutine checking for a sigterm
	select {}

}

// Main is the entrypoint for the Gophercoin Daemon
func Main() error {

	// Work around defer not working after os.Exit()
	if err := gcdMain(nil); err != nil {
		return err
	}

	return nil
}
