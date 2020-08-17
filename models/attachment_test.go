package models

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

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
		URL:         "http://testurl.com/?rid=1234567",
		TrackingURL: "http://testurl.local/track?rid=1234567",
		Tracker:     "<img alt='' style='display: none' src='http://testurl.local/track?rid=1234567'/>",
		From:        "From Address",
		RId:         "1234567",
	}

	files, err := ioutil.ReadDir("testdata")
	if err != nil {
		log.Fatalf("Failed to open attachment folder 'testdata': %v\n", err)
	}
	for _, ff := range files {
		if !ff.IsDir() {
			fname := ff.Name()
			fmt.Printf("Checking attachment file -> %s\n", fname)
			f, err := os.Open("testdata/" + fname)
			if err != nil {
				log.Fatalf("Failed to open attachment test file '%s': %v\n", fname, err)
			}
			reader := bufio.NewReader(f)
			content, err := ioutil.ReadAll(reader)
			if err != nil {
				log.Fatalf("Failed to read attachment test file '%s': %v\n", fname, err)
			}

			data := ""
			if filepath.Ext(fname) == ".b64" {
				data = string(content)
				fname = fname[:len(fname)-4]
			} else {
				data = base64.StdEncoding.EncodeToString(content)
			}

			a := Attachment{
				Content: data,
				Name:    fname,
			}

			_, err = a.ApplyTemplate(ptx)
			c.Assert(err, check.Equals, nil)

		}
	}

}
