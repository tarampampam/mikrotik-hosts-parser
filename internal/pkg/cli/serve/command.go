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
	appHttp "github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http"
	"go.uber.org/zap"
)

const (
	envNameListen       string = "LISTEN_ADDR"
	envNamePort         string = "LISTEN_PORT"
	envNameResourcesDir string = "RESOURCES_DIR"
	envNameConfigPath   string = "CONFIG_PATH"
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
		Short:   "Start HTTP server",
		PreRunE: func(*cobra.Command, []string) error {
			// lookup for environment variables
			if envListen, exists := os.LookupEnv(envNameListen); exists {
				listen = envListen
			}

			if envPort, exists := os.LookupEnv(envNamePort); exists {
				if p, err := strconv.ParseUint(envPort, 10, 16); err == nil {
					port = uint16(p)
				} else {
					return fmt.Errorf("wrong TCP port environment variable [%s] value (cannot be parsed)", envPort)
				}
			}

			if envResourcesDir, exists := os.LookupEnv(envNameResourcesDir); exists {
				resourcesDir = envResourcesDir
			}

			if envConfigPath, exists := os.LookupEnv(envNameConfigPath); exists {
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
		fmt.Sprintf("IP address to listen on [$%s]", envNameListen),
	)
	cmd.Flags().Uint16VarP(
		&port,
		"port",
		"p",
		8080,
		fmt.Sprintf("TCP port number [$%s]", envNamePort),
	)
	cmd.Flags().StringVarP(
		&resourcesDir,
		"resources-dir",
		"r",
		filepath.Join(wd, "web"),
		fmt.Sprintf("path to the directory with public assets [$%s]", envNameResourcesDir),
	)
	cmd.Flags().StringVarP(
		&configPath,
		"config",
		"c",
		filepath.Join(wd, "configs", "config.yml"),
		fmt.Sprintf("config file path [$%s]", envNameConfigPath),
	)

	return cmd
}

// run current command.
func run(log *zap.Logger, listen string, port uint16, resourcesDir string, cfg *config.Config) error {
	var (
		ctx, cancel = context.WithCancel(context.Background()) // main context creation
		oss         = breaker.NewOSSignals(ctx)                // OS signals listener
	)

	oss.Subscribe(func(sig os.Signal) {
		log.Warn("Stopping by OS signal..", zap.String("signal", sig.String()))

		cancel()
	})

	defer func() {
		cancel()   // call cancellation function after all (for context "leak" avoiding)
		oss.Stop() // stop system signals listening
	}()

	server := appHttp.NewServer(ctx, log, fmt.Sprintf("%s:%d", listen, port), resourcesDir, cfg)

	if err := server.Register(); err != nil {
		return err
	}

	var startingErrCh = make(chan error)

	go func(errCh chan<- error) {
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

	select {
	case <-ctx.Done(): // wait for context cancellation
	case err := <-startingErrCh: // or server starting error
		close(startingErrCh)
		log.Error(err.Error())
		cancel()
	}

	log.Debug("Server stopping")

	ctxShutdown, ctxCancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancelShutdown()

	if err := server.Stop(ctxShutdown); err != nil {
		return err
	} else {
		log.Info("Server stopped")
	}

	return nil
}
