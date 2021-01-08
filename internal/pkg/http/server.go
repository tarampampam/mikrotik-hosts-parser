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
		resourcesDir string
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
			Addr:    listen,
			Handler: &router,
			//ErrorLog:     errLog, // TODO zap.NewStdLog
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

func (s *Server) RegisterGlobalMiddlewares() {
	s.router.Use(
		logreq.NewMiddleware(s.log),
		panic.NewMiddleware(s.log),
	)
}

// RegisterHandlers register server http handlers.
func (s *Server) RegisterHandlers() error {
	s.registerScriptGeneratorHandlers()
	s.registerAPIHandlers()

	if err := s.registerFileServerHandler(); err != nil {
		return err
	}

	return nil
}

// RegisterCustomMimeTypes registers custom mime types.
func (*Server) RegisterCustomMimeTypes() error {
	return mime.AddExtensionType(".vue", "text/html; charset=utf-8")
}

// Stop server.
func (s *Server) Stop(ctx context.Context) error { return s.srv.Shutdown(ctx) }
