package http

import (
	"context"
	"mime"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/config"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type (
	ServerSettings struct {
		WriteTimeout     time.Duration
		ReadTimeout      time.Duration
		KeepAliveEnabled bool
	}

	Server struct {
		Settings      *ServerSettings
		ServeSettings *config.Config
		Server        *http.Server
		Router        *mux.Router
		startTime     time.Time
	}
)

// NewServer creates new server instance.
func NewServer(settings *ServerSettings, serveSettings *config.Config) *Server {
	var (
		router     = *mux.NewRouter()
		httpServer = &http.Server{
			Addr:    serveSettings.Listen.Address + ":" + strconv.Itoa(int(serveSettings.Listen.Port)),
			Handler: handlers.LoggingHandler(os.Stdout, &router),
			//ErrorLog:     errLog, // TODO zap.NewStdLog
			WriteTimeout: settings.WriteTimeout,
			ReadTimeout:  settings.ReadTimeout,
		}
	)

	httpServer.SetKeepAlivesEnabled(settings.KeepAliveEnabled)

	return &Server{
		Settings:      settings,
		ServeSettings: serveSettings,
		Server:        httpServer,
		Router:        &router,
	}
}

// Start proxy Server.
func (s *Server) Start() error {
	s.startTime = time.Now()
	if err := s.registerCustomMimeTypes(); err != nil {
		panic(err)
	}

	return s.Server.ListenAndServe()
}

// Register custom mime types.
func (*Server) registerCustomMimeTypes() error {
	return mime.AddExtensionType(".vue", "text/html; charset=utf-8")
}

// Stop the Server.
func (s *Server) Stop(ctx context.Context) error { return s.Server.Shutdown(ctx) }
