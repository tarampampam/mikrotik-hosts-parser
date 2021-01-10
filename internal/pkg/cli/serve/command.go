// Package version contains CLI `serve` command implementation.
package serve

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/breaker"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/config"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/env"
	appHttp "github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http"
	"go.uber.org/zap"
)

// NewCommand creates `serve` command.
func NewCommand(log *zap.Logger) *cobra.Command {
	var (
		listen       string
		port         uint16
		resourcesDir string // can be empty
		configPath   string
	)

	cmd := &cobra.Command{
		Use:     "serve",
		Aliases: []string{"s", "server"},
		Short:   "Start HTTP server.\n\nEnvironment variables have higher priority then flag values.",
		PreRunE: func(*cobra.Command, []string) error {
			// lookup for environment variables
			if envListen, exists := os.LookupEnv(env.ListenAddr); exists {
				listen = envListen
			}

			if envPort, exists := os.LookupEnv(env.ListenPort); exists {
				if p, err := strconv.ParseUint(envPort, 10, 16); err == nil {
					port = uint16(p)
				} else {
					return fmt.Errorf("wrong TCP port environment variable [%s] value (cannot be parsed)", envPort)
				}
			}

			if envResourcesDir, exists := os.LookupEnv(env.ResourcesDir); exists {
				resourcesDir = envResourcesDir
			}

			if envConfigPath, exists := os.LookupEnv(env.ConfigPath); exists {
				configPath = envConfigPath
			}

			// validate important values
			if net.ParseIP(listen) == nil {
				return fmt.Errorf("wrong IP address [%s] for listening", listen)
			}

			if resourcesDir != "" {
				if info, err := os.Stat(resourcesDir); err != nil || !info.Mode().IsDir() {
					return fmt.Errorf("wrong resources directory [%s] path", resourcesDir)
				}
			}

			if info, err := os.Stat(configPath); err != nil || !info.Mode().IsRegular() {
				return fmt.Errorf("config file [%s] was not found", configPath)
			}

			return nil
		},
		RunE: func(*cobra.Command, []string) error {
			cfg, err := config.FromYamlFile(configPath, true)
			if err != nil {
				return err
			}

			return run(log, listen, port, resourcesDir, cfg)
		},
	}

	wd, _ := os.Getwd()

	cmd.Flags().StringVarP(
		&listen,
		"listen",
		"l",
		"0.0.0.0",
		fmt.Sprintf("IP address to listen on [$%s]", env.ListenAddr),
	)
	cmd.Flags().Uint16VarP(
		&port,
		"port",
		"p",
		8080,
		fmt.Sprintf("TCP port number [$%s]", env.ListenPort),
	)
	cmd.Flags().StringVarP(
		&resourcesDir,
		"resources-dir",
		"r",
		filepath.Join(wd, "web"),
		fmt.Sprintf("path to the directory with public assets [$%s]", env.ResourcesDir),
	)
	cmd.Flags().StringVarP(
		&configPath,
		"config",
		"c",
		filepath.Join(wd, "configs", "config.yml"),
		fmt.Sprintf("config file path [$%s]", env.ConfigPath),
	)

	return cmd
}

// run current command.
func run(log *zap.Logger, listen string, port uint16, resourcesDir string, cfg *config.Config) error {
	var (
		ctx, cancel = context.WithCancel(context.Background()) // main context creation
		oss         = breaker.NewOSSignals(ctx)                // OS signals listener
	)

	// subscribe for system signals
	oss.Subscribe(func(sig os.Signal) {
		log.Warn("Stopping by OS signal..", zap.String("signal", sig.String()))

		cancel()
	})

	defer func() {
		cancel()   // call the cancellation function after all
		oss.Stop() // stop system signals listening
	}()

	// create HTTP server
	server := appHttp.NewServer(ctx, log, fmt.Sprintf("%s:%d", listen, port), resourcesDir, cfg)

	// register server routes, middlewares, etc.
	if err := server.Register(); err != nil {
		return err
	}

	startingErrCh := make(chan error, 1) // channel for server starting error

	// start HTTP server in separate goroutine
	go func(errCh chan<- error) {
		defer close(errCh)

		log.Info("Server starting",
			zap.String("addr", listen),
			zap.Uint16("port", port),
			zap.String("resources", resourcesDir),
		)

		if resourcesDir == "" {
			log.Warn("Resources directory was not provided")
		}

		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}(startingErrCh)

	// and wait for..
	select {
	case err := <-startingErrCh: // ..server starting error
		return err

	case <-ctx.Done(): // ..or context cancellation
		log.Debug("Server stopping")

		// create context for server graceful shutdown
		ctxShutdown, ctxCancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer ctxCancelShutdown()

		// and stop the server using created context above
		if err := server.Stop(ctxShutdown); err != nil {
			return err
		}
	}

	return nil
}
