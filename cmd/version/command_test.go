package version

import (
	"bytes"
	"io"
	ver "mikrotik-hosts-parser/version"
	"os"
	"strings"
	"testing"
)

func TestCommand_Execute(t *testing.T) {
	t.Parallel()

	captureOutput := func(f func()) string {
		t.Helper()

		r, w, err := os.Pipe()
		if err != nil {
			panic(err)
		}

		stdout := os.Stdout
		os.Stdout = w
		defer func() {
			os.Stdout = stdout
		}()
		f()
		_ = w.Close()

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		return buf.String()
	}

	tests := []struct {
		name             string
		giveVersion      string
		giveArgs         []string
		wantOutput       []string
		wantErr          bool
		wantErrorMessage string
	}{
		{
			name:             "Without version set",
			giveVersion:      "",
			giveArgs:         []string{},
			wantOutput:       []string{"Version:", "\n"},
			wantErr:          false,
			wantErrorMessage: "",
		},
		{
			name:             "With version set",
			giveVersion:      "6.6.6.RC1",
			giveArgs:         []string{},
			wantOutput:       []string{"Version:", "6.6.6.RC1", "\n"},
			wantErr:          false,
			wantErrorMessage: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver.Version = tt.giveVersion
			var err error
			var cmd = Command{}

			output := captureOutput(func() {
				err = cmd.Execute(tt.giveArgs)
			})

			if tt.wantOutput != nil {
				for _, line := range tt.wantOutput {
					if !strings.Contains(output, line) {
						t.Errorf("Expected line [%s] in output [%s] was not found", line, output)
					}
				}
			}

			if tt.wantErr && err.Error() != tt.wantErrorMessage {
				t.Errorf("Expected error message [%s] was not found in %v", tt.wantErrorMessage, err)
			}
		})
	}
}
