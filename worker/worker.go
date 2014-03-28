package worker

import (
	"log"
	"os"

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
	err := models.UpdateCampaignStatus(c, models.IN_PROGRESS)
	if err != nil {
		Logger.Println(err)
	}
	for _, t := range c.Results {
		Logger.Println("Creating email using template")
		/*e := email.Email{
			Text: []byte(c.Template.Text),
			HTML: []byte(c.Template.Html),
		}*/
		Logger.Printf("Sending Email to %s\n", t.Email)
	}
}
