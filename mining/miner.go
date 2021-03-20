package mining

import (
	"encoding/hex"
	"github.com/murlokito/gophercoin/peer"
	"github.com/murlokito/gophercoin/transaction"
	"os"
	"sync"
	"time"

	"github.com/murlokito/gophercoin/blockchain"
	"github.com/murlokito/gophercoin/log"
)

// MinerServer is the structure which defines the
// mining server
type MinerServer struct {
	MinerChan     chan []byte
	logger        log.Logger
	wg            *sync.WaitGroup
	peerServer    *peer.PeerServer
	chainMgr      *blockchain.ChainManager
	quitChan      chan int
	timeChan      chan int64
	miningAddress string
	miningTxs     bool
}

// StartMiner is the function used to start the gophercoin miner
func (s *MinerServer) StartMiner() {
	defer s.wg.Done()
	s.logger.Info("Miner ready")
	go s.timeAdjustment()
	s.wg.Add(1)

	go s.Mine()
	s.wg.Add(1)
}

func (s *MinerServer) Mine() {
	for {
		select {
		case <-s.quitChan:
			s.logger.Info("Received stop signal")
			break
		case msg := <-s.MinerChan:
			s.logger.Info("Received tx with ID %v", msg)

			if len(s.chainMgr.MemPool) > 2 && !s.miningTxs {
				s.miningTxs = true
				t := time.Now().Unix()
				s.mineTxs()
				now := time.Now().Unix()
				diff := (t - now)
				s.logger.Info("Mined new block after %v seconds.", diff)
				s.miningTxs = false
			}

		case msg := <-s.timeChan:
			if msg > 15 && s.miningTxs {
				t := time.Now().Unix()
				s.mineTxs()
				now := time.Now().Unix()
				diff := (t - now)
				s.logger.Info("Mined new block after %v seconds.", diff)
				s.miningTxs = false
			}

		}

	}

}

func (s *MinerServer) mineTxs() {
	var txs []*transaction.Transaction
	for id := range s.chainMgr.MemPool {
		tx := s.chainMgr.MemPool[id]
		s.logger.Info("Verifying transaction: %s\n", id)
		if s.chainMgr.Chain.VerifyTransaction(&tx) {
			s.logger.Info("Verified transaction: %s\n", id)
			txs = append(txs, &tx)
		}
	}

	if len(txs) == 0 {
		s.logger.Info("No valid transactions in mempool")
	}
	var cbTx *transaction.Transaction

	cbTx = transaction.NewCoinbaseTX(s.miningAddress, "")

	txs = append(txs, cbTx)
	s.logger.Info("Block transactions aggregated: \n%v", txs)
	newBlock := s.chainMgr.Chain.MineBlock(txs)
	s.logger.Info("New block is mined!")

	go func() {
		s.logger.Info("Reindexing UTXO Set.")

		s.chainMgr.UTXOSet.Reindex()
		ctx := s.chainMgr.UTXOSet.CountTransactions()
		s.logger.Info("Finished reindexing UTXO Set, there are %d transactions in it.", ctx)
	}()

	s.logger.Info("Reindexing UTXO Set.")

	for _, tx := range txs {
		txID := hex.EncodeToString(tx.ID)
		delete(s.chainMgr.MemPool, txID)
	}
	for _, node := range s.peerServer.KnownNodes {
		if node.Address != s.peerServer.NodeAddress {
			s.peerServer.SendInv(node.Address, "block", [][]byte{newBlock.Hash})
		}
	}
}

func (s *MinerServer) timeAdjustment() {
	defer s.wg.Done()

	if !s.miningTxs {
		if s.chainMgr.Chain != nil {
			tip := s.chainMgr.Chain.Tip
			block, err := s.chainMgr.Chain.GetBlock(tip)

			if err != nil {
				s.logger.Info("Unable to fetch blockchain tip.")
				os.Exit(1)
			}

			now := time.Now().Unix()
			diff := now - block.Timestamp

			if diff > 15 && !s.miningTxs {
				s.logger.Info("Elapsed since last block: %v seconds.", diff)
				s.miningTxs = true
				s.timeChan <- diff
			}

		}
	}

	time.Sleep(1000)
	go s.timeAdjustment()
	s.wg.Add(1)

	return
}

func NewMinerServer(chainMgr *blockchain.ChainManager, wg *sync.WaitGroup, miningAddr string, peerServer *peer.PeerServer) *MinerServer {
	return &MinerServer{
		MinerChan:     make(chan []byte, 5),
		peerServer:    peerServer,
		wg:            wg,
		logger:        log.NewLogger(log.InfoLevel),
		chainMgr:      chainMgr,
		quitChan:      make(chan int),
		timeChan:      make(chan int64, 5),
		miningTxs:     false,
		miningAddress: miningAddr,
	}
}
