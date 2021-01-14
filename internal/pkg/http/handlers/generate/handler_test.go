package generate

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/cache"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/config"
	"go.uber.org/zap"
)

type fakeHTTPClientFunc func(*http.Request) (*http.Response, error)

func (f fakeHTTPClientFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

var httpMock fakeHTTPClientFunc = func(req *http.Request) (*http.Response, error) { //nolint:gochecknoglobals
	path, absErr := filepath.Abs(testDataPath + req.URL.RequestURI())
	if absErr != nil {
		panic(absErr)
	}

	if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
		raw, readingErr := ioutil.ReadFile(path)
		if readingErr != nil {
			panic(readingErr)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type":   []string{"text/plain; charset=utf-8"},
				"Content-Length": []string{strconv.FormatInt(info.Size(), 10)},
			},
			Body: ioutil.NopCloser(bytes.NewReader(raw)),
		}, nil
	}

	return &http.Response{
		StatusCode: http.StatusNotFound,
		Header:     http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}},
		Body:       ioutil.NopCloser(bytes.NewReader([]byte("Requested file was not found: " + path))),
	}, nil
}

const testDataPath = "../../../../../test/testdata/hosts"

func createConfig() *config.Config {
	cfg := &config.Config{}
	cfg.RouterScript.MaxSourcesCount = 10
	cfg.RouterScript.Comment = "foo"
	cfg.RouterScript.MaxSourceSizeBytes = 2097152

	return cfg
}

func BenchmarkHandler_ServeHTTP(b *testing.B) {
	b.ReportAllocs()

	cacher := cache.NewInMemoryCache(time.Minute, time.Second)
	defer cacher.Close()

	h, _ := NewHandler(context.Background(), zap.NewNop(), cacher, createConfig())

	h.(*handler).httpClient = httpMock

	var (
		req, _ = http.NewRequest(http.MethodGet, "http://testing?"+
			"format=routeros"+ //nolint:misspell
			"&version=v0.0.666@1a0339c"+
			"&redirect_to=127.0.0.5"+
			"&limit=1234"+
			"&sources_urls="+
			"https%3A%2F%2Fmock%2Fad_servers.txt"+
			",http://mock/hosts_adaway.txt"+
			",http://non-existing-file.txt"+
			"&excluded_hosts="+
			"d.com"+
			",c.org"+
			",localhost"+
			",localhost.localdomain"+
			",broadcasthost"+
			",local", http.NoBody)
		rr = httptest.NewRecorder()
	)

	for n := 0; n < b.N; n++ {
		h.ServeHTTP(rr, req)
	}
}

func TestHandler_ServeHTTP(t *testing.T) {
	cacher := cache.NewInMemoryCache(time.Minute, time.Second)
	defer cacher.Close()

	h, err := NewHandler(context.Background(), zap.NewNop(), cacher, createConfig())
	assert.NoError(t, err)

	h.(*handler).httpClient = httpMock

	var (
		req, _ = http.NewRequest(http.MethodGet, "http://testing?"+
			"format=routeros"+ //nolint:misspell
			"&version=v0.0.666@1a0339c"+
			"&redirect_to=127.0.0.5"+
			"&limit=1234"+
			"&sources_urls="+
			"https%3A%2F%2Fmock%2Fad_servers.txt"+
			",http://mock/hosts_adaway.txt"+
			",http://non-existing-file.txt"+
			"&excluded_hosts="+
			"aaa.com"+
			",bbb.org"+
			",localhost", http.NoBody)
		rr = httptest.NewRecorder()
	)

	h.ServeHTTP(rr, req) // first run

	body := rr.Body.String()

	assert.Regexp(t, `Cache.+miss.+http:\/\/mock\/hosts_adaway\.txt`, body)
	assert.Regexp(t, `Cache.+miss.+https:\/\/mock\/ad_servers\.txt`, body)
	assert.Regexp(t, `Source.+non-existing-file\.txt.+404`, body)
	assert.Regexp(t, `(?sU)Excluded hosts.+aaa\.com.+bbb\.org.+localhost`, rr.Body.String())
	assert.Contains(t, body, "/ip dns static")
	assert.Equal(t, strings.Count(body, "add address=127.0.0.5 comment=\"foo\" disabled=no"), 1234)

	// assert non-comments and non-empty lines count
	var lineWithoutCommentsRegex = regexp.MustCompile(`(?mU)^([^#\n]+.*)\n`) // <https://regex101.com/r/O23cel/1>

	assert.Equal(t, 1234+1, len(lineWithoutCommentsRegex.FindAllStringIndex(body, -1))) //nolint:wsl

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req) // second run

	body = rr.Body.String()

	assert.Regexp(t, `Cache.+HIT.+http:\/\/mock\/hosts_adaway\.txt`, body)
	assert.Regexp(t, `Cache.+HIT.+https:\/\/mock\/ad_servers\.txt`, body)
	assert.Regexp(t, `Source.+non-existing-file\.txt.+404`, body)

	assert.Equal(t, 1234+1, len(lineWithoutCommentsRegex.FindAllStringIndex(body, -1)))
}

func TestHandler_ServeHTTPHostnamesExcluding(t *testing.T) {
	cacher := cache.NewInMemoryCache(time.Minute, time.Second)
	defer cacher.Close()

	h, err := NewHandler(context.Background(), zap.NewNop(), cacher, createConfig())
	assert.NoError(t, err)

	var customHTTPMock fakeHTTPClientFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"text/plain; charset=utf-8"},
			},
			Body: ioutil.NopCloser(bytes.NewReader([]byte(`
4.3.2.1 ___id___.c.mystat-in.net		# comment with double tab
1.1.1.1 a.cn b.cn a.cn # "a.cn" is duplicate

::1  localfoo
2606:4700:4700::1111 cloudflare #[cf]

broken line format

0.0.0.1	example.com
0.0.0.1 example.com # duplicate
`))),
		}, nil
	}

	h.(*handler).httpClient = customHTTPMock

	var (
		req, _ = http.NewRequest(http.MethodGet, "http://testing?"+
			"&sources_urls="+
			"https%3A%2F%2Fmock%2Fad_servers.txt"+
			"&excluded_hosts="+
			"a.cn", http.NoBody)
		rr = httptest.NewRecorder()
	)

	h.ServeHTTP(rr, req)

	body := rr.Body.String()

	assert.NotContains(t, body, "name=\"a.cn\"")
	assert.Contains(t, body, "name=\"___id___.c.mystat-in.net\"")
	assert.Contains(t, body, "name=\"b.cn\"")
	assert.Contains(t, body, "name=\"localfoo\"")
	assert.Contains(t, body, "name=\"cloudflare\"")
	assert.Contains(t, body, "name=\"example.com\"")
	assert.Contains(t, body, "/ip dns static")
	assert.Equal(t, strings.Count(body, "add address=127.0.0.1 comment=\"foo\" disabled=no"), 5)
}

func TestHandler_ServeHTTPWithoutRequest(t *testing.T) { //nolint:dupl
	cacher := cache.NewInMemoryCache(time.Minute, time.Second)
	defer cacher.Close()

	h, err := NewHandler(context.Background(), zap.NewNop(), cacher, createConfig())
	assert.NoError(t, err)

	var rr = httptest.NewRecorder()

	h.ServeHTTP(rr, nil)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "## Empty request or query parameters\n", rr.Body.String())
}

func TestHandler_ServeHTTPRequestWithoutSourcesURLs(t *testing.T) { //nolint:dupl
	cacher := cache.NewInMemoryCache(time.Minute, time.Second)
	defer cacher.Close()

	h, err := NewHandler(context.Background(), zap.NewNop(), cacher, createConfig())
	assert.NoError(t, err)

	var (
		req, _ = http.NewRequest(http.MethodGet, "http://testing", http.NoBody)
		rr     = httptest.NewRecorder()
	)

	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Regexp(t, `(?mU)## Query parameters error.*sources_urls`, rr.Body.String())
}

func TestHandler_ServeHTTPRequestEmptySourcesURLs(t *testing.T) { //nolint:dupl
	cacher := cache.NewInMemoryCache(time.Minute, time.Second)
	defer cacher.Close()

	h, err := NewHandler(context.Background(), zap.NewNop(), cacher, createConfig())
	assert.NoError(t, err)

	var (
		req, _ = http.NewRequest(http.MethodGet, "http://testing?sources_urls=", http.NoBody)
		rr     = httptest.NewRecorder()
	)

	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Regexp(t, `(?mU)## Query parameters.*fail.*empty.*sources`, rr.Body.String())
}

func TestHandler_ServeHTTPRequestWrongFormat(t *testing.T) { //nolint:dupl
	cacher := cache.NewInMemoryCache(time.Minute, time.Second)
	defer cacher.Close()

	h, err := NewHandler(context.Background(), zap.NewNop(), cacher, createConfig())
	assert.NoError(t, err)

	var (
		req, _ = http.NewRequest(http.MethodGet, "http://testing?sources_urls=http://foo&format=foobar", http.NoBody)
		rr     = httptest.NewRecorder()
	)

	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Regexp(t, `(?mU)## Unsupported format.*foobar`, rr.Body.String())
}
