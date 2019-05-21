package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

const (
	defaultRESTHostPort = "9000"
)

func fatal(err error) {

	log.Printf("[gccli] %v\n", err)
	os.Exit(1)
}

func main() {
	app := cli.NewApp()
	app.Name = "Website Lookup CLI"
	app.Usage = "Let's you query IPs, CNAMEs, MX records and Name Servers!"

	// We'll be using the same flag for all our commands
	// so we'll define it up here
	flags := []cli.Flag{
		cli.StringFlag{
			Name:  "apiport",
			Value: defaultRESTHostPort,
		},
	}

	// we create our commands
	app.Commands = []cli.Command{
		{
			Name:  "createwallet",
			Usage: "Creates the user's wallet",
			Flags: flags,
			Action: func(c *cli.Context) error {

				resp, err := RequestURL(":" + defaultRESTHostPort + "/create_wallet")

				if err != nil {
					log.Printf("Wallet create successfully. New address: %+v", resp["Address"])
				}

				return err
			},
		},
		{
			Name:  "newaddress",
			Usage: "Generates a new address from the user wallet",
			Flags: flags,
			Action: func(c *cli.Context) error {

				resp, err := RequestURL(":" + defaultRESTHostPort + "/new_address")

				if err != nil {
					log.Printf("Wallet create successfully. New address: %+v", resp["Address"])
				}

				return err
			},
		},
	}
}
