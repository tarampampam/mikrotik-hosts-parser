package http

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/config"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/api"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/http/script"
	"go.uber.org/zap"
)

func TestServer_RegisterHandlers(t *testing.T) {
	compareHandlers := func(h1, h2 interface{}) bool {
		t.Helper()
		return reflect.ValueOf(h1).Pointer() == reflect.ValueOf(h2).Pointer()
	}

	var s = NewServer(context.Background(), zap.NewNop(), "", ".", &config.Config{})

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
			handler: script.RouterOsScriptSourceGenerationHandlerFunc(s.cfg),
		},
		{
			name:    "api_get_settings",
			route:   "/api/settings",
			methods: []string{"GET"},
			handler: api.GetSettingsHandlerFunc(s.cfg),
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
			handler: api.GetRoutesHandlerFunc(s.router),
		},
	}

	for _, testCase := range cases {
		if s.router.Get(testCase.name) != nil {
			t.Errorf("Handler for route [%s] must be not registered before RegisterHandlers() calling", testCase.name)
		}
	}

	assert.NoError(t, s.RegisterHandlers())

	for _, testCase := range cases {
		if route, _ := s.router.Get(testCase.name).GetPathTemplate(); route != testCase.route {
			t.Errorf("wrong route for [%s] route: want %v, got %v", testCase.name, testCase.route, route)
		}
		if methods, _ := s.router.Get(testCase.name).GetMethods(); !reflect.DeepEqual(methods, testCase.methods) {
			t.Errorf("wrong method(s) for [%s] route: want %v, got %v", testCase.name, testCase.methods, methods)
		}
		if !compareHandlers(testCase.handler, s.router.Get(testCase.name).GetHandler()) {
			t.Errorf("wrong handler for [%s] route", testCase.name)
		}
	}

	// Test static files handler registration
	staticRoute := s.router.Get("static")

	if prefix, _ := staticRoute.GetPathTemplate(); prefix != "/" {
		t.Errorf("Wrong prefix for static files handler. Got: %s", prefix)
	}
}
