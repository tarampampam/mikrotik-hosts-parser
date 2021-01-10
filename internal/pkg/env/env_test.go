package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	assert.Equal(t, "LISTEN_ADDR", ListenAddr)
	assert.Equal(t, "LISTEN_PORT", ListenPort)
	assert.Equal(t, "RESOURCES_DIR", ResourcesDir)
	assert.Equal(t, "CONFIG_PATH", ConfigPath)
}
