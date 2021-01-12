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
		{giveName: "caching-engine", wantShorthand: "", wantDefault: "memory"},
		{giveName: "redis-dsn", wantShorthand: "", wantDefault: "redis://127.0.0.1:6379/0"},
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

func executeCommandWithoutRunning(t *testing.T, args []string) string {
	cmd := NewCommand(context.Background(), zap.NewNop())
	cmd.SetArgs(args)

	var executed bool

	cmd.RunE = func(*cobra.Command, []string) error {
		executed = true

		return nil
	}

	output := capturer.CaptureStderr(func() {
		assert.Error(t, cmd.Execute())
	})

	assert.False(t, executed)

	return output
}

func TestListenFlagWrongArgument(t *testing.T) {
	output := executeCommandWithoutRunning(t, []string{
		"-r", "",
		"-c", configFilePath,
		"-l", "256.256.256.256", // 255 is max
	})

	assert.Contains(t, output, "wrong IP address")
	assert.Contains(t, output, "256.256.256.256")
}

func TestListenFlagWrongEnvValue(t *testing.T) {
	assert.NoError(t, os.Setenv("LISTEN_ADDR", "256.256.256.256")) // 255 is max

	defer func() { assert.NoError(t, os.Unsetenv("LISTEN_ADDR")) }()

	output := executeCommandWithoutRunning(t, []string{
		"-r", "",
		"-c", configFilePath,
		"-l", "0.0.0.0", // `-l` flag must be ignored
	})

	assert.Contains(t, output, "wrong IP address")
	assert.Contains(t, output, "256.256.256.256")
}

func TestPortFlagWrongArgument(t *testing.T) {
	output := executeCommandWithoutRunning(t, []string{
		"-r", "",
		"-c", configFilePath,
		"-p", "65536", // 65535 is max
	})

	assert.Contains(t, output, "invalid argument")
	assert.Contains(t, output, "65536")
	assert.Contains(t, output, "value out of range")
}

func TestPortFlagWrongEnvValue(t *testing.T) {
	assert.NoError(t, os.Setenv("LISTEN_PORT", "65536")) // 65535 is max

	defer func() { assert.NoError(t, os.Unsetenv("LISTEN_PORT")) }()

	output := executeCommandWithoutRunning(t, []string{
		"-r", "",
		"-c", configFilePath,
		"-p", "8090", // `-p` flag must be ignored
	})

	assert.Contains(t, output, "wrong TCP port")
	assert.Contains(t, output, "environment variable")
	assert.Contains(t, output, "65536")
}

func TestResourcesDirFlagWrongArgument(t *testing.T) {
	output := executeCommandWithoutRunning(t, []string{
		"-r", "/tmp/nonexistent/bar/baz",
		"-c", configFilePath,
	})

	assert.Contains(t, output, "wrong resources directory")
	assert.Contains(t, output, "/tmp/nonexistent/bar/baz")
}

func TestResourcesDirFlagWrongEnvValue(t *testing.T) {
	assert.NoError(t, os.Setenv("RESOURCES_DIR", "/tmp/nonexistent/bar/baz"))

	defer func() { assert.NoError(t, os.Unsetenv("RESOURCES_DIR")) }()

	output := executeCommandWithoutRunning(t, []string{
		"-c", configFilePath,
		"-r", ".", // `-r` flag must be ignored
	})

	assert.Contains(t, output, "wrong resources directory")
	assert.Contains(t, output, "/tmp/nonexistent/bar/baz")
}

func TestCachingEngineFlagWrongArgument(t *testing.T) {
	output := executeCommandWithoutRunning(t, []string{
		"-r", "",
		"-c", configFilePath,
		"--caching-engine", "foobarEngine",
	})

	assert.Contains(t, output, "unsupported caching engine")
	assert.Contains(t, output, "foobarEngine")
}

func TestCachingEngineFlagWrongEnvValue(t *testing.T) {
	assert.NoError(t, os.Setenv("CACHING_ENGINE", "barEngine"))

	defer func() { assert.NoError(t, os.Unsetenv("CACHING_ENGINE")) }()

	output := executeCommandWithoutRunning(t, []string{
		"-r", "",
		"-c", configFilePath,
		"--caching-engine", "foobarEngine",
	})

	assert.Contains(t, output, "unsupported caching engine")
	assert.Contains(t, output, "barEngine")
}

func TestRedisDSNFlagWrongArgument(t *testing.T) {
	output := executeCommandWithoutRunning(t, []string{
		"-r", "",
		"-c", configFilePath,
		"--caching-engine", "redis",
		"--redis-dsn", "foo://bar",
	})

	assert.Contains(t, output, "wrong redis DSN")
	assert.Contains(t, output, "foo://bar")
}

func TestRedisDSNFlagWrongEnvValue(t *testing.T) {
	assert.NoError(t, os.Setenv("REDIS_DSN", "bar://baz"))

	defer func() { assert.NoError(t, os.Unsetenv("REDIS_DSN")) }()

	output := executeCommandWithoutRunning(t, []string{
		"-r", "",
		"-c", configFilePath,
		"--caching-engine", "redis",
		"--redis-dsn", "foo://bar", // `--redis-dsn` flag must be ignored
	})

	assert.Contains(t, output, "wrong redis DSN")
	assert.Contains(t, output, "bar://baz")
}

func TestConfigFlagWrongArgument(t *testing.T) {
	output := executeCommandWithoutRunning(t, []string{
		"-r", "",
		"-c", "/tmp/nonexistent/bar.baz",
	})

	assert.Contains(t, output, "config file")
	assert.Contains(t, output, "/tmp/nonexistent/bar.baz")
	assert.Contains(t, output, "not found")
}

func TestConfigFlagWrongEnvValue(t *testing.T) {
	assert.NoError(t, os.Setenv("CONFIG_PATH", "/tmp/nonexistent/foo.baz"))

	defer func() { assert.NoError(t, os.Unsetenv("CONFIG_PATH")) }()

	output := executeCommandWithoutRunning(t, []string{
		"-r", "",
		"-c", configFilePath, // `-c` flag must be ignored
	})

	assert.Contains(t, output, "config file")
	assert.Contains(t, output, "/tmp/nonexistent/foo.baz")
	assert.Contains(t, output, "not found")
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

func startAndStopServer(t *testing.T, port int, args []string) string {
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
			cmd.SetArgs(args)

			assert.NoError(t, cmd.Execute())
		})

		ch <- struct{}{}
	}(executedCh)

	portBusyCh := make(chan struct{})

	// check port "busy" (by HTTP server) state
	go func(ch chan<- struct{}) {
		defer close(ch)

		for i := 0; i < 2000; i++ {
			if checkTCPPortIsBusy(t, port) {
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

	return output
}

func TestSuccessfulCommandRunningUsingRedisCacheEngine(t *testing.T) {
	// get TCP port number for a test
	port, err := getRandomTCPPort(t)
	assert.NoError(t, err)

	// start mini-redis
	mini, err := miniredis.Run()
	assert.NoError(t, err)

	defer mini.Close()

	output := startAndStopServer(t, port, []string{
		"-r", "",
		"--port", strconv.Itoa(port),
		"-c", configFilePath,
		"--caching-engine", "redis",
		"--redis-dsn", fmt.Sprintf("redis://127.0.0.1:%s/0", mini.Port()),
	})

	assert.Contains(t, output, "Server starting")
	assert.Contains(t, output, "Stopping by OS signal")
	assert.Contains(t, output, "Server stopping")
}

func TestSuccessfulCommandRunningUsingDefaultCacheEngine(t *testing.T) {
	// get TCP port number for a test
	port, err := getRandomTCPPort(t)
	assert.NoError(t, err)

	output := startAndStopServer(t, port, []string{
		"-r", "",
		"--port", strconv.Itoa(port),
		"-c", configFilePath,
	})

	assert.Contains(t, output, "Server starting")
	assert.Contains(t, output, "Stopping by OS signal")
	assert.Contains(t, output, "Server stopping")
}

func TestRunningUsingBusyPortFailing(t *testing.T) {
	port, err := getRandomTCPPort(t)
	assert.NoError(t, err)

	// occupy a TCP port
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	assert.NoError(t, err)

	defer func() { assert.NoError(t, l.Close()) }()

	// create command with valid flags to run
	cmd := NewCommand(context.Background(), zap.NewNop())
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"-r", "", "--port", strconv.Itoa(port), "-c", configFilePath})

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
