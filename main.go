package main

import "time"

const VERSION = "3.0.0" // Do not forget update this value before new version releasing

func main() {
	server := NewServer(&HttpServerSettings{
		Host:             "0.0.0.0",
		Port:             8080,
		PublicDir:        "./resources/public",
		IndexFile:        "index.html",
		Error404File:     "404.html",
		WriteTimeout:     time.Second * 15,
		ReadTimeout:      time.Second * 15,
		KeepAliveEnabled: false,
	})

	server.RegisterHandlers()

	server.Start()
}
