// Package env contains all about environment variables, that can be used by current application.
package env

const (
	// ListenAddr is IP address for listening.
	ListenAddr string = "LISTEN_ADDR"

	// ListenPort is port number for listening.
	ListenPort string = "LISTEN_PORT"

	// ResourcesDir is a directory with resources.
	ResourcesDir string = "RESOURCES_DIR"

	// ConfigPath is a path to the configuration file.
	ConfigPath string = "CONFIG_PATH"
)
