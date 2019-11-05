package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestHttpFileServer_ServeHTTP(t *testing.T) {
	t.Parallel()

	// Create directory in temporary
	createTempDir := func() string {
		t.Helper()
		if dir, err := ioutil.TempDir("", "test-"); err != nil {
			panic(err)
		} else {
			return dir
		}
	}

	tests := []struct {
		name                string
		giveDirs            []string
		giveFiles           map[string][]byte
		giveResources       map[string][]byte
		giveNotFoundHandler HttpFileNotFoundHandler
		giveIndexFile       string
		giveResourcesPrefix string
		giveError404file    string
		giveRequestUri      string
		giveRequestMethod   string
		wantResponseCode    int
		wantResponseBody    []byte
		wantContentType     string
		wantRedirectTo      string
	}{
		{
			name: "Static TEXT file serving from local FS",
			giveFiles: map[string][]byte{
				"test1.txt": []byte("test content"),
			},
			giveRequestUri:    "/test1.txt",
			giveRequestMethod: "GET",
			wantResponseCode:  http.StatusOK,
			wantResponseBody:  []byte("test content"),
			wantContentType:   "text/plain; charset=utf-8",
		},
		{
			name: "Static HTML file serving from local FS",
			giveFiles: map[string][]byte{
				"test1.html": []byte("<html>test content</html>"),
			},
			giveRequestUri:    "/test1.html",
			giveRequestMethod: "GET",
			wantResponseCode:  http.StatusOK,
			wantResponseBody:  []byte("<html>test content</html>"),
			wantContentType:   "text/html; charset=utf-8",
		},
		{
			name: "Static TEXT file serving from resources",
			giveResources: map[string][]byte{
				"/test1.txt": []byte("test content"),
			},
			giveRequestUri:    "/test1.txt",
			giveRequestMethod: "GET",
			wantResponseCode:  http.StatusOK,
			wantResponseBody:  []byte("test content"),
			wantContentType:   "text/plain; charset=utf-8",
		},
		{
			name: "Static HTML file serving from resources",
			giveResources: map[string][]byte{
				"/test1.html": []byte("<html>test content</html>"),
			},
			giveRequestUri:    "/test1.html",
			giveRequestMethod: "GET",
			wantResponseCode:  http.StatusOK,
			wantResponseBody:  []byte("<html>test content</html>"),
			wantContentType:   "text/html; charset=utf-8",
		},
		{
			name: "File on local FS have priority ABOVE file from resource",
			giveFiles: map[string][]byte{
				"test1.txt": []byte("from file"),
			},
			giveResources: map[string][]byte{
				"/test1.txt": []byte("from resource"),
			},
			giveRequestUri:    "/test1.txt",
			giveRequestMethod: "GET",
			wantResponseCode:  http.StatusOK,
			wantResponseBody:  []byte("from file"),
			wantContentType:   "text/plain; charset=utf-8",
		},
		{
			name:              "Redirect from .../index.html to .../",
			giveIndexFile:     "indx.html",
			giveRequestUri:    "/indx.html",
			giveRequestMethod: "GET",
			wantResponseCode:  http.StatusMovedPermanently,
			wantRedirectTo:    "/",
		},
		{
			name:              "Redirect from .../index.html to .../ insime some directory",
			giveIndexFile:     "indx.html",
			giveRequestUri:    "/some/indx.html",
			giveRequestMethod: "GET",
			wantResponseCode:  http.StatusMovedPermanently,
			wantRedirectTo:    "/some/",
		},
		{
			name: "Request root",
			giveFiles: map[string][]byte{
				"indx.html": []byte("test content"),
			},
			giveIndexFile:     "indx.html",
			giveRequestUri:    "",
			giveRequestMethod: "GET",
			wantResponseBody:  []byte("test content"),
			wantResponseCode:  http.StatusOK,
			wantContentType:   "text/html; charset=utf-8",
		},
		{
			name:     "Index file from some directory",
			giveDirs: []string{"foo"},
			giveFiles: map[string][]byte{
				"indx.html":                       []byte("index in root"),
				filepath.Join("foo", "indx.html"): []byte("index in foo"),
			},
			giveIndexFile:     "indx.html",
			giveRequestUri:    "/foo/",
			giveRequestMethod: "GET",
			wantResponseBody:  []byte("index in foo"),
			wantResponseCode:  http.StatusOK,
			wantContentType:   "text/html; charset=utf-8",
		},
		{
			name:     "404 on directory request",
			giveDirs: []string{"foo"},
			giveFiles: map[string][]byte{
				"indx.html":                       []byte("index in root"),
				filepath.Join("foo", "indx.html"): []byte("index in foo"),
			},
			giveIndexFile:     "indx.html",
			giveRequestUri:    "/foo",
			giveRequestMethod: "GET",
			wantResponseCode:  http.StatusNotFound,
		},
		{
			name:              "NotFoundHandler handling",
			giveIndexFile:     "indx.html",
			giveRequestUri:    "/foo",
			giveRequestMethod: "GET",
			giveNotFoundHandler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(444)
				w.Write([]byte("foo bar"))
				w.Header().Set("Content-Type", "blah blah")
			},
			wantResponseCode: 444,
			wantResponseBody: []byte("foo bar"),
			wantContentType:  "blah blah",
		},
		{
			name: "Error 404 file serving from local FS",
			giveFiles: map[string][]byte{
				"404.html": []byte("error 404 file"),
			},
			giveRequestUri:    "/foo",
			giveError404file:  "404.html",
			giveRequestMethod: "GET",
			wantResponseCode:  http.StatusNotFound,
			wantResponseBody:  []byte("error 404 file"),
			wantContentType:   "text/html; charset=utf-8",
		},
		{
			name: "Error 404 file serving from resources",
			giveResources: map[string][]byte{
				"/404.html": []byte("error 404 resource"),
			},
			giveRequestUri:    "/foo",
			giveError404file:  "404.html",
			giveRequestMethod: "GET",
			wantResponseCode:  http.StatusNotFound,
			wantResponseBody:  []byte("error 404 resource"),
			wantContentType:   "text/html; charset=utf-8",
		},
		{
			name: "Error 404 file on local FS have priority ABOVE file from resource",
			giveFiles: map[string][]byte{
				"404.html": []byte("from file"),
			},
			giveResources: map[string][]byte{
				"/404.html": []byte("from resource"),
			},
			giveRequestUri:    "/foo",
			giveError404file:  "404.html",
			giveRequestMethod: "GET",
			wantResponseCode:  http.StatusNotFound,
			wantResponseBody:  []byte("from file"),
			wantContentType:   "text/html; charset=utf-8",
		},
		{
			name:              "Error 404 fail-back",
			giveRequestUri:    "/foo",
			giveError404file:  "404.html",
			giveRequestMethod: "GET",
			wantResponseCode:  http.StatusNotFound,
			wantResponseBody:  []byte("<html><body><h1>ERROR 404</h1><h2>Requested file was not found</h2></body></html>"),
			wantContentType:   "text/html; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var root http.Dir

			if len(tt.giveDirs) > 0 || len(tt.giveFiles) > 0 {
				tmpDir := createTempDir()
				root = http.Dir(tmpDir)

				defer func(d string) {
					if err := os.RemoveAll(d); err != nil {
						panic(err)
					}
				}(tmpDir)

				// Create directories
				for _, d := range tt.giveDirs {
					if err := os.Mkdir(filepath.Join(tmpDir, d), 0777); err != nil {
						panic(err)
					}
				}

				// Create files
				for name, content := range tt.giveFiles {
					if f, err := os.Create(filepath.Join(tmpDir, name)); err != nil {
						panic(err)
					} else {
						if _, err := f.Write(content); err != nil {
							panic(err)
						}
						if err := f.Close(); err != nil {
							panic(err)
						}
					}
				}
			} else {
				root = ""
			}

			resources := newResourceBox()

			// Create resources
			for name, content := range tt.giveResources {
				resources.Add(name, content)
			}

			fileServer := &HttpFileServer{
				root:            root,
				resources:       resources,
				NotFoundHandler: tt.giveNotFoundHandler,
				indexFile:       tt.giveIndexFile,
				resourcesPrefix: tt.giveResourcesPrefix,
				error404file:    tt.giveError404file,
			}

			var (
				req, _ = http.NewRequest(tt.giveRequestMethod, tt.giveRequestUri, nil)
				rr     = httptest.NewRecorder()
			)

			fileServer.ServeHTTP(rr, req)

			if rr.Code != tt.wantResponseCode {
				t.Errorf("Wrong response HTTP code. Want %d, got %d", tt.wantResponseCode, rr.Code)
			}

			if len(tt.wantResponseBody) > 0 && !reflect.DeepEqual(rr.Body.Bytes(), tt.wantResponseBody) {
				t.Errorf("Wrong HTTP response. Want [%s], got [%s]", tt.wantResponseBody, rr.Body.String())
			}

			if ct := rr.Header().Get("Content-Type"); tt.wantContentType != "" && ct != tt.wantContentType {
				t.Errorf("Wrong response content type header. Want %s, got %s", tt.wantContentType, ct)
			}

			if rt := rr.Header().Get("Location"); tt.wantRedirectTo != "" && tt.wantRedirectTo != rt {
				t.Errorf("Wrong redirect to location. Want %s, got %s", tt.wantRedirectTo, rt)
			}

			//t.Log(rr)
		})
	}
}
