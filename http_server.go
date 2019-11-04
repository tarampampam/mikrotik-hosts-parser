package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type (
	HttpServerSettings struct {
		Host             string
		Port             int
		PublicDir        string
		IndexFile        string
		Error404File     string
		WriteTimeout     time.Duration
		ReadTimeout      time.Duration
		KeepAliveEnabled bool
	}

	HttpServer struct {
		Settings  *HttpServerSettings
		Server    *http.Server
		Router    *mux.Router
		stdLog    *log.Logger
		errLog    *log.Logger
		startTime time.Time
	}
)

// HttpServer constructor.
func NewServer(settings *HttpServerSettings) *HttpServer {
	var (
		router     = *mux.NewRouter()
		stdLog     = log.New(os.Stderr, "", log.Ldate|log.Lmicroseconds)
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

	return &HttpServer{
		Settings: settings,
		Server:   httpServer,
		Router:   &router,
		stdLog:   stdLog,
		errLog:   errLog,
	}
}

// Register server http handlers.
func (s *HttpServer) RegisterHandlers() {
	s.Router.HandleFunc("/script/source", s.scriptSourceHandler).
		Methods("GET").
		Name("script_source")

	s.Router.PathPrefix("/").
		Handler(&HttpFileServer{
			root:                     http.Dir(s.Settings.PublicDir),
			NotFoundHandler:          s.fileNotFoundHandler,
			DirectoryListingDisabled: s.directoryListingDisabledHandler,
		}).
		Name("static")
}

// Start proxy Server.
func (s *HttpServer) Start() error {
	s.startTime = time.Now()
	s.stdLog.Println("Starting Server on " + s.Server.Addr)
	return s.Server.ListenAndServe()
}

// Stop proxy Server.
func (s *HttpServer) Stop() error {
	s.stdLog.Println("Stopping Server")
	return s.Server.Shutdown(context.Background())
}

// Error handler - 404
func (s *HttpServer) fileNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	const errFileName = "404.html"

	s.stdLog.Printf("Request from '%s' failed: '%s' was not found\n", r.RemoteAddr, r.RequestURI)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)

	var errFilePath = filepath.Join(s.Settings.PublicDir, errFileName)
	if _, statErr := os.Stat(errFilePath); statErr == nil {
		if f, err := os.Open(errFilePath); f != nil {
			defer f.Close()
			if err == nil {
				_, _ = io.Copy(w, f)
			}
		}
	} else {
		// default error page content
		if _, err := w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
	<title>Error 404</title>
	<style>html,body{background-color:#1a1a1a;color:#fff;font-family:sans-serif;}h1,h2,h3{text-align:center}</style>
</head>
<body>
	<h1>Error 404</h1>
	<h2>page not found</h2>
</body>
</html>`)); err != nil {
			panic(err)
		}
	}
}

// Error handler - directory listing is disabled
func (s *HttpServer) directoryListingDisabledHandler(w http.ResponseWriter, r *http.Request) {
	const errFileName = "403_listing_disabled.html"

	s.stdLog.Printf("Request from '%s' failed: '%s' directory listing disabled\n", r.RemoteAddr, r.RequestURI)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)

	var errFilePath = filepath.Join(s.Settings.PublicDir, errFileName)
	if _, statErr := os.Stat(errFilePath); statErr == nil {
		if f, err := os.Open(errFilePath); f != nil {
			defer f.Close()
			if err == nil {
				_, _ = io.Copy(w, f)
			}
		}
	} else {
		// default error page content
		if _, err := w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
	<title>Error 403</title>
	<style>html,body{background-color:#1a1a1a;color:#fff;font-family:sans-serif;}h1,h2,h3{text-align:center}</style>
</head>
<body>
	<h1>Error 403</h1>
	<h2>directory listing disabled</h2>
</body>
</html>`)); err != nil {
			panic(err)
		}
	}
}

// Metrics request handler.
func (s *HttpServer) scriptSourceHandler(w http.ResponseWriter, _ *http.Request) {
	res := make(map[string]interface{})
	// Append version
	res["version"] = VERSION

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(res)
}
