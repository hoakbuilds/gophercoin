package gcd

import (
	"fmt"
	"os"
)

// gcdMain is the real main function for btcd.  It is necessary to work around
// the fact that deferred functions do not run when os.Exit() is called.  The
// optional serverChan parameter is mainly used by the service code to be
// notified with the server once it is setup so it can gracefully stop it when
// requested from the service control manager.
func gcdMain(serverChan chan<- *gcdServer) error {
	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.
	s := &GrpcServer{}

	s.startGRPCServer()

	db := NewBlockchain()
	ws, err := NewWallets()
	if err != nil {
		return fmt.Errorf("Failed to create new Wallet: %v", err)
	}
	utxoset := &UTXOSet{
		chain: db,
	}

	gcd := &gcdServer{
		db:        db,
		ws:        ws,
		utxoSet:   utxoset,
		rpcServer: s,
	}

	gcd.StartServer()
	if serverChan != nil {
		serverChan <- gcd
	}

	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	//<-interrupt
	return nil
}

// Main is the entrypoint for the Gophercoin Daemon
func Main() {

	// Work around defer not working after os.Exit()
	if err := gcdMain(nil); err != nil {
		os.Exit(1)
	}

}
