package main

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type HttpFileServer struct {
	root                     http.Dir
	resources                ResourcesBox
	NotFoundHandler          func(http.ResponseWriter, *http.Request)
	DirectoryListingDisabled func(http.ResponseWriter, *http.Request)
}

// Serve requests to the "public" files and directories.
func (fileServer *HttpFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const (
		indexFileName       string = "index.html"
		resourcesPathPrefix string = "/public"
	)

	// if empty, set current directory
	dir := string(fileServer.root)
	if dir == "" {
		dir = "."
	}

	// add prefix and clean
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	upath = path.Clean(upath)

	// path to file
	name := path.Join(dir, filepath.FromSlash(upath))

	if fileServer.NotFoundHandler != nil {
		// check if file exists
		f, err := os.Open(name)
		if err != nil {
			if os.IsNotExist(err) {
				fileServer.NotFoundHandler(w, r)
				return
			}
		}
		defer func() {
			if err := f.Close(); err != nil {
				panic(err)
			}
		}()
	}

	// check if requested directory (not regular file)
	if info, err := os.Stat(name); err == nil {
		if info.Mode().IsDir() {
			indexFilePath := path.Join(name, indexFileName)
			if _, err := os.Stat(indexFilePath); err == nil {
				// index file exists - "rewrite" requested name
				name = indexFilePath
			} else if os.IsNotExist(err) {
				// index file does not exists - check for handler
				if fileServer.DirectoryListingDisabled != nil {
					fileServer.DirectoryListingDisabled(w, r)
					return
				}
			}
		}
	}

	http.ServeFile(w, r, name)
}
