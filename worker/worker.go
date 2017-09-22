package worker

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

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
		cs, err := models.GetQueuedCampaigns(t.UTC())
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
		e.SetAddressHeader("From", f.Address, f.Name)
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
			"<img alt='' style='display: none' src='" + c.URL + "/track?rid=" + t.RId + "'/>",
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
			e.SetHeader(parsedHeader.Key.String(), parsedHeader.Value.String())
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
		e.SetHeader("Subject", subjBuff.String())
		Logger.Println("Creating email using template")
		e.SetHeader("To", t.FormatAddress())
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
			e.Attach(func(a models.Attachment) (string, gomail.FileSetting, gomail.FileSetting) {
				h := map[string][]string{"Content-ID": {fmt.Sprintf("<%s>", a.Name)}}
				return a.Name, gomail.SetCopyFunc(func(w io.Writer) error {
					decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(a.Content))
					_, err = io.Copy(w, decoder)
					return err
				}), gomail.SetHeader(h)
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

func SendTestEmail(s *models.SendTestEmailRequest) error {
	f, err := mail.ParseAddress(s.SMTP.FromAddress)
	if err != nil {
		Logger.Println(err)
		return err
	}
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
	Logger.Println("Creating email using template")
	e := gomail.NewMessage()
	// Parse the customHeader templates
	for _, header := range s.SMTP.Headers {
		parsedHeader := struct {
			Key   bytes.Buffer
			Value bytes.Buffer
		}{}
		keytmpl, err := template.New("text_template").Parse(header.Key)
		if err != nil {
			Logger.Println(err)
		}
		err = keytmpl.Execute(&parsedHeader.Key, s)
		if err != nil {
			Logger.Println(err)
		}

		valtmpl, err := template.New("text_template").Parse(header.Value)
		if err != nil {
			Logger.Println(err)
		}
		err = valtmpl.Execute(&parsedHeader.Value, s)
		if err != nil {
			Logger.Println(err)
		}

		// Add our header immediately
		e.SetHeader(parsedHeader.Key.String(), parsedHeader.Value.String())
	}
	e.SetAddressHeader("From", f.Address, f.Name)
	e.SetHeader("To", s.FormatAddress())
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
