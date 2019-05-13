package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

const (
	defaultRPCHostPort = "7777"
)

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "[gccli] %v\n", err)
	os.Exit(1)
}

func main() {
	app := cli.NewApp()
	app.Name = "gccli"
	app.Usage = "control plane for your Gophercoin Daemon (gcd)"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "rpcserver",
			Value: defaultRPCHostPort,
			Usage: "host:port of gophercoin daemon",
		},
	}
	app.Commands = []cli.Command{
		getBalanceCommand,
	}
	if err := app.Run(os.Args); err != nil {
		fatal(err)
	}
}
