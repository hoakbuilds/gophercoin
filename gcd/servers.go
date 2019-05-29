package gcd

import (
	"encoding/hex"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

// MinerServer is the structure which defines the
// mining server
type MinerServer struct {
	db            *Blockchain
	server        *Server
	quitChan      chan int
	minerChan     chan []byte
	timeChan      chan int64
	miningAddress string
	miningTxs     bool
}

// Server is the structure which defines the Gophercoin
// Daemon
type Server struct {
	cfg     Config
	db      *Blockchain
	wallet  *Wallet
	utxoSet *UTXOSet

	miner  *MinerServer
	Router *mux.Router

	knownNodes      []Peer
	nodeAddress     string
	blocksInTransit [][]byte
	memPool         map[string]Transaction

	wg *sync.WaitGroup

	nodeServChan chan interface{}
}

// StartServer is the function used to start the gcd Server
func (s *Server) StartServer() {
	defer s.wg.Done()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Printf("[PRSRV] Catching signal, terminating gracefully.")
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

	log.Printf("[PRSRV] PeerServer listening on: %s", s.nodeAddress)

	if len(s.knownNodes) > 0 {
		if s.nodeAddress != s.knownNodes[0].Address {
			log.Printf("[PRSRV] sending version message to %s\n", s.knownNodes[0].Address)
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
func (s *MinerServer) StartMiner() {
	defer s.server.wg.Done()
	log.Printf("[GCMNR] Miner ready")
	go s.timeAdjustment()
	s.server.wg.Add(1)
	for {
		select {
		case <-s.quitChan:
			log.Printf("[GCMNR] Received stop signal")
			break
		case msg := <-s.minerChan:
			log.Printf("[GCMNR] Received tx with ID %v", msg)

			if len(s.server.memPool) > 2 && !s.miningTxs {
				s.miningTxs = true
				t := time.Now().Unix()
				s.mineTxs()
				now := time.Now().Unix()
				diff := (t - now)
				log.Printf("[GCMNR] Mined new block after %v seconds.", diff)
				s.miningTxs = false
			}

		case msg := <-s.timeChan:
			if msg > 2 && s.miningTxs {
				t := time.Now().Unix()
				s.mineTxs()
				now := time.Now().Unix()
				diff := (t - now)
				log.Printf("[GCMNR] Mined new block after %v seconds.", diff)
				s.miningTxs = false
			}

		}

	}

}

func getExternalAddress() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		log.Printf("[PRSRV] Unable to fetch external ip address.")
		os.Exit(1)
	}
	defer resp.Body.Close()
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[PRSRV] Unable to read external ip address response.")
		os.Exit(1)
	}

	return string(r)
}

func (s *MinerServer) mineTxs() {
	var txs []*Transaction
	for id := range s.server.memPool {
		tx := s.server.memPool[id]
		log.Printf("[GCMNR] Verifying transaction: %s\n", id)
		if s.server.db.VerifyTransaction(&tx) {
			log.Printf("[GCMNR] Verified transaction: %s\n", id)
			txs = append(txs, &tx)
		}
	}

	if len(txs) == 0 {
		log.Println("[GCMNR] No valid transactions in mempool")
	}
	var cbTx *Transaction
	if s.miningAddress == "" {
		log.Println("[GCMNR] No mining address from config structure")
		s.miningAddress = s.server.wallet.CreateAddress()
		log.Printf("[GCMNR] New mining address: %v", s.miningAddress)
		cbTx = NewCoinbaseTX(s.miningAddress, "")
	} else {
		cbTx = NewCoinbaseTX(s.miningAddress, "")
	}
	txs = append(txs, cbTx)
	log.Printf("[GCMNR] Block transactions aggregated: \n%v", txs)
	newBlock := s.server.db.MineBlock(txs)
	log.Println("[GCMNR] New block is mined!")

	go func() {
		log.Printf("[GCDB] Reindexing UTXO Set.")

		s.server.utxoSet.Reindex()
		ctx := s.server.utxoSet.CountTransactions()
		log.Printf("[GCDB] Finished reindexing UTXO Set, there are %d transactions in it.", ctx)

	}()

	log.Println("[GCMNR] Reindexing UTXO Set.")

	for _, tx := range txs {
		txID := hex.EncodeToString(tx.ID)
		delete(s.server.memPool, txID)
	}

	for _, node := range s.server.knownNodes {
		if node.Address != s.server.nodeAddress {
			s.server.sendInv(node.Address, "block", [][]byte{newBlock.Hash})
		}
	}
}

func (s *MinerServer) timeAdjustment() {
	defer s.server.wg.Done()

	if !s.miningTxs {
		if s.server.db != nil {
			tip := s.server.db.tip
			block, err := s.server.db.GetBlock(tip)

			if err != nil {
				log.Printf("[GCMNR] Unable to fetch blockchain tip.")
				os.Exit(1)
			}

			now := time.Now().Unix()
			diff := (now - block.Timestamp) / 60

			if diff > 2 && !s.miningTxs {
				log.Printf("[GCMNR] Elapsed since last block: %v minutes.", diff)
				s.miningTxs = true
				s.timeChan <- diff
			}

		}
	}

	time.Sleep(1000000000)
	go s.timeAdjustment()
	s.server.wg.Add(1)

	return

}
