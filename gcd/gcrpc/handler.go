package gcrpc

import (
	"log"

	"golang.org/x/net/context"
)

// Server represents the gRPC server
type Server struct {
}

// GetBalance generates response to a Ping request
func (s *Server) GetBalance(ctx context.Context, in *GetBalanceRequest) (*GetBalanceResponse, error) {
	log.Printf("Received GetBalance for Address: %s", in.Address)

	return &GetBalanceResponse{Balance: balance}, nil
}
