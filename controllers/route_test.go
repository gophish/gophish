package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (s *ControllersSuite) TestLoginCSRF() {
	resp, err := http.PostForm(fmt.Sprintf("%s/login", s.adminServer.URL),
		url.Values{
			"username": {"admin"},
			"password": {"gophish"},
		})

	s.Equal(resp.StatusCode, http.StatusForbidden)
	fmt.Println(err)
}

func (s *ControllersSuite) TestInvalidCredentials() {
	resp, err := http.Get(fmt.Sprintf("%s/login", s.adminServer.URL))
	s.Equal(err, nil)
	s.Equal(resp.StatusCode, http.StatusOK)

	doc, err := goquery.NewDocumentFromResponse(resp)
	s.Equal(err, nil)
	elem := doc.Find("input[name='csrf_token']").First()
	token, ok := elem.Attr("value")
	s.Equal(ok, true)

	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/login", s.adminServer.URL), strings.NewReader(url.Values{
		"username":   {"admin"},
		"password":   {"invalid"},
		"csrf_token": {token},
	}.Encode()))
	s.Equal(err, nil)

	req.Header.Set("Cookie", resp.Header.Get("Set-Cookie"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err = client.Do(req)
	s.Equal(err, nil)
	s.Equal(resp.StatusCode, http.StatusUnauthorized)
}

func (s *ControllersSuite) TestSuccessfulLogin() {
	resp, err := http.Get(fmt.Sprintf("%s/login", s.adminServer.URL))
	s.Equal(err, nil)
	s.Equal(resp.StatusCode, http.StatusOK)

	doc, err := goquery.NewDocumentFromResponse(resp)
	s.Equal(err, nil)
	elem := doc.Find("input[name='csrf_token']").First()
	token, ok := elem.Attr("value")
	s.Equal(ok, true)

	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/login", s.adminServer.URL), strings.NewReader(url.Values{
		"username":   {"admin"},
		"password":   {"gophish"},
		"csrf_token": {token},
	}.Encode()))
	s.Equal(err, nil)

	req.Header.Set("Cookie", resp.Header.Get("Set-Cookie"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err = client.Do(req)
	s.Equal(err, nil)
	s.Equal(resp.StatusCode, http.StatusOK)
}

func (s *ControllersSuite) TestSuccessfulRedirect() {
	next := "/campaigns"
	resp, err := http.Get(fmt.Sprintf("%s/login", s.adminServer.URL))
	s.Equal(err, nil)
	s.Equal(resp.StatusCode, http.StatusOK)

	doc, err := goquery.NewDocumentFromResponse(resp)
	s.Equal(err, nil)
	elem := doc.Find("input[name='csrf_token']").First()
	token, ok := elem.Attr("value")
	s.Equal(ok, true)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/login?next=%s", s.adminServer.URL, next), strings.NewReader(url.Values{
		"username":   {"admin"},
		"password":   {"gophish"},
		"csrf_token": {token},
	}.Encode()))
	s.Equal(err, nil)

	req.Header.Set("Cookie", resp.Header.Get("Set-Cookie"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err = client.Do(req)
	s.Equal(err, nil)
	s.Equal(resp.StatusCode, http.StatusFound)
	url, err := resp.Location()
	s.Equal(err, nil)
	s.Equal(url.Path, next)
}
