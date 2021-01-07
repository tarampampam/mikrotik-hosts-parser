package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServingConfigFromYaml(t *testing.T) {
	var cases = []struct {
		name          string
		giveYaml      []byte
		giveExpandEnv bool
		giveEnv       map[string]string
		wantErr       bool
		checkResultFn func(*testing.T, *ServingConfig)
		wantConfig    *ServingConfig
	}{

		{
			name:          "With all possible values",
			giveExpandEnv: true,
			giveYaml: []byte(`
# Some comment

listen:
 address: '1.2.3.4'
 port: 321

resources:
 dir: /tmp
 index_name: idx.html
 error_404_name: err404.asp

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

cache:
 files:
   dir: /foo/bar
 lifetime_sec: 10

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
			checkResultFn: func(t *testing.T, config *ServingConfig) {
				assert.Equal(t, "1.2.3.4", config.Listen.Address)
				assert.Equal(t, uint16(321), config.Listen.Port)

				assert.Equal(t, "/tmp", config.Resources.DirPath)
				assert.Equal(t, "idx.html", config.Resources.IndexName)
				assert.Equal(t, "err404.asp", config.Resources.Error404Name)

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

				assert.Equal(t, "/foo/bar", config.Cache.File.DirPath)
				assert.Equal(t, uint32(10), config.Cache.LifetimeSec)

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
			giveEnv:       map[string]string{"__TEST_LISTEN_ADDR": "1.2.3.4", "__TEST_LISTEN_PORT": "567"},
			giveYaml: []byte(`
listen:
 address: ${__TEST_LISTEN_ADDR}
 port: ${__TEST_LISTEN_PORT}
`),
			wantErr: false,
			checkResultFn: func(t *testing.T, config *ServingConfig) {
				assert.Equal(t, "1.2.3.4", config.Listen.Address)
				assert.Equal(t, uint16(567), config.Listen.Port)
			},
		},
		{
			name:          "ENV variables NOT expanded",
			giveExpandEnv: false,
			giveYaml: []byte(`
listen:
 address: ${__TEST_LISTEN_ADDR}
`),
			wantErr: false,
			checkResultFn: func(t *testing.T, config *ServingConfig) {
				assert.Equal(t, "${__TEST_LISTEN_ADDR}", config.Listen.Address)
			},
		},
		{
			name:          "ENV variables defaults",
			giveExpandEnv: true,
			giveYaml: []byte(`
listen:
 address: ${__TEST_LISTEN_ADDR:-2.3.4.5}
 port: ${__TEST_LISTEN_PORT:-666}
`),
			wantErr: false,
			checkResultFn: func(t *testing.T, config *ServingConfig) {
				assert.Equal(t, "2.3.4.5", config.Listen.Address)
				assert.Equal(t, uint16(666), config.Listen.Port)
			},
		},
		{
			name:     "broken yaml",
			giveYaml: []byte(`foo bar`),
			wantErr:  true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.giveEnv != nil {
				for key, value := range tt.giveEnv {
					assert.NoError(t, os.Setenv(key, value))
				}
			}

			conf, err := ServingConfigFromYaml(tt.giveYaml, tt.giveExpandEnv)

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

func TestServingConfigFromYamlFile(t *testing.T) {
	var cases = []struct {
		name          string
		giveYaml      []byte
		giveExpandEnv bool
		wantError     bool
		checkResultFn func(*testing.T, *ServingConfig)
	}{
		{
			name:          "Using correct yaml",
			giveExpandEnv: true,
			giveYaml: []byte(`
listen:
 address: '1.2.3.4'
 port: 321
`),
			checkResultFn: func(t *testing.T, config *ServingConfig) {
				assert.Equal(t, "1.2.3.4", config.Listen.Address)
				assert.Equal(t, uint16(321), config.Listen.Port)
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
		t.Run(tt.name, func(t *testing.T) {
			file, err := ioutil.TempFile("", "unit-test-")
			assert.NoError(t, err)

			_, err = file.Write(tt.giveYaml)
			assert.NoError(t, err)
			assert.NoError(t, file.Close())

			defer func() { assert.NoError(t, os.Remove(file.Name())) }() // cleanup

			conf, loadingErr := ServingConfigFromYamlFile(file.Name(), tt.giveExpandEnv)

			if tt.wantError {
				assert.Error(t, loadingErr)
			} else {
				assert.NoError(t, loadingErr)
				tt.checkResultFn(t, conf)
			}
		})
	}
}
