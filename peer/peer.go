package peer

import (
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/murlokito/gophercoin/log"

	"github.com/murlokito/gophercoin/blockchain"
)

// Peer is the structure that defines a peer
// in the peer to peer network
type Peer struct {
	// address defines the peer's IP address
	Address string `json:"Address"`
	// version defines the peer's best block height
	Version int64 `json:"Version"`
}

// PeerServer is the structure that defines the peer server
// in the peer to peer network
type PeerServer struct {
	Config      Config
	KnownNodes  []Peer
	NodeAddress string
	MinerChan   chan []byte

	listener        net.Listener
	chainMgr        *blockchain.ChainManager
	blocksInTransit [][]byte
	wg              *sync.WaitGroup
	logger          log.Logger
}

// Start is the function used to start the PeerServer
func (s PeerServer) Start() {
	defer s.wg.Done()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		s.logger.Info("Catching signal, terminating gracefully.")

		os.Exit(1)
	}()

	if s.Config.Port != "" {
		s.NodeAddress = ":" + s.Config.Port
	} else {
		s.NodeAddress = ":" + DefaultProtocolPort
	}

	lis, err := net.Listen(protocol, s.NodeAddress)
	if err != nil {
		s.logger.WithError(err)
		return
	}
	s.listener = lis
	s.logger.Info("PeerServer listening on port %s", s.NodeAddress)

	go s.Listen()
	s.wg.Add(1)
}

// Listen to peer connections
func (s PeerServer) Listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.logger.WithError(err)
		}
		go s.handleConnection(conn)
		s.wg.Add(1)
	}
}

// NewPeerServer creates a new peer server with the passed config
func NewPeerServer(config Config, wg *sync.WaitGroup, chainMgr *blockchain.ChainManager) *PeerServer {
	server := &PeerServer{
		KnownNodes:      make([]Peer, 0),
		MinerChan:       nil,
		chainMgr:        chainMgr,
		blocksInTransit: make([][]byte, 0),
		wg:              wg,
		logger:          log.NewLogger(config.LogLevel),
	}

	go server.Start()
	wg.Add(1)

	return server
}
