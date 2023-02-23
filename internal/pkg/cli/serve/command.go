// Package serve contains CLI `serve` command implementation.
package serve

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/breaker"
	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/cache"
	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/config"
	appHttp "gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/http"
)

const cachingEngineMemory, cachingEngineRedis = "memory", "redis"

// NewCommand creates `serve` command.
func NewCommand(ctx context.Context, log *zap.Logger) *cobra.Command {
	var f flags

	cmd := &cobra.Command{
		Use:     "serve",
		Aliases: []string{"s", "server"},
		Short:   "Start HTTP server",
		Long:    "Environment variables have higher priority then flags",
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
func run(parentCtx context.Context, log *zap.Logger, cfg *config.Config, f *flags) error { //nolint:funlen,gocyclo
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

	var (
		cacheTTL time.Duration
		rdb      *redis.Client // optional, can be nil
		cacher   cache.Cacher
	)

	cacheTTL, _ = time.ParseDuration(f.cache.ttl)

	switch f.cache.engine {
	case cachingEngineMemory:
		inmemory := cache.NewInMemoryCache(cacheTTL, time.Second)

		defer func() { _ = inmemory.Close() }()

		cacher = inmemory

	case cachingEngineRedis:
		opt, _ := redis.ParseURL(f.redisDSN)
		rdb = redis.NewClient(opt).WithContext(ctx)

		defer func() { _ = rdb.Close() }()

		if pingErr := rdb.Ping(ctx).Err(); pingErr != nil {
			return pingErr
		}

		cacher = cache.NewRedisCache(ctx, rdb, cacheTTL)

	default:
		return errors.New("unsupported caching engine")
	}

	// create HTTP server
	server := appHttp.NewServer(ctx, log, cacher, f.resourcesDir, cfg, rdb)

	// register server routes, middlewares, etc.
	if err := server.Register(); err != nil {
		return err
	}

	startingErrCh := make(chan error, 1) // channel for server starting error

	// start HTTP server in separate goroutine
	go func(errCh chan<- error) {
		defer close(errCh)

		fields := []zap.Field{
			zap.String("addr", f.listen.ip),
			zap.Uint16("port", f.listen.port),
			zap.String("resources", f.resourcesDir),
			zap.String("config file", f.configPath),
			zap.String("caching engine", f.cache.engine),
			zap.Duration("cache ttl", cacheTTL),
		}

		if f.cache.engine == cachingEngineRedis {
			fields = append(fields, zap.String("redis dsn", f.redisDSN))
		}

		log.Info("Server starting", fields...)

		if f.resourcesDir == "" {
			log.Warn("Resources directory was not provided")
		}

		if err := server.Start(f.listen.ip, f.listen.port); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

		// stop the server using created context above
		if err := server.Stop(ctxShutdown); err != nil {
			return err
		}

		// close cacher (if it is possible)
		if c, ok := cacher.(io.Closer); ok {
			if err := c.Close(); err != nil {
				return err
			}
		}

		// and close redis connection
		if rdb != nil {
			if err := rdb.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}
