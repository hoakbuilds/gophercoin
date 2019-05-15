package main

import (
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/murlokito/gophercoin/gcd"
)

func main() {

	// Call the "real" main in a nested manner so the defers will properly
	// be executed in the case of a graceful shutdown.
	if err := gcd.Main(); err != nil {
		if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
		} else {
			log.Println(err)
		}
		os.Exit(1)
	}
}
