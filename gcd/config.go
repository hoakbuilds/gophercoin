package gcd

import (
	"errors"
	"flag"
	"os"
)

// Config is used as a structure to hold information
// passed via the CLI upon GCD startup. It's passed
// into the Server structure to define some of it's
// parameters
type Config struct {
	walletPath    string
	dbPath        string
	peerPort      string
	restPort      string
	restPassword  string
	miningAddr    string
	miningNode    bool
	restProtected bool
}

func loadConfig() (*Config, error) {

	var (
		walletvar    string
		dbvar        string
		peervar      string
		restvar      string
		protectedvar string
		passwordvar  string
		miningvar    string
		addrvar      string
		mining       = false
		protected    = false
	)

	flag.StringVar(&walletvar, "wallet", "", "Path to the wallet.dat file.")
	flag.StringVar(&dbvar, "db", "", "Path to the blockchain.db file.")
	flag.StringVar(&peervar, "listen", "", "Port for the daemon to use to listen for peer connections.")
	flag.StringVar(&restvar, "rest", "", "Port to use for the REST API server.")
	flag.StringVar(&protectedvar, "protected", "", "If the REST API should be protected by password.")
	flag.StringVar(&passwordvar, "password", "", "Password to protect the REST API.")
	flag.StringVar(&miningvar, "mining", "", "Set to `true` to mine, `false` not to.")
	flag.StringVar(&addrvar, "addr", "", "Address used for mining reward.")

	flag.Parse()
	if len(os.Args) == 0 {
		return nil, errors.New("no arguments given")
	}

	if protectedvar == "true" && passwordvar == "" {
		return nil, errors.New("must specify password")
	}

	if miningvar == "true" {
		mining = true
	}

	if protectedvar == "true" {
		protected = true
	}

	return &Config{
		walletPath:    walletvar,
		dbPath:        dbvar,
		peerPort:      peervar,
		restPort:      restvar,
		miningNode:    mining,
		miningAddr:    addrvar,
		restProtected: protected,
		restPassword:  passwordvar,
	}, nil
}
