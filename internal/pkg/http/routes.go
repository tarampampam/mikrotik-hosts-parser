package http

import (
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/api"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/fileserver"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/middlewares/nocache"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/script"
)

func (s *Server) registerScriptGeneratorHandlers() {
	s.router.
		HandleFunc("/script/source", script.RouterOsScriptSourceGenerationHandlerFunc(s.cfg)).
		Methods("GET").
		Name("script_generator")
}

// Register API handlers.
func (s *Server) registerAPIHandlers() {
	apiRouter := s.router.
		PathPrefix("/api").
		Subrouter()

	apiRouter.Use(nocache.Middleware)

	apiRouter.
		HandleFunc("/settings", api.GetSettingsHandlerFunc(s.cfg)).
		Methods("GET").
		Name("api_get_settings")

	apiRouter.
		HandleFunc("/version", api.GetVersionHandler).
		Methods("GET").
		Name("api_get_version")

	apiRouter.
		HandleFunc("/routes", api.GetRoutesHandlerFunc(s.router)).
		Methods("GET").
		Name("api_get_routes")
}

// Register file server handler.
func (s *Server) registerFileServerHandler(resourcesDir string) error {
	fs, err := fileserver.NewFileServer(fileserver.Settings{
		FilesRoot:               resourcesDir,
		IndexFileName:           "index.html",
		ErrorFileName:           "__error__.html",
		RedirectIndexFileToRoot: true,
	})
	if err != nil {
		return err
	}

	s.router.
		PathPrefix("/").
		Handler(fs).
		Name("static")

	return nil
}
