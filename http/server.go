package http

import (
	"context"
	"log"
	"mikrotik-hosts-parser/settings/serve"
	"mime"
	"net/http"
	"os"
	"strconv"
	"time"

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
		ServeSettings *serve.Settings
		Server        *http.Server
		Router        *mux.Router
		stdLog        *log.Logger
		errLog        *log.Logger
		startTime     time.Time
	}
)

// NewServer creates new server instance.
func NewServer(settings *ServerSettings, serveSettings *serve.Settings) *Server {
	var (
		router     = *mux.NewRouter()
		stdLog     = log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds)
		errLog     = log.New(os.Stderr, "[error] ", log.LstdFlags)
		httpServer = &http.Server{
			Addr:         serveSettings.Listen.Address + ":" + strconv.Itoa(serveSettings.Listen.Port),
			Handler:      &router,
			ErrorLog:     errLog,
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
		stdLog:        stdLog,
		errLog:        errLog,
	}
}

// Start proxy Server.
func (s *Server) Start() error {
	s.startTime = time.Now()
	if err := s.registerCustomMimeTypes(); err != nil {
		panic(err)
	}
	s.stdLog.Println("Starting Server on " + s.Server.Addr)

	return s.Server.ListenAndServe()
}

// Register custom mime types.
func (*Server) registerCustomMimeTypes() error {
	return mime.AddExtensionType(".vue", "text/html; charset=utf-8")
}

// Stop proxy Server.
func (s *Server) Stop() error {
	s.stdLog.Println("Stopping Server")

	return s.Server.Shutdown(context.Background())
}
