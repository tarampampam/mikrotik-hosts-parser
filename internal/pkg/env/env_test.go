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

	assert.Equal(t, "REDIS_HOST", string(RedisHost))
	assert.Equal(t, "REDIS_PORT", string(RedisPort))
	assert.Equal(t, "REDIS_PASSWORD", string(RedisPassword))
	assert.Equal(t, "REDIS_DB_NUM", string(RedisDBNum))
	assert.Equal(t, "REDIS_MAX_CONN", string(RedisMaxConn))
}

func TestEnvVariable_Lookup(t *testing.T) {
	cases := []struct {
		giveEnv envVariable
	}{
		{giveEnv: ListenAddr},
		{giveEnv: ListenPort},
		{giveEnv: ResourcesDir},
		{giveEnv: ConfigPath},

		{giveEnv: RedisHost},
		{giveEnv: RedisPort},
		{giveEnv: RedisPassword},
		{giveEnv: RedisDBNum},
		{giveEnv: RedisMaxConn},
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
