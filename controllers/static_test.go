package controllers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

var fileContent = []byte("Hello world")

func mustRemoveAll(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		panic(err)
	}
}

func createTestFile(dir, filename string) error {
	return ioutil.WriteFile(filepath.Join(dir, filename), fileContent, 0644)
}

func (s *ControllersSuite) TestGetStaticFile() {
	dir, err := ioutil.TempDir("static/endpoint", "test-")
	tempFolder := filepath.Base(dir)

	s.Nil(err)
	defer mustRemoveAll(dir)

	err = createTestFile(dir, "foo.txt")
	s.Nil(nil, err)

	resp, err := http.Get(fmt.Sprintf("%s/static/%s/foo.txt", ps.URL, tempFolder))
	s.Nil(err)

	defer resp.Body.Close()
	got, err := ioutil.ReadAll(resp.Body)
	s.Nil(err)

	s.Equal(bytes.Compare(fileContent, got), 0, fmt.Sprintf("Got %s", got))
}

func (s *ControllersSuite) TestStaticFileListing() {
	dir, err := ioutil.TempDir("static/endpoint", "test-")
	tempFolder := filepath.Base(dir)

	s.Nil(err)
	defer mustRemoveAll(dir)

	err = createTestFile(dir, "foo.txt")
	s.Nil(nil, err)

	resp, err := http.Get(fmt.Sprintf("%s/static/%s/", ps.URL, tempFolder))
	s.Nil(err)

	defer resp.Body.Close()
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNotFound)
}

func (s *ControllersSuite) TestStaticIndex() {
	dir, err := ioutil.TempDir("static/endpoint", "test-")
	tempFolder := filepath.Base(dir)

	s.Nil(err)
	defer mustRemoveAll(dir)

	err = createTestFile(dir, "index.html")
	s.Nil(nil, err)

	resp, err := http.Get(fmt.Sprintf("%s/static/%s/", ps.URL, tempFolder))
	s.Nil(err)

	defer resp.Body.Close()
	got, err := ioutil.ReadAll(resp.Body)
	s.Nil(err)

	s.Equal(bytes.Compare(fileContent, got), 0, fmt.Sprintf("Got %s", got))
}
