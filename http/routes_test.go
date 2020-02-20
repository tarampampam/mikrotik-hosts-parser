package http

import (
	"mikrotik-hosts-parser/http/api"
	"mikrotik-hosts-parser/http/fileserver"
	"mikrotik-hosts-parser/http/script"
	"mikrotik-hosts-parser/settings/serve"
	"net/http"
	"reflect"
	"testing"
)

func TestServer_RegisterHandlers(t *testing.T) {
	t.Parallel()

	compareHandlers := func(h1, h2 interface{}) bool {
		t.Helper()
		return reflect.ValueOf(h1).Pointer() == reflect.ValueOf(h2).Pointer()
	}

	// @link: <https://stackoverflow.com/a/35791105/2252921>
	getType := func(myvar interface{}) string {
		t.Helper()

		if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
			return "*" + t.Elem().Name()
		}

		return t.Name()
	}

	var s = NewServer(&ServerSettings{}, &serve.Settings{})

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

	staticHandler := staticRoute.GetHandler()

	if handlerType := getType(staticHandler); handlerType != "*FileServer" {
		t.Errorf("Wrong handler (%s) for static route", handlerType)
	}

	if staticHandler.(*fileserver.FileServer).Settings.Root != http.Dir(s.ServeSettings.Resources.DirPath) {
		t.Error("Wrong resources root path is set for file server")
	}

	if staticHandler.(*fileserver.FileServer).Settings.IndexFile != s.ServeSettings.Resources.IndexName {
		t.Error("Wrong resources index file name is set for file server")
	}

	if staticHandler.(*fileserver.FileServer).Settings.Error404file != s.ServeSettings.Resources.Error404Name {
		t.Error("Wrong resources 404 error file name is set for file server")
	}
}
