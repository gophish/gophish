package models

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"gopkg.in/check.v1"
)

func (s *ModelsSuite) TestPostPage(c *check.C) {
	html := `<html>
			<head></head>
			<body><form action="example.com">
				<input name="username"/>
				<input name="password" type="password"/>
			</form></body>
		  </html>`
	p := Page{
		Name:        "Test Page",
		HTML:        html,
		RedirectURL: "http://example.com",
	}
	// Check the capturing credentials and passwords
	p.CaptureCredentials = true
	p.CapturePasswords = true
	err := PostPage(&p)
	c.Assert(err, check.Equals, nil)
	c.Assert(p.RedirectURL, check.Equals, "http://example.com")
	d, err := goquery.NewDocumentFromReader(strings.NewReader(p.HTML))
	c.Assert(err, check.Equals, nil)
	forms := d.Find("form")
	forms.Each(func(i int, f *goquery.Selection) {
		// Check the action has been set
		a, _ := f.Attr("action")
		c.Assert(a, check.Equals, "")
		// Check the password still has a name
		_, ok := f.Find("input[type=\"password\"]").Attr("name")
		c.Assert(ok, check.Equals, true)
		// Check the username is still correct
		u, ok := f.Find("input").Attr("name")
		c.Assert(ok, check.Equals, true)
		c.Assert(u, check.Equals, "username")
	})

	// Check what happens when we don't capture passwords
	p.CapturePasswords = false
	p.HTML = html
	p.RedirectURL = ""
	err = PutPage(&p)
	c.Assert(err, check.Equals, nil)
	c.Assert(p.RedirectURL, check.Equals, "")
	d, err = goquery.NewDocumentFromReader(strings.NewReader(p.HTML))
	c.Assert(err, check.Equals, nil)
	forms = d.Find("form")
	forms.Each(func(i int, f *goquery.Selection) {
		// Check the action has been set
		a, _ := f.Attr("action")
		c.Assert(a, check.Equals, "")
		// Check the password name has been removed
		_, ok := f.Find("input[type=\"password\"]").Attr("name")
		c.Assert(ok, check.Equals, false)
		// Check the username is still correct
		u, ok := f.Find("input").Attr("name")
		c.Assert(ok, check.Equals, true)
		c.Assert(u, check.Equals, "username")
	})

	// Check when we don't capture credentials
	p.CaptureCredentials = false
	p.HTML = html
	err = PutPage(&p)
	c.Assert(err, check.Equals, nil)
	d, err = goquery.NewDocumentFromReader(strings.NewReader(p.HTML))
	c.Assert(err, check.Equals, nil)
	forms = d.Find("form")
	forms.Each(func(i int, f *goquery.Selection) {
		// Check the action has been set
		a, _ := f.Attr("action")
		c.Assert(a, check.Equals, "")
		// Check the password name has been removed
		_, ok := f.Find("input[type=\"password\"]").Attr("name")
		c.Assert(ok, check.Equals, false)
		// Check the username name has been removed
		_, ok = f.Find("input").Attr("name")
		c.Assert(ok, check.Equals, false)
	})

	// Finally, re-enable capturing passwords (ref: #1267)
	p.CaptureCredentials = true
	p.CapturePasswords = true
	err = PutPage(&p)
	c.Assert(err, check.Equals, nil)
	d, err = goquery.NewDocumentFromReader(strings.NewReader(p.HTML))
	c.Assert(err, check.Equals, nil)
	forms = d.Find("form")
	forms.Each(func(i int, f *goquery.Selection) {
		// Check the password still has a name
		_, ok := f.Find("input[type=\"password\"]").Attr("name")
		c.Assert(ok, check.Equals, true)
	})
}

func (s *ModelsSuite) TestPageValidation(c *check.C) {
	html := `<html>
			<head></head>
			<body>{{.BaseURL}}</body>
		  </html>`
	p := Page{
		HTML:        html,
		RedirectURL: "http://example.com",
	}
	// Validate that a name is required
	err := p.Validate()
	c.Assert(err, check.Equals, ErrPageNameNotSpecified)

	p.Name = "Test Page"

	// Validate that CaptureCredentials is automatically set if somehow the
	// user fails to set it, but does indicate that passwords should be
	// captured
	p.CapturePasswords = true
	c.Assert(p.CaptureCredentials, check.Equals, false)
	err = p.Validate()
	c.Assert(err, check.Equals, nil)
	c.Assert(p.CaptureCredentials, check.Equals, true)

	// Validate that if the HTML contains an invalid template tag, that we
	// catch it
	p.HTML = `<html>
		<head></head>
		<body>{{.INVALIDTAG}}</body>
	  </html>`
	err = p.Validate()
	c.Assert(err, check.NotNil)

	// Validate that if the RedirectURL contains an invalid template tag, that
	// we catch it
	p.HTML = "valid data"
	p.RedirectURL = "http://example.com/{{.INVALIDTAG}}"
	err = p.Validate()
	c.Assert(err, check.NotNil)
}
