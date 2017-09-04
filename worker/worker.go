package worker

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gophish/gophish/mailer"
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
		ms, err := models.GetQueuedMailLogs(t.UTC())
		if err != nil {
			Logger.Println(err)
			continue
		}
		// We'll group the maillogs by campaign ID to (sort of) group
		// them by sending profile. This lets us re-use the Sender
		// instead of having to re-connect to the SMTP server for every
		// email.
		msg := make(map[int64][]models.MailLog)
		for _, m := range ms {
			msg[m.CampaignId] = append(msg[m.CampaignId], m)
		}

		// Next, we process each group of maillogs in parallel
		for cid, msc := range msg {
			go func(cid int64, msc []models.MailLog) {
				uid := msc[0].UserId
				c, err := models.GetCampaign(cid, uid)
				if err != nil {
					failMailLogs(err, ms)
				}
				if c.Status == models.CAMPAIGN_QUEUED {
					err := c.UpdateStatus(models.CAMPAIGN_IN_PROGRESS)
					if err != nil {
						Logger.Println(err)
					}
				}
				// Create the dialer and connect to the SMTP server
				d := mailer.NewDialer(c.SMTP)
				sc, err := d.Dial()
				if err != nil {
					failMailLogs(err, ms)
					return
				}
				processMailLogs(cid, uid, msc, d)
			}(cid, msc)
		}
	}
}

// failMailLogs automatically marks each mail log as a failed sending
// attempt. This occurs when fatal errors such as database/smtp
// connection errors occur.
func failMailLogs(err error, ms []models.MailLog) {
	for _, m := range ms {
		err := m.Error(err)
		if err != nil {
			Logger.Println(err)
		}
	}
}

func processMailLogs(cid int64, uid int64, ms []models.MailLog, d mailer.Dialer) {
	c, err := models.GetCampaign(cid, uid)
	if err != nil {
		failMailLogs(err, ms)
	}
	sc, err := d.Dial()
	if err != nil {
		failMailLogs(err, ms)
	}
	msg := gomail.NewMessage()
	for i := range ms {
		m := ms[i]
		err = m.GenerateMessage(msg)
		if err != nil {
			Logger.Println(err)
			m.Error(err)
			msg.Reset()
			continue
		}
		err := gomail.Send(sc, msg)
		if err != nil {
			if te, ok := err.(*textproto.Error); ok {
				Logger.Println(te)
				switch {
				// In the case of a temporary (4xx) error, we will
				// set a backoff on the maillog and reconnect
				// to be nice.
				// I don't believe we need to reset the connection
				// for deferred errors
				case te.Code >= 400 && te.Code <= 499:
					err = m.Backoff()
					if err != nil {
						m.Error(err)
					}
				// If we have a different permanent error, let's be
				// considerate and just establish a new connection.
				case te.Code >= 500 && te.Code <= 599:
					m.Error(te)
					sc, err = d.Dial()
					if err != nil {
						failMailLogs(err, ms[i:])
						return
					}
				}
			} else {
				// Generic error, let's just log it and try to
				// re-connect
				Logger.Println(err)
				sc, err = d.Dial()
				if err != nil {
					failMailLogs(err, ms[i:])
					return
				}
			}

		}
		msg.Reset()
	}
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
	e.SetHeader("From", s.SMTP.FromAddress)
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
