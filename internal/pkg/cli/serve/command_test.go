package serve

import (
	"bufio"
	"bytes"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/kami-zh/go-capturer"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func TestSuccessfulCommandRunning(t *testing.T) {
	getRandomTCPPort := func() (int, error) {
		t.Helper()

		// zero port means randomly (os) chosen port
		listener, err := net.Listen("tcp", ":0") //nolint:gosec
		if err != nil {
			return 0, err
		}

		port := listener.Addr().(*net.TCPAddr).Port

		if err := listener.Close(); err != nil {
			return 0, err
		}

		return port, nil
	}

	// create logger instance with output capturing
	var (
		logBuf  bytes.Buffer
		encoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		writer  = bufio.NewWriter(&logBuf)
		log     = zap.New(zapcore.NewCore(encoder, zapcore.AddSync(writer), zapcore.DebugLevel))
	)

	// get TCP port number for a test
	tcpPort, err := getRandomTCPPort()
	assert.NoError(t, err)

	// create command with valid flags to run
	cmd := NewCommand(log)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"-r", "", "--port", strconv.Itoa(tcpPort), "-c", configFilePath})
	var output string

	executed := make(chan struct{})

	// start HTTP server
	go func() {
		defer close(executed)

		output = capturer.CaptureOutput(func() {
			assert.NoError(t, cmd.Execute())
		})

		executed <- struct{}{}
	}()

	portBusy := make(chan struct{})

	// check port "busy" (by HTTP server) state
	go func() {
		defer close(portBusy)

		for i := 0; i < 3000; i++ {
			listener, e := net.Listen("tcp", ":"+strconv.Itoa(tcpPort))
			if e != nil {
				portBusy <- struct{}{}
				return
			}
			assert.NoError(t, listener.Close())
			<-time.After(time.Millisecond)
		}

		t.Error("port opening timeout exceeded")
	}()

	<-portBusy // wait for server starting

	// send OS signal for server stopping
	proc, err := os.FindProcess(os.Getpid())
	assert.NoError(t, err)
	assert.NoError(t, proc.Signal(syscall.SIGINT)) // send the signal

	<-executed // wait until server has been stopped

	// flush the logger buffer
	assert.NoError(t, writer.Flush())
	logged := logBuf.String()

	assert.Empty(t, output) // there is no output, all must be inside logger buffer

	// log asserts is a very bed practice, but i have no idea how to test command execution better
	assert.Contains(t, logged, "Server starting")
	assert.Contains(t, logged, "Stopping by OS signal")
	assert.Contains(t, logged, "Server stopping")
}

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
