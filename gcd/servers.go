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

// Server is the structure which defines the Gophercoin
// Daemon
type Server struct {
	cfg     Config
	db      *Blockchain
	wallet  *Wallet
	utxoSet *UTXOSet

	Router *mux.Router

	knownNodes      []Peer
	nodeAddress     string
	blocksInTransit [][]byte
	memPool         map[string]Transaction
	miningAddress   string

	wg *sync.WaitGroup
}

// StartServer is the function used to start the gcd Server
func (s *Server) StartServer() {
	defer s.wg.Done()
	// create a listener on TCP port
	var lis net.Listener

	if s.cfg.peerPort != "" {

		lst, err := net.Listen(protocol, "127.0.0.1:"+s.cfg.peerPort)
		if err != nil {
			log.Printf("failed to listen: %v", err)
			return
		}
		lis = lst
	} else {
		log.Printf("failed to listen: %v", s.cfg)
		lst, err := net.Listen(protocol, "127.0.0.1:"+defaultProtocolPort)
		if err != nil {
			log.Printf("failed to listen: %v", err)
			return
		}
		lis = lst
	}

	log.Printf("[GCD] PeerServer listening on port %s", s.nodeAddress)

	if len(s.knownNodes) > 0 {
		if s.nodeAddress != s.knownNodes[0].Address {
			log.Printf("[PRSV] sending version message to %s\n", s.knownNodes[0].Address)
			s.sendVersion(s.knownNodes[0].Address, s.db)
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

// StartMiner is the function used to start the gcd Server
func (s *Server) StartMiner(msgChan chan interface{}, nodeServ chan interface{}, quitChan chan int) {
	defer s.wg.Done()
	log.Printf("[GCMNR] Miner ready")

	for {
		select {
		case <-quitChan:
			break
		case msg := <-msgChan:
			log.Printf("[GCMNR] Received %v", msg)
		}

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
