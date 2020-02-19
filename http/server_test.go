package http

import (
	"log"
	"mikrotik-hosts-parser/settings/serve"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	t.Parallel()

	t.Skip("Not implemented yet") // @todo: implement

	settings := ServerSettings{
		WriteTimeout:     10 * time.Second,
		ReadTimeout:      13 * time.Second,
		KeepAliveEnabled: false,
	}

	server := NewServer(&settings, &serve.Settings{})

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
}
