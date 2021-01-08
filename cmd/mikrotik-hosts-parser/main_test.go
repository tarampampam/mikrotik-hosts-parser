package main

import (
	"os"
	"testing"

	"github.com/kami-zh/go-capturer"
	"github.com/stretchr/testify/assert"
)

func Test_Main(t *testing.T) {
	origFlags := make([]string, 0)
	origFlags = append(origFlags, os.Args...)

	defer func() { os.Args = origFlags }()

	os.Args = []string{"", "--help"}

	output := capturer.CaptureStdout(func() {
		main()
	})

	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "Available Commands:")
	assert.Contains(t, output, "Flags:")
}
