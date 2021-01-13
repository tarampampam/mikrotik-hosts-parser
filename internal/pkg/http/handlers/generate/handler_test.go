package generate

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/cache"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/config"
	"go.uber.org/zap"
)

type fakeHttpClientFunc func(*http.Request) (*http.Response, error)

func (f fakeHttpClientFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestHandler_ServeHTTP(t *testing.T) {
	cfg := &config.Config{}
	cfg.RouterScript.MaxSourcesCount = 10
	cfg.RouterScript.Comment = "foo"
	cfg.RouterScript.MaxSourceSizeBytes = 2097152

	var httpMock fakeHttpClientFunc = func(req *http.Request) (*http.Response, error) {
		raw, err := ioutil.ReadFile("../../../../../test/testdata/hosts/ad_servers.txt")
		if err != nil {
			panic(err)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}},
			Body:       ioutil.NopCloser(bytes.NewReader(raw)),
		}, nil
	}

	h, err := NewHandler(context.Background(), zap.NewNop(), cache.NewInMemoryCache(time.Minute, time.Second), cfg)
	assert.NoError(t, err)

	h.(*handler).httpClient = httpMock

	var (
		req, _ = http.NewRequest(http.MethodGet, "http://testing?"+
			"format=routeros&"+
			"version=v0.0.666@1a0339c&"+
			"redirect_to=127.0.0.5&"+
			"limit=1234&"+
			"sources_urls=https%3A%2F%2Ffoo.com%2Fbar.txt"+
			",http://bar.com/baz.asp"+
			",http://baz.com/blah.list"+
			"&excluded_hosts="+
			"d.com,c.org,"+
			"localhost,"+
			"localhost.localdomain,"+
			"broadcasthost,"+
			"local,"+
			"ip6-localhost,"+
			"ip6-loopback,"+
			"ip6-localnet,"+
			"ip6-mcastprefix,"+
			"ip6-allnodes,"+
			"ip6-allrouters,"+
			"ip6-allhosts", http.NoBody)
		rr = httptest.NewRecorder()
	)

	h.ServeHTTP(rr, req)

	body := rr.Body.String()

	assert.Contains(t, body, "/ip dns static")
	assert.Contains(t, body, "add address=127.0.0.5 comment=\"foo\" disabled=no")

	// TODO write more asserts

	t.Log(body)
}

func BenchmarkHandler_ServeHTTP(b *testing.B) {
	b.ReportAllocs()

	cfg := &config.Config{}
	cfg.RouterScript.MaxSourcesCount = 10
	cfg.RouterScript.Comment = "foo"
	cfg.RouterScript.MaxSourceSizeBytes = 2097152

	var httpMock fakeHttpClientFunc = func(req *http.Request) (*http.Response, error) {
		raw, err := ioutil.ReadFile("../../../../../test/testdata/hosts/ad_servers.txt")
		if err != nil {
			panic(err)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}},
			Body:       ioutil.NopCloser(bytes.NewReader(raw)),
		}, nil
	}

	cacher := cache.NewInMemoryCache(time.Minute, time.Second)
	defer cacher.Close()

	h, _ := NewHandler(context.Background(), zap.NewNop(), cacher, cfg)

	h.(*handler).httpClient = httpMock

	var (
		req, _ = http.NewRequest(http.MethodGet, "http://testing?"+
			"format=routeros&"+
			"version=v0.0.666@1a0339c&"+
			"redirect_to=127.0.0.5&"+
			"limit=1234&"+
			"sources_urls=https%3A%2F%2Ffoo.com%2Fbar.txt"+
			//",http://bar.com/baz.asp"+
			//",http://baz.com/blah.list"+
			"&excluded_hosts="+
			"d.com,c.org,"+
			"localhost,"+
			"localhost.localdomain,"+
			"broadcasthost,"+
			"local,"+
			"ip6-localhost,"+
			"ip6-loopback,"+
			"ip6-localnet,"+
			"ip6-mcastprefix,"+
			"ip6-allnodes,"+
			"ip6-allrouters,"+
			"ip6-allhosts", http.NoBody)
		rr = httptest.NewRecorder()
	)

	for n := 0; n < b.N; n++ {
		h.ServeHTTP(rr, req)
	}
}

func BenchmarkHandlerOld_ServeHTTP(b *testing.B) {
	b.ReportAllocs()

	cfg := &config.Config{}
	cfg.RouterScript.MaxSourcesCount = 10
	cfg.RouterScript.Comment = "foo"
	cfg.RouterScript.MaxSourceSizeBytes = 2097152

	var httpMock fakeHttpClientFunc = func(req *http.Request) (*http.Response, error) {
		raw, err := ioutil.ReadFile("../../../../../test/testdata/hosts/ad_servers.txt")
		if err != nil {
			panic(err)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}},
			Body:       ioutil.NopCloser(bytes.NewReader(raw)),
		}, nil
	}

	cacher := cache.NewInMemoryCache(time.Minute, time.Second)
	defer cacher.Close()

	h := RouterOsScriptSourceGenerationHandlerFunc(cfg, cacher)

	defaultHTTPClient.client = httpMock

	var (
		req, _ = http.NewRequest(http.MethodGet, "http://testing?"+
			"format=routeros&"+
			"version=v0.0.666@1a0339c&"+
			"redirect_to=127.0.0.5&"+
			"limit=1234&"+
			"sources_urls=https%3A%2F%2Ffoo.com%2Fbar.txt"+
			//",http://bar.com/baz.asp"+
			//",http://baz.com/blah.list"+
			"&excluded_hosts="+
			"d.com,c.org,"+
			"localhost,"+
			"localhost.localdomain,"+
			"broadcasthost,"+
			"local,"+
			"ip6-localhost,"+
			"ip6-loopback,"+
			"ip6-localnet,"+
			"ip6-mcastprefix,"+
			"ip6-allnodes,"+
			"ip6-allrouters,"+
			"ip6-allhosts", http.NoBody)
		rr = httptest.NewRecorder()
	)

	for n := 0; n < b.N; n++ {
		h(rr, req)
	}
}
