package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

const (
	defaultRESTHostPort = "127.0.0.1:9000"
)

func fatal(err error) {

	log.Printf("[gccli] %v\n", err)
	os.Exit(1)
}

func init() {
	if len(os.Args) == 1 {
		log.Printf("Invalid usage, daemon host and port. Please use gccli -h")
		os.Exit(1)
	}
}

func main() {

	var app = cli.NewApp()
	app.Name = "gccli"
	app.Usage = "The control plane for the gophercoin daemon"
	app.Version = "0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "rest",
			Value: defaultRESTHostPort,
			Usage: "host:port of ln daemon REST API",
		},
	}
	app.Commands = []cli.Command{
		newAddressCommand,
	}

	if err := app.Run(os.Args); err != nil {
		fatal(err)
	}
}
