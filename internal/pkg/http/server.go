// Package http contains HTTP server and all required stuff for HTTP server working.
package http

import (
	"context"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/cache"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/config"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/middlewares/logreq"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/middlewares/panic"
	"go.uber.org/zap"
)

type (
	// Server is HTTP server.
	Server struct {
		ctx          context.Context
		log          *zap.Logger
		cacher       cache.Cacher
		resourcesDir string // can be empty
		cfg          *config.Config
		srv          *http.Server
		router       *mux.Router
		rdb          *redis.Client // optional, can be nil
	}
)

const (
	defaultWriteTimeout = time.Second * 15
	defaultReadTimeout  = time.Second * 15
)

// NewServer creates new server instance.
func NewServer(
	ctx context.Context,
	log *zap.Logger,
	cacher cache.Cacher,
	resourcesDir string, // can be empty
	cfg *config.Config,
	rdb *redis.Client, // optional, can be nil
) Server {
	var (
		router     = mux.NewRouter()
		httpServer = &http.Server{
			Handler:      router,
			ErrorLog:     zap.NewStdLog(log),
			WriteTimeout: defaultWriteTimeout,
			ReadTimeout:  defaultReadTimeout,
		}
	)

	return Server{
		ctx:          ctx,
		log:          log,
		cacher:       cacher,
		resourcesDir: resourcesDir,
		cfg:          cfg,
		srv:          httpServer,
		router:       router,
		rdb:          rdb,
	}
}

// Start server.
func (s *Server) Start(ip string, port uint16) error {
	s.srv.Addr = ip + ":" + strconv.Itoa(int(port))

	return s.srv.ListenAndServe()
}

// Register server routes, middlewares, etc.
func (s *Server) Register() error {
	s.registerGlobalMiddlewares()

	if err := s.registerHandlers(); err != nil {
		return err
	}

	if err := s.registerCustomMimeTypes(); err != nil {
		return err
	}

	return nil
}

func (s *Server) registerGlobalMiddlewares() {
	s.router.Use(
		logreq.New(s.log),
		panic.New(s.log),
	)
}

// registerHandlers register server http handlers.
func (s *Server) registerHandlers() error {
	if err := s.registerScriptGeneratorHandlers(); err != nil {
		return err
	}

	s.registerAPIHandlers()
	s.registerServiceHandlers()

	if s.resourcesDir != "" {
		if err := s.registerFileServerHandler(s.resourcesDir); err != nil {
			return err
		}
	}

	return nil
}

// registerCustomMimeTypes registers custom mime types.
func (*Server) registerCustomMimeTypes() error {
	return mime.AddExtensionType(".vue", "text/html; charset=utf-8")
}

// Stop server.
func (s *Server) Stop(ctx context.Context) error { return s.srv.Shutdown(ctx) }
