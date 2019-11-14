package version

import (
	"errors"
	"fmt"
)

type Command struct{}

// Version value (must be set before command execution outside of current package)
var Version string

// Execute version command.
func (*Command) Execute(_ []string) error {
	if Version == "" {
		return errors.New("version value must be initialized outside current package BEFORE command execution")
	}

	fmt.Printf("Version: %s\n", Version)

	return nil
}
