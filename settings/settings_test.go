package settings

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestFromYaml(t *testing.T) { //nolint:funlen
	t.Parallel()

	var cases = []struct {
		name          string
		giveYaml      []byte
		giveExpandEnv bool
		wantSettings  *Settings
	}{
		{
			name:          "Using yaml part",
			giveExpandEnv: true,
			giveYaml: []byte(`
listen:
 address: '1.2.3.4'
 port: 321
`),
			wantSettings: &Settings{
				Listen: listen{
					Address: "1.2.3.4",
					Port:    321,
				},
			},
		},
		{
			name:          "ENV variables expanding",
			giveExpandEnv: true,
			giveYaml: []byte(`
listen:
 address: ${__TEST_IP_ADDR}
 port: ${__TEST_PORT_NUM}
`),
			wantSettings: &Settings{
				Listen: listen{
					Address: "8.7.8.7",
					Port:    4567,
				},
			},
		},
		{
			name:          "ENV variables not expanding",
			giveExpandEnv: false,
			giveYaml: []byte(`
listen:
 address: ${__TEST_IP_ADDR}
 port: 0
`),
			wantSettings: &Settings{
				Listen: listen{
					Address: "${__TEST_IP_ADDR}",
					Port:    0,
				},
			},
		},
		{
			name:          "Default env values is set",
			giveExpandEnv: true,
			giveYaml: []byte(`
listen:
  address: ${__NON_EXISTING_VALUE_FOR_ADDR__:-3.4.5.6}
  port: ${__NON_EXISTING_VALUE_FOR_PORT__:-1234}
`),
			wantSettings: &Settings{
				Listen: listen{
					Address: "3.4.5.6",
					Port:    1234,
				},
			},
		},
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
   count: -2

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
 max_source_size: -4
`),
			wantSettings: &Settings{
				Listen: listen{
					Address: "1.2.3.4",
					Port:    321,
				},
				Resources: resources{
					DirPath:      "/tmp",
					IndexName:    "idx.html",
					Error404Name: "err404.asp",
				},
				Sources: []source{{
					URI:              "http://goo.gl/hosts.txt",
					Name:             "Foo name",
					Description:      "Foo desc",
					EnabledByDefault: true,
					RecordsCount:     123,
				}, {
					URI:              "http://example.com/txt.stsoh",
					Name:             "Bar name",
					Description:      "Bar desc",
					EnabledByDefault: false,
					RecordsCount:     321,
				}, {
					URI:          "http://goo.gl/txt.stsoh",
					RecordsCount: -2,
				}},
				Cache: cache{
					File: cacheFiles{
						DirPath: "/foo/bar",
					},
					LifetimeSec: 10,
				},
				RouterScript: routerScript{
					Redirect: redirect{
						Address: "0.1.1.0",
					},
					Exclude: excludes{
						Hosts: []string{"foo", "bar"},
					},
					MaxSources:    1,
					MaxSourceSize: -4,
					Comment:       " [ blah ] ",
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("__TEST_IP_ADDR", "8.7.8.7")
			_ = os.Setenv("__TEST_PORT_NUM", "4567")

			settings, err := FromYaml(tt.giveYaml, tt.giveExpandEnv)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(settings, tt.wantSettings) {
				t.Errorf(
					`Wrong yaml (as a string) decoding result. Want: %+v, got: %+v`,
					tt.wantSettings,
					settings,
				)
			}

			_ = os.Unsetenv("__TEST_IP_ADDR")
			_ = os.Unsetenv("__TEST_PORT_NUM")
		})
	}
}

func TestFromYamlFile(t *testing.T) {
	t.Parallel()

	var cases = []struct {
		name          string
		giveYaml      []byte
		giveExpandEnv bool
		wantError     bool
		wantSettings  *Settings
	}{
		{
			name:          "Using correct yaml",
			giveExpandEnv: true,
			giveYaml: []byte(`
listen:
  address: '1.2.3.4'
  port: 321
`),
			wantSettings: &Settings{
				Listen: listen{
					Address: "1.2.3.4",
					Port:    321,
				},
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
			file, _ := ioutil.TempFile("", "unit-test-")
			defer func() {
				if err := file.Close(); err != nil {
					panic(err)
				}
				if err := os.Remove(file.Name()); err != nil {
					panic(err)
				}
			}()

			if _, err := file.Write(tt.giveYaml); err != nil {
				t.Fatal(err)
			}
			settings, err := FromYamlFile(file.Name(), tt.giveExpandEnv)

			if tt.wantError {
				if err == nil {
					t.Error(`Expected error not returned`)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}

				if tt.wantSettings != nil && !reflect.DeepEqual(settings, tt.wantSettings) {
					t.Errorf(
						`Wrong yaml (as a string) decoding result. Want: %+v, got: %+v`,
						tt.wantSettings,
						settings,
					)
				}
			}
		})
	}
}
