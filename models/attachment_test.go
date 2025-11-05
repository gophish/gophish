package models

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/check.v1"
)

func (s *ModelsSuite) TestAttachment(c *check.C) {
	ptx := PhishingTemplateContext{
		BaseRecipient: BaseRecipient{
			FirstName: "Foo",
			LastName:  "Bar",
			Email:     "foo@bar.com",
			Position:  "Space Janitor",
		},
		BaseURL:     "http://testurl.com",
		URL:         "http://testurl.com/?keyname=1234567",
		TrackingURL: "http://testurl.local/track?keyname=1234567",
		Tracker:     "<img alt='' style='display: none' src='http://testurl.local/track?keyname=1234567'/>",
		From:        "From Address",
		RId:         "1234567",
	}

	files, err := ioutil.ReadDir("testdata")
	if err != nil {
		log.Fatalf("Failed to open attachment folder 'testdata': %v\n", err)
	}
	for _, ff := range files {
		if !ff.IsDir() && !strings.Contains(ff.Name(), "templated") {
			fname := ff.Name()
			fmt.Printf("Checking attachment file -> %s\n", fname)
			data := readFile("testdata/" + fname)
			if filepath.Ext(fname) == ".b64" {
				fname = fname[:len(fname)-4]
			}
			a := Attachment{
				Content: data,
				Name:    fname,
			}
			t, err := a.ApplyTemplate(ptx)
			c.Assert(err, check.Equals, nil)
			c.Assert(a.vanillaFile, check.Equals, strings.Contains(fname, "without-vars"))
			c.Assert(a.vanillaFile, check.Not(check.Equals), strings.Contains(fname, "with-vars"))

			// Verfify template was applied as expected
			tt, err := ioutil.ReadAll(t)
			if err != nil {
				log.Fatalf("Failed to parse templated file '%s': %v\n", fname, err)
			}
			templatedB64 := base64.StdEncoding.EncodeToString(tt)
			expectedB64 := readFile("testdata/" + strings.TrimSuffix(ff.Name(), filepath.Ext(ff.Name())) + ".templated" + filepath.Ext(ff.Name()))

			if filepath.Ext(fname) == ".ics" {
				gotRaw, err := base64.StdEncoding.DecodeString(templatedB64)
				c.Assert(err, check.IsNil)
				wantRaw, err := base64.StdEncoding.DecodeString(expectedB64)
				c.Assert(err, check.IsNil)
				c.Assert(normalizeICS(string(gotRaw)), check.Equals, normalizeICS(string(wantRaw)))
			} else {
				gotRaw, err := base64.StdEncoding.DecodeString(templatedB64)
				c.Assert(err, check.IsNil)
				wantRaw, err := base64.StdEncoding.DecodeString(expectedB64)
				c.Assert(err, check.IsNil)
				c.Assert(trimEOFNL(gotRaw), check.DeepEquals, trimEOFNL(wantRaw))
			}
		}
	}
}

func readFile(fname string) string {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatalf("Failed to open file '%s': %v\n", fname, err)
	}
	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatalf("Failed to read file '%s': %v\n", fname, err)
	}
	data := ""
	if filepath.Ext(fname) == ".b64" {
		data = string(content)
	} else {
		data = base64.StdEncoding.EncodeToString(content)
	}
	return data
}

var tzidRe = regexp.MustCompile(`(?m)^TZID:\s+`)

func normalizeICS(s string) string {
	// RFC 5545 allows folded lines indicated by CRLF followed by space or tab.
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n ", "")
	s = strings.ReplaceAll(s, "\n\t", "")
	return tzidRe.ReplaceAllString(s, "TZID:")
}

func trimEOFNL(b []byte) []byte {
	if bytes.HasSuffix(b, []byte("\r\n")) {
		return b[:len(b)-2]
	}
	if len(b) > 0 && b[len(b)-1] == '\n' {
		return b[:len(b)-1]
	}
	return b
}
