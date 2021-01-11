// Package serve contains CLI `serve` command implementation.
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

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/breaker"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/config"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/env"
	appHttp "github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http"
	"go.uber.org/zap"
)

type flags struct { //nolint:maligned
	listen       string // TODO use just string for "ip:port"?
	port         uint16
	resourcesDir string // can be empty
	configPath   string
	redis        struct { //nolint:maligned
		host     string // TODO use redis DSN?
		port     uint16
		password string // can be empty
		dbNum    uint16
		maxConn  uint16
	}
}

func (f *flags) overrideUsingEnv() error {
	if envVar, exists := env.ListenAddr.Lookup(); exists {
		f.listen = envVar
	}

	if envVar, exists := env.ListenPort.Lookup(); exists {
		if p, err := strconv.ParseUint(envVar, 10, 16); err == nil {
			f.port = uint16(p)
		} else {
			return fmt.Errorf("wrong TCP port environment variable [%s] value", envVar)
		}
	}

	if envVar, exists := env.ResourcesDir.Lookup(); exists {
		f.resourcesDir = envVar
	}

	if envVar, exists := env.ConfigPath.Lookup(); exists {
		f.configPath = envVar
	}

	if envVar, exists := env.RedisHost.Lookup(); exists {
		f.redis.host = envVar
	}

	if envVar, exists := env.RedisPort.Lookup(); exists {
		if p, err := strconv.ParseUint(envVar, 10, 16); err == nil {
			f.redis.port = uint16(p)
		} else {
			return fmt.Errorf("wrong redis server TCP port environment variable [%s] value", envVar)
		}
	}

	if envVar, exists := env.RedisPassword.Lookup(); exists {
		f.redis.password = envVar
	}

	if envVar, exists := env.RedisDBNum.Lookup(); exists {
		if p, err := strconv.ParseUint(envVar, 10, 16); err == nil {
			f.redis.dbNum = uint16(p)
		} else {
			return fmt.Errorf("wrong redis DB number environment variable [%s] value", envVar)
		}
	}

	if envVar, exists := env.RedisMaxConn.Lookup(); exists {
		if p, err := strconv.ParseUint(envVar, 10, 16); err == nil {
			f.redis.maxConn = uint16(p)
		} else {
			return fmt.Errorf("wrong maximum redis connections environment variable [%s] value", envVar)
		}
	}

	return nil
}

func (f *flags) validate() error {
	if net.ParseIP(f.listen) == nil {
		return fmt.Errorf("wrong IP address [%s] for listening", f.listen)
	}

	if f.resourcesDir != "" {
		if info, err := os.Stat(f.resourcesDir); err != nil || !info.Mode().IsDir() {
			return fmt.Errorf("wrong resources directory [%s] path", f.resourcesDir)
		}
	}

	if info, err := os.Stat(f.configPath); err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("config file [%s] was not found", f.configPath)
	}

	if f.redis.host == "" {
		return errors.New("empty redis server hostname or IP address")
	}

	return nil
}

// NewCommand creates `serve` command.
func NewCommand(ctx context.Context, log *zap.Logger) *cobra.Command {
	var f flags

	cmd := &cobra.Command{
		Use:     "serve",
		Aliases: []string{"s", "server"},
		Short:   "Start HTTP server.\n\nEnvironment variables have higher priority then flags.",
		PreRunE: func(*cobra.Command, []string) error {
			if err := f.overrideUsingEnv(); err != nil {
				return err
			}

			return f.validate()
		},
		RunE: func(*cobra.Command, []string) error {
			return run(ctx, log, &f)
		},
	}

	wd, _ := os.Getwd()

	cmd.Flags().StringVarP(&f.listen, "listen", "l", "0.0.0.0", fmt.Sprintf("IP address to listen on [$%s]", env.ListenAddr)) //nolint:lll
	cmd.Flags().Uint16VarP(&f.port, "port", "p", 8080, fmt.Sprintf("TCP port number [$%s]", env.ListenPort))                  //nolint:lll

	cmd.Flags().StringVarP(&f.resourcesDir, "resources-dir", "r", filepath.Join(wd, "web"), fmt.Sprintf("path to the directory with public assets [$%s]", env.ResourcesDir)) //nolint:lll
	cmd.Flags().StringVarP(&f.configPath, "config", "c", filepath.Join(wd, "configs", "config.yml"), fmt.Sprintf("config file path [$%s]", env.ConfigPath))                  //nolint:lll

	cmd.Flags().StringVarP(&f.redis.host, "redis-host", "", "", fmt.Sprintf("redis server hostname or IP address [$%s]", env.RedisHost))          //nolint:lll
	cmd.Flags().Uint16VarP(&f.redis.port, "redis-port", "", 6379, fmt.Sprintf("redis server TCP port number [$%s]", env.RedisPort))               //nolint:lll
	cmd.Flags().StringVarP(&f.redis.password, "redis-password", "", "", fmt.Sprintf("redis server password (optional) [$%s]", env.RedisPassword)) //nolint:lll
	cmd.Flags().Uint16VarP(&f.redis.dbNum, "redis-db-num", "", 0, fmt.Sprintf("redis database number [$%s]", env.RedisDBNum))                     //nolint:lll
	cmd.Flags().Uint16VarP(&f.redis.maxConn, "redis-max-conn", "", 10, fmt.Sprintf("maximum redis connections [$%s]", env.RedisMaxConn))          //nolint:lll

	return cmd
}

const serverShutdownTimeout = 5 * time.Second

// run current command.
func run(parentCtx context.Context, log *zap.Logger, f *flags) error { //nolint:funlen
	cfg, cfgErr := config.FromYamlFile(f.configPath, true)
	if cfgErr != nil {
		return cfgErr
	}

	var (
		ctx, cancel = context.WithCancel(parentCtx) // serve context creation
		oss         = breaker.NewOSSignals(ctx)     // OS signals listener
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

	// establish connection to the redis server
	redisConn := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", f.redis.host, f.redis.port),
		Username: "",
		Password: f.redis.password,
		DB:       int(f.redis.dbNum),
		PoolSize: int(f.redis.maxConn),
	})
	if err := redisConn.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("cannot establish connection to the redis server: %w", err)
	}

	defer func() { _ = redisConn.Close() }() // for dead connections preventing

	// create HTTP server // TODO pass redis connection for the server
	server := appHttp.NewServer(ctx, log, fmt.Sprintf("%s:%d", f.listen, f.port), f.resourcesDir, cfg)

	// register server routes, middlewares, etc.
	if err := server.Register(); err != nil {
		return err
	}

	startingErrCh := make(chan error, 1) // channel for server starting error

	// start HTTP server in separate goroutine
	go func(errCh chan<- error) {
		defer close(errCh)

		log.Info("Server starting",
			zap.String("addr", f.listen),
			zap.Uint16("port", f.port),
			zap.String("resources", f.resourcesDir),
			zap.String("config file", f.configPath),
			zap.String("redis host", f.redis.host),
			zap.Uint16("redis port", f.redis.port),
			zap.Uint16("redis db", f.redis.dbNum),
		)

		if f.resourcesDir == "" {
			log.Warn("Resources directory was not provided")
		}

		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
		ctxShutdown, ctxCancelShutdown := context.WithTimeout(context.Background(), serverShutdownTimeout)
		defer ctxCancelShutdown()

		// and stop the server using created context above
		if err := server.Stop(ctxShutdown); err != nil {
			return err
		}

		// do not forget to close connection to the redis server
		if err := redisConn.Close(); err != nil {
			return err
		}
	}

	return nil
}
