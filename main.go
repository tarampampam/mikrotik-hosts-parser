package main

import (
	"log"
	"mikrotik-hosts-parser/options"
	"os"
	"time"
)

const VERSION = "3.0.0" // Do not forget update this value before new version releasing

func main() {
	var (
		stdLog = log.New(os.Stderr, "", 0)
		errLog = log.New(os.Stderr, "", log.LstdFlags)
	)

	// Precess CLI options
	opts := options.NewOptions(stdLog, errLog, func(code int) {
		os.Exit(code)
	})

	opts.Version = VERSION

	// Parse options and make all checks
	opts.Parse()

	server := NewServer(&HttpServerSettings{
		Host:             opts.Address,
		Port:             opts.Port,
		PublicDir:        opts.ResourcesDir,
		IndexFile:        opts.IndexFileName,
		Error404File:     opts.Error404FileName,
		WriteTimeout:     time.Second * 15,
		ReadTimeout:      time.Second * 15,
		KeepAliveEnabled: false,
	})

	server.RegisterHandlers()

	server.Start()
}
