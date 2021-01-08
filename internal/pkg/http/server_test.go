package http

import (
	"context"
	"mime"
	"testing"

	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/config"
	"go.uber.org/zap"
)

/*
func TestNewServer(t *testing.T) {
	settings := ServerSettings{
		WriteTimeout:     10 * time.Second,
		ReadTimeout:      13 * time.Second,
		KeepAliveEnabled: false,
	}

	server := NewServer(&settings, &settings2.Config{
		Listen: settings2.listen{Address: "1.2.3.4", Port: 321},
	})

	if !reflect.DeepEqual(&settings, server.Settings) {
		t.Errorf("Wrong config set. Expected: %v, got: %v", settings, server.Settings)
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

	if server.srv.Addr != "1.2.3.4:321" {
		t.Errorf("Wrong HTTP server addr set. Want [%s], got [%s]", "1.2.3.4:321", server.srv.Addr)
	}

	if server.srv.WriteTimeout != 10*time.Second {
		t.Error("Wrong server write timeout value is set")
	}

	if server.srv.ReadTimeout != 13*time.Second {
		t.Error("Wrong server read timeout value is set")
	}
}
*/

func Test_registerCustomMimeTypes(t *testing.T) {
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

	srv := NewServer(context.Background(), zap.NewNop(), "", ".", &config.Config{})

	if err := srv.RegisterCustomMimeTypes(); err != nil {
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
