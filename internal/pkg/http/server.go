package http

import (
	"context"
	"mime"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/config"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/middlewares/logreq"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/middlewares/panic"
	"go.uber.org/zap"
)

type (
	Server struct {
		ctx          context.Context
		log          *zap.Logger
		resourcesDir string // can be empty
		cfg          *config.Config
		srv          *http.Server
		router       *mux.Router
	}
)

// NewServer creates new server instance.
func NewServer(ctx context.Context, log *zap.Logger, listen, resourcesDir string, cfg *config.Config) Server {
	var (
		router     = *mux.NewRouter()
		httpServer = &http.Server{
			Addr:         listen,
			Handler:      &router,
			ErrorLog:     zap.NewStdLog(log),
			WriteTimeout: time.Second * 15,
			ReadTimeout:  time.Second * 15,
		}
	)

	return Server{
		ctx:          ctx,
		log:          log,
		resourcesDir: resourcesDir,
		cfg:          cfg,
		srv:          httpServer,
		router:       &router,
	}
}

// Start server.
func (s *Server) Start() error { return s.srv.ListenAndServe() }

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
	s.registerScriptGeneratorHandlers()
	s.registerAPIHandlers()

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
