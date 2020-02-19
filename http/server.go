package http

import (
	"context"
	"encoding/json"
	"log"
	"mikrotik-hosts-parser/http/fileserver"
	"mime"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type (
	ServerSettings struct {
		Host             string
		Port             int
		PublicDir        string
		IndexFile        string
		Error404File     string
		WriteTimeout     time.Duration
		ReadTimeout      time.Duration
		KeepAliveEnabled bool
	}

	Server struct {
		Settings  *ServerSettings
		Server    *http.Server
		Router    *mux.Router
		stdLog    *log.Logger
		errLog    *log.Logger
		startTime time.Time
	}
)

// Server constructor.
func NewServer(settings *ServerSettings) *Server {
	var (
		router     = *mux.NewRouter()
		stdLog     = log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds)
		errLog     = log.New(os.Stderr, "[error] ", log.LstdFlags)
		httpServer = &http.Server{
			Addr:         settings.Host + ":" + strconv.Itoa(settings.Port), // TCP address and port to listen on
			Handler:      &router,
			ErrorLog:     errLog,
			WriteTimeout: settings.WriteTimeout,
			ReadTimeout:  settings.ReadTimeout,
		}
	)

	httpServer.SetKeepAlivesEnabled(settings.KeepAliveEnabled)

	return &Server{
		Settings: settings,
		Server:   httpServer,
		Router:   &router,
		stdLog:   stdLog,
		errLog:   errLog,
	}
}

// RegisterHandlers register server http handlers.
func (s *Server) RegisterHandlers() {
	s.Router.HandleFunc("/script/source", s.scriptSourceHandler).
		Methods("GET").
		Name("script_source")

	s.Router.PathPrefix("/").
		Handler(&fileserver.FileServer{
			Root:            http.Dir(s.Settings.PublicDir),
			IndexFile:       "index.html",
			Error404file:    "404.html",
		}).
		Name("static")
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

// Metrics request handler.
func (s *Server) scriptSourceHandler(w http.ResponseWriter, _ *http.Request) {
	res := make(map[string]interface{})
	// Append version
	res["version"] = "UNSET"

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(res)
}
