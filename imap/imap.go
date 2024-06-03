package imap

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/charset"
	"github.com/gophish/gophish/dialer"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"

	"github.com/jordan-wright/email"
)

// Client interface for IMAP interactions
type Client interface {
	Login(username, password string) (cmd *imap.Command, err error)
	Logout(timeout time.Duration) (cmd *imap.Command, err error)
	Select(name string, readOnly bool) (mbox *imap.MailboxStatus, err error)
	Store(seq *imap.SeqSet, item imap.StoreItem, value interface{}, ch chan *imap.Message) (err error)
	Fetch(seqset *imap.SeqSet, items []imap.FetchItem, ch chan *imap.Message) (err error)
}

// Email represents an email.Email with an included IMAP Sequence Number
type Email struct {
	SeqNum uint32 `json:"seqnum"`
	*email.Email
}

// Mailbox holds onto the credentials and other information
// needed for connecting to an IMAP server.
type Mailbox struct {
	Host             string
	TLS              bool
	IgnoreCertErrors bool
	User             string
	Pwd              string
	Folder           string
	// Read only mode, false (original logic) if not initialized
	ReadOnly bool
}

// Validate validates supplied IMAP model by connecting to the server
func Validate(s *models.IMAP) error {
	err := s.Validate()
	if err != nil {
		log.Error(err)
		return err
	}

	s.Host = s.Host + ":" + strconv.Itoa(int(s.Port)) // Append port
	mailServer := Mailbox{
		Host:             s.Host,
		TLS:              s.TLS,
		IgnoreCertErrors: s.IgnoreCertErrors,
		User:             s.Username,
		Pwd:              s.Password,
		Folder:           s.Folder}

	imapClient, err := mailServer.newClient()
	if err != nil {
		log.Error(err.Error())
	} else {
		imapClient.Logout()
	}
	return err
}

// MarkAsUnread will set the UNSEEN flag on a supplied slice of SeqNums
func (mbox *Mailbox) MarkAsUnread(seqs []uint32) error {
	imapClient, err := mbox.newClient()
	if err != nil {
		return err
	}

	defer imapClient.Logout()

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqs...)

	item := imap.FormatFlagsOp(imap.RemoveFlags, true)
	err = imapClient.Store(seqSet, item, imap.SeenFlag, nil)
	if err != nil {
		return err
	}

	return nil

}

// DeleteEmails will delete emails from the supplied slice of SeqNums
func (mbox *Mailbox) DeleteEmails(seqs []uint32) error {
	imapClient, err := mbox.newClient()
	if err != nil {
		return err
	}

	defer imapClient.Logout()

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqs...)

	item := imap.FormatFlagsOp(imap.AddFlags, true)
	err = imapClient.Store(seqSet, item, imap.DeletedFlag, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetUnread will find all unread emails in the folder and return them as a list.
func (mbox *Mailbox) GetUnread(markAsRead, delete bool) ([]Email, error) {
	imap.CharsetReader = charset.Reader
	var emails []Email

	imapClient, err := mbox.newClient()
	if err != nil {
		return emails, fmt.Errorf("failed to create IMAP connection: %s", err)
	}

	defer imapClient.Logout()

	// Search for unread emails
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	seqs, err := imapClient.Search(criteria)
	if err != nil {
		return emails, err
	}

	if len(seqs) == 0 {
		return emails, nil
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(seqs...)
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, section.FetchItem()}
	messages := make(chan *imap.Message)

	go func() {
		if err := imapClient.Fetch(seqset, items, messages); err != nil {
			log.Error("Error fetching emails: ", err.Error()) // TODO: How to handle this, need to propogate error out
		}
	}()

	// Step through each email
	for msg := range messages {
		// Extract raw message body. I can't find a better way to do this with the emersion library
		var em *email.Email
		var buf []byte
		for _, value := range msg.Body {
			buf = make([]byte, value.Len())
			value.Read(buf)
			break // There should only ever be one item in this map, but I'm not 100% sure
		}

		//Remove CR characters, see https://github.com/jordan-wright/email/issues/106
		tmp := string(buf)
		re := regexp.MustCompile(`\r`)
		tmp = re.ReplaceAllString(tmp, "")
		buf = []byte(tmp)

		rawBodyStream := bytes.NewReader(buf)
		em, err = email.NewEmailFromReader(rawBodyStream) // Parse with @jordanwright's library
		if err != nil {
			return emails, err
		}

		emtmp := Email{Email: em, SeqNum: msg.SeqNum} // Not sure why msg.Uid is always 0, so swapped to sequence numbers
		emails = append(emails, emtmp)

	}
	return emails, nil
}

// newClient will initiate a new IMAP connection with the given creds.
func (mbox *Mailbox) newClient() (*client.Client, error) {
	var imapClient *client.Client
	var err error
	restrictedDialer := dialer.Dialer()
	if mbox.TLS {
		config := new(tls.Config)
		config.InsecureSkipVerify = mbox.IgnoreCertErrors
		imapClient, err = client.DialWithDialerTLS(restrictedDialer, mbox.Host, config)
	} else {
		imapClient, err = client.DialWithDialer(restrictedDialer, mbox.Host)
	}
	if err != nil {
		return imapClient, err
	}

	err = imapClient.Login(mbox.User, mbox.Pwd)
	if err != nil {
		return imapClient, err
	}

	_, err = imapClient.Select(mbox.Folder, mbox.ReadOnly)
	if err != nil {
		return imapClient, err
	}

	return imapClient, nil
}
