package gcd

import (
	"flag"
	"os"
)

// Config is used as a structure to hold information
// passed via the CLI upon GCD startup. It's passed
// into the Server structure to define some of it's
// parameters
type Config struct {
	walletPath string
	dbPath     string

	peerPort string
	restPort string

	miningNode bool
}

func loadConfig() Config {

	var (
		walletvar string
		dbvar     string
		peervar   string
		restvar   string
		miningvar string
	)

	flag.StringVar(&walletvar, "wallet", "", "Path to the wallet.dat file.")
	flag.StringVar(&dbvar, "db", "", "Path to the blockchain.db file.")
	flag.StringVar(&peervar, "listen", "", "Port for the daemon to use to listen for peer connections")
	flag.StringVar(&restvar, "rest", "", "Port to use for the REST API server.")
	flag.StringVar(&miningvar, "mining", "", "Set to `true` to mine, `false` not to.")

	flag.Parse()
	if len(os.Args) == 0 {
		return Config{}
	}

	if miningvar == "true" {
		return Config{
			walletPath: walletvar,
			dbPath:     dbvar,
			peerPort:   peervar,
			restPort:   restvar,
			miningNode: true,
		}
	}

	return Config{
		walletPath: walletvar,
		dbPath:     dbvar,
		peerPort:   peervar,
		restPort:   restvar,
		miningNode: false,
	}
}
