package worker

import (
	"bytes"
	"crypto/tls"
	"errors"
	"log"
	"net"
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
		Logger.Printf("Sending Email to %s\n", t.Email)
		err = e.Send(c.SMTP.Host, auth)
		if err != nil {
			Logger.Println(err)
			err = t.UpdateStatus(models.ERROR)
			if err != nil {
				Logger.Println(err)
			}
			err = c.AddEvent(models.Event{Email: t.Email, Message: models.EVENT_SENDING_ERROR})
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
	f, err := mail.ParseAddress(s.SMTP.FromAddress)
	if err != nil {
		Logger.Println(err)
		return err
	}
	ft := f.Name
	if ft == "" {
		ft = f.Address
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
	Logger.Printf("Sending Email to %s\n", s.Email)
	err = sendMail(e, s.SMTP)
	if err != nil {
		Logger.Println(err)
		// For now, let's split the error and return
		// the last element (the most descriptive error message)
		serr := strings.Split(err.Error(), ":")
		return errors.New(serr[len(serr)-1])
	}
	return err
}

// sendEmail is a copy of the net/smtp#SendMail function
// that has the option to ignore TLS errors
// TODO: Find a more elegant way (maybe in the email lib?) to do this
func sendMail(e email.Email, s models.SMTP) error {
	var auth smtp.Auth
	if s.Username != "" && s.Password != "" {
		auth = smtp.PlainAuth("", s.Username, s.Password, strings.Split(s.Host, ":")[0])
	}
	// Taken from the email library
	// Merge the To, Cc, and Bcc fields
	to := make([]string, 0, len(e.To)+len(e.Cc)+len(e.Bcc))
	to = append(append(append(to, e.To...), e.Cc...), e.Bcc...)
	for i := 0; i < len(to); i++ {
		addr, err := mail.ParseAddress(to[i])
		if err != nil {
			return err
		}
		to[i] = addr.Address
	}
	// Check to make sure there is at least one recipient and one "From" address
	if e.From == "" || len(to) == 0 {
		return errors.New("Must specify at least one From address and one To address")
	}
	from, err := mail.ParseAddress(e.From)
	if err != nil {
		return err
	}
	msg, err := e.Bytes()
	if err != nil {
		return err
	}
	// Taken from the standard library
	// https://github.com/golang/go/blob/master/src/net/smtp/smtp.go#L300
	c, err := smtp.Dial(s.Host)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Hello("localhost"); err != nil {
		return err
	}
	// Use TLS if available
	if ok, _ := c.Extension("STARTTLS"); ok {
		host, _, _ := net.SplitHostPort(s.Host)
		config := &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: s.IgnoreCertErrors,
		}
		if err = c.StartTLS(config); err != nil {
			return err
		}
	}

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				return err
			}
		}
	}
	if err = c.Mail(from.Address); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
