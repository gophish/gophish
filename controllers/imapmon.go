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
	"time"

	log "github.com/gophish/gophish/logger"

	"github.com/glennzw/eazye"
	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/models"
)

// Pattern for GoPhish emails e.g ?rid=AbC123
var goPhishRegex = regexp.MustCompile("(\\?rid=[A-Za-z0-9]{7})")

// ImapMonitor is a worker that monitors IMAP servers for reported campaign emails
type ImapMonitor struct {
	cancel    func()
	reportURL string
}

// ImapMonitor.start() checks for campaign emails
// As each account can have its own polling frequency set we need to run one Go routine for
// each, as well as keeping an eye on newly created user accounts.
func (im *ImapMonitor) start(ctx context.Context) {

	usermap := make(map[int64]int) // Keep track of running go routines, one per user. We assume incrementing non-repeating UIDs (for the case where users are deleted and re-added).

	for {
		select {
		case <-ctx.Done():
			return
		default:
			dbusers, err := models.GetUsers() //Slice of all user ids. Each user gets their own IMAP monitor routine.
			if err != nil {
				log.Error(err)
				break
			}
			for _, dbuser := range dbusers {
				if _, ok := usermap[dbuser.Id]; !ok { // If we don't currently have a running Go routine for this user, start one.
					log.Info("Starting new IMAP monitor for user ", dbuser.Username)
					usermap[dbuser.Id] = 1
					go monitorIMAP(dbuser.Id, ctx, im.reportURL)
				}
			}
			time.Sleep(10 * time.Second) // Every ten seconds we check if a new user has been created
		}
	}
}

// monitorIMAP will continuously login to the IMAP settings associated to the supplied user id (if the user account has IMAP settings, and they're enabled.)
// It also verifies the user account exists, and returns if not (for the case of a user being deleted).
func monitorIMAP(uid int64, ctx context.Context, reportURL string) {

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 1. Check if user exists, if not, return.
			_, err := models.GetUser(uid)
			if err != nil { // Not sure if there's a better way to determine user existence via id.
				log.Info("User ", uid, " seems to have been deleted. Stopping IMAP monitor for this user.")
				return
			}
			// 2. Check if user has IMAP settings.
			imapSettings, err := models.GetIMAP(uid)
			if err != nil {
				log.Error(err)
				break
			}
			if len(imapSettings) > 0 {
				im := imapSettings[0]
				// 3. Check if IMAP is enabled
				if im.Enabled {
					log.Debug("Checking IMAP for user ", uid, ": ", im.Username, "@", im.Host)
					checkForNewEmails(im, reportURL)
					time.Sleep((time.Duration(im.IMAPFreq) - 10) * time.Second) // Subtract 10 to compensate for the default sleep of 10 at the bottom
				}
			}
		}
		time.Sleep(10 * time.Second)
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
	}
	return im
}

// Start launches the IMAP campaign monitor
func (im *ImapMonitor) Start() error {
	log.Info("Starting IMAP monitor manager")
	ctx, cancel := context.WithCancel(context.Background()) // ctx is the derivedContext
	im.cancel = cancel
	go im.start(ctx)
	return nil
}

// Shutdown attempts to gracefully shutdown the IMAP monitor.
func (im *ImapMonitor) Shutdown() error {
	log.Info("Shutting down IMAP monitor manager")
	im.cancel()
	return nil
}

// checkForNewEmails logs into an IMAP account and checks unread emails
//  for the rid campaign identifier.
func checkForNewEmails(im models.IMAP, reportURL string) {

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
		var campaignEmails []uint32  // UIDs of campaign emails. If DeleteReportedCampaignEmail is true, we will delete these
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
						if im.DeleteReportedCampaignEmail == true {
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
			if im.DeleteReportedCampaignEmail == true && len(campaignEmails) > 0 {
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
