package http

import (
	"log"
	"mikrotik-hosts-parser/settings/serve"
	"mime"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	t.Parallel()

	settings := ServerSettings{
		WriteTimeout:     10 * time.Second,
		ReadTimeout:      13 * time.Second,
		KeepAliveEnabled: false,
	}

	server := NewServer(&settings, &serve.Settings{
		Listen: serve.Listen{Address: "1.2.3.4", Port: 321},
	})

	if !reflect.DeepEqual(&settings, server.Settings) {
		t.Errorf("Wrong settings set. Expected: %v, got: %v", settings, server.Settings)
	}

	if server.stdLog.Writer() != os.Stdout {
		t.Error("Wrong 'stdLog' writer set")
	}

	if server.stdLog.Flags() != log.Ldate|log.Lmicroseconds {
		t.Error("Wrong 'stdLog' flags set")
	}

	if server.errLog.Flags() != log.LstdFlags {
		t.Error("Wrong 'errLog' flags set")
	}

	if server.Server.Addr != "1.2.3.4:321" {
		t.Errorf("Wrong HTTP server addr set. Want [%s], got [%s]", "1.2.3.4:321", server.Server.Addr)
	}

	if server.Server.WriteTimeout != 10*time.Second {
		t.Error("Wrong server write timeout value is set")
	}

	if server.Server.ReadTimeout != 13*time.Second {
		t.Error("Wrong server read timeout value is set")
	}
}

func Test_registerCustomMimeTypes(t *testing.T) {
	t.Parallel()

	testSliceContainsString := func(t *testing.T, slice []string, expects string) {
		t.Helper()

		for _, n := range slice {
			if expects == n {
				return
			}
		}

		t.Errorf("Slice %v does not contains %s", slice, expects)
	}

	testSliceNotContainsString := func(t *testing.T, slice []string, expects string) {
		t.Helper()

		for _, n := range slice {
			if expects == n {
				t.Errorf("Slice %v contains %s (but should not)", slice, expects)
			}
		}
	}

	types, _ := mime.ExtensionsByType("text/html; charset=utf-8")
	testSliceNotContainsString(t, types, ".vue")

	if err := NewServer(&ServerSettings{}, &serve.Settings{}).registerCustomMimeTypes(); err != nil {
		t.Error(err)
	}

	types, _ = mime.ExtensionsByType("text/html; charset=utf-8")
	testSliceContainsString(t, types, ".vue")
}

func TestServer_Start(t *testing.T) {
	t.Skip("Not implemented yet")
}

func TestServer_Stop(t *testing.T) {
	t.Skip("Not implemented yet")
}
