package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	assert.Equal(t, "LISTEN_ADDR", string(ListenAddr))
	assert.Equal(t, "LISTEN_PORT", string(ListenPort))
	assert.Equal(t, "RESOURCES_DIR", string(ResourcesDir))
	assert.Equal(t, "CONFIG_PATH", string(ConfigPath))
	assert.Equal(t, "CACHING_ENGINE", string(CachingEngine))
	assert.Equal(t, "CACHE_TTL", string(CacheTTL))
	assert.Equal(t, "REDIS_DSN", string(RedisDSN))
}

func TestEnvVariable_Lookup(t *testing.T) {
	cases := []struct {
		giveEnv envVariable
	}{
		{giveEnv: ListenAddr},
		{giveEnv: ListenPort},
		{giveEnv: ResourcesDir},
		{giveEnv: ConfigPath},
		{giveEnv: CachingEngine},
		{giveEnv: CacheTTL},
		{giveEnv: RedisDSN},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.giveEnv.String(), func(t *testing.T) {
			defer func() { assert.NoError(t, os.Unsetenv(tt.giveEnv.String())) }()

			value, exists := tt.giveEnv.Lookup()
			assert.False(t, exists)
			assert.Empty(t, value)

			assert.NoError(t, os.Setenv(tt.giveEnv.String(), "foo"))

			value, exists = tt.giveEnv.Lookup()
			assert.True(t, exists)
			assert.Equal(t, "foo", value)
		})
	}
}
