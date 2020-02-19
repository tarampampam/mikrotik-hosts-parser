package version

import "testing"

func TestVersion(t *testing.T) {
	t.Parallel()

	if Version != "undefined@undefined" {
		t.Errorf("Unexpected default version value: %s", Version)
	}
}
