package version

import "fmt"

type Command struct{}

var Version string

func (*Command) Execute(_ []string) error {
	if Version == "" {
		panic("Version value must be initialized outside current package BEFORE command execution")
	}

	fmt.Printf("Version: %s\n", Version)

	return nil
}
