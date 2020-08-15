package models

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gophish/gomail"
	"github.com/gophish/gophish/config"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/mailer"
)

// MaxSendAttempts set to 8 since we exponentially backoff after each failed send
// attempt. This will give us a maximum send delay of 256 minutes, or about 4.2 hours.
var MaxSendAttempts = 8

// ErrMaxSendAttempts is thrown when the maximum number of sending attempts for a given
// MailLog is exceeded.
var ErrMaxSendAttempts = errors.New("max send attempts exceeded")

// processAttachment is used to to keep track of which email attachments have templated values.
// This allows us to skip re-templating attach
var processAttachment = map[[20]byte]bool{} // Considered using attachmentLookup[campaignid][filehash] but given the low number of files current approach should be fine

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

	cachedCampaign *Campaign
}

// GenerateMailLog creates a new maillog for the given campaign and
// result. It sets the initial send date to match the campaign's launch date.
func GenerateMailLog(c *Campaign, r *Result, sendDate time.Time) error {
	m := &MailLog{
		UserId:     c.UserId,
		CampaignId: c.Id,
		RId:        r.RId,
		SendDate:   sendDate,
	}
	return db.Save(m).Error
}

// Backoff sets the MailLog SendDate to be the next entry in an exponential
// backoff. ErrMaxRetriesExceeded is thrown if this maillog has been retried
// too many times. Backoff also unlocks the maillog so that it can be processed
// again in the future.
func (m *MailLog) Backoff(reason error) error {
	r, err := GetResult(m.RId)
	if err != nil {
		return err
	}
	if m.SendAttempt == MaxSendAttempts {
		r.HandleEmailError(ErrMaxSendAttempts)
		return ErrMaxSendAttempts
	}
	// Add an error, since we had to backoff because of a
	// temporary error of some sort during the SMTP transaction
	m.SendAttempt++
	backoffDuration := math.Pow(2, float64(m.SendAttempt))
	m.SendDate = m.SendDate.Add(time.Minute * time.Duration(backoffDuration))
	err = db.Save(m).Error
	if err != nil {
		return err
	}
	err = r.HandleEmailBackoff(reason, m.SendDate)
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

// Error sets the error status on the models.Result that the
// maillog refers to. Since MailLog errors are permanent,
// this action also deletes the maillog.
func (m *MailLog) Error(e error) error {
	r, err := GetResult(m.RId)
	if err != nil {
		log.Warn(err)
		return err
	}
	err = r.HandleEmailError(e)
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
	err = r.HandleEmailSent()
	if err != nil {
		return err
	}
	err = db.Delete(m).Error
	return err
}

// GetDialer returns a dialer based on the maillog campaign's SMTP configuration
func (m *MailLog) GetDialer() (mailer.Dialer, error) {
	c := m.cachedCampaign
	if c == nil {
		campaign, err := GetCampaignMailContext(m.CampaignId, m.UserId)
		if err != nil {
			return nil, err
		}
		c = &campaign
	}
	return c.SMTP.GetDialer()
}

// CacheCampaign allows bulk-mail workers to cache the otherwise expensive
// campaign lookup operation by providing a pointer to the campaign here.
func (m *MailLog) CacheCampaign(campaign *Campaign) error {
	if campaign.Id != m.CampaignId {
		return fmt.Errorf("incorrect campaign provided for caching. expected %d got %d", m.CampaignId, campaign.Id)
	}
	m.cachedCampaign = campaign
	return nil
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
	c := m.cachedCampaign
	if c == nil {
		campaign, err := GetCampaignMailContext(m.CampaignId, m.UserId)
		if err != nil {
			return err
		}
		c = &campaign
	}

	f, err := mail.ParseAddress(c.SMTP.FromAddress)
	if err != nil {
		return err
	}
	msg.SetAddressHeader("From", f.Address, f.Name)

	ptx, err := NewPhishingTemplateContext(c, r.BaseRecipient, r.RId)
	if err != nil {
		return err
	}

	// Add the transparency headers
	msg.SetHeader("X-Mailer", config.ServerName)
	if conf.ContactAddress != "" {
		msg.SetHeader("X-Gophish-Contact", conf.ContactAddress)
	}

	// Add Message-Id header as described in RFC 2822.
	messageID, err := m.generateMessageID()
	if err != nil {
		return err
	}
	msg.SetHeader("Message-Id", messageID)

	// Parse the customHeader templates
	for _, header := range c.SMTP.Headers {
		key, err := ExecuteTemplate(header.Key, ptx)
		if err != nil {
			log.Warn(err)
		}

		value, err := ExecuteTemplate(header.Value, ptx)
		if err != nil {
			log.Warn(err)
		}

		// Add our header immediately
		msg.SetHeader(key, value)
	}

	// Parse remaining templates
	subject, err := ExecuteTemplate(c.Template.Subject, ptx)

	if err != nil {
		log.Warn(err)
	}
	// don't set Subject header if the subject is empty
	if len(subject) != 0 {
		msg.SetHeader("Subject", subject)
	}

	msg.SetHeader("To", r.FormatAddress())
	if c.Template.Text != "" {
		text, err := ExecuteTemplate(c.Template.Text, ptx)
		if err != nil {
			log.Warn(err)
		}
		msg.SetBody("text/plain", text)
	}
	if c.Template.HTML != "" {
		html, err := ExecuteTemplate(c.Template.HTML, ptx)
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
				//decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(a.Content))
				decoder, err := applyAttachmentTemplate(a, ptx)
				if err != nil {
					return err
				}
				_, err = io.Copy(w, decoder)
				return err
			}), gomail.SetHeader(h)
		}(a))
	}

	return nil
}

// applyAttachmentTemplate parses different attachment files and applies the supplied phishing template.
func applyAttachmentTemplate(a Attachment, ptx PhishingTemplateContext) (io.Reader, error) {

	fileContentsHash := sha1.Sum([]byte(a.Content)) // Hash of the file content
	var processedAttachment string                  // Attachment content to return

	decodedAttachment, err := base64.StdEncoding.DecodeString(a.Content) // Decode the attachment
	if err != nil {
		return nil, err
	}

	// Keep track of which files have no template variables so we don't parse them repeatidly
	if _, ok := processAttachment[fileContentsHash]; !ok {
		processAttachment[fileContentsHash] = true // Default to true to process a file
	}

	if processAttachment[fileContentsHash] == true {

		// Decided to use the file extension rather than the content type, as there seems to be quite
		//  a bit of variability with types. e.g sometimes a Word docx file would have:
		//   "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		fileExtension := filepath.Ext(a.Name)

		switch fileExtension {

		case ".docx", ".docm", ".pptx", ".xlsx", ".xlsm":
			// Most modern office formats are xml based and can be unarchived.
			// .docm and .xlsm files are comprised of xml, and a binary blob for the macro code

			// Create a new zip reader from the file
			zipReader, err := zip.NewReader(bytes.NewReader(decodedAttachment), int64(len(decodedAttachment)))
			if err != nil {
				return nil, err
			}

			newZipArchive := new(bytes.Buffer)
			zipWriter := zip.NewWriter(newZipArchive) // For writing the new archive

			// i. Read each file from the Word document archive
			// ii. Apply the template to it
			// iii. Add the templated content to a new zip Word archive
			fileContainedTemplatesVars := false
			for _, zipFile := range zipReader.File {
				ff, err := zipFile.Open()
				if err != nil {
					return nil, err
				}
				defer ff.Close()
				contents, err := ioutil.ReadAll(ff)
				if err != nil {
					return nil, err
				}
				subFileExtension := filepath.Ext(zipFile.Name)
				var tFile string
				if subFileExtension == ".xml" || subFileExtension == ".rels" { // Ignore other files, e.g binary ones and images
					// For each file apply the template.
					tFile, err = ExecuteTemplate(string(contents), ptx)
					if err != nil {
						return nil, err
					}
					// Check if the subfile changed. We only need this to be set once to know in the future to check the 'parent' file
					if tFile != string(contents) {
						fileContainedTemplatesVars = true
					}

				} else {
					tFile = string(contents) // Could move this to the declaration of tFile, but might be confusing to read
				}
				// Write new Word archive
				newZipFile, err := zipWriter.Create(zipFile.Name)
				if err != nil {
					zipWriter.Close() // Don't use defer when writing files https://www.joeshaw.org/dont-defer-close-on-writable-files/
					return nil, err
				}
				_, err = newZipFile.Write([]byte(tFile))
				if err != nil {
					zipWriter.Close()
					return nil, err
				}

			}

			// If no files in the archive had template variables, we set the 'parent' file to not be checked in the future
			if fileContainedTemplatesVars == false {
				processAttachment[fileContentsHash] = false
			}

			zipWriter.Close()
			processedAttachment = newZipArchive.String()

		case ".txt", ".html":
			processedAttachment, err = ExecuteTemplate(string(decodedAttachment), ptx)
		case ".pdf":
			// Todo.
			// See: https://stackoverflow.com/questions/8099927/tracking-code-into-a-pdf-or-postscript-file
		case ".exe":
			// Todo. Perhaps we ignore the .exe and build our own, with a simple callback to the server
			//  A special extension of 'exef' or some such might be useful in case users want to attach
			//  an actual exe file. Does anyone email exe files in 2020 ?
		default:
			// We have two options here; either apply template to all files, or none. Probably safer to err on the side of none.
			processedAttachment = string(decodedAttachment) // Option one: Do nothing
			//processedAttachment, err = ExecuteTemplate(string(decodedAttachment), ptx) // Option two: Template all files
		}
		// Handle err from all the switch statement ExecuteTemplate functions
		if err != nil {
			return nil, err
		}

		// Check if applying the template altered the file contents. If not, let's not apply the template again to that file.
		// This doesn't work very well with .docx etc files, as the unzipping and rezipping seems to alter them.
		if processedAttachment == string(decodedAttachment) {
			processAttachment[fileContentsHash] = false
		}

	} else {
		processedAttachment = string(decodedAttachment)
	}

	decoder := strings.NewReader(processedAttachment)
	return decoder, nil
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
	return db.Model(&MailLog{}).Update("processing", false).Error
}

var maxBigInt = big.NewInt(math.MaxInt64)

// generateMessageID generates and returns a string suitable for an RFC 2822
// compliant Message-ID, e.g.:
// <1444789264909237300.3464.1819418242800517193@DESKTOP01>
//
// The following parameters are used to generate a Message-ID:
// - The nanoseconds since Epoch
// - The calling PID
// - A cryptographically random int64
// - The sending hostname
func (m *MailLog) generateMessageID() (string, error) {
	t := time.Now().UnixNano()
	pid := os.Getpid()
	rint, err := rand.Int(rand.Reader, maxBigInt)
	if err != nil {
		return "", err
	}
	h, err := os.Hostname()
	// If we can't get the hostname, we'll use localhost
	if err != nil {
		h = "localhost.localdomain"
	}
	msgid := fmt.Sprintf("<%d.%d.%d@%s>", t, pid, rint, h)
	return msgid, nil
}
