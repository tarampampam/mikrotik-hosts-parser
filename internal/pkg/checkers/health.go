package checkers

import (
	"fmt"
	"net/http"
	"time"
)

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type HealthChecker struct {
	httpClient httpClient
}

func NewHealthChecker(client ...httpClient) *HealthChecker {
	var c httpClient

	if len(client) == 1 {
		c = client[0]
	} else {
		c = &http.Client{Timeout: time.Second * 3} // default
	}

	return &HealthChecker{httpClient: c}
}

func (c *HealthChecker) Check(port uint16) error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/live", port), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "HealthChecker/internal")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()

	if code := resp.StatusCode; code != http.StatusOK {
		return fmt.Errorf("wrong status code [%d] from live endpoint", code)
	}

	return nil
}
