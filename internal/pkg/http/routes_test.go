package http

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/config"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/api"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/script"
)

func TestServer_RegisterHandlers(t *testing.T) {
	compareHandlers := func(h1, h2 interface{}) bool {
		t.Helper()
		return reflect.ValueOf(h1).Pointer() == reflect.ValueOf(h2).Pointer()
	}

	var s = NewServer(&ServerSettings{}, &config.Config{})

	var cases = []struct {
		name    string
		route   string
		methods []string
		handler func(http.ResponseWriter, *http.Request)
	}{
		{
			name:    "script_generator",
			route:   "/script/source",
			methods: []string{"GET"},
			handler: script.RouterOsScriptSourceGenerationHandlerFunc(s.ServeSettings),
		},
		{
			name:    "api_get_settings",
			route:   "/api/settings",
			methods: []string{"GET"},
			handler: api.GetSettingsHandlerFunc(s.ServeSettings),
		},
		{
			name:    "api_get_version",
			route:   "/api/version",
			methods: []string{"GET"},
			handler: api.GetVersionHandler,
		},
		{
			name:    "api_get_routes",
			route:   "/api/routes",
			methods: []string{"GET"},
			handler: api.GetRoutesHandlerFunc(s.Router),
		},
	}

	for _, testCase := range cases {
		if s.Router.Get(testCase.name) != nil {
			t.Errorf("Handler for route [%s] must be not registered before RegisterHandlers() calling", testCase.name)
		}
	}

	s.RegisterHandlers()

	for _, testCase := range cases {
		if route, _ := s.Router.Get(testCase.name).GetPathTemplate(); route != testCase.route {
			t.Errorf("wrong route for [%s] route: want %v, got %v", testCase.name, testCase.route, route)
		}
		if methods, _ := s.Router.Get(testCase.name).GetMethods(); !reflect.DeepEqual(methods, testCase.methods) {
			t.Errorf("wrong method(s) for [%s] route: want %v, got %v", testCase.name, testCase.methods, methods)
		}
		if !compareHandlers(testCase.handler, s.Router.Get(testCase.name).GetHandler()) {
			t.Errorf("wrong handler for [%s] route", testCase.name)
		}
	}

	// Test static files handler registration
	staticRoute := s.Router.Get("static")

	if prefix, _ := staticRoute.GetPathTemplate(); prefix != "/" {
		t.Errorf("Wrong prefix for static files handler. Got: %s", prefix)
	}
}
