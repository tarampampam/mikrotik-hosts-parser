package script

import (
	"mikrotik-hosts-parser/settings/serve"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouterOsScriptSourceGenerationHandlerFunc(t *testing.T) {
	var (
		// @todo: rewrite using mocked http client
		req, _ = http.NewRequest("GET", "http://testing/script/source?"+
			"format=routeros&"+
			"version=v0.0.666@1a0339c&"+
			"redirect_to=127.0.0.1&"+
			"limit=5000&"+
			"sources_urls=https%3A%2F%2Fcdn.jsdelivr.net%2Fgh%2Ftarampampam%2Fmikrotik-hosts-parser%40master%2F.hosts%2Fbasic.txt,"+
			"https://raw.githubusercontent.com/crazy-max/WindowsSpyBlocker/master/data/hosts/spy.txt,"+
			"https%3A%2F%2Fwww.malwaredomainlist.com%2Fhostslist%2Fhosts.txt,"+
			"https%3A%2F%2Fpgl.yoyo.org%2Fadservers%2Fserverlist.php%3Fhostformat%3Dhosts%26showintro%3D0%26mimetype%3Dplaintext"+
			"&excluded_hosts=localhost,"+
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
			}, {
				URI:              "http://face.book/hosts.txt",
				Name:             "Bar name",
				Description:      "Bar desc",
				EnabledByDefault: false,
				RecordsCount:     -321,
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
		}
	)

	RouterOsScriptSourceGenerationHandlerFunc(&serveSettings)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Wrong response HTTP code. Want %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()

	for _, substring := range []string{"/ip dns static"} {
		if !strings.Contains(body, substring) {
			t.Errorf("Expected substring '%s' was not found in response (%s)", substring, body)
		}
	}
}
