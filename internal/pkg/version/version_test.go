package version

import "testing"

func TestVersion(t *testing.T) {
	if value := Version(); value != "undefined@undefined" {
		t.Errorf("Unexpected default version value: %s", value)
	}
}
