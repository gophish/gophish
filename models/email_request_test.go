package models

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/gophish/gomail"
	"github.com/jordan-wright/email"
	check "gopkg.in/check.v1"
)

func (s *ModelsSuite) TestEmailNotPresent(ch *check.C) {
	req := &SendTestEmailRequest{}
	ch.Assert(req.Validate(), check.Equals, ErrEmailNotSpecified)
	req.Email = "test@example.com"
	ch.Assert(req.Validate(), check.Equals, nil)
}

func (s *ModelsSuite) TestEmailRequestBackoff(ch *check.C) {
	req := &SendTestEmailRequest{
		ErrorChan: make(chan error),
	}
	expected := errors.New("Temporary Error")
	go func() {
		err = req.Backoff(expected)
		ch.Assert(err, check.Equals, nil)
	}()
	ch.Assert(<-req.ErrorChan, check.Equals, expected)
}

func (s *ModelsSuite) TestEmailRequestError(ch *check.C) {
	req := &SendTestEmailRequest{
		ErrorChan: make(chan error),
	}
	expected := errors.New("Temporary Error")
	go func() {
		err = req.Error(expected)
		ch.Assert(err, check.Equals, nil)
	}()
	ch.Assert(<-req.ErrorChan, check.Equals, expected)
}

func (s *ModelsSuite) TestEmailRequestSuccess(ch *check.C) {
	req := &SendTestEmailRequest{
		ErrorChan: make(chan error),
	}
	go func() {
		err = req.Success()
		ch.Assert(err, check.Equals, nil)
	}()
	ch.Assert(<-req.ErrorChan, check.Equals, nil)
}

func (s *ModelsSuite) TestEmailRequestGenerate(ch *check.C) {
	smtp := SMTP{
		FromAddress: "from@example.com",
	}
	template := Template{
		Name:    "Test Template",
		Subject: "{{.FirstName}} - Subject",
		Text:    "{{.Email}} - Text",
		HTML:    "{{.Email}} - HTML",
	}
	target := Target{
		FirstName: "First",
		LastName:  "Last",
		Email:     "firstlast@example.com",
	}
	req := &SendTestEmailRequest{
		SMTP:     smtp,
		Template: template,
		Target:   target,
	}

	msg := gomail.NewMessage()
	err = req.Generate(msg)
	ch.Assert(err, check.Equals, nil)

	expected := &email.Email{
		Subject: fmt.Sprintf("%s - Subject", req.FirstName),
		Text:    []byte(fmt.Sprintf("%s - Text", req.Email)),
		HTML:    []byte(fmt.Sprintf("%s - HTML", req.Email)),
	}

	msgBuff := &bytes.Buffer{}
	_, err = msg.WriteTo(msgBuff)
	ch.Assert(err, check.Equals, nil)

	got, err := email.NewEmailFromReader(msgBuff)
	ch.Assert(err, check.Equals, nil)
	ch.Assert(got.Subject, check.Equals, expected.Subject)
	ch.Assert(string(got.Text), check.Equals, string(expected.Text))
	ch.Assert(string(got.HTML), check.Equals, string(expected.HTML))
}

func (s *ModelsSuite) TestEmailRequestURLTemplating(ch *check.C) {
	smtp := SMTP{
		FromAddress: "from@example.com",
	}
	template := Template{
		Name:    "Test Template",
		Subject: "{{.URL}}",
		Text:    "{{.URL}}",
		HTML:    "{{.URL}}",
	}
	target := Target{
		FirstName: "First",
		LastName:  "Last",
		Email:     "firstlast@example.com",
	}
	req := &SendTestEmailRequest{
		SMTP:     smtp,
		Template: template,
		Target:   target,
		URL: "http://127.0.0.1/{{.Email}}",
	}

	msg := gomail.NewMessage()
	err = req.Generate(msg)
	ch.Assert(err, check.Equals, nil)

	expectedURL := fmt.Sprintf("http://127.0.0.1/%s", target.Email)

	msgBuff := &bytes.Buffer{}
	_, err = msg.WriteTo(msgBuff)
	ch.Assert(err, check.Equals, nil)

	got, err := email.NewEmailFromReader(msgBuff)
	ch.Assert(err, check.Equals, nil)
	ch.Assert(got.Subject, check.Equals, expectedURL)
	ch.Assert(string(got.Text), check.Equals, expectedURL)
	ch.Assert(string(got.HTML), check.Equals, expectedURL)
}