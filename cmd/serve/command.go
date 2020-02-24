package serve

import (
	"errors"
	"fmt"
	"mikrotik-hosts-parser/http"
	serveSettings "mikrotik-hosts-parser/settings/serve"
	"net"
	"os"
	"strconv"
	"time"
)

type (
	listenAddress    string
	listenPort       int
	resourcesDirPath string
	configFilePath   string

	listenOptions struct {
		Address listenAddress `short:"l" long:"listen" env:"LISTEN_ADDR" description:"Address (IP) to listen on"`
		Port    listenPort    `short:"p" long:"port" env:"LISTEN_PORT" description:"TCP port number"`
	}

	resourcesOptions struct {
		ResourcesDir resourcesDirPath `short:"r" long:"resources-dir" env:"RESOURCES_DIR" description:"Resources directory path"`
	}
)

type (
	Command struct {
		ConfigFile configFilePath `short:"c" long:"config" env:"CONFIG_PATH" required:"true" description:"Config file path"`

		ResourcesOptions resourcesOptions `group:"Resources"`
		ServingOptions   listenOptions    `group:"Listening"`
	}
)

// Convert struct into string representation.
func (s listenAddress) String() string {
	return string(s)
}

// Convert struct into string representation.
func (s resourcesDirPath) String() string {
	return string(s)
}

// Convert struct into string representation.
func (s configFilePath) String() string {
	return string(s)
}

// Validate address for listening on.
func (listenAddress) IsValidValue(ip string) error {
	if net.ParseIP(ip) == nil {
		return errors.New("wrong address for listening value (invalid IP address)")
	}
	return nil
}

// Validate config file path
func (configFilePath) IsValidValue(value string) error {
	if info, err := os.Stat(value); err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("config file [%s] was not found", value)
	}

	return nil
}

// Validate port for listening
func (listenPort) IsValidValue(value string) error {
	portNum, err := strconv.Atoi(value)
	if err != nil {
		return errors.New("wrong port value (cannot be converted into number)")
	}

	if portNum <= 0 || portNum > 65535 {
		return errors.New("wrong port number (must be in interval 1..65535)")
	}

	return nil
}

// Validate resources directory path.
func (resourcesDirPath) IsValidValue(value string) error {
	if info, err := os.Stat(value); err != nil || !info.Mode().IsDir() {
		return fmt.Errorf("resources directory [%s] was not found", value)
	}

	return nil
}

// Get serving settings
func (c *Command) getSettings(filepath string) (*serveSettings.Settings, error) {
	sets, err := serveSettings.FromYamlFile(filepath, true)
	if err != nil {
		return nil, err
	}

	// override settings using passed command options
	if len(c.ServingOptions.Address) > 0 {
		sets.Listen.Address = c.ServingOptions.Address.String()
	}
	if c.ServingOptions.Port != 0 {
		sets.Listen.Port = int(c.ServingOptions.Port)
	}
	if len(c.ResourcesOptions.ResourcesDir) > 0 {
		sets.Resources.DirPath = c.ResourcesOptions.ResourcesDir.String()
	}

	return sets, nil
}

// Execute the command.
func (c *Command) Execute(_ []string) error {
	settings, err := c.getSettings(c.ConfigFile.String())
	if err != nil {
		return err
	}

	server := http.NewServer(&http.ServerSettings{
		WriteTimeout:     time.Second * 15,
		ReadTimeout:      time.Second * 15,
		KeepAliveEnabled: false,
	}, settings)

	server.RegisterHandlers()

	_ = settings.PrintInfo(os.Stdout)

	return server.Start()
}
