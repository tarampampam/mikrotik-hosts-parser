package serve

import (
	"github.com/kami-zh/go-capturer"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"testing"
)

func TestProperties(t *testing.T) {
	cmd := NewCommand(zap.NewNop())

	assert.Equal(t, "serve", cmd.Use)
	assert.ElementsMatch(t, []string{"s", "server"}, cmd.Aliases)
	assert.NotNil(t, cmd.RunE)
}

func TestFlags(t *testing.T) {
	cmd := NewCommand(zap.NewNop())
	wd, _ := os.Getwd()

	cases := []struct {
		giveName      string
		wantShorthand string
		wantDefault   string
	}{
		{giveName: "listen", wantShorthand: "l", wantDefault: "0.0.0.0"},
		{giveName: "port", wantShorthand: "p", wantDefault: "8080"},
		{giveName: "resources-dir", wantShorthand: "r", wantDefault: filepath.Join(wd, "web")},
		{giveName: "config", wantShorthand: "c", wantDefault: filepath.Join(wd, "configs", "config.yml")},
	}

	for _, tt := range cases {
		t.Run(tt.giveName, func(t *testing.T) {
			flag := cmd.Flag(tt.giveName)

			if flag == nil {
				assert.Failf(t, "flag not found", "flag [%s] was not found", tt.giveName)

				return
			}

			assert.Equal(t, tt.wantShorthand, flag.Shorthand)
			assert.Equal(t, tt.wantDefault, flag.DefValue)
		})
	}
}

const configFilePath = "../../../../configs/config.yml"

func TestSuccessfulFlagsPreparing(t *testing.T) {
	cmd := NewCommand(zap.NewNop())
	cmd.SetArgs([]string{"-r", "", "-c", configFilePath})

	var executed bool
	cmd.RunE = func(*cobra.Command, []string) error {
		executed = true
		return nil
	}

	output := capturer.CaptureOutput(func() {
		assert.NoError(t, cmd.Execute())
	})

	assert.Empty(t, output)
	assert.True(t, executed)
}

func TestListenFlagWrongArgument(t *testing.T) {
	cmd := NewCommand(zap.NewNop())
	cmd.SetArgs([]string{"-r", "", "-c", configFilePath, "-l", "256.256.256.256"}) // 255 is max

	var executed bool
	cmd.RunE = func(*cobra.Command, []string) error {
		executed = true
		return nil
	}

	output := capturer.CaptureStderr(func() {
		assert.Error(t, cmd.Execute())
	})

	assert.Contains(t, output, "wrong IP address")
	assert.Contains(t, output, "256.256.256.256")
	assert.False(t, executed)
}

func TestListenFlagWrongEnvValue(t *testing.T) {
	cmd := NewCommand(zap.NewNop())
	cmd.SetArgs([]string{"-r", "", "-c", configFilePath, "-l", "0.0.0.0"}) // `-l` flag must be ignored

	assert.NoError(t, os.Setenv("LISTEN_ADDR", "256.256.256.256")) // 255 is max
	defer func() { assert.NoError(t, os.Unsetenv("LISTEN_ADDR")) }()

	var executed bool
	cmd.RunE = func(*cobra.Command, []string) error {
		executed = true
		return nil
	}

	output := capturer.CaptureStderr(func() {
		assert.Error(t, cmd.Execute())
	})

	assert.Contains(t, output, "wrong IP address")
	assert.Contains(t, output, "256.256.256.256")
	assert.False(t, executed)
}

func TestPortFlagWrongArgument(t *testing.T) {
	cmd := NewCommand(zap.NewNop())
	cmd.SetArgs([]string{"-r", "", "-c", configFilePath, "-p", "65536"}) // 65535 is max

	var executed bool
	cmd.RunE = func(*cobra.Command, []string) error {
		executed = true
		return nil
	}

	output := capturer.CaptureStderr(func() {
		assert.Error(t, cmd.Execute())
	})

	assert.Contains(t, output, "invalid argument")
	assert.Contains(t, output, "65536")
	assert.Contains(t, output, "value out of range")
	assert.False(t, executed)
}

func TestPortFlagWrongEnvValue(t *testing.T) {
	cmd := NewCommand(zap.NewNop())
	cmd.SetArgs([]string{"-r", "", "-c", configFilePath, "-p", "8090"}) // `-p` flag must be ignored

	assert.NoError(t, os.Setenv("LISTEN_PORT", "65536")) // 65535 is max
	defer func() { assert.NoError(t, os.Unsetenv("LISTEN_PORT")) }()

	var executed bool
	cmd.RunE = func(*cobra.Command, []string) error {
		executed = true
		return nil
	}

	output := capturer.CaptureStderr(func() {
		assert.Error(t, cmd.Execute())
	})

	assert.Contains(t, output, "wrong TCP port")
	assert.Contains(t, output, "environment variable")
	assert.Contains(t, output, "65536")
	assert.Contains(t, output, "cannot be parsed")
	assert.False(t, executed)
}

func TestResourcesDirFlagWrongArgument(t *testing.T) {
	cmd := NewCommand(zap.NewNop())
	cmd.SetArgs([]string{"-r", "/foo/bar/baz", "-c", configFilePath})

	var executed bool
	cmd.RunE = func(*cobra.Command, []string) error {
		executed = true
		return nil
	}

	output := capturer.CaptureStderr(func() {
		assert.Error(t, cmd.Execute())
	})

	assert.Contains(t, output, "wrong resources directory")
	assert.Contains(t, output, "/foo/bar/baz")
	assert.False(t, executed)
}

func TestResourcesDirFlagWrongEnvValue(t *testing.T) {
	cmd := NewCommand(zap.NewNop())
	cmd.SetArgs([]string{"-c", configFilePath, "-r", "."}) // `-r` flag must be ignored

	assert.NoError(t, os.Setenv("RESOURCES_DIR", "/foo/bar/baz"))
	defer func() { assert.NoError(t, os.Unsetenv("RESOURCES_DIR")) }()

	var executed bool
	cmd.RunE = func(*cobra.Command, []string) error {
		executed = true
		return nil
	}

	output := capturer.CaptureStderr(func() {
		assert.Error(t, cmd.Execute())
	})

	assert.Contains(t, output, "wrong resources directory")
	assert.Contains(t, output, "/foo/bar/baz")
	assert.False(t, executed)
}

func TestConfigFlagWrongArgument(t *testing.T) {
	cmd := NewCommand(zap.NewNop())
	cmd.SetArgs([]string{"-r", "", "-c", "/foo/bar.baz"})

	var executed bool
	cmd.RunE = func(*cobra.Command, []string) error {
		executed = true
		return nil
	}

	output := capturer.CaptureStderr(func() {
		assert.Error(t, cmd.Execute())
	})

	assert.Contains(t, output, "config file")
	assert.Contains(t, output, "/foo/bar.baz")
	assert.Contains(t, output, "not found")
	assert.False(t, executed)
}

func TestConfigFlagWrongEnvValue(t *testing.T) {
	cmd := NewCommand(zap.NewNop())
	cmd.SetArgs([]string{"-r", "", "-c", configFilePath}) // `-c` flag must be ignored

	assert.NoError(t, os.Setenv("CONFIG_PATH", "/foo/bar.baz"))
	defer func() { assert.NoError(t, os.Unsetenv("CONFIG_PATH")) }()

	var executed bool
	cmd.RunE = func(*cobra.Command, []string) error {
		executed = true
		return nil
	}

	output := capturer.CaptureStderr(func() {
		assert.Error(t, cmd.Execute())
	})

	assert.Contains(t, output, "config file")
	assert.Contains(t, output, "/foo/bar.baz")
	assert.Contains(t, output, "not found")
	assert.False(t, executed)
}
