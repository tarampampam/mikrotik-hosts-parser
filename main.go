package main

import (
	"mikrotik-hosts-parser/cmd"
	"mikrotik-hosts-parser/cmd/version"
	"os"

	"github.com/jessevdk/go-flags"
)

var Version string = "undefined@undefined"

func main() {
	version.Version = Version

	// parse the arguments
	if _, err := flags.NewParser(&cmd.Options{}, flags.Default).Parse(); err != nil {
		// make error type checking
		if e, ok := err.(*flags.Error); (ok && e.Type != flags.ErrHelp) || !ok {
			os.Exit(1)
		}
	}
}
