package http

import (
	"context"
	"errors"
	"mime"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/cache"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/config"
	"go.uber.org/zap"
)

func getRandomTCPPort(t *testing.T) (int, error) {
	t.Helper()

	// zero port means randomly (os) chosen port
	l, err := net.Listen("tcp", ":0") //nolint:gosec
	if err != nil {
		return 0, err
	}

	port := l.Addr().(*net.TCPAddr).Port

	if closingErr := l.Close(); closingErr != nil {
		return 0, closingErr
	}

	return port, nil
}

func checkTCPPortIsBusy(t *testing.T, port int) bool {
	t.Helper()

	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return true
	}

	_ = l.Close()

	return false
}

func TestServer_StartAndStop(t *testing.T) {
	port, err := getRandomTCPPort(t)
	assert.NoError(t, err)

	cacheConn := cache.NewInMemoryConnector(time.Second, time.Second)
	defer cacheConn.Close()

	srv := NewServer(context.Background(), zap.NewNop(), cacheConn, ":"+strconv.Itoa(port), ".", &config.Config{})

	assert.False(t, checkTCPPortIsBusy(t, port))

	go func() {
		startingErr := srv.Start()

		if !errors.Is(startingErr, http.ErrServerClosed) {
			assert.NoError(t, startingErr)
		}
	}()

	for i := 0; ; i++ {
		if !checkTCPPortIsBusy(t, port) {
			if i > 100 {
				t.Error("too many attempts for server start checking")
			}

			<-time.After(time.Microsecond * 10)
		} else {
			break
		}
	}

	assert.True(t, checkTCPPortIsBusy(t, port))
	assert.NoError(t, srv.Stop(context.Background()))
	assert.False(t, checkTCPPortIsBusy(t, port))
}

func TestServer_Register(t *testing.T) {
	var routes = []struct {
		name    string
		route   string
		methods []string
	}{
		{name: "script_generator", route: "/script/source", methods: []string{http.MethodGet}},
		{name: "api_get_settings", route: "/api/settings", methods: []string{http.MethodGet}},
		{name: "api_get_version", route: "/api/version", methods: []string{http.MethodGet}},
		{name: "ready", route: "/ready", methods: []string{http.MethodGet, http.MethodHead}},
		{name: "live", route: "/live", methods: []string{http.MethodGet, http.MethodHead}},
		{name: "static", route: "/", methods: []string{http.MethodGet, http.MethodHead}},
	}

	cacheConn := cache.NewInMemoryConnector(time.Second, time.Second)
	defer cacheConn.Close()

	srv := NewServer(context.Background(), zap.NewNop(), cacheConn, ":0", ".", &config.Config{})
	router := srv.router // dirty hack, yes, i know

	// state *before* registration
	types, err := mime.ExtensionsByType("text/html; charset=utf-8")
	assert.NoError(t, err)
	assert.NotContains(t, types, ".vue") // mime types registration can be executed only once

	for _, r := range routes {
		assert.Nil(t, router.Get(r.name))
	}

	// call register fn
	assert.NoError(t, srv.Register())

	// state *after* registration
	types, _ = mime.ExtensionsByType("text/html; charset=utf-8") // reload
	assert.Contains(t, types, ".vue")

	for _, r := range routes {
		route, _ := router.Get(r.name).GetPathTemplate()
		assert.Equal(t, r.route, route)
		methods, _ := router.Get(r.name).GetMethods()
		assert.Equal(t, r.methods, methods)
	}
}

func TestServer_RegisterWithoutResourcesDir(t *testing.T) {
	cacheEng := cache.NewInMemoryConnector(time.Second, time.Second)
	defer cacheEng.Close()

	srv := NewServer(context.Background(), zap.NewNop(), cacheEng, ":0", "", &config.Config{}) // empty resources dir
	router := srv.router                                                                       // dirty hack, yes, i know

	assert.Nil(t, router.Get("static"))
	assert.NoError(t, srv.Register())
	assert.Nil(t, router.Get("static"))
}
