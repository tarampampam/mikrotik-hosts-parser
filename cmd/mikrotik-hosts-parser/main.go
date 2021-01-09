// Main CLI application entrypoint.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/cli"
)

// exitFn is a function for application exiting.
var exitFn = os.Exit

// main CLI application entrypoint.
func main() { exitFn(run()) }

// run this CLI application.
// Exit codes documentation: <https://tldp.org/LDP/abs/html/exitcodes.html>
func run() int {
	cmd := cli.NewCommand(filepath.Base(os.Args[0]))

	if err := cmd.Execute(); err != nil {
		if _, outErr := fmt.Fprintln(os.Stderr, err.Error()); outErr != nil {
			panic(outErr)
		}

		return 1
	}

	return 0
}
