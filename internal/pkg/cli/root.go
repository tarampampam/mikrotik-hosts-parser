// Package cli contains CLI command handlers.
package cli

import (
	"github.com/spf13/cobra"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/cli/serve"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/cli/version"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewCommand creates root command.
func NewCommand(appName string) (*cobra.Command, func(error)) {
	var (
		verbose bool
		debug   bool
		logJson bool
	)

	log, _ := zap.NewProduction() // important: logger should be used only inside (sub)command Run* actions

	cmd := &cobra.Command{
		Use: appName,
		PersistentPreRun: func(*cobra.Command, []string) {
			*log = *createLogger(verbose, debug, logJson) // override "default" logger with customized
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
	cmd.PersistentFlags().BoolVarP(&logJson, "log-json", "", false, "logs in JSON format")

	cmd.AddCommand(
		version.NewCommand(),
		serve.NewCommand(log),
	)

	cmd.FlagErrorFunc()

	execErrHandler := func(err error) {
		log.Fatal(err.Error()) // `os.Exit(1)` here
	}

	return cmd, execErrHandler
}

func createLogger(verbose, debug, logJson bool) *zap.Logger {
	var config zap.Config

	if logJson {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	}

	// default configuration for all encoders
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config.Development = false
	config.DisableStacktrace = true
	config.DisableCaller = true

	if debug {
		config.Development = true
		config.DisableStacktrace = false
		config.DisableCaller = false
	}

	if verbose || debug {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	log, err := config.Build()
	if err != nil {
		panic(err)
	}

	return log
}
