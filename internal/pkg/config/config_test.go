package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_AddSource(t *testing.T) {
	cfg := Config{}

	assert.Len(t, cfg.Sources, 0)
	cfg.AddSource("https://foo", "foo", "foo desc", true, 123)
	assert.Len(t, cfg.Sources, 1)
	assert.Equal(t, "https://foo", cfg.Sources[0].URI)
	assert.Equal(t, "foo", cfg.Sources[0].Name)
	assert.Equal(t, "foo desc", cfg.Sources[0].Description)
	assert.True(t, cfg.Sources[0].EnabledByDefault)
	assert.Equal(t, uint(123), cfg.Sources[0].RecordsCount)
}

func TestFromYaml(t *testing.T) {
	var cases = []struct { //nolint:maligned
		name          string
		giveYaml      []byte
		giveExpandEnv bool
		giveEnv       map[string]string
		wantErr       bool
		checkResultFn func(*testing.T, *Config)
		wantConfig    *Config
	}{

		{
			name:          "With all possible values",
			giveExpandEnv: true,
			giveYaml: []byte(`
# Some comment

sources:
 - uri: http://goo.gl/hosts.txt
   name: Foo name
   description: Foo desc
   enabled: true
   count: 123
 - uri: http://example.com/txt.stsoh # inline comment
   name: Bar name
   description: Bar desc
   enabled: false
   count: 321
 - uri: http://goo.gl/txt.stsoh
   count: 2

router_script:
 redirect:
   address: 0.1.1.0
 exclude:
   hosts:
     - "foo"
     - bar
 comment: " [ blah ] "
 max_sources: 1
 max_source_size: 4
`),
			wantErr: false,
			checkResultFn: func(t *testing.T, config *Config) {
				assert.Equal(t, "http://goo.gl/hosts.txt", config.Sources[0].URI)
				assert.Equal(t, "Foo name", config.Sources[0].Name)
				assert.Equal(t, "Foo desc", config.Sources[0].Description)
				assert.True(t, config.Sources[0].EnabledByDefault)
				assert.Equal(t, uint(123), config.Sources[0].RecordsCount)

				assert.Equal(t, "http://example.com/txt.stsoh", config.Sources[1].URI)
				assert.Equal(t, "Bar name", config.Sources[1].Name)
				assert.Equal(t, "Bar desc", config.Sources[1].Description)
				assert.False(t, config.Sources[1].EnabledByDefault)
				assert.Equal(t, uint(321), config.Sources[1].RecordsCount)

				assert.Equal(t, "http://goo.gl/txt.stsoh", config.Sources[2].URI)
				assert.Equal(t, uint(2), config.Sources[2].RecordsCount)

				assert.Equal(t, "0.1.1.0", config.RouterScript.Redirect.Address)
				assert.ElementsMatch(t, []string{"foo", "bar"}, config.RouterScript.Exclude.Hosts)
				assert.Equal(t, " [ blah ] ", config.RouterScript.Comment)
				assert.Equal(t, uint16(1), config.RouterScript.MaxSourcesCount)
				assert.Equal(t, uint32(4), config.RouterScript.MaxSourceSizeBytes)
			},
		},

		{
			name:          "ENV variables expanded",
			giveExpandEnv: true,
			giveEnv:       map[string]string{"__TEST_ADDR": "1.2.3.4", "__TEST_COMMENT": "foo"},
			giveYaml: []byte(`
router_script:
 redirect:
   address: ${__TEST_ADDR}
 comment: ${__TEST_COMMENT}
`),
			wantErr: false,
			checkResultFn: func(t *testing.T, config *Config) {
				assert.Equal(t, "1.2.3.4", config.RouterScript.Redirect.Address)
				assert.Equal(t, "foo", config.RouterScript.Comment)
			},
		},
		{
			name:          "ENV variables NOT expanded",
			giveExpandEnv: false,
			giveYaml: []byte(`
router_script:
 redirect:
   address: ${__TEST_ADDR}
`),
			wantErr: false,
			checkResultFn: func(t *testing.T, config *Config) {
				assert.Equal(t, "${__TEST_ADDR}", config.RouterScript.Redirect.Address)
			},
		},
		{
			name:          "ENV variables defaults",
			giveExpandEnv: true,
			giveYaml: []byte(`
router_script:
 redirect:
   address: ${__TEST_ADDR:-2.3.4.5}
 comment: ${__TEST_COMMENT:-foo}
`),
			wantErr: false,
			checkResultFn: func(t *testing.T, config *Config) {
				assert.Equal(t, "2.3.4.5", config.RouterScript.Redirect.Address)
				assert.Equal(t, "foo", config.RouterScript.Comment)
			},
		},
		{
			name:     "broken yaml",
			giveYaml: []byte(`foo bar`),
			wantErr:  true,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.giveEnv != nil {
				for key, value := range tt.giveEnv {
					assert.NoError(t, os.Setenv(key, value))
				}
			}

			conf, err := FromYaml(tt.giveYaml, tt.giveExpandEnv)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
				tt.checkResultFn(t, conf)
			}

			if tt.giveEnv != nil {
				for key := range tt.giveEnv {
					assert.NoError(t, os.Unsetenv(key))
				}
			}
		})
	}
}

func TestFromYamlFile(t *testing.T) {
	var cases = []struct {
		name          string
		giveYaml      []byte
		giveExpandEnv bool
		wantError     bool
		checkResultFn func(*testing.T, *Config)
	}{
		{
			name:          "Using correct yaml",
			giveExpandEnv: true,
			giveYaml: []byte(`
router_script:
 redirect:
   address: 0.1.1.0
 comment: "foo"
`),
			checkResultFn: func(t *testing.T, config *Config) {
				assert.Equal(t, "0.1.1.0", config.RouterScript.Redirect.Address)
				assert.Equal(t, "foo", config.RouterScript.Comment)
			},
		},
		{
			name:          "Using broken file (wrong format)",
			giveExpandEnv: true,
			giveYaml:      []byte(`!foo bar`),
			wantError:     true,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "unit-test-")
			assert.NoError(t, err)

			_, err = file.Write(tt.giveYaml)
			assert.NoError(t, err)
			assert.NoError(t, file.Close())

			defer func() { assert.NoError(t, os.Remove(file.Name())) }() // cleanup

			conf, loadingErr := FromYamlFile(file.Name(), tt.giveExpandEnv)

			if tt.wantError {
				assert.Error(t, loadingErr)
			} else {
				assert.NoError(t, loadingErr)
				tt.checkResultFn(t, conf)
			}
		})
	}
}
