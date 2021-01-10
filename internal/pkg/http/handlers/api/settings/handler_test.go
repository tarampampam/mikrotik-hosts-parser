package settings

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/config"
)

func TestNewHandler(t *testing.T) {
	var (
		req, _ = http.NewRequest("GET", "http://testing", nil)
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
	cfg.Cache.LifetimeSec = 222

	NewHandler(cfg)(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)

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
			"lifetime_sec":222
		}
	}`, rr.Body.String())
}
