package http

import (
	"mikrotik-hosts-parser/http/api"
	"mikrotik-hosts-parser/http/fileserver"
	"mikrotik-hosts-parser/http/script"
	"net/http"
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
	s.Router.
		PathPrefix("/").
		Handler(&fileserver.FileServer{Settings: fileserver.Settings{
			Root:         http.Dir(s.ServeSettings.Resources.DirPath),
			IndexFile:    s.ServeSettings.Resources.IndexName,
			Error404file: s.ServeSettings.Resources.Error404Name,
		}}).
		Name("static")
}
