package controllers

import (
	"net/http"
	"strings"
)

// UnindexedFileSystem is an implementation of a standard http.FileSystem
// without the ability to list files in the directory.
// This implementation is largely inspired by
// https://www.alexedwards.net/blog/disable-http-fileserver-directory-listings
type UnindexedFileSystem struct {
	fs http.FileSystem
}

// Open returns a file from the static directory. If the requested path ends
// with a slash, there is a check for an index.html file. If none exists, then
// an error is returned.
func (ufs UnindexedFileSystem) Open(name string) (http.File, error) {
	f, err := ufs.fs.Open(name)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := strings.TrimSuffix(name, "/") + "/index.html"
		indexFile, err := ufs.fs.Open(index)
		if err != nil {
			return nil, err
		}
		return indexFile, nil
	}
	return f, nil
}
