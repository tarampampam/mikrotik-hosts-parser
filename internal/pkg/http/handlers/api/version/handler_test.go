package version

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	var (
		req, _ = http.NewRequest("GET", "http://testing", nil)
		rr     = httptest.NewRecorder()
	)

	NewHandler("1.2.3@foo")(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)
	assert.JSONEq(t, `{"version":"1.2.3@foo"}`, rr.Body.String())
}
