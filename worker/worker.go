package worker

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/models"
	"gopkg.in/gomail.v2"
)

// Logger is the logger for the worker
var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

// Worker is the background worker that handles watching for new campaigns and sending emails appropriately.
type Worker struct{}

// New creates a new worker object to handle the creation of campaigns
func New() *Worker {
	return &Worker{}
}

// Start launches the worker to poll the database every minute for any jobs.
// If a job is found, it launches the job
func (w *Worker) Start() {
	Logger.Println("Background Worker Started Successfully - Waiting for Campaigns")
	for t := range time.Tick(1 * time.Minute) {
		cs, err := models.GetQueuedCampaigns(t)
		// Not really sure of a clean way to catch errors per campaign...
		if err != nil {
			Logger.Println(err)
			continue
		}
		for _, c := range cs {
			go func(c models.Campaign) {
				processCampaign(&c)
			}(c)
		}
	}
	// Set seed here for SMTP send delay randomness
	rand.Seed(time.Now().Unix())
}

func processCampaign(c *models.Campaign) {
	Logger.Printf("Worker received: %s", c.Name)
	err := c.UpdateStatus(models.CAMPAIGN_IN_PROGRESS)
	if err != nil {
		Logger.Println(err)
	}
	f, err := mail.ParseAddress(c.SMTP.FromAddress)
	if err != nil {
		Logger.Println(err)
	}
	fn := f.Name
	if fn == "" {
		fn = f.Address
	}
	// Setup the message and dial
	hp := strings.Split(c.SMTP.Host, ":")
	if len(hp) < 2 {
		hp = append(hp, "25")
	}
	// Any issues should have been caught in validation, so we just log
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		Logger.Println(err)
	}
	d := gomail.NewDialer(hp[0], port, c.SMTP.Username, c.SMTP.Password)
	d.TLSConfig = &tls.Config{
		ServerName:         c.SMTP.Host,
		InsecureSkipVerify: c.SMTP.IgnoreCertErrors,
	}
	hostname, err := os.Hostname()
	if err != nil {
		Logger.Println(err)
		hostname = "localhost"
	}
	d.LocalName = hostname
	s, err := d.Dial()
	// Short circuit if we have an err
	// However, we still need to update each target
	if err != nil {
		Logger.Println(err)
		for _, t := range c.Results {
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
		}
		return
	}
	// Send each email
	e := gomail.NewMessage()
	for _, t := range c.Results {
		if config.Conf.SMTPConf.UseDelay == true {
			SendingDelay()
		}

		e.SetHeader("From", c.SMTP.FromAddress)
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
			"<img style='display: none' src='" + c.URL + "/track?rid=" + t.RId + "'/>",
			fn,
		}
		// Parse the templates
		var subjBuff bytes.Buffer
		tmpl, err := template.New("text_template").Parse(c.Template.Subject)
		if err != nil {
			Logger.Println(err)
		}
		err = tmpl.Execute(&subjBuff, td)
		if err != nil {
			Logger.Println(err)
		}
		e.SetHeader("Subject", subjBuff.String())
		Logger.Println("Creating email using template")
		e.SetHeader("To", t.Email)
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
			e.SetBody("text/plain", textBuff.String())
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
				e.SetBody("text/html", htmlBuff.String())
			} else {
				e.AddAlternative("text/html", htmlBuff.String())
			}
		}
		// Attach the files
		for _, a := range c.Template.Attachments {
			e.Attach(func(a models.Attachment) (string, gomail.FileSetting) {
				return a.Name, gomail.SetCopyFunc(func(w io.Writer) error {
					decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(a.Content))
					_, err = io.Copy(w, decoder)
					return err
				})
			}(a))
		}
		Logger.Printf("Sending Email to %s\n", t.Email)
		err = gomail.Send(s, e)
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
		e.Reset()
	}
	err = c.UpdateStatus(models.CAMPAIGN_EMAILS_SENT)
	if err != nil {
		Logger.Println(err)
	}
}

// Creates a simple constant delay between sending each email message.  Later
// we can expand this function to include randomization of delay time, min and
// max delay times and possibly other features.  Returns nothing for now just
// delays the num seconds specified in the main config file.
func SendingDelay() {
	strSendDelay := strconv.Itoa(RandWithinRange(config.Conf.SMTPConf.MinDelay, config.Conf.SMTPConf.MaxDelay))
	Logger.Printf("Next email will be sent in %s seconds\n", strSendDelay)
	tSendDelay, _ := time.ParseDuration(strSendDelay + "s")
	time.Sleep(tSendDelay)
}

func RandWithinRange(min, max int) int {
	if min >= max { // Avoids a panic with rand if min is same value as max
		max = min + 1
	}
	return rand.Intn(max-min) + min
}

func SendTestEmail(s *models.SendTestEmailRequest) error {
	hp := strings.Split(s.SMTP.Host, ":")
	if len(hp) < 2 {
		hp = append(hp, "25")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		Logger.Println(err)
		return err
	}
	d := gomail.NewDialer(hp[0], port, s.SMTP.Username, s.SMTP.Password)
	d.TLSConfig = &tls.Config{
		ServerName:         s.SMTP.Host,
		InsecureSkipVerify: s.SMTP.IgnoreCertErrors,
	}
	hostname, err := os.Hostname()
	if err != nil {
		Logger.Println(err)
		hostname = "localhost"
	}
	d.LocalName = hostname
	dc, err := d.Dial()
	if err != nil {
		Logger.Println(err)
		return err
	}
	e := gomail.NewMessage()
	e.SetHeader("From", s.SMTP.FromAddress)
	e.SetHeader("To", s.Email)
	Logger.Println("Creating email using template")
	// Parse the templates
	var subjBuff bytes.Buffer
	tmpl, err := template.New("text_template").Parse(s.Template.Subject)
	if err != nil {
		Logger.Println(err)
	}
	err = tmpl.Execute(&subjBuff, s)
	if err != nil {
		Logger.Println(err)
	}
	e.SetHeader("Subject", subjBuff.String())
	if s.Template.Text != "" {
		var textBuff bytes.Buffer
		tmpl, err = template.New("text_template").Parse(s.Template.Text)
		if err != nil {
			Logger.Println(err)
		}
		err = tmpl.Execute(&textBuff, s)
		if err != nil {
			Logger.Println(err)
		}
		e.SetBody("text/plain", textBuff.String())
	}
	if s.Template.HTML != "" {
		var htmlBuff bytes.Buffer
		tmpl, err = template.New("html_template").Parse(s.Template.HTML)
		if err != nil {
			Logger.Println(err)
		}
		err = tmpl.Execute(&htmlBuff, s)
		if err != nil {
			Logger.Println(err)
		}
		// If we don't have a text part, make the html the root part
		if s.Template.Text == "" {
			e.SetBody("text/html", htmlBuff.String())
		} else {
			e.AddAlternative("text/html", htmlBuff.String())
		}
	}
	// Attach the files
	for _, a := range s.Template.Attachments {
		e.Attach(func(a models.Attachment) (string, gomail.FileSetting) {
			return a.Name, gomail.SetCopyFunc(func(w io.Writer) error {
				decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(a.Content))
				_, err = io.Copy(w, decoder)
				return err
			})
		}(a))
	}
	Logger.Printf("Sending Email to %s\n", s.Email)
	err = gomail.Send(dc, e)
	if err != nil {
		Logger.Println(err)
		// For now, let's split the error and return
		// the last element (the most descriptive error message)
		serr := strings.Split(err.Error(), ":")
		return errors.New(serr[len(serr)-1])
	}
	return err
}
