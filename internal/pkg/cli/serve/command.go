// Package serve contains CLI `serve` command implementation.
package serve

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/breaker"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/cache"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/config"
	appHttp "github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http"
	"go.uber.org/zap"
)

const cachingEngineMemory, cachingEngineRedis = "memory", "redis"

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
			cfg, err := config.FromYamlFile(f.configPath, true)
			if err != nil {
				return err
			}

			return run(ctx, log, cfg, &f)
		},
	}

	f.init(cmd.Flags())

	return cmd
}

const serverShutdownTimeout = 5 * time.Second

// run current command.
func run(parentCtx context.Context, log *zap.Logger, cfg *config.Config, f *flags) error { //nolint:funlen
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

	var cachingEngine cache.Engine

	switch f.cachingEngine {
	case cachingEngineMemory:
		cachingEngine = cache.NewInMemoryEngine(time.Second*time.Duration(cfg.Cache.LifetimeSec), time.Second)

	case cachingEngineRedis:
		opt, err := redis.ParseURL(f.redisDSN)
		if err != nil {
			return err
		}

		cachingEngine = cache.NewRedisEngine(ctx, opt, time.Second*time.Duration(cfg.Cache.LifetimeSec))

	default:
		return errors.New("unsupported caching engine")
	}

	// try to open caching engine
	if err := cachingEngine.Open(); err != nil {
		return err
	}

	defer func() { _ = cachingEngine.Close() }()

	// create HTTP server
	server := appHttp.NewServer(
		ctx,
		log,
		cachingEngine,
		fmt.Sprintf("%s:%d", f.listen.ip, f.listen.port),
		f.resourcesDir,
		cfg,
	)

	// register server routes, middlewares, etc.
	if err := server.Register(); err != nil {
		return err
	}

	startingErrCh := make(chan error, 1) // channel for server starting error

	// start HTTP server in separate goroutine
	go func(errCh chan<- error) {
		defer close(errCh)

		log.Info("Server starting",
			zap.String("addr", f.listen.ip),
			zap.Uint16("port", f.listen.port),
			zap.String("resources", f.resourcesDir),
			zap.String("config file", f.configPath),
			zap.String("caching engine", f.cachingEngine),
			zap.String("redis dsn", f.redisDSN),
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

		if err := cachingEngine.Close(); err != nil {
			return err
		}
	}

	return nil
}
