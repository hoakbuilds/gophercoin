package gcd

import (
	"fmt"
	"log"
	"net"

	"github.com/murlokito/gophercoin/gcd/gcrpc"
	"google.golang.org/grpc"
)

const (
	defaultProtocolPort = 3000
	defaultRPCHostPort  = 7777
	defaultRESTHostPort = 7778
	protocol            = "tcp"
	nodeVersion         = 1
	commandLength       = 12
)

type GrpcServer struct {
}

type RestServer struct {
}

type PeerServer struct {
}

type gcdServer struct {
	db        *Blockchain
	ws        *Wallets
	utxoSet   *UTXOSet
	rpcServer *GrpcServer
}

func (s *GrpcServer) startGRPCServer() error {
	// create a listener on TCP port
	lis, err := net.Listen(protocol, address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	log.Print("[gcd] Listening on port 7777")

	// create a server instance
	gcServer := gcrpc.Server{}

	// create a gRPC server object
	grpcServer := grpc.NewServer()

	// attach the Ping service to the server
	gcrpc.RegisterGCDServer(grpcServer, &gcServer)

	// start the server
	log.Printf("[gcd] Starting HTTP/2 gRPC server on %s", address)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %s", err)
	}

	return nil
}

// StartServer is the function used to start a server listening for peers.
func (s gcdServer) StartServer() error {

}
