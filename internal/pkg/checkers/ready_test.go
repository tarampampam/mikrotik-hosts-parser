package checkers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadyChecker_Check(t *testing.T) {
	assert.NoError(t, NewReadyChecker().Check())
}
