package email

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

type staticFilesFile struct {
	data  string
	mime  string
	mtime time.Time
	// size is the size before compression. If 0, it means the data is uncompressed
	size int
	// hash is a sha256 hash of the file contents. Used for the Etag, and useful for caching
	hash string
}

var staticFiles = map[string]*staticFilesFile{
	"newpassword.html": {
		data:  "<h1>Seems you dropped something...</h1>\n<p>No worries pick up your new password <a href=\"{{.Link}}\">here</a>.</p>",
		hash:  "38838cf9c3fad0861817912e7ed4ef23b693a4c832ba98055ac91c39b0fc024b",
		mime:  "text/html; charset=utf-8",
		mtime: time.Unix(1496283935, 0),
		size:  0,
	},
	"test.html": {
		data:  "<h1>\n\tHey {{.NameOne}}\n</h1>\n<h1>\n\tHey {{.NameTwo}}\n</h1>",
		hash:  "87b1a1ef3ef8b8ea69ed2f3d54ed24bb6d37d4f73d9863c999e98eefcd379e93",
		mime:  "text/html; charset=utf-8",
		mtime: time.Unix(1496282690, 0),
		size:  0,
	},
	"verify.html": {
		data:  "<h1>Verify Email</h1>\n<p>Please click on <a href=\"{{.Link}}\">here</a> to verify your email.</p>",
		hash:  "4ad49cdd22218d0ea21ecbdcd78816b72e816ce48513dfa0fa156daf4ac7b992",
		mime:  "text/html; charset=utf-8",
		mtime: time.Unix(1496282212, 0),
		size:  0,
	},
}

// NotFound is called when no asset is found.
// It defaults to http.NotFound but can be overwritten
var NotFound = http.NotFound

// ServeHTTP serves a request, attempting to reply with an embedded file.
func ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	f, ok := staticFiles[strings.TrimPrefix(req.URL.Path, "/")]
	if !ok {
		NotFound(rw, req)
		return
	}
	header := rw.Header()
	if f.hash != "" {
		if hash := req.Header.Get("If-None-Match"); hash == f.hash {
			rw.WriteHeader(http.StatusNotModified)
			return
		}
		header.Set("ETag", f.hash)
	}
	if !f.mtime.IsZero() {
		if t, err := time.Parse(http.TimeFormat, req.Header.Get("If-Modified-Since")); err == nil && f.mtime.Before(t.Add(1*time.Second)) {
			rw.WriteHeader(http.StatusNotModified)
			return
		}
		header.Set("Last-Modified", f.mtime.UTC().Format(http.TimeFormat))
	}
	header.Set("Content-Type", f.mime)

	// Check if the asset is compressed in the binary
	if f.size == 0 {
		header.Set("Content-Length", strconv.Itoa(len(f.data)))
		io.WriteString(rw, f.data)
	} else {
		if header.Get("Content-Encoding") == "" && strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			header.Set("Content-Encoding", "gzip")
			header.Set("Content-Length", strconv.Itoa(len(f.data)))
			io.WriteString(rw, f.data)
		} else {
			header.Set("Content-Length", strconv.Itoa(f.size))
			reader, _ := gzip.NewReader(strings.NewReader(f.data))
			io.Copy(rw, reader)
			reader.Close()
		}
	}
}

// Server is simply ServeHTTP but wrapped in http.HandlerFunc so it can be passed into net/http functions directly.
var Server http.Handler = http.HandlerFunc(ServeHTTP)

// Open allows you to read an embedded file directly. It will return a decompressing Reader if the file is embedded in compressed format.
// You should close the Reader after you're done with it.
func Open(name string) (io.ReadCloser, error) {
	f, ok := staticFiles[name]
	if !ok {
		return nil, fmt.Errorf("Asset %s not found", name)
	}

	if f.size == 0 {
		return ioutil.NopCloser(strings.NewReader(f.data)), nil
	} else {
		return gzip.NewReader(strings.NewReader(f.data))
	}
}

// ModTime returns the modification time of the original file.
// Useful for caching purposes
// Returns zero time if the file is not in the bundle
func ModTime(file string) (t time.Time) {
	if f, ok := staticFiles[file]; ok {
		t = f.mtime
	}
	return
}

// Hash returns the hex-encoded SHA256 hash of the original file
// Used for the Etag, and useful for caching
// Returns an empty string if the file is not in the bundle
func Hash(file string) (s string) {
	if f, ok := staticFiles[file]; ok {
		s = f.hash
	}
	return
}

// Slower than Open as it must cycle through every element in map. Open all files that match glob.
func OpenGlob(name string) ([]io.ReadCloser, error) {
	readers := make([]io.ReadCloser, 0)
	for file := range staticFiles {
		matches, err := path.Match(name, file)
		if err != nil {
			continue
		}
		if matches {
			reader, err := Open(file)
			if err == nil && reader != nil {
				readers = append(readers, reader)
			}
		}
	}
	if len(readers) == 0 {
		return nil, fmt.Errorf("No assets found that match.")
	}
	return readers, nil
}
