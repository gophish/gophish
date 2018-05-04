package models

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/mail"
	"strings"

	"github.com/gophish/gomail"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/mailer"
)

// SendTestEmailRequest is the structure of a request
// to send a test email to test an SMTP connection.
// This type implements the mailer.Mail interface.
type SendTestEmailRequest struct {
	Template    Template `json:"template"`
	Page        Page     `json:"page"`
	SMTP        SMTP     `json:"smtp"`
	URL         string   `json:"url"`
	Tracker     string   `json:"tracker"`
	TrackingURL string   `json:"tracking_url"`
	From        string   `json:"from"`
	Target
	ErrorChan chan (error) `json:"-"`
}

// Validate ensures the SendTestEmailRequest structure
// is valid.
func (s *SendTestEmailRequest) Validate() error {
	switch {
	case s.Email == "":
		return ErrEmailNotSpecified
	}
	return nil
}

// Backoff treats temporary errors as permanent since this is expected to be a
// synchronous operation. It returns any errors given back to the ErrorChan
func (s *SendTestEmailRequest) Backoff(reason error) error {
	s.ErrorChan <- reason
	return nil
}

// Error returns an error on the ErrorChan.
func (s *SendTestEmailRequest) Error(err error) error {
	s.ErrorChan <- err
	return nil
}

// Success returns nil on the ErrorChan to indicate that the email was sent
// successfully.
func (s *SendTestEmailRequest) Success() error {
	s.ErrorChan <- nil
	return nil
}

// Generate fills in the details of a gomail.Message with the contents
// from the SendTestEmailRequest.
func (s *SendTestEmailRequest) Generate(msg *gomail.Message) error {
	f, err := mail.ParseAddress(s.SMTP.FromAddress)
	if err != nil {
		return err
	}
	fn := f.Name
	if fn == "" {
		fn = f.Address
	}
	msg.SetAddressHeader("From", f.Address, f.Name)

	url, err := buildTemplate(s.URL, s)
	if err != nil {
		return err
	}
	s.URL = url

	// Parse the customHeader templates
	for _, header := range s.SMTP.Headers {
		key, err := buildTemplate(header.Key, s)
		if err != nil {
			log.Error(err)
		}

		value, err := buildTemplate(header.Value, s)
		if err != nil {
			log.Error(err)
		}

		// Add our header immediately
		msg.SetHeader(key, value)
	}

	// Parse remaining templates
	subject, err := buildTemplate(s.Template.Subject, s)
	if err != nil {
		log.Error(err)
	}
	// don't set the Subject header if it is blank
	if len(subject) != 0 {
		msg.SetHeader("Subject", subject)
	}

	msg.SetHeader("To", s.FormatAddress())
	if s.Template.Text != "" {
		text, err := buildTemplate(s.Template.Text, s)
		if err != nil {
			log.Error(err)
		}
		msg.SetBody("text/plain", text)
	}
	if s.Template.HTML != "" {
		html, err := buildTemplate(s.Template.HTML, s)
		if err != nil {
			log.Error(err)
		}
		if s.Template.Text == "" {
			msg.SetBody("text/html", html)
		} else {
			msg.AddAlternative("text/html", html)
		}
	}
	// Attach the files
	for _, a := range s.Template.Attachments {
		msg.Attach(func(a Attachment) (string, gomail.FileSetting, gomail.FileSetting) {
			h := map[string][]string{"Content-ID": {fmt.Sprintf("<%s>", a.Name)}}
			return a.Name, gomail.SetCopyFunc(func(w io.Writer) error {
				decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(a.Content))
				_, err = io.Copy(w, decoder)
				return err
			}), gomail.SetHeader(h)
		}(a))
	}

	return nil
}

// GetDialer returns the mailer.Dialer for the underlying SMTP object
func (s *SendTestEmailRequest) GetDialer() (mailer.Dialer, error) {
	return s.SMTP.GetDialer()
}
