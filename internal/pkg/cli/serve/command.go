// Package version contains CLI `serve` command implementation.
package serve

import (
	"context"
	"errors"
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
		resourcesDir string
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
					return errors.New("wrong TCP port environment variable value (cannot be parsed)")
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

			if info, err := os.Stat(resourcesDir); err != nil || !info.Mode().IsDir() {
				return fmt.Errorf("wrong resources directory [%s] path", resourcesDir)
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

			return execute(log, listen, port, resourcesDir, cfg)
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
		"resources",
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

func execute(
	log *zap.Logger,
	listen string,
	port uint16,
	resourcesDir string,
	cfg *config.Config,
) error {
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

	server := appHttp.NewServer(&appHttp.ServerSettings{
		WriteTimeout:     time.Second * 15,
		ReadTimeout:      time.Second * 15,
		KeepAliveEnabled: false,
	}, cfg)

	server.RegisterHandlers()

	go func() {
		log.Info("HTTP server starting",
			zap.String("addr", listen),
			zap.Uint16("port", port),
			zap.String("resources", resourcesDir),
		)

		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server cannot be started", zap.Error(err))
		}
	}()

	<-ctx.Done()

	log.Debug("HTTP server stopping")

	ctxShutdown, ctxCancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() { ctxCancelShutdown() }()

	if err := server.Stop(ctxShutdown); err != nil {
		return err
	}

	log.Info("HTTP server stopped")

	if true {
		return errors.New("foo")
	}

	return nil
}
