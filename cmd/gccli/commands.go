package main

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli"
)

func printRespJSON(resp []byte) {
	json, err := json.MarshalIndent(resp, "", "\t")

	if err != nil {
		fmt.Println("unable to decode response: ", err)
		return
	}

	fmt.Println(json)
}

var newAddressCommand = cli.Command{
	Name:     "newaddress",
	Category: "Wallet",
	Usage:    "Generates a new address.",
	Description: `
	Generate a wallet new address.`,
	Action: actionDecorator(newAddress),
}

func newAddress(ctx *cli.Context) error {

	args := ctx.Args()

	fmt.Printf("%s", args)

	resp, err := RequestURL("/new_address", defaultRESTHostPort)

	if err != nil {
		return err
	}
	printRespJSON(resp)
	return nil
}
