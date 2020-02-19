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
		HandleFunc("/script/source", script.RouterOsScriptSourceGenerationHandler).
		Methods("GET").
		Name("script_generator")
}

// Register API handlers.
func (s *Server) registerAPIHandlers() {
	apiRouter := s.Router.
		PathPrefix("/api").
		Subrouter()

	apiRouter.Use(disableAPICachingMiddleware)

	apiRouter.
		HandleFunc("/settings", func(w http.ResponseWriter, r *http.Request) {
			api.GetSettingsHandler(s.ServeSettings, w, r) // additionally passes serving settings into handler
		}).
		Methods("GET").
		Name("api_get_settings")

	apiRouter.
		HandleFunc("/version", api.GetVersion).
		Methods("GET").
		Name("api_get_version")

	apiRouter.
		HandleFunc("/routes", func(w http.ResponseWriter, r *http.Request) {
			api.GetRoutes(s.Router, w, r) // additionally passes router into handler
		}).
		Methods("GET").
		Name("api_get_routes")
}

// Register file server handler.
func (s *Server) registerFileServerHandler() {
	s.Router.
		PathPrefix("/").
		Handler(&fileserver.FileServer{
			Root:         http.Dir(s.ServeSettings.Resources.DirPath),
			IndexFile:    "index.html",
			Error404file: "404.html",
		}).
		Name("static")
}
