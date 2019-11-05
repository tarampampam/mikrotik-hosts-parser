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
			resources:                Resources,
			NotFoundHandler:          s.fileNotFoundHandler,
			DirectoryListingDisabled: s.accessDeniedHandler,
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

// HTML errors rendering function.
func (s *HttpServer) renderHtmlErrorPage(w http.ResponseWriter, code int, fName, failBack string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)

	var errFilePath = filepath.Join(s.Settings.PublicDir, fName)
	if _, statErr := os.Stat(errFilePath); statErr == nil {
		// file exists
		if f, err := os.Open(errFilePath); f != nil {
			// file opened
			defer f.Close()
			if err == nil {
				if _, writeErr := io.Copy(w, f); writeErr != nil {
					panic(writeErr)
				}
			}
		}
	} else {
		// fail-back to passed error page content
		if _, err := w.Write([]byte(failBack)); err != nil {
			panic(err)
		}
	}
}

// Error handler - 404
func (s *HttpServer) fileNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	const errFileName = "404.html"

	s.stdLog.Printf("Request from '%s' failed: '%s' was not found\n", r.RemoteAddr, r.RequestURI)

	s.renderHtmlErrorPage(w, http.StatusNotFound, errFileName, `<!DOCTYPE html>
<html lang="en">
<head>
	<title>Error 404</title>
	<style>html,body{background-color:#1a1a1a;color:#fff;font-family:sans-serif;}h1,h2,h3{text-align:center}</style>
</head>
<body>
	<h1>Error 404</h1>
	<h2>page not found</h2>
</body>
</html>`)
}

// Error handler - access denied
func (s *HttpServer) accessDeniedHandler(w http.ResponseWriter, r *http.Request) {
	const errFileName = "403.html"

	s.stdLog.Printf("Request from '%s' failed: '%s' access denied\n", r.RemoteAddr, r.RequestURI)

	s.renderHtmlErrorPage(w, http.StatusForbidden, errFileName, `<!DOCTYPE html>
<html lang="en">
<head>
	<title>Error 403</title>
	<style>html,body{background-color:#1a1a1a;color:#fff;font-family:sans-serif;}h1,h2,h3{text-align:center}</style>
</head>
<body>
	<h1>Error 403</h1>
	<h2>access denied</h2>
</body>
</html>`)
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
