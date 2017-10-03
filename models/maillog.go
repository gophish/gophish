package models

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/mail"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
)

var MaxSendAttempts = 10
var ErrMaxSendAttempts = errors.New("max send attempts exceeded")

// MailLog is a struct that holds information about an email that is to be
// sent out.
type MailLog struct {
	Id          int64     `json:"-"`
	UserId      int64     `json:"-"`
	CampaignId  int64     `json:"campaign_id"`
	RId         string    `json:"id"`
	SendDate    time.Time `json:"send_date"`
	SendAttempt int       `json:"send_attempt"`
}

// GenerateMailLog creates a new maillog for the given campaign and
// result.
func GenerateMailLog(c *Campaign, r *Result) error {
	m := &MailLog{
		UserId:   c.UserId,
		RId:      r.RId,
		SendDate: c.LaunchDate,
	}
	err = db.Save(m).Error
	return err
}

// Backoff sets the MailLog SendDate to be the next entry in an exponential
// backoff. ErrMaxRetriesExceeded is thrown if this maillog has been retried
// too many times.
func (m MailLog) Backoff() error {
	if m.SendAttempt == MaxSendAttempts {
		return ErrMaxSendAttempts
	}
	m.SendAttempt += 1
	m.SendDate.Add(time.Minute * time.Duration(2*m.SendAttempt))
	db.Save(&m)
	return nil
}

// Error sets the error status on the models.Result that the
// maillog refers to.
func (m MailLog) Error(err error) error {
	r, err := GetResult(m.RId)
	if err != nil {
		return err
	}
	c, err := GetCampaign(m.CampaignId, m.UserId)
	if err != nil {
		return err
	}
	// Update the result
	r.UpdateStatus(ERROR)
	// Update the campaign events
	es := struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	}
	ej, err := json.Marshal(es)
	if err != nil {
		Logger.Println(err)
	}
	err = c.AddEvent(Event{Email: r.Email, Message: EVENT_SENDING_ERROR, Details: string(ej)})
	if err != nil {
		return err
	}
	return nil
}

// GenerateMessage fills in the details of a gomail.Message instance with
// the correct headers and body from the campaign and recipient listed in
// the maillog. We accept the gomail.Message as an argument so that the caller
// can choose to re-use the message across recipients.
func (m MailLog) GenerateMessage(msg *gomail.Message) error {
	r, err := GetResult(m.RId)
	if err != nil {
		return err
	}
	c, err := GetCampaign(m.CampaignId, m.UserId)
	if err != nil {
		return err
	}
	f, err := mail.ParseAddress(c.SMTP.FromAddress)
	if err != nil {
		return err
	}
	fn := f.Name
	if fn == "" {
		fn = f.Address
	}
	msg.SetAddressHeader("From", c.SMTP.FromAddress)
	td := struct {
		Result
		URL         string
		TrackingURL string
		Tracker     string
		From        string
	}{
		r,
		c.URL + "?rid=" + r.RId,
		c.URL + "/track?rid=" + r.RId,
		"<img alt='' style='display: none' src='" + c.URL + "/track?rid=" + r.RId + "'/>",
		fn,
	}

	// Parse the customHeader templates
	for _, header := range c.SMTP.Headers {
		parsedHeader := struct {
			Key   bytes.Buffer
			Value bytes.Buffer
		}{}
		keytmpl, err := template.New("text_template").Parse(header.Key)
		if err != nil {
			Logger.Println(err)
		}
		err = keytmpl.Execute(&parsedHeader.Key, td)
		if err != nil {
			Logger.Println(err)
		}

		valtmpl, err := template.New("text_template").Parse(header.Value)
		if err != nil {
			Logger.Println(err)
		}
		err = valtmpl.Execute(&parsedHeader.Value, td)
		if err != nil {
			Logger.Println(err)
		}

		// Add our header immediately
		msg.SetHeader(parsedHeader.Key.String(), parsedHeader.Value.String())
	}

	// Parse remaining templates
	var subjBuff bytes.Buffer
	tmpl, err := template.New("text_template").Parse(c.Template.Subject)
	if err != nil {
		Logger.Println(err)
	}
	err = tmpl.Execute(&subjBuff, td)
	if err != nil {
		Logger.Println(err)
	}
	msg.SetHeader("Subject", subjBuff.String())
	Logger.Println("Creating email using template")
	msg.SetHeader("To", r.FormatAddress())
	if c.Template.Text != "" {
		var textBuff bytes.Buffer
		tmpl, err = template.New("text_template").Parse(c.Template.Text)
		if err != nil {
			Logger.Println(err)
		}
		err = tmpl.Execute(&textBuff, td)
		if err != nil {
			Logger.Println(err)
		}
		msg.SetBody("text/plain", textBuff.String())
	}
	if c.Template.HTML != "" {
		var htmlBuff bytes.Buffer
		tmpl, err = template.New("html_template").Parse(c.Template.HTML)
		if err != nil {
			Logger.Println(err)
		}
		err = tmpl.Execute(&htmlBuff, td)
		if err != nil {
			Logger.Println(err)
		}
		if c.Template.Text == "" {
			msg.SetBody("text/html", htmlBuff.String())
		} else {
			msg.AddAlternative("text/html", htmlBuff.String())
		}
	}
	// Attach the files
	for _, a := range c.Template.Attachments {
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

// GetQueuedMailLogs returns the mail logs that are queued up for the given minute.
func GetQueuedMailLogs(t time.Time) ([]MailLog, error) {
	ms := []MailLog{}
	err := db.Where("send_at <= ?", t).
		Where("status = ?", CAMPAIGN_QUEUED).Find(&ms).Error
	if err != nil {
		Logger.Println(err)
	}
	return ms, err
}
