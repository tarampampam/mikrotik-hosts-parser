// Package env contains all about environment variables, that can be used by current application.
package env

import "os"

type envVariable string

const (
	// ListenAddr is IP address for listening.
	ListenAddr envVariable = "LISTEN_ADDR"

	// ListenPort is port number for listening.
	ListenPort envVariable = "LISTEN_PORT"

	// ResourcesDir is a directory with resources.
	ResourcesDir envVariable = "RESOURCES_DIR"

	// ConfigPath is a path to the configuration file.
	ConfigPath envVariable = "CONFIG_PATH"
)

// String returns environment variable name in the string representation.
func (e envVariable) String() string { return string(e) }

// Lookup retrieves the value of the environment variable. If the variable is present in the environment the value
// (which may be empty) is returned and the boolean is true. Otherwise the returned value will be empty and the
// boolean will be false.
func (e envVariable) Lookup() (string, bool) { return os.LookupEnv(e.String()) }
