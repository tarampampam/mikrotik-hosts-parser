package serve

import (
	"bufio"
	"bytes"
	"net"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/kami-zh/go-capturer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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
	// create logger instance with output capturing
	var (
		logBuf  bytes.Buffer // FIXME sometimes test fails (race detector)
		encoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		writer  = bufio.NewWriter(&logBuf)
		log     = zap.New(zapcore.NewCore(encoder, zapcore.AddSync(writer), zapcore.DebugLevel))
	)

	// get TCP port number for a test
	tcpPort, err := getRandomTCPPort(t)
	assert.NoError(t, err)

	// create command with valid flags to run
	cmd := NewCommand(log)
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"-r", "", "--port", strconv.Itoa(tcpPort), "-c", configFilePath})
	var output string

	executedCh := make(chan struct{})

	// start HTTP server
	go func(ch chan<- struct{}) {
		defer close(ch)

		output = capturer.CaptureOutput(func() {
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

	// flush the logger buffer
	assert.NoError(t, writer.Flush())
	logged := logBuf.String()

	assert.Empty(t, output) // there is no output, all must be inside logger buffer

	// log asserts is a very bed practice, but i have no idea how to test command execution better
	assert.Contains(t, logged, "Server starting")
	assert.Contains(t, logged, "Stopping by OS signal")
	assert.Contains(t, logged, "Server stopping")
}
