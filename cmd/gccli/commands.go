package main

import (
	"context"
	"fmt"
	"log"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/murlokito/gophercoin/gcd/gcrpc"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func printRespJSON(resp proto.Message) {
	jsonMarshaler := &jsonpb.Marshaler{
		EmitDefaults: true,
		Indent:       "    ",
	}

	jsonStr, err := jsonMarshaler.MarshalToString(resp)
	if err != nil {
		log.Println("unable to decode response: ", err)
		return
	}

	log.Println(jsonStr)
}

// actionDecorator is used to add additional information and error handling
// to command actions.
func actionDecorator(f func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if err := f(c); err != nil {
			s, ok := status.FromError(err)
			if s != nil {
				log.Printf("[gccli] %v", s)
			}
			if ok != false {
				log.Printf("[gccli] %v", s)
			}
		}
		return nil
	}
}

var getBalanceCommand = cli.Command{
	Name:      "getbalance",
	Category:  "Wallet",
	Usage:     "Fetches address balance.",
	ArgsUsage: "address",
	Description: `
				Fetches balance for given address.
				`,
	Action: actionDecorator(getBalance),
}

func getClient(ctx *cli.Context) (gcrpc.GCDClient, func()) {

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial("localhost:7777", opts...)
	cleanUp := func() {
		conn.Close()
	}
	if err != nil {
		log.Printf("failure while dialing: %v", err)
	}
	defer conn.Close()

	client := gcrpc.NewGCDClient(conn)

	return client, cleanUp
}

func getBalance(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	stringAddr := ctx.Args().First()

	ctxb := context.Background()
	addr, err := client.GetBalance(ctxb, &gcrpc.GetBalanceRequest{
		Address: stringAddr,
	})
	if err != nil {
		return err
	}

	printRespJSON(addr)
	return nil
}

var newAddressCommand = cli.Command{
	Name:      "newaddress",
	Category:  "Wallet",
	Usage:     "Gets a new address from the wallet",
	ArgsUsage: "address",
	Description: `
				Fetches a new address from the wallet.
				`,
	Action: actionDecorator(newAddress),
}

func newAddress(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

	ctxb := context.Background()
	addr, err := client.NewAddress(ctxb, &gcrpc.NewAddressRequest{})

	fmt.Printf("New Address:\n")
	if err != nil {
		return err
	}
	printRespJSON(addr)
	return nil
}
