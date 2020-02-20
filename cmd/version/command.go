package version

import (
	"fmt"
	ver "mikrotik-hosts-parser/version"
)

type Command struct{}

// Execute version command.
func (*Command) Execute(_ []string) error {
	fmt.Printf("Version: %s\n", ver.Version())

	return nil
}
