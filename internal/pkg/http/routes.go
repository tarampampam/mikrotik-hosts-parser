package http

import (
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/api"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/fileserver"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/script"
)

// RegisterHandlers register server http handlers.
func (s *Server) RegisterHandlers() {
	s.registerStaticHandlers()
	s.registerAPIHandlers()
	s.registerFileServerHandler()
}

// Register static route handlers.
func (s *Server) registerStaticHandlers() {
	s.Router.
		HandleFunc("/script/source", script.RouterOsScriptSourceGenerationHandlerFunc(s.ServeSettings)).
		Methods("GET").
		Name("script_generator")
}

// Register API handlers.
func (s *Server) registerAPIHandlers() {
	apiRouter := s.Router.
		PathPrefix("/api").
		Subrouter()

	apiRouter.Use(DisableAPICachingMiddleware)

	apiRouter.
		HandleFunc("/settings", api.GetSettingsHandlerFunc(s.ServeSettings)).
		Methods("GET").
		Name("api_get_settings")

	apiRouter.
		HandleFunc("/version", api.GetVersionHandler).
		Methods("GET").
		Name("api_get_version")

	apiRouter.
		HandleFunc("/routes", api.GetRoutesHandlerFunc(s.Router)).
		Methods("GET").
		Name("api_get_routes")
}

// Register file server handler.
func (s *Server) registerFileServerHandler() {
	fs, _ := fileserver.NewFileServer(fileserver.Settings{ // FIXME handle an error
		FilesRoot:     s.ServeSettings.Resources.DirPath,
		IndexFileName: s.ServeSettings.Resources.IndexName,
		ErrorFileName: s.ServeSettings.Resources.Error404Name,
	})

	s.Router.
		PathPrefix("/").
		Handler(fs).
		Name("static")
}
