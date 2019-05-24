package gcd

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

	quitChan     chan int
	minerChan    chan interface{}
	nodeServChan chan interface{}
}

// StartServer is the function used to start the gcd Server
func (s *Server) StartServer() {
	defer s.wg.Done()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Printf("[GCD] Catching signal, terminating gracefully.")
		if s.wallet != nil {
			s.wallet.SaveToFile()
		}

		os.Exit(1)
	}()
	// create a listener on TCP port
	var lis net.Listener

	if s.cfg.peerPort != "" {

		lst, err := net.Listen(protocol, ":"+s.cfg.peerPort)
		if err != nil {
			log.Printf("failed to listen: %v", err)
			return
		}
		lis = lst
	} else {
		lst, err := net.Listen(protocol, ":"+defaultProtocolPort)
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
			s.sendVersion(s.knownNodes[0].Address)
		}
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Panic(err)
		}
		go s.handleConnection(conn)
		s.wg.Add(1)
	}

}

// StartMiner is the function used to start the gcd Server
func (s *Server) StartMiner() {
	defer s.wg.Done()
	log.Printf("[GCMNR] Miner ready")

	for {
		select {
		case <-s.quitChan:
			log.Printf("[GCMNR] Received stop signal")
			break
		case msg := <-s.minerChan:
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
