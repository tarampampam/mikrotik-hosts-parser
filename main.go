package main

import (
	"log"
	"os"
)

const VERSION = "3.0.0" // Do not forget update this value before new version releasing

func main() {
	var (
		stdLog = log.New(os.Stderr, "", 0)
		errLog = log.New(os.Stderr, "", log.LstdFlags)
	)

	server := NewServer("0.0.0.0", 8080, "./public", stdLog, errLog)

	server.RegisterHandlers()

	server.Start()
}
