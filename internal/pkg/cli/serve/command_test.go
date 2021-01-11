package serve

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/kami-zh/go-capturer"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestProperties(t *testing.T) {
	cmd := NewCommand(context.Background(), zap.NewNop())

	assert.Equal(t, "serve", cmd.Use)
	assert.ElementsMatch(t, []string{"s", "server"}, cmd.Aliases)
	assert.NotNil(t, cmd.RunE)
}

func TestFlags(t *testing.T) {
	cmd := NewCommand(context.Background(), zap.NewNop())
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
		tt := tt
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
	cmd := NewCommand(context.Background(), zap.NewNop())
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
	cmd := NewCommand(context.Background(), zap.NewNop())
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
	cmd := NewCommand(context.Background(), zap.NewNop())
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
	cmd := NewCommand(context.Background(), zap.NewNop())
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
	cmd := NewCommand(context.Background(), zap.NewNop())

	// `-p` flag must be ignored
	cmd.SetArgs([]string{"-r", "", "-c", configFilePath, "-p", "8090"})

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
	assert.False(t, executed)
}

func TestResourcesDirFlagWrongArgument(t *testing.T) {
	cmd := NewCommand(context.Background(), zap.NewNop())
	cmd.SetArgs([]string{"-r", "/tmp/nonexistent/bar/baz", "-c", configFilePath})

	var executed bool

	cmd.RunE = func(*cobra.Command, []string) error {
		executed = true

		return nil
	}

	output := capturer.CaptureStderr(func() {
		assert.Error(t, cmd.Execute())
	})

	assert.Contains(t, output, "wrong resources directory")
	assert.Contains(t, output, "/tmp/nonexistent/bar/baz")
	assert.False(t, executed)
}

func TestResourcesDirFlagWrongEnvValue(t *testing.T) {
	cmd := NewCommand(context.Background(), zap.NewNop())
	cmd.SetArgs([]string{"-c", configFilePath, "-r", "."}) // `-r` flag must be ignored

	assert.NoError(t, os.Setenv("RESOURCES_DIR", "/tmp/nonexistent/bar/baz"))

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
	assert.Contains(t, output, "/tmp/nonexistent/bar/baz")
	assert.False(t, executed)
}

func TestConfigFlagWrongArgument(t *testing.T) {
	cmd := NewCommand(context.Background(), zap.NewNop())
	cmd.SetArgs([]string{"-r", "", "-c", "/tmp/nonexistent/bar.baz"})

	var executed bool

	cmd.RunE = func(*cobra.Command, []string) error {
		executed = true

		return nil
	}

	output := capturer.CaptureStderr(func() {
		assert.Error(t, cmd.Execute())
	})

	assert.Contains(t, output, "config file")
	assert.Contains(t, output, "/tmp/nonexistent/bar.baz")
	assert.Contains(t, output, "not found")
	assert.False(t, executed)
}

func TestConfigFlagWrongEnvValue(t *testing.T) {
	cmd := NewCommand(context.Background(), zap.NewNop())
	cmd.SetArgs([]string{"-r", "", "-c", configFilePath}) // `-c` flag must be ignored

	assert.NoError(t, os.Setenv("CONFIG_PATH", "/tmp/nonexistent/bar.baz"))

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
	assert.Contains(t, output, "/tmp/nonexistent/bar.baz")
	assert.Contains(t, output, "not found")
	assert.False(t, executed)
}

func getRandomTCPPort(t *testing.T) (int, error) {
	t.Helper()

	// zero port means randomly (os) chosen port
	l, err := net.Listen("tcp", ":0") //nolint:gosec
	if err != nil {
		return 0, err
	}

	port := l.Addr().(*net.TCPAddr).Port

	if closingErr := l.Close(); closingErr != nil {
		return 0, closingErr
	}

	return port, nil
}

func checkTCPPortIsBusy(t *testing.T, port int) bool {
	t.Helper()

	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return true
	}

	_ = l.Close()

	return false
}

func TestSuccessfulCommandRunning(t *testing.T) {
	// get TCP port number for a test
	tcpPort, err := getRandomTCPPort(t)
	assert.NoError(t, err)

	// start mini-redis
	mini, err := miniredis.Run()
	assert.NoError(t, err)

	defer mini.Close()

	var (
		output     string
		executedCh = make(chan struct{})
	)

	// start HTTP server
	go func(ch chan<- struct{}) {
		defer close(ch)

		output = capturer.CaptureStderr(func() {
			// create command with valid flags to run
			log, _ := zap.NewDevelopment()
			cmd := NewCommand(context.Background(), log)
			cmd.SilenceUsage = true
			cmd.SetArgs([]string{"-r", "", "--port", strconv.Itoa(tcpPort), "-c", configFilePath, "--redis-dsn", fmt.Sprintf("redis://127.0.0.1:%s/0", mini.Port())}) //nolint:lll

			assert.NoError(t, cmd.Execute())
		})

		ch <- struct{}{}
	}(executedCh)

	portBusyCh := make(chan struct{})

	// check port "busy" (by HTTP server) state
	go func(ch chan<- struct{}) {
		defer close(ch)

		for i := 0; i < 2000; i++ {
			if checkTCPPortIsBusy(t, tcpPort) {
				ch <- struct{}{}

				return
			}

			<-time.After(time.Millisecond * 2)
		}

		t.Error("port opening timeout exceeded")
	}(portBusyCh)

	<-portBusyCh // wait for server starting

	// send OS signal for server stopping
	proc, err := os.FindProcess(os.Getpid())
	assert.NoError(t, err)
	assert.NoError(t, proc.Signal(syscall.SIGINT)) // send the signal

	<-executedCh // wait until server has been stopped

	// next asserts is a very bed practice, but i have no idea how to test command execution better
	assert.Contains(t, output, "Server starting")
	assert.Contains(t, output, "Stopping by OS signal")
	assert.Contains(t, output, "Server stopping")
}

func TestRunningUsingBusyPortFailing(t *testing.T) {
	port, err := getRandomTCPPort(t)
	assert.NoError(t, err)

	// start mini-redis
	mini, err := miniredis.Run()
	assert.NoError(t, err)

	defer mini.Close()

	// occupy a TCP port
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	assert.NoError(t, err)

	defer func() { assert.NoError(t, l.Close()) }()

	// create command with valid flags to run
	cmd := NewCommand(context.Background(), zap.NewNop())
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"-r", "", "--port", strconv.Itoa(port), "-c", configFilePath, "--redis-dsn", fmt.Sprintf("redis://127.0.0.1:%s/0", mini.Port())}) //nolint:lll

	executedCh := make(chan struct{})

	// start HTTP server
	go func(ch chan<- struct{}) {
		defer close(ch)

		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "address already in use")

		ch <- struct{}{}
	}(executedCh)

	<-executedCh // wait until server has been stopped
}
