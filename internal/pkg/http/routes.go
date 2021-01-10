package http

import (
	"net/http"

	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/checkers"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/handlers/healthz"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/version"

	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/fileserver"
	apiSettings "github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/handlers/api/settings"
	apiVersion "github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/handlers/api/version"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/middlewares/nocache"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/script"
)

func (s *Server) registerScriptGeneratorHandlers() {
	s.router.
		HandleFunc("/script/source", script.RouterOsScriptSourceGenerationHandlerFunc(s.cfg)).
		Methods(http.MethodGet).
		Name("script_generator")
}

func (s *Server) registerAPIHandlers() {
	apiRouter := s.router.
		PathPrefix("/api").
		Subrouter()

	apiRouter.Use(nocache.New())

	apiRouter.
		HandleFunc("/settings", apiSettings.NewHandler(*s.cfg)).
		Methods(http.MethodGet).
		Name("api_get_settings")

	apiRouter.
		HandleFunc("/version", apiVersion.NewHandler(version.Version())).
		Methods(http.MethodGet).
		Name("api_get_version")
}

func (s *Server) registerServiceHandlers() {
	s.router.
		HandleFunc("/ready", healthz.NewHandler(checkers.NewReadyChecker())).
		Methods(http.MethodGet, http.MethodHead).
		Name("ready")

	s.router.
		HandleFunc("/live", healthz.NewHandler(checkers.NewLiveChecker())).
		Methods(http.MethodGet, http.MethodHead).
		Name("live")
}

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
		Methods(http.MethodGet, http.MethodHead).
		Handler(fs).
		Name("static")

	return nil
}
