package worker

import (
	"bytes"
	"log"
	"net/smtp"
	"os"
	"strings"
	"text/template"

	"github.com/jordan-wright/email"
	"github.com/jordan-wright/gophish/models"
)

var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

type Worker struct {
	Queue chan *models.Campaign
}

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
	for _, t := range c.Results {
		// Parse the templates
		var buff bytes.Buffer
		tmpl, err := template.New("html_template").Parse(c.Template.HTML)
		if err != nil {
			Logger.Println(err)
		}
		err = tmpl.Execute(&buff, t)
		if err != nil {
			Logger.Println(err)
		}
		e.HTML = buff.Bytes()
		buff.Reset()
		tmpl, err = template.New("text_template").Parse(c.Template.Text)
		if err != nil {
			Logger.Println(err)
		}
		err = tmpl.Execute(&buff, t)
		if err != nil {
			Logger.Println(err)
		}
		e.Text = buff.Bytes()
		buff.Reset()
		Logger.Println("Creating email using template")
		e.To = []string{t.Email}
		err = e.Send(c.SMTP.Host, auth)
		if err != nil {
			Logger.Println(err)
			err = t.UpdateStatus("Error")
			if err != nil {
				Logger.Println(err)
			}
		} else {
			err = t.UpdateStatus(models.EVENT_SENT)
			if err != nil {
				Logger.Println(err)
			}
		}
		Logger.Printf("Sending Email to %s\n", t.Email)
	}
}
