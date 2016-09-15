package controllers

import (
	"fmt"
	"net/http"
	"net/url"
)

func (s *ControllersSuite) TestLoginCSRF() {
	resp, err := http.PostForm(fmt.Sprintf("%s/login", as.URL),
		url.Values{
			"username": {"admin"},
			"password": {"gophish"},
		})

	s.Equal(resp.StatusCode, 403)
	fmt.Println(err)
}
