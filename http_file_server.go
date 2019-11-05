package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type (
	HttpFileNotFoundHandler func(http.ResponseWriter, *http.Request)

	HttpFileServer struct {
		root            http.Dir
		resources       ResourcesBox            // optionally, but strongly recommended
		NotFoundHandler HttpFileNotFoundHandler // optionally
		indexFile       string
		resourcesPrefix string
		error404file    string
	}
)

// Serve requests to the "public" files and directories.
func (fileServer *HttpFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// redirect .../index.html to .../
	if strings.HasSuffix(r.URL.Path, "/"+fileServer.indexFile) {
		http.Redirect(w, r, r.URL.Path[0:len(r.URL.Path)-len(fileServer.indexFile)], http.StatusMovedPermanently)
		return
	}

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
	// add index file name if requested directory (or server root)
	if upath[len(upath)-1] == '/' {
		upath += fileServer.indexFile
	}
	// make path clean
	upath = path.Clean(upath)

	// path to file
	name := path.Join(dir, filepath.FromSlash(upath))

	// if files server root directory is set - try to find file and serve them
	if len(fileServer.root) > 0 {
		// check for file exists
		if f, err := os.Open(name); err == nil {
			// file exists and opened
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()
			// file (or directory) exists
			if stat, statErr := os.Stat(name); statErr == nil && stat.Mode().IsRegular() {
				// requested file is file (not directory)
				var modTime time.Time
				// Try to extract file modified time
				if info, err := f.Stat(); err == nil {
					modTime = info.ModTime()
				} else {
					modTime = time.Now() // fail-back
				}
				// serve fie content
				http.ServeContent(
					w,
					r,
					filepath.Base(upath),
					modTime,
					f,
				)
				return
			}
		}
	}

	// requested file exists in resources
	if fileServer.resources != nil {
		if content, ok := fileServer.resources.Get(fileServer.resourcesPrefix + upath); ok {
			http.ServeContent(
				w,
				r,
				filepath.Base(upath),
				time.Now(), // @todo: set build time, not time.Now()
				bytes.NewReader(content),
			)
			return
		}
	}

	// If all tries for content serving above has been failed - file was not found (HTTP 404)
	if fileServer.NotFoundHandler != nil {
		// If "file not found" handler is set - call them
		fileServer.NotFoundHandler(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusNotFound)

	// at first - we try to find local file with error content
	if len(fileServer.root) > 0 {
		var errPage = string(fileServer.root) + "/" + fileServer.error404file
		if f, err := os.Open(errPage); err == nil {
			// file exists and opened
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()
			// file (or directory) exists
			if stat, statErr := os.Stat(errPage); statErr == nil && stat.Mode().IsRegular() {
				// requested file is file (not directory)
				if _, writeErr := io.Copy(w, f); writeErr != nil {
					panic(writeErr)
				}
				return
			}
		}
	}

	if fileServer.resources != nil {
		// if local file was not found - try to extract data from resources
		if content, ok := fileServer.resources.Get(fileServer.resourcesPrefix + "/" + fileServer.error404file); ok {
			// write content into response
			if _, writeErr := w.Write(content); writeErr != nil {
				panic(writeErr)
			}
			return
		}
	}

	// fail-back
	if _, err := fmt.Fprint(w, "<html><body><h1>ERROR 404</h1><h2>Requested file was not found</h2></body></html>"); err != nil {
		panic(err)
	}
}
