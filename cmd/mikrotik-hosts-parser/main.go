// Main CLI application entrypoint.
package main

import (
	"os"
	"path/filepath"

	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/cli"
)

func main() {
	cmd, errorHandler := cli.NewCommand(filepath.Base(os.Args[0]))

	if err := cmd.Execute(); err != nil {
		errorHandler(err)
	}
}
