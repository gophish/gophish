package imap

//package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
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
	Host   string
	TLS    bool
	User   string
	Pwd    string
	Folder string
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
		Host:   s.Host,
		TLS:    s.TLS,
		User:   s.Username,
		Pwd:    s.Password,
		Folder: s.Folder}

	imapclient, err := mailServer.newClient()
	if err != nil {
		log.Error(err.Error())
	} else {
		imapclient.Logout()
	}
	return err
}

// MarkAsUnread will set the UNSEEN flag on a supplied slice of SeqNums
func (mbox *Mailbox) MarkAsUnread(seqs []uint32) error {
	imapclient, err := mbox.newClient()
	if err != nil {
		return err
	}

	defer imapclient.Logout()

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqs...)

	item := imap.FormatFlagsOp(imap.RemoveFlags, true)
	flags := []interface{}{imap.SeenFlag}
	err = imapclient.Store(seqSet, item, flags, nil)
	if err != nil {
		return err
	}

	return nil

}

// DeleteEmails will delete emails from the supplied slice of SeqNums
func (mbox *Mailbox) DeleteEmails(seqs []uint32) error {
	imapclient, err := mbox.newClient()
	if err != nil {
		return err
	}

	defer imapclient.Logout()

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqs...)

	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.DeletedFlag}
	err = imapclient.Store(seqSet, item, flags, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetUnread will find all unread emails in the folder and return them as a list.
func (mbox *Mailbox) GetUnread(markAsRead, delete bool) ([]Email, error) {
	var emails []Email

	imapclient, err := mbox.newClient()
	if err != nil {
		return emails, fmt.Errorf("failed to create IMAP connection: %s", err)
	}

	defer imapclient.Logout()

	// Search for unread emails
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{"\\Seen"}
	seqs, err := imapclient.Search(criteria)
	if err != nil {
		return emails, err
	}

	if len(seqs) > 0 {
		seqset := new(imap.SeqSet)
		seqset.AddNum(seqs...)
		section := &imap.BodySectionName{}
		items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, section.FetchItem()} // Check this
		messages := make(chan *imap.Message)

		go func() {
			if err := imapclient.Fetch(seqset, items, messages); err != nil {
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
			tmp = strings.ReplaceAll(tmp, "\r", "")
			buf = []byte(tmp)

			rawBodyStream := bytes.NewReader(buf)
			em, err = email.NewEmailFromReader(rawBodyStream) // Parse with @jordanwright's library
			if err != nil {
				return emails, err
			}

			// Reload the reader 🔫
			rawBodyStream = bytes.NewReader(buf)
			mr, err := mail.CreateReader(rawBodyStream)
			if err != nil {
				return emails, err
			}

			// Step over each part of the email, parsing attachments and attaching them to Jordan's email
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				} else if err != nil {
					return emails, err
				}
				h := p.Header

				s, ok := h.(*mail.AttachmentHeader)
				if ok {
					filename, _ := s.Filename()
					typ, _, _ := s.ContentType()
					_, err := em.Attach(p.Body, filename, typ)
					if err != nil {
						return emails, err //Unable to attach file
					}
				}
			}

			emtmp := Email{Email: em, SeqNum: msg.SeqNum} // Not sure why msg.Uid is always 0, so swapped to sequence numbers
			emails = append(emails, emtmp)

		} // On to the next email
	} else {
		//log.Println("No new messages")
	}

	return emails, nil
}

// newClient will initiate a new IMAP connection with the given creds.
func (mbox *Mailbox) newClient() (*client.Client, error) {
	var imapclient *client.Client
	var err error
	if mbox.TLS {
		imapclient, err = client.DialTLS(mbox.Host, new(tls.Config))
		if err != nil {
			return imapclient, err
		}
	} else {
		imapclient, err = client.Dial(mbox.Host)
		if err != nil {
			return imapclient, err
		}
	}

	err = imapclient.Login(mbox.User, mbox.Pwd)
	if err != nil {
		return imapclient, err
	}

	_, err = imapclient.Select(mbox.Folder, mbox.ReadOnly)
	if err != nil {
		return imapclient, err
	}

	return imapclient, nil
}
