package controllers

/* TODO:
*		 - Have a counter per config for number of consecutive login errors and backoff (e.g if supplied creds are incorrect)
*		 - Have a DB field "last_login_error" if last login failed
*		 - DB counter for non-campaign emails that the admin should investigate
*		 - Add field to User for numner of non-campaign emails reported
 */
import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	log "github.com/gophish/gophish/logger"

	"github.com/glennzw/eazye"
	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/models"
)

// Pattern for GoPhish emails e.g ?rid=AbC123
var goPhishRegex = regexp.MustCompile("(\\?rid=[A-Za-z0-9]{7})")

// ImapMonitor is a worker that monitors IMAP serverd for reported campaign emails
type ImapMonitor struct {
	cancel    func()
	reportURL string
	checkFreq int
}

// PeriodicIMAPMonitor will periodically check the database for IMAP instances to login to and check for campaigns
func PeriodicIMAPMonitor(ctx context.Context, im *ImapMonitor) {

	for {
		select {
		case <-ctx.Done():
			return
		default:
			imaps, err := models.GetEnabledIMAPs()
			if err != nil {
				log.Error(err)
				break
			}
			var wg sync.WaitGroup
			wg.Add(len(imaps))
			for _, i := range imaps {
				// We launch a goroutine for each IMAP account. In the future should use some more sophisticated worker/queues
				go checkForNewEmails(i, &wg, im.reportURL)
			}
			wg.Wait()
			time.Sleep(time.Duration(im.checkFreq) * time.Second)
		}
	}

}

// NewImapMonitor returns a new instance of the ImapMonitor
func NewImapMonitor(config *config.Config) *ImapMonitor {
	// Make sure database connection exists. Not sure why I have to do this here, but
	//  otherwise db is <nil> when calling models.GetEnabledIMAPs()
	err := models.Setup(config)
	if err != nil {
		log.Fatal(err)
	}
	reportURL := "http://" + config.PhishConf.ListenURL
	if config.PhishConf.UseTLS {
		reportURL = "https://" + config.PhishConf.ListenURL
	}

	im := &ImapMonitor{
		reportURL: reportURL,
		checkFreq: config.IMAPFreq,
	}
	return im
}

// Start launches the IMAP campaign monitor
func (im *ImapMonitor) Start() error {
	log.Info("Starting IMAP monitor")
	ctx, cancel := context.WithCancel(context.Background()) // ctx is the derivedContext
	im.cancel = cancel
	go PeriodicIMAPMonitor(ctx, im)
	return nil
}

// Shutdown attempts to gracefully shutdown the IMAP monitor.
func (im *ImapMonitor) Shutdown() error {
	log.Info("Shutting down IMAP monitor")
	im.cancel()
	return nil
}

// checkForNewEmails logs into an IMAP account and checks unread emails
//  for the rid campaign identifier.
func checkForNewEmails(im models.IMAP, wg *sync.WaitGroup, reportURL string) {
	defer wg.Done()
	mailSettings := eazye.MailboxInfo{
		Host:   im.Host,
		TLS:    im.TLS,
		User:   im.Username,
		Pwd:    im.Password,
		Folder: im.Folder}

	// Make sure server ends with a trailing slash
	if reportURL[len(reportURL)-1:] != "/" {
		reportURL = reportURL + "/"
	}
	// Set http timeout to 10 seconds, or routine may hang
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	msgs, err := eazye.GetUnread(mailSettings, true, false)
	if err != nil {
		log.Error(err)
		return
	}
	// Update last_succesful_login here via im.Host
	err = models.SuccessfulLogin(&im)

	if len(msgs) > 0 {
		var reportingFailed []uint32 // UIDs of emails that were unable to be reported to phishing server, mark as unread
		var campaignEmails []uint32  // UIDs of campaign emails. If deleteCampaignEmails is true, we will delete these
		for _, m := range msgs {
			// Check if sender is from company's domain, if enabled. TODO: Make this an IMAP filter
			if im.RestrictDomain != "" { // e.g domainResitct = widgets.com
				splitEmail := strings.Split(m.From.Address, "@")
				senderDomain := splitEmail[len(splitEmail)-1]
				if senderDomain != im.RestrictDomain {
					log.Debug("Ignoring email as not from company domain: ", senderDomain)
					continue
				}
			}

			body := string(append(m.Text, m.HTML...)) // Empty body is being returned for emails forwarded from ProtonMail. Need to investigate more. Perhaps it's being attached, and eazye doesn't support attachments.
			rid := goPhishRegex.FindString(body)

			if rid != "" {
				rid := strings.TrimSpace(rid)
				reportURL := fmt.Sprintf("%sreport%s", reportURL, rid)
				response, err := netClient.Get(reportURL)
				if err != nil {
					log.Error("Error reporting GoPhish email, marking as unread. ", err.Error())
					reportingFailed = append(reportingFailed, m.UID)

				} else {
					if response.StatusCode == 204 {
						if im.DeleteCampaign == true {
							campaignEmails = append(campaignEmails, m.UID)
						}
						log.Debugf("User '%s' reported GoPhish campaign email with subject '%s'\n", m.From, m.Subject)
					} else {
						log.Error("Failed to report campaign. Server did not recogise rid: " + rid)
					}
				}
			} else {
				log.Debugf("User '%s' reported email with subject '%s'. This is not a GoPhish campaign; you should investigate it.\n", m.From, m.Subject)
			}
			if len(reportingFailed) > 0 {
				log.Debugf("Marking %d emails as unread as failed to report\n", len(reportingFailed))
				err := eazye.MarkAsUnread(mailSettings, reportingFailed) // Set emails as unread that we failed to report to GoPhish
				if err != nil {
					log.Error("Unable to mark emails as unread: ", err.Error())
				}
			}
			if im.DeleteCampaign == true && len(campaignEmails) > 0 {
				fmt.Printf("Deleting %d campaign emails\n", len(campaignEmails))
				log.Debugf("Deleting %d campaign emails\n", len(campaignEmails))
				err := eazye.DeleteEmails(mailSettings, campaignEmails) // Delete GoPhish campaign emails.
				if err != nil {
					log.Error("Failed to delete emails: ", err.Error())
				}
			}
		}

	} else {
		log.Debug("No new emails for ", im.Username)
	}
}
