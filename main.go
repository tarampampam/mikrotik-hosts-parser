package main

import (
	"log"
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
	options := NewOptions(stdLog, errLog, func(code int) {
		os.Exit(code)
	})

	// Parse options and make all checks
	options.Parse()

	server := NewServer(&HttpServerSettings{
		Host:             options.Address,
		Port:             options.Port,
		PublicDir:        options.ResourcesDir,
		IndexFile:        options.IndexFileName,
		Error404File:     options.Error404FileName,
		WriteTimeout:     time.Second * 15,
		ReadTimeout:      time.Second * 15,
		KeepAliveEnabled: false,
	})

	server.RegisterHandlers()

	server.Start()
}
