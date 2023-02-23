package http

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/checkers"
	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/http/fileserver"
	apiSettings "gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/http/handlers/api/settings"
	apiVersion "gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/http/handlers/api/version"
	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/http/handlers/generate"
	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/http/handlers/healthz"
	metricsHandler "gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/http/handlers/metrics"
	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/http/middlewares/nocache"
	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/metrics"
	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/version"
)

func (s *Server) registerScriptGeneratorHandlers(registerer prometheus.Registerer) error {
	m := metrics.NewGenerator()
	if err := m.Register(registerer); err != nil {
		return err
	}

	h, err := generate.NewHandler(s.ctx, s.log, s.cacher, s.cfg, &m)
	if err != nil {
		return err
	}

	s.router.
		Handle("/script/source", h).
		Methods(http.MethodGet).
		Name("script_generator")

	return nil
}

func (s *Server) registerAPIHandlers() {
	apiRouter := s.router.
		PathPrefix("/api").
		Subrouter()

	apiRouter.Use(nocache.New())

	apiRouter.
		HandleFunc("/settings", apiSettings.NewHandler(*s.cfg, s.cacher)).
		Methods(http.MethodGet).
		Name("api_get_settings")

	apiRouter.
		HandleFunc("/version", apiVersion.NewHandler(version.Version())).
		Methods(http.MethodGet).
		Name("api_get_version")
}

func (s *Server) registerServiceHandlers(registry prometheus.Gatherer) {
	s.router.
		HandleFunc("/metrics", metricsHandler.NewHandler(registry)).
		Methods(http.MethodGet).
		Name("metrics")

	s.router.
		HandleFunc("/ready", healthz.NewHandler(checkers.NewReadyChecker(s.ctx, s.rdb))).
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
