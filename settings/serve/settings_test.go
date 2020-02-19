package serve

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestFromYaml(t *testing.T) {
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
				Listen: Listen{
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
				Listen: Listen{
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
				Listen: Listen{
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
				Listen: Listen{
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
				Listen: Listen{
					Address: "1.2.3.4",
					Port:    321,
				},
				Resources: Resources{
					DirPath:      "/tmp",
					IndexName:    "idx.html",
					Error404Name: "err404.asp",
				},
				Sources: []Source{{
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
				Cache: Cache{
					File: CacheFiles{
						DirPath: "/foo/bar",
					},
					LifetimeSec: 10,
				},
				RouterScript: RouterScript{
					Redirect: Redirect{
						Address: "0.1.1.0",
					},
					Exclude: Excludes{
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
				Listen: Listen{
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

func TestSettings_PrintInfo(t *testing.T) {
	t.Parallel()

	var cases = []struct {
		name           string
		giveSettings   *Settings
		wantEntries    []string
		wantLinesCount byte
	}{
		{
			name: "Regular use-case",
			giveSettings: &Settings{
				Listen: Listen{
					Address: "1.2.3.4",
					Port:    112233,
				},
				Resources: Resources{
					DirPath:      "FooDirPath",
					IndexName:    "FooIndexName",
					Error404Name: "FooError404Name",
				},
				Sources: []Source{
					{URI: "source URI", Name: "source name", Description: "source desc", EnabledByDefault: true, RecordsCount: 123},
				},
				Cache: Cache{
					File: CacheFiles{
						DirPath: "/tmp/foo/bar",
					},
					LifetimeSec: 321,
				},
				RouterScript: RouterScript{
					Redirect: Redirect{
						Address: "",
					},
					Exclude: Excludes{
						Hosts: []string{},
					},
					Comment:       "",
					MaxSources:    222,
					MaxSourceSize: 333,
				},
			},
			wantEntries: []string{
				"Listen address", "1.2.3.4",
				"Listen port", "112233",
				"Resources dir", "FooDirPath",
				"Index file name", "FooIndexName",
				"Error 404 file name", "FooError404Name",
				"Sources count", "1",
				"Cache lifetime (sec)", "321",
				"Cache files directory", "/tmp/foo/bar",
				"Max sources count", "222",
				"Max source response size (bytes)", "333",
			},
			wantLinesCount: 10,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer

			err := tt.giveSettings.PrintInfo(&b)

			if err != nil {
				t.Fatalf("Got an error: %v", err)
			}

			if linesCount := strings.Count(b.String(), "\n"); byte(linesCount) != tt.wantLinesCount {
				t.Errorf("Want lines count %d, got: %d", tt.wantLinesCount, linesCount)
			}

			for _, line := range tt.wantEntries {
				if !strings.Contains(b.String(), line) {
					t.Errorf("Result [%s] does not contains required substring: [%s]", b.String(), line)
				}
			}
		})
	}
}
