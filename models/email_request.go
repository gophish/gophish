package models

import (
	"fmt"
	"net/mail"

	"github.com/gophish/gomail"
	"github.com/gophish/gophish/config"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/mailer"
)

// PreviewPrefix is the standard prefix added to the rid parameter when sending
// test emails.
const PreviewPrefix = "preview-"

// EmailRequest is the structure of a request
// to send a test email to test an SMTP connection.
// This type implements the mailer.Mail interface.
type EmailRequest struct {
	Id          int64        `json:"-"`
	Template    Template     `json:"template"`
	TemplateId  int64        `json:"-"`
	Page        Page         `json:"page"`
	PageId      int64        `json:"-"`
	SMTP        SMTP         `json:"smtp"`
	URL         string       `json:"url"`
	Tracker     string       `json:"tracker" gorm:"-"`
	TrackingURL string       `json:"tracking_url" gorm:"-"`
	UserId      int64        `json:"-"`
	ErrorChan   chan (error) `json:"-" gorm:"-"`
	RId         string       `json:"id"`
	FromAddress string       `json:"-"`
	BaseRecipient
}

func (s *EmailRequest) getBaseURL() string {
	return s.URL
}

func (s *EmailRequest) getFromAddress() string {
	return s.FromAddress
}

// Validate ensures the SendTestEmailRequest structure
// is valid.
func (s *EmailRequest) Validate() error {
	switch {
	case s.Email == "":
		return ErrEmailNotSpecified
	case s.FromAddress == "" && s.SMTP.FromAddress == "":
		return ErrFromAddressNotSpecified
	}
	return nil
}

// Backoff treats temporary errors as permanent since this is expected to be a
// synchronous operation. It returns any errors given back to the ErrorChan
func (s *EmailRequest) Backoff(reason error) error {
	s.ErrorChan <- reason
	return nil
}

// Error returns an error on the ErrorChan.
func (s *EmailRequest) Error(err error) error {
	s.ErrorChan <- err
	return nil
}

// Success returns nil on the ErrorChan to indicate that the email was sent
// successfully.
func (s *EmailRequest) Success() error {
	s.ErrorChan <- nil
	return nil
}

func (s *EmailRequest) GetSmtpFrom() (string, error) {
	return s.SMTP.FromAddress, nil
}

// PostEmailRequest stores a SendTestEmailRequest in the database.
func PostEmailRequest(s *EmailRequest) error {
	// Generate an ID to be used in the underlying Result object
	rid, err := generateResultId()
	if err != nil {
		return err
	}
	s.RId = fmt.Sprintf("%s%s", PreviewPrefix, rid)
	return db.Save(&s).Error
}

// GetEmailRequestByResultId retrieves the EmailRequest by the underlying rid
// parameter.
func GetEmailRequestByResultId(id string) (EmailRequest, error) {
	s := EmailRequest{}
	err := db.Table("email_requests").Where("r_id=?", id).First(&s).Error
	return s, err
}

// Generate fills in the details of a gomail.Message with the contents
// from the SendTestEmailRequest.
func (s *EmailRequest) Generate(msg *gomail.Message) error {
	f, err := mail.ParseAddress(s.getFromAddress())
	if err != nil {
		return err
	}
	msg.SetAddressHeader("From", f.Address, f.Name)

	ptx, err := NewPhishingTemplateContext(s, s.BaseRecipient, s.RId)
	if err != nil {
		return err
	}

	url, err := ExecuteTemplate(s.URL, ptx)
	if err != nil {
		return err
	}
	s.URL = url

	// Add the transparency headers
	msg.SetHeader("X-Mailer", config.ServerName)
	if conf.ContactAddress != "" {
		msg.SetHeader("X-Gophish-Contact", conf.ContactAddress)
	}

	// Parse the customHeader templates
	for _, header := range s.SMTP.Headers {
		key, err := ExecuteTemplate(header.Key, ptx)
		if err != nil {
			log.Error(err)
		}

		value, err := ExecuteTemplate(header.Value, ptx)
		if err != nil {
			log.Error(err)
		}

		// Add our header immediately
		msg.SetHeader(key, value)
	}

	// Parse remaining templates
	subject, err := ExecuteTemplate(s.Template.Subject, ptx)
	if err != nil {
		log.Error(err)
	}
	// don't set the Subject header if it is blank
	if subject != "" {
		msg.SetHeader("Subject", subject)
	}

	msg.SetHeader("To", s.FormatAddress())
	if s.Template.Text != "" {
		text, err := ExecuteTemplate(s.Template.Text, ptx)
		if err != nil {
			log.Error(err)
		}
		msg.SetBody("text/plain", text)
	}
	if s.Template.HTML != "" {
		html, err := ExecuteTemplate(s.Template.HTML, ptx)
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
		addAttachment(msg, a, ptx)
	}

	return nil
}

// GetDialer returns the mailer.Dialer for the underlying SMTP object
func (s *EmailRequest) GetDialer() (mailer.Dialer, error) {
	return s.SMTP.GetDialer()
}
