package settings

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/cache"
	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/config"
)

func TestNewHandler(t *testing.T) {
	var (
		req, _ = http.NewRequest(http.MethodGet, "http://testing", nil)
		rr     = httptest.NewRecorder()
		cfg    config.Config
	)

	cfg.AddSource("http://goo.gl/hosts.txt", "Foo", "Foo desc", false, 123)
	cfg.AddSource("http://face.book/hosts.txt", "Bar", "Bar desc", true, 321)
	cfg.RouterScript.Redirect.Address = "0.1.1.0"
	cfg.RouterScript.Exclude.Hosts = []string{"foo", "bar"}
	cfg.RouterScript.MaxSourcesCount = 1
	cfg.RouterScript.Comment = " [ blah ] "
	cfg.RouterScript.MaxSourceSizeBytes = 666

	cacher := cache.NewInMemoryCache(time.Second*123, time.Second)

	NewHandler(cfg, cacher)(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, rr.Header().Get("Content-Type"), "application/json")

	assert.JSONEq(t, `{
		"sources":{
			"provided":[
				{"uri":"http://goo.gl/hosts.txt","name":"Foo","description":"Foo desc","default":false,"count":123},
				{"uri":"http://face.book/hosts.txt","name":"Bar","description":"Bar desc","default":true,"count":321}
			],
			"max":1,
			"max_source_size":666
		},
		"redirect":{
			"addr":"0.1.1.0"
		},
		"records":{
			"comment":" [ blah ] "
		},
		"excludes":{
			"hosts":["foo", "bar"]
		},
		"cache":{
			"lifetime_sec":123
		}
	}`, rr.Body.String())
}
