// Package cli contains CLI command handlers.
package cli

import (
	"context"

	"github.com/spf13/cobra"

	healthcheckCmd "gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/cli/healthcheck"
	serveCmd "gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/cli/serve"
	versionCmd "gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/cli/version"
	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/logger"
	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/version"

	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/checkers"
)

// NewCommand creates root command.
func NewCommand(appName string) *cobra.Command {
	var (
		verbose bool
		debug   bool
		logJSON bool
	)

	ctx := context.Background() // main CLI context

	// create "default" logger (will be overwritten later with customized)
	log, err := logger.New(false, false, false)
	if err != nil {
		panic(err)
	}

	cmd := &cobra.Command{
		Use: appName,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			_ = log.Sync() // sync previous logger instance

			customizedLog, e := logger.New(verbose, debug, logJSON)
			if e != nil {
				return e
			}

			*log = *customizedLog // override "default" logger with customized

			return nil
		},
		PersistentPostRun: func(*cobra.Command, []string) {
			// error ignoring reasons:
			// - <https://github.com/uber-go/zap/issues/772>
			// - <https://github.com/uber-go/zap/issues/328>
			_ = log.Sync()
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	cmd.PersistentFlags().BoolVarP(&debug, "debug", "", false, "debug output")
	cmd.PersistentFlags().BoolVarP(&logJSON, "log-json", "", false, "logs in JSON format")

	cmd.AddCommand(
		versionCmd.NewCommand(version.Version()),
		serveCmd.NewCommand(ctx, log),
		healthcheckCmd.NewCommand(checkers.NewHealthChecker(ctx)),
	)

	return cmd
}
