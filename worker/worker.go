package worker

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"text/template"

	"github.com/gophish/gophish/models"
	"github.com/jordan-wright/email"
)

// Logger is the logger for the worker
var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

// Worker is the background worker that handles watching for new campaigns and sending emails appropriately.
type Worker struct {
	Queue chan *models.Campaign
}

// New creates a new worker object to handle the creation of campaigns
func New() *Worker {
	return &Worker{
		Queue: make(chan *models.Campaign),
	}
}

// Start launches the worker to monitor the database for any jobs.
// If a job is found, it launches the job
func (w *Worker) Start() {
	Logger.Println("Background Worker Started Successfully - Waiting for Campaigns")
	for {
		processCampaign(<-w.Queue)
	}
}

func processCampaign(c *models.Campaign) {
	Logger.Printf("Worker received: %s", c.Name)
	err := c.UpdateStatus(models.CAMPAIGN_IN_PROGRESS)
	if err != nil {
		Logger.Println(err)
	}
	e := email.Email{
		Subject: c.Template.Subject,
		From:    c.SMTP.FromAddress,
	}
	var auth smtp.Auth
	if c.SMTP.Username != "" && c.SMTP.Password != "" {
		auth = smtp.PlainAuth("", c.SMTP.Username, c.SMTP.Password, strings.Split(c.SMTP.Host, ":")[0])
	}
	tc := &tls.Config{
		ServerName:         c.SMTP.Host,
		InsecureSkipVerify: c.SMTP.IgnoreCertErrors,
	}
	f, err := mail.ParseAddress(c.SMTP.FromAddress)
	if err != nil {
		Logger.Println(err)
	}
	ft := f.Name
	if ft == "" {
		ft = f.Address
	}
	for _, t := range c.Results {
		td := struct {
			models.Result
			URL         string
			TrackingURL string
			Tracker     string
			From        string
		}{
			t,
			c.URL + "?rid=" + t.RId,
			c.URL + "/track?rid=" + t.RId,
			"<img src='" + c.URL + "/track?rid=" + t.RId + "'/>",
			ft,
		}
		// Parse the templates
		var subjBuff bytes.Buffer
		var htmlBuff bytes.Buffer
		var textBuff bytes.Buffer
		tmpl, err := template.New("html_template").Parse(c.Template.HTML)
		if err != nil {
			Logger.Println(err)
		}
		err = tmpl.Execute(&htmlBuff, td)
		if err != nil {
			Logger.Println(err)
		}
		e.HTML = htmlBuff.Bytes()
		tmpl, err = template.New("text_template").Parse(c.Template.Text)
		if err != nil {
			Logger.Println(err)
		}
		err = tmpl.Execute(&textBuff, td)
		if err != nil {
			Logger.Println(err)
		}
		e.Text = textBuff.Bytes()
		tmpl, err = template.New("text_template").Parse(c.Template.Subject)
		if err != nil {
			Logger.Println(err)
		}
		err = tmpl.Execute(&subjBuff, td)
		if err != nil {
			Logger.Println(err)
		}
		e.Subject = string(subjBuff.Bytes())
		Logger.Println("Creating email using template")
		e.To = []string{t.Email}
		e.Attachments = []*email.Attachment{}
		// Attach the files
		for _, a := range c.Template.Attachments {
			decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(a.Content))
			_, err = e.Attach(decoder, a.Name, a.Type)
			if err != nil {
				Logger.Println(err)
			}
		}
		Logger.Printf("Sending Email to %s\n", t.Email)
		err = e.SendWithTLS(c.SMTP.Host, auth, tc)
		if err != nil {
			Logger.Println(err)
			es := struct {
				Error string `json:"error"`
			}{
				Error: err.Error(),
			}
			ej, err := json.Marshal(es)
			if err != nil {
				Logger.Println(err)
			}
			err = t.UpdateStatus(models.ERROR)
			if err != nil {
				Logger.Println(err)
			}
			err = c.AddEvent(models.Event{Email: t.Email, Message: models.EVENT_SENDING_ERROR, Details: string(ej)})
			if err != nil {
				Logger.Println(err)
			}
		} else {
			err = t.UpdateStatus(models.EVENT_SENT)
			if err != nil {
				Logger.Println(err)
			}
			err = c.AddEvent(models.Event{Email: t.Email, Message: models.EVENT_SENT})
			if err != nil {
				Logger.Println(err)
			}
		}
	}
	err = c.UpdateStatus(models.CAMPAIGN_EMAILS_SENT)
	if err != nil {
		Logger.Println(err)
	}
}

func SendTestEmail(s *models.SendTestEmailRequest) error {
	e := email.Email{
		Subject: s.Template.Subject,
		From:    s.SMTP.FromAddress,
	}
	var auth smtp.Auth
	if s.SMTP.Username != "" && s.SMTP.Password != "" {
		auth = smtp.PlainAuth("", s.SMTP.Username, s.SMTP.Password, strings.Split(s.SMTP.Host, ":")[0])
	}
	t := &tls.Config{
		ServerName:         s.SMTP.Host,
		InsecureSkipVerify: s.SMTP.IgnoreCertErrors,
	}
	f, err := mail.ParseAddress(s.SMTP.FromAddress)
	if err != nil {
		Logger.Println(err)
		return err
	}
	s.From = f.Name
	if s.From == "" {
		s.From = f.Address
	}
	Logger.Println("Creating email using template")
	// Parse the templates
	var subjBuff bytes.Buffer
	var htmlBuff bytes.Buffer
	var textBuff bytes.Buffer
	tmpl, err := template.New("html_template").Parse(s.Template.HTML)
	if err != nil {
		Logger.Println(err)
	}
	err = tmpl.Execute(&htmlBuff, s)
	if err != nil {
		Logger.Println(err)
	}
	e.HTML = htmlBuff.Bytes()
	tmpl, err = template.New("text_template").Parse(s.Template.Text)
	if err != nil {
		Logger.Println(err)
	}
	err = tmpl.Execute(&textBuff, s)
	if err != nil {
		Logger.Println(err)
	}
	e.Text = textBuff.Bytes()
	tmpl, err = template.New("text_template").Parse(s.Template.Subject)
	if err != nil {
		Logger.Println(err)
	}
	err = tmpl.Execute(&subjBuff, s)
	if err != nil {
		Logger.Println(err)
	}
	e.Subject = string(subjBuff.Bytes())
	e.To = []string{s.Email}
	// Attach the files
	for _, a := range s.Template.Attachments {
		decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(a.Content))
		_, err = e.Attach(decoder, a.Name, a.Type)
		if err != nil {
			Logger.Println(err)
		}
	}
	Logger.Printf("Sending Email to %s\n", s.Email)
	err = e.SendWithTLS(s.SMTP.Host, auth, t)
	if err != nil {
		Logger.Println(err)
		// For now, let's split the error and return
		// the last element (the most descriptive error message)
		serr := strings.Split(err.Error(), ":")
		return errors.New(serr[len(serr)-1])
	}
	return err
}
