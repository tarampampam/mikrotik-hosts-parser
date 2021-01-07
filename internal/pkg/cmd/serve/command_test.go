package serve

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCommand_Structures(t *testing.T) {
	tests := []struct {
		element         func() reflect.StructField
		wantShort       string
		wantLong        string
		wantEnv         string
		wantDescription string
		wantRequired    string
		wantGroup       string
	}{
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(listenOptions{}).FieldByName("Address")
				return field
			},
			wantShort:       "l",
			wantLong:        "listen",
			wantEnv:         "LISTEN_ADDR",
			wantDescription: "Address (IP) to listen on",
		},
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(listenOptions{}).FieldByName("Port")
				return field
			},
			wantShort:       "p",
			wantLong:        "port",
			wantEnv:         "LISTEN_PORT",
			wantDescription: "TCP port number",
		},

		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(resourcesOptions{}).FieldByName("ResourcesDir")
				return field
			},
			wantShort:       "r",
			wantLong:        "resources-dir",
			wantEnv:         "RESOURCES_DIR",
			wantDescription: "resources directory path",
		},

		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(Command{}).FieldByName("ConfigFile")
				return field
			},
			wantShort:       "c",
			wantLong:        "config",
			wantEnv:         "CONFIG_PATH",
			wantDescription: "Config file path",
			wantRequired:    "true",
		},
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(Command{}).FieldByName("ResourcesOptions")
				return field
			},
			wantGroup: "resources",
		},
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(Command{}).FieldByName("ServingOptions")
				return field
			},
			wantGroup: "Listening",
		},
	}

	for _, tt := range tests {
		t.Run(tt.wantDescription, func(t *testing.T) {
			el := tt.element()
			if tt.wantShort != "" {
				value, _ := el.Tag.Lookup("short")
				if value != tt.wantShort {
					t.Errorf("Wrong value for 'short' tag. Want: %v, got: %v", tt.wantShort, value)
				}
			}

			if tt.wantLong != "" {
				value, _ := el.Tag.Lookup("long")
				if value != tt.wantLong {
					t.Errorf("Wrong value for 'long' tag. Want: %v, got: %v", tt.wantLong, value)
				}
			}

			if tt.wantEnv != "" {
				value, _ := el.Tag.Lookup("env")
				if value != tt.wantEnv {
					t.Errorf("Wrong value for 'env' tag. Want: %v, got: %v", tt.wantEnv, value)
				}
			}

			if tt.wantDescription != "" {
				value, _ := el.Tag.Lookup("description")
				if value != tt.wantDescription {
					t.Errorf("Wrong value for 'description' tag. Want: %v, got: %v", tt.wantDescription, value)
				}
			}

			if tt.wantRequired != "" {
				value, _ := el.Tag.Lookup("required")
				if value != tt.wantRequired {
					t.Errorf("Wrong value for 'required' tag. Want: %v, got: %v", tt.wantRequired, value)
				}
			}

			if tt.wantGroup != "" {
				value, _ := el.Tag.Lookup("group")
				if value != tt.wantGroup {
					t.Errorf("Wrong value for 'group' tag. Want: %v, got: %v", tt.wantGroup, value)
				}
			}
		})
	}
}

func TestStringableStruct_String(t *testing.T) {
	if listenAddress("foo").String() != "foo" {
		t.Error("Wrong convertation into string")
	}

	if resourcesDirPath("bar").String() != "bar" {
		t.Error("Wrong convertation into string")
	}

	if configFilePath("baz").String() != "baz" {
		t.Error("Wrong convertation into string")
	}
}

func TestConfigFilePath_IsValidValue(t *testing.T) {
	// create temp dir (and delete if after test)
	tmpDir, dirErr := ioutil.TempDir("", "test-")
	if dirErr != nil {
		t.Fatal(dirErr)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatal(err)
		}
	}()

	// create temp file in temp dir
	tmpFile, fileErr := os.Create(filepath.Join(tmpDir, "test-file"))
	if fileErr != nil {
		t.Fatal(fileErr)
	}
	_ = tmpFile.Close() // is not needed

	tests := []struct {
		name      string
		giveValue string
		wantError error
	}{
		{
			name:      "Correct path",
			giveValue: tmpFile.Name(),
			wantError: nil,
		},
		{
			name:      "Some directory path passed",
			giveValue: tmpDir,
			wantError: fmt.Errorf("config file [%s] was not found", tmpDir),
		},
		{
			name:      "Wrong file path",
			giveValue: "abracadabra !",
			wantError: errors.New("config file [abracadabra !] was not found"),
		},
		{
			name:      "Empty value passed",
			giveValue: "",
			wantError: errors.New("config file [] was not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := configFilePath("").IsValidValue(tt.giveValue)

			if res != nil {
				if tt.wantError == nil {
					t.Errorf("Unexpected error %v returned", res)
				} else if res.Error() != tt.wantError.Error() {
					t.Errorf("Wrong error returned. Want: %v, got: %v", tt.wantError, res)
				}
			}
		})
	}
}

func TestListenPort_IsValidValue(t *testing.T) {
	var defaultErrorMessage = "wrong port number (must be in interval 1..65535)"

	tests := []struct {
		name      string
		giveValue string
		wantError error
	}{
		{
			name:      "Correct port",
			giveValue: "8080",
			wantError: nil,
		},
		{
			name:      "Too much port number",
			giveValue: "65536",
			wantError: errors.New(defaultErrorMessage),
		},
		{
			name:      "Too low port number",
			giveValue: "-1",
			wantError: errors.New(defaultErrorMessage),
		},
		{
			name:      "Empty value passed",
			giveValue: "",
			wantError: errors.New("wrong port value (cannot be converted into number)"),
		},
		{
			name:      "Alpha-string",
			giveValue: "foo bar",
			wantError: errors.New("wrong port value (cannot be converted into number)"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := listenPort(0).IsValidValue(tt.giveValue)

			if res != nil {
				if tt.wantError == nil {
					t.Errorf("Unexpected error %v returned", res)
				} else if res.Error() != tt.wantError.Error() {
					t.Errorf("Wrong error returned. Want: %v, got: %v", tt.wantError, res)
				}
			}
		})
	}
}

func TestResourcesDirPath_IsValidValue(t *testing.T) {
	// create temp dir (and delete if after test)
	tmpDir, dirErr := ioutil.TempDir("", "test-")
	if dirErr != nil {
		t.Fatal(dirErr)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatal(err)
		}
	}()

	// create temp file in temp dir
	tmpFile, fileErr := os.Create(filepath.Join(tmpDir, "test-file"))
	if fileErr != nil {
		t.Fatal(fileErr)
	}
	_ = tmpFile.Close() // is not needed

	tests := []struct {
		name      string
		giveValue string
		wantError error
	}{
		{
			name:      "Correct path",
			giveValue: tmpDir,
			wantError: nil,
		},
		{
			name:      "Some file path passed",
			giveValue: tmpFile.Name(),
			wantError: fmt.Errorf("resources directory [%s] was not found", tmpFile.Name()),
		},
		{
			name:      "Wrong file path",
			giveValue: "abracadabra !",
			wantError: errors.New("resources directory [abracadabra !] was not found"),
		},
		{
			name:      "Empty value passed",
			giveValue: "",
			wantError: errors.New("resources directory [] was not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := resourcesDirPath("").IsValidValue(tt.giveValue)

			if res != nil {
				if tt.wantError == nil {
					t.Errorf("Unexpected error %v returned", res)
				} else if res.Error() != tt.wantError.Error() {
					t.Errorf("Wrong error returned. Want: %v, got: %v", tt.wantError, res)
				}
			}
		})
	}
}

func TestCommand_Execute(t *testing.T) {
	t.Skip("Not implemented yet")
}