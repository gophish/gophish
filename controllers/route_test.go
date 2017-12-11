package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (s *ControllersSuite) TestLoginCSRF() {
	resp, err := http.PostForm(fmt.Sprintf("%s/login", as.URL),
		url.Values{
			"username": {"admin"},
			"password": {"gophish"},
		})

	s.Equal(resp.StatusCode, http.StatusForbidden)
	fmt.Println(err)
}

func (s *ControllersSuite) TestInvalidCredentials() {
	resp, err := http.Get(fmt.Sprintf("%s/login", as.URL))
	s.Equal(err, nil)
	s.Equal(resp.StatusCode, http.StatusOK)

	doc, err := goquery.NewDocumentFromResponse(resp)
	s.Equal(err, nil)
	elem := doc.Find("input[name='csrf_token']").First()
	token, ok := elem.Attr("value")
	s.Equal(ok, true)

	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/login", as.URL), strings.NewReader(url.Values{
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
	resp, err := http.Get(fmt.Sprintf("%s/login", as.URL))
	s.Equal(err, nil)
	s.Equal(resp.StatusCode, http.StatusOK)

	doc, err := goquery.NewDocumentFromResponse(resp)
	s.Equal(err, nil)
	elem := doc.Find("input[name='csrf_token']").First()
	token, ok := elem.Attr("value")
	s.Equal(ok, true)

	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/login", as.URL), strings.NewReader(url.Values{
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
