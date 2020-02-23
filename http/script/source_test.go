package script

import (
	"bytes"
	"io/ioutil"
	"mikrotik-hosts-parser/settings/serve"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls.
func NewTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(fn), //nolint:unconvert
	}
}

func TestRouterOsScriptSourceGenerationHandlerFunc(t *testing.T) {
	// Create directory in temporary
	createTempDir := func() string {
		t.Helper()
		if dir, err := ioutil.TempDir("", "test-"); err != nil {
			panic(err)
		} else {
			return dir
		}
	}

	tmpDir := createTempDir()

	defer func(d string) {
		if err := os.RemoveAll(d); err != nil {
			panic(err)
		}
	}(tmpDir)

	var (
		req, _ = http.NewRequest("GET", "http://testing/script/source?"+
			"format=routeros&"+
			"version=v0.0.666@1a0339c&"+
			"redirect_to=127.0.0.5&"+
			"limit=1234&"+
			"sources_urls=https%3A%2F%2Ffoo.com%2Fbar.txt,"+
			"http://bar.com/baz.asp,"+
			"http://baz.com/blah.list"+
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
			"ip6-allhosts", nil)
		rr            = httptest.NewRecorder()
		serveSettings = serve.Settings{
			Sources: []serve.Source{{
				URI:              "http://goo.gl/hosts.txt",
				Name:             "Foo name",
				Description:      "Foo desc",
				EnabledByDefault: true,
				RecordsCount:     123,
			}},
			RouterScript: serve.RouterScript{
				Redirect: serve.Redirect{
					Address: "0.1.1.0",
				},
				Exclude: serve.Excludes{
					Hosts: []string{"foo", "bar"},
				},
				MaxSources:    4,
				MaxSourceSize: 2097152,
				Comment:       "AdBlockTest",
			},
			Cache: serve.Cache{
				File: serve.CacheFiles{DirPath: tmpDir},
			},
		}
	)

	// mock default http client
	defaultHTTPClient = &httpClient{
		client: NewTestClient(func(req *http.Request) *http.Response {
			switch req.URL.String() {
			case "https://foo.com/bar.txt":
				return &http.Response{
					StatusCode: 200,
					Body: ioutil.NopCloser(bytes.NewBufferString(`127.0.0.1 a.com
127.0.0.1 b.com
127.0.0.1 c.com
127.0.0.1 d.com`)),
					Header: http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}},
				}
			case "http://bar.com/baz.asp":
				return &http.Response{
					StatusCode: 200,
					Body: ioutil.NopCloser(bytes.NewBufferString(
						"\n\n0.0.0.0 a.org\n127.0.0.1 b.org\n127.0.0.1 c.org",
					)),
					Header: http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}},
				}
			}

			return &http.Response{
				StatusCode: 404,
				Body:       ioutil.NopCloser(bytes.NewBufferString("404 ERROR")),
				Header:     make(http.Header),
			}
		}),
	}

	RouterOsScriptSourceGenerationHandlerFunc(&serveSettings)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Wrong response HTTP code. Want %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()

	for _, substring := range []string{
		`'d.com', 'c.org'`,
		`Limit: 1234`,
		"/ip dns static",
		`add address=127.0.0.5 comment="AdBlockTest" disabled=no name="a.com"`,
		`add address=127.0.0.5 comment="AdBlockTest" disabled=no name="b.com"`,
		`add address=127.0.0.5 comment="AdBlockTest" disabled=no name="c.com"`,
		`add address=127.0.0.5 comment="AdBlockTest" disabled=no name="a.org"`,
		`add address=127.0.0.5 comment="AdBlockTest" disabled=no name="b.org"`,
	} {
		if !strings.Contains(body, substring) {
			t.Errorf("Expected substring '%s' was not found in response (%s)", substring, body)
		}
	}

	for _, substring := range []string{
		`add address=127.0.0.5 comment="AdBlockTest" disabled=no name="d.com"`,
		`add address=127.0.0.5 comment="AdBlockTest" disabled=no name="c.org"`,
	} {
		if strings.Contains(body, substring) {
			t.Errorf("Unexpected substring '%s' was found in response (%s)", substring, body)
		}
	}
}
