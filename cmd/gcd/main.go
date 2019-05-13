package main

import (
	"fmt"
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
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}
