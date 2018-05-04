package models

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/mail"
	"net/url"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/gophish/gomail"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/mailer"
)

// MaxSendAttempts set to 8 since we exponentially backoff after each failed send
// attempt. This will give us a maximum send delay of 256 minutes, or about 4.2 hours.
var MaxSendAttempts = 8

// ErrMaxSendAttempts is thrown when the maximum number of sending attemps for a given
// MailLog is exceeded.
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
	Processing  bool      `json:"-"`
}

// GenerateMailLog creates a new maillog for the given campaign and
// result. It sets the initial send date to match the campaign's launch date.
func GenerateMailLog(c *Campaign, r *Result) error {
	m := &MailLog{
		UserId:     c.UserId,
		CampaignId: c.Id,
		RId:        r.RId,
		SendDate:   c.LaunchDate,
	}
	err = db.Save(m).Error
	return err
}

// Backoff sets the MailLog SendDate to be the next entry in an exponential
// backoff. ErrMaxRetriesExceeded is thrown if this maillog has been retried
// too many times. Backoff also unlocks the maillog so that it can be processed
// again in the future.
func (m *MailLog) Backoff(reason error) error {
	if m.SendAttempt == MaxSendAttempts {
		err = m.addError(ErrMaxSendAttempts)
		return ErrMaxSendAttempts
	}
	r, err := GetResult(m.RId)
	if err != nil {
		return err
	}
	// Add an error, since we had to backoff because of a
	// temporary error of some sort during the SMTP transaction
	err = m.addError(reason)
	if err != nil {
		return err
	}
	m.SendAttempt++
	backoffDuration := math.Pow(2, float64(m.SendAttempt))
	m.SendDate = m.SendDate.Add(time.Minute * time.Duration(backoffDuration))
	err = db.Save(m).Error
	if err != nil {
		return err
	}
	r.Status = STATUS_RETRY
	r.SendDate = m.SendDate
	err = db.Save(r).Error
	if err != nil {
		return err
	}
	err = m.Unlock()
	return err
}

// Unlock removes the processing flag so the maillog can be processed again
func (m *MailLog) Unlock() error {
	m.Processing = false
	return db.Save(&m).Error
}

// Lock sets the processing flag so that other processes cannot modify the maillog
func (m *MailLog) Lock() error {
	m.Processing = true
	return db.Save(&m).Error
}

// addError adds an error to the associated campaign
func (m *MailLog) addError(e error) error {
	c, err := GetCampaign(m.CampaignId, m.UserId)
	if err != nil {
		return err
	}
	// This is redundant in the case of permanent
	// errors, but the extra query makes for
	// a cleaner API.
	r, err := GetResult(m.RId)
	if err != nil {
		return err
	}
	es := struct {
		Error string `json:"error"`
	}{
		Error: e.Error(),
	}
	ej, err := json.Marshal(es)
	if err != nil {
		log.Warn(err)
	}
	err = c.AddEvent(Event{Email: r.Email, Message: EVENT_SENDING_ERROR, Details: string(ej)})
	return err
}

// Error sets the error status on the models.Result that the
// maillog refers to. Since MailLog errors are permanent,
// this action also deletes the maillog.
func (m *MailLog) Error(e error) error {
	r, err := GetResult(m.RId)
	if err != nil {
		log.Warn(err)
		return err
	}
	// Update the result
	err = r.UpdateStatus(ERROR)
	if err != nil {
		log.Warn(err)
		return err
	}
	// Update the campaign events
	err = m.addError(e)
	if err != nil {
		log.Warn(err)
		return err
	}
	err = db.Delete(m).Error
	return err
}

// Success deletes the maillog from the database and updates the underlying
// campaign result.
func (m *MailLog) Success() error {
	r, err := GetResult(m.RId)
	if err != nil {
		return err
	}
	err = r.UpdateStatus(EVENT_SENT)
	if err != nil {
		return err
	}
	c, err := GetCampaign(m.CampaignId, m.UserId)
	if err != nil {
		return err
	}
	err = c.AddEvent(Event{Email: r.Email, Message: EVENT_SENT})
	if err != nil {
		return err
	}
	err = db.Delete(m).Error
	return nil
}

// GetDialer returns a dialer based on the maillog campaign's SMTP configuration
func (m *MailLog) GetDialer() (mailer.Dialer, error) {
	c, err := GetCampaign(m.CampaignId, m.UserId)
	if err != nil {
		return nil, err
	}
	return c.SMTP.GetDialer()
}

// buildTemplate creates a templated string based on the provided
// template body and data.
func buildTemplate(text string, data interface{}) (string, error) {
	buff := bytes.Buffer{}
	tmpl, err := template.New("template").Parse(text)
	if err != nil {
		return buff.String(), err
	}
	err = tmpl.Execute(&buff, data)
	return buff.String(), err
}

// Generate fills in the details of a gomail.Message instance with
// the correct headers and body from the campaign and recipient listed in
// the maillog. We accept the gomail.Message as an argument so that the caller
// can choose to re-use the message across recipients.
func (m *MailLog) Generate(msg *gomail.Message) error {
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
	msg.SetAddressHeader("From", f.Address, f.Name)
	campaignURL, err := buildTemplate(c.URL, r)
	if err != nil {
		return err
	}

	phishURL, _ := url.Parse(campaignURL)
	q := phishURL.Query()
	q.Set("rid", r.RId)
	phishURL.RawQuery = q.Encode()

	trackingURL, _ := url.Parse(campaignURL)
	trackingURL.Path = path.Join(trackingURL.Path, "/track")
	trackingURL.RawQuery = q.Encode()

	td := struct {
		Result
		URL         string
		TrackingURL string
		Tracker     string
		From        string
	}{
		r,
		phishURL.String(),
		trackingURL.String(),
		"<img alt='' style='display: none' src='" + trackingURL.String() + "'/>",
		fn,
	}

	// Parse the customHeader templates
	for _, header := range c.SMTP.Headers {
		key, err := buildTemplate(header.Key, td)
		if err != nil {
			log.Warn(err)
		}

		value, err := buildTemplate(header.Value, td)
		if err != nil {
			log.Warn(err)
		}

		// Add our header immediately
		msg.SetHeader(key, value)
	}

	// Parse remaining templates
	subject, err := buildTemplate(c.Template.Subject, td)
	if err != nil {
		log.Warn(err)
	}
	// don't set Subject header if the subject is empty
	if len(subject) != 0 {
		msg.SetHeader("Subject", subject)
	}

	msg.SetHeader("To", r.FormatAddress())
	if c.Template.Text != "" {
		text, err := buildTemplate(c.Template.Text, td)
		if err != nil {
			log.Warn(err)
		}
		msg.SetBody("text/plain", text)
	}
	if c.Template.HTML != "" {
		html, err := buildTemplate(c.Template.HTML, td)
		if err != nil {
			log.Warn(err)
		}
		if c.Template.Text == "" {
			msg.SetBody("text/html", html)
		} else {
			msg.AddAlternative("text/html", html)
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
func GetQueuedMailLogs(t time.Time) ([]*MailLog, error) {
	ms := []*MailLog{}
	err := db.Where("send_date <= ? AND processing = ?", t, false).
		Find(&ms).Error
	if err != nil {
		log.Warn(err)
	}
	return ms, err
}

// GetMailLogsByCampaign returns all of the mail logs for a given campaign.
func GetMailLogsByCampaign(cid int64) ([]*MailLog, error) {
	ms := []*MailLog{}
	err := db.Where("campaign_id = ?", cid).Find(&ms).Error
	return ms, err
}

// LockMailLogs locks or unlocks a slice of maillogs for processing.
func LockMailLogs(ms []*MailLog, lock bool) error {
	tx := db.Begin()
	for i := range ms {
		ms[i].Processing = lock
		err := tx.Save(ms[i]).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

// UnlockAllMailLogs removes the processing lock for all maillogs
// in the database. This is intended to be called when Gophish is started
// so that any previously locked maillogs can resume processing.
func UnlockAllMailLogs() error {
	err = db.Model(&MailLog{}).Update("processing", false).Error
	return err
}
