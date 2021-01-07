package version

import (
	"fmt"
	"os"
	"runtime"

	ver "github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/version"
)

type Command struct{}

// Execute version command.
func (*Command) Execute(_ []string) error {
	_, _ = fmt.Fprintf(os.Stdout, "Version:\t%s (%s)\n", ver.Version(), runtime.Version())

	return nil
}
