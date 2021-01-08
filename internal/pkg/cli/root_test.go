package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSubcommands(t *testing.T) {
	cmd, _ := NewCommand("unit test")

	cases := []struct {
		giveName string
	}{
		{giveName: "serve"},
		{giveName: "version"},
	}

	// get all existing subcommands and put into the map
	subcommands := make(map[string]*cobra.Command)
	for _, sub := range cmd.Commands() {
		subcommands[sub.Name()] = sub
	}

	for _, tt := range cases {
		t.Run(tt.giveName, func(t *testing.T) {
			if _, exists := subcommands[tt.giveName]; !exists {
				assert.Failf(t, "command not found", "command %s was not found", tt.giveName)
			}
		})
	}
}

func TestFlags(t *testing.T) {
	cmd, _ := NewCommand("unit test")

	cases := []struct {
		giveName      string
		wantShorthand string
	}{
		{giveName: "verbose", wantShorthand: "v"},
	}

	for _, tt := range cases {
		t.Run(tt.giveName, func(t *testing.T) {
			flag := cmd.Flag(tt.giveName)

			if flag == nil {
				assert.Failf(t, "flag not found", "flag %s was not found", tt.giveName)

				return
			}

			assert.Equal(t, tt.wantShorthand, flag.Shorthand)
		})
	}
}
