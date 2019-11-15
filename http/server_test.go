package http

import (
	"log"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	t.Parallel()

	settings := ServerSettings{
		Host:             "1.2.3.4",
		Port:             321,
		PublicDir:        "/tmp/foo",
		IndexFile:        "idx.html",
		Error404File:     "err404.asp",
		WriteTimeout:     10 * time.Second,
		ReadTimeout:      13 * time.Second,
		KeepAliveEnabled: false,
	}

	server := NewServer(&settings)

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
