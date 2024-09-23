package models

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/gophish/gomail"
	"github.com/gophish/gophish/config"
	"github.com/jordan-wright/email"
	check "gopkg.in/check.v1"
)

func (s *ModelsSuite) TestEmailNotPresent(ch *check.C) {
	req := &EmailRequest{}
	ch.Assert(req.Validate(), check.Equals, ErrEmailNotSpecified)
	req.Email = "test@example.com"
	ch.Assert(req.Validate(), check.Equals, ErrFromAddressNotSpecified)
	req.FromAddress = "from@example.com"
	ch.Assert(req.Validate(), check.Equals, nil)
}

func (s *ModelsSuite) TestEmailRequestBackoff(ch *check.C) {
	req := &EmailRequest{
		ErrorChan: make(chan error),
	}
	expected := errors.New("Temporary Error")
	go func() {
		err := req.Backoff(expected)
		ch.Assert(err, check.Equals, nil)
	}()
	ch.Assert(<-req.ErrorChan, check.Equals, expected)
}

func (s *ModelsSuite) TestEmailRequestError(ch *check.C) {
	req := &EmailRequest{
		ErrorChan: make(chan error),
	}
	expected := errors.New("Temporary Error")
	go func() {
		err := req.Error(expected)
		ch.Assert(err, check.Equals, nil)
	}()
	ch.Assert(<-req.ErrorChan, check.Equals, expected)
}

func (s *ModelsSuite) TestEmailRequestSuccess(ch *check.C) {
	req := &EmailRequest{
		ErrorChan: make(chan error),
	}
	go func() {
		err := req.Success()
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
	req := &EmailRequest{
		SMTP:     smtp,
		Template: template,
		BaseRecipient: BaseRecipient{
			FirstName: "First",
			LastName:  "Last",
			Email:     "firstlast@example.com",
		},
		FromAddress: smtp.FromAddress,
	}

	s.config.ContactAddress = "test@test.com"
	expectedHeaders := map[string]string{
		"X-Mailer":          config.ServerName,
		"X-Gophish-Contact": s.config.ContactAddress,
	}

	msg := gomail.NewMessage()
	err := req.Generate(msg)
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
	for k, v := range expectedHeaders {
		ch.Assert(got.Headers.Get(k), check.Equals, v)
	}
}

func (s *ModelsSuite) TestGetSmtpFrom(ch *check.C) {
	smtp := SMTP{
		FromAddress: "from@example.com",
	}
	template := Template{
		Name:    "Test Template",
		Subject: "{{.FirstName}} - Subject",
		Text:    "{{.Email}} - Text",
		HTML:    "{{.Email}} - HTML",
	}
	req := &EmailRequest{
		SMTP:     smtp,
		Template: template,
		URL:      "http://127.0.0.1/{{.Email}}",
		BaseRecipient: BaseRecipient{
			FirstName: "First",
			LastName:  "Last",
			Email:     "firstlast@example.com",
		},
		FromAddress: smtp.FromAddress,
		RId:         fmt.Sprintf("%s-foobar", PreviewPrefix),
	}

	msg := gomail.NewMessage()
	err := req.Generate(msg)
	smtp_from, err := req.GetSmtpFrom()

	ch.Assert(err, check.Equals, nil)
	ch.Assert(smtp_from, check.Equals, "from@example.com")
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
	req := &EmailRequest{
		SMTP:     smtp,
		Template: template,
		URL:      "http://127.0.0.1/{{.Email}}",
		BaseRecipient: BaseRecipient{
			FirstName: "First",
			LastName:  "Last",
			Email:     "firstlast@example.com",
		},
		FromAddress: smtp.FromAddress,
		RId:         fmt.Sprintf("%s-foobar", PreviewPrefix),
	}

	msg := gomail.NewMessage()
	err := req.Generate(msg)
	ch.Assert(err, check.Equals, nil)

	expectedURL := fmt.Sprintf("http://127.0.0.1/%s?%s=%s", req.Email, RecipientParameter, req.RId)

	msgBuff := &bytes.Buffer{}
	_, err = msg.WriteTo(msgBuff)
	ch.Assert(err, check.Equals, nil)

	got, err := email.NewEmailFromReader(msgBuff)
	ch.Assert(err, check.Equals, nil)
	ch.Assert(got.Subject, check.Equals, expectedURL)
	ch.Assert(string(got.Text), check.Equals, expectedURL)
	ch.Assert(string(got.HTML), check.Equals, expectedURL)
}
func (s *ModelsSuite) TestEmailRequestGenerateEmptySubject(ch *check.C) {
	smtp := SMTP{
		FromAddress: "from@example.com",
	}
	template := Template{
		Name:    "Test Template",
		Subject: "",
		Text:    "{{.Email}} - Text",
		HTML:    "{{.Email}} - HTML",
	}
	req := &EmailRequest{
		SMTP:     smtp,
		Template: template,
		BaseRecipient: BaseRecipient{
			FirstName: "First",
			LastName:  "Last",
			Email:     "firstlast@example.com",
		},
		FromAddress: smtp.FromAddress,
	}

	msg := gomail.NewMessage()
	err := req.Generate(msg)
	ch.Assert(err, check.Equals, nil)

	expected := &email.Email{
		Subject: "",
		Text:    []byte(fmt.Sprintf("%s - Text", req.Email)),
		HTML:    []byte(fmt.Sprintf("%s - HTML", req.Email)),
	}

	msgBuff := &bytes.Buffer{}
	_, err = msg.WriteTo(msgBuff)
	ch.Assert(err, check.Equals, nil)

	got, err := email.NewEmailFromReader(msgBuff)
	ch.Assert(err, check.Equals, nil)
	ch.Assert(got.Subject, check.Equals, expected.Subject)
}

func (s *ModelsSuite) TestPostSendTestEmailRequest(ch *check.C) {
	smtp := SMTP{
		FromAddress: "from@example.com",
	}
	template := Template{
		Name:    "Test Template",
		Subject: "",
		Text:    "{{.Email}} - Text",
		HTML:    "{{.Email}} - HTML",
		UserId:  1,
	}
	err := PostTemplate(&template)
	ch.Assert(err, check.Equals, nil)

	page := Page{
		Name:   "Test Page",
		HTML:   "test",
		UserId: 1,
	}
	err = PostPage(&page)
	ch.Assert(err, check.Equals, nil)

	req := &EmailRequest{
		SMTP:       smtp,
		TemplateId: template.Id,
		PageId:     page.Id,
		BaseRecipient: BaseRecipient{
			FirstName: "First",
			LastName:  "Last",
			Email:     "firstlast@example.com",
		},
	}
	err = PostEmailRequest(req)
	ch.Assert(err, check.Equals, nil)

	got, err := GetEmailRequestByResultId(req.RId)
	ch.Assert(err, check.Equals, nil)
	ch.Assert(got.RId, check.Equals, req.RId)
	ch.Assert(got.Email, check.Equals, req.Email)
}
