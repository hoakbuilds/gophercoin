package gcd

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/mux"
)

const (
	defaultProtocolPort = "3000"
	defaultRPCHostPort  = "7777"
	defaultRESTHostPort = "7778"
	protocol            = "tcp"
	nodeVersion         = 1
	commandLength       = 12
)

// GcdServer is the structure which defines the Gophercoin
// Daemon
type GcdServer struct {
	db              *Blockchain
	wallet          *Wallet
	utxoSet         *UTXOSet
	Router          *mux.Router
	knownNodes      []Peer
	nodeAddress     string
	blocksInTransit [][]byte
	memPool         map[string]Transaction
	miningAddress   string

	wg *sync.WaitGroup
}

// StartServer is the function used to start the gcd Server
func (s *GcdServer) StartServer() {
	// create a listener on TCP port
	lis, err := net.Listen(protocol, "127.0.0.1:"+defaultProtocolPort)
	if err != nil {
		log.Printf("failed to listen: %v", err)
	}
	log.Printf("[GCD] PeerServer listening on port %s", defaultProtocolPort)

	if len(s.knownNodes) > 0 {
		if s.nodeAddress != s.knownNodes[0].address {
			log.Printf("[PRSV] sending version message to %s\n", s.knownNodes[0].address)
			s.sendVersion(s.knownNodes[0].address, s.db)
		}
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Panic(err)
		}
		go s.handleConnection(conn, s.db)
		s.wg.Add(1)
	}

}

func getExternalAddress() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		log.Printf("[PRSV] Unable to fetch external ip address.")
		os.Exit(1)
	}
	defer resp.Body.Close()
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[PRSV] Unable to read external ip address response.")
		os.Exit(1)
	}

	return string(r)
}
