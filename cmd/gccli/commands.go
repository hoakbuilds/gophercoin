package main

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
<<<<<<< HEAD
	"github.com/murlokito/gophercoin/gcd/gcrpc"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
=======
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/urfave/cli"
>>>>>>> d3233990347c6be6c9d1316dbc6bc74557aa1242
	"google.golang.org/grpc/status"
)

func printRespJSON(resp proto.Message) {
	jsonMarshaler := &jsonpb.Marshaler{
		EmitDefaults: true,
		Indent:       "    ",
	}

	jsonStr, err := jsonMarshaler.MarshalToString(resp)
	if err != nil {
		fmt.Println("unable to decode response: ", err)
		return
	}

	fmt.Println(jsonStr)
}

// actionDecorator is used to add additional information and error handling
// to command actions.
func actionDecorator(f func(*cli.Context) error) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if err := f(c); err != nil {
			s, ok := status.FromError(err)
<<<<<<< HEAD
			if s != nil {
				fmt.Printf("[gccli] %v", s)
			}
			if ok != false {
				fmt.Printf("[gccli] %v", s)
			}
=======
>>>>>>> d3233990347c6be6c9d1316dbc6bc74557aa1242
		}
		return nil
	}
}

var getBalanceCommand = cli.Command{
	Name:      "getbalance",
	Category:  "Wallet",
	Usage:     "Fetches address balance.",
<<<<<<< HEAD
	ArgsUsage: "address",
=======
	ArgsUsage: "address-type",
>>>>>>> d3233990347c6be6c9d1316dbc6bc74557aa1242
	Description: `
				Fetches balance for given address.
				`,
	Action: actionDecorator(getBalance),
}

<<<<<<< HEAD
func getClient(ctx *cli.Context) (gcrpc.GCDClient, func()) {

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial("localhost:7777", opts...)
	cleanUp := func() {
		conn.Close()
	}
	if err != nil {
		fmt.Printf("failure while dialing: %v", err)
	}
	defer conn.Close()

	client := gcrpc.NewGCDClient(conn)

	return client, cleanUp
}

=======
>>>>>>> d3233990347c6be6c9d1316dbc6bc74557aa1242
func getBalance(ctx *cli.Context) error {
	client, cleanUp := getClient(ctx)
	defer cleanUp()

<<<<<<< HEAD
	stringAddr := ctx.Args().First()

	ctxb := context.Background()
	addr, err := client.GetBalance(ctxb, &gcrpc.GetBalanceRequest{
		Address: stringAddr,
=======
	stringAddrType := ctx.Args().First()

	// Map the string encoded address type, to the concrete typed address
	// type enum. An unrecognized address type will result in an error.
	var addrType lnrpc.AddressType
	switch stringAddrType { // TODO(roasbeef): make them ints on the cli?
	case "p2wkh":
		addrType = lnrpc.AddressType_WITNESS_PUBKEY_HASH
	case "np2wkh":
		addrType = lnrpc.AddressType_NESTED_PUBKEY_HASH
	default:
		return fmt.Errorf("invalid address type %v, support address type "+
			"are: p2wkh and np2wkh", stringAddrType)
	}

	ctxb := context.Background()
	addr, err := client.NewAddress(ctxb, &lnrpc.NewAddressRequest{
		Type: addrType,
>>>>>>> d3233990347c6be6c9d1316dbc6bc74557aa1242
	})
	if err != nil {
		return err
	}

	printRespJSON(addr)
	return nil
}
