package gcrpc

import (
	fmt "fmt"

	"golang.org/x/net/context"
)

// Server represents the gRPC server
type Server struct {
}

// GetBalance generates response to a Ping request
func (s *Server) GetBalance(ctx context.Context, in *GetBalanceRequest) (*GetBalanceResponse, error) {
	fmt.Printf("Received GetBalance for Address: %s", in.Address)
	balance := int64(0)
	return &GetBalanceResponse{Balance: balance}, nil
}

// NewAddress generates response to a Ping request
func (s *Server) NewAddress(ctx context.Context, in *NewAddressRequest) (*NewAddresResponse, error) {
	fmt.Printf("Received request for new address")
	address := string("satoshi ples!s")
	return &NewAddresResponse{Address: address}, nil
}
