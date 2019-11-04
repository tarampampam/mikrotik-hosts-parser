package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type (
	Server struct {
		srv               *http.Server
		fileSrv           http.Handler
		router            *mux.Router
		stdLog            *log.Logger
		errLog            *log.Logger
		startTime         time.Time
		publicDir         string
		originStdLogFlags int
	}

	ServerFileSystem struct {
		fs http.FileSystem
	}
)

// Opens file.
func (fs ServerFileSystem) Open(path string) (http.File, error) {
	f, err := fs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	fmt.Println(path)

	if s, _ := f.Stat(); s != nil && s.IsDir() { // @todo: Handle f.Stat() error
		index := strings.TrimSuffix(path, "/") + "/index.html"
		if _, err := fs.fs.Open(index); err != nil {
			return nil, err
		}
	}

	return f, nil
}

// Server constructor.
func NewServer(host string, port int, publicDir string, stdLog, errLog *log.Logger) *Server {
	var router = *mux.NewRouter()

	return &Server{
		srv: &http.Server{
			Addr:         host + ":" + strconv.Itoa(port), // TCP address and port to listen on
			Handler:      &router,
			ErrorLog:     errLog,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		},
		router:    &router,
		publicDir: publicDir,
		stdLog:    stdLog,
	}
}

// Register server http handlers.
func (s *Server) RegisterHandlers() {
	//s.router.Handle("/", http.FileServer(ServerFileSystem{fs: http.Dir(s.publicDir)})) // Handle static content

	s.router.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(s.publicDir))))

	//s.router.NotFoundHandler = s.notFoundHandler()
	//s.router.MethodNotAllowedHandler = s.methodNotAllowedHandler()
}

// Start proxy srv.
func (s *Server) Start() error {
	s.startTime = time.Now()
	s.originStdLogFlags = s.stdLog.Flags()
	s.stdLog.SetFlags(log.Ldate | log.Lmicroseconds)
	s.srv.SetKeepAlivesEnabled(false) // Disable keep alive
	s.stdLog.Println("Starting srv on " + s.srv.Addr)
	return s.srv.ListenAndServe()
}

// Stop proxy srv.
func (s *Server) Stop() error {
	s.stdLog.Println("Stopping srv")
	s.stdLog.SetFlags(s.originStdLogFlags)
	return s.srv.Shutdown(context.Background())
}
