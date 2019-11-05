package imap

// Functionality taken from https://github.com/jprobinson/eazye
// TODO: Remove any functions not used by monitor.go

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/mail"
	"strconv"
	"strings"
	"time"

	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/mxk/go-imap/imap"
	"github.com/paulrosania/go-charset/charset"

	_ "github.com/paulrosania/go-charset/data"
	qprintable "github.com/sloonz/go-qprintable"
	"golang.org/x/net/html"
)

// MailboxInfo holds onto the credentials and other information
// needed for connecting to an IMAP server.
type MailboxInfo struct {
	Host   string
	TLS    bool
	User   string
	Pwd    string
	Folder string
	// Read only mode, false (original logic) if not initialized
	ReadOnly bool
}

// GetAll will pull all emails from the email folder and return them as a list.
func GetAll(info MailboxInfo, markAsRead, delete bool) ([]Email, error) {
	// call chan, put 'em in a list, return
	var emails []Email
	responses, err := GenerateAll(info, markAsRead, delete)
	if err != nil {
		return emails, err
	}

	for resp := range responses {
		if resp.Err != nil {
			return emails, resp.Err
		}
		emails = append(emails, resp.Email)
	}

	return emails, nil
}

// GenerateAll will find all emails in the email folder and pass them along to the responses channel.
func GenerateAll(info MailboxInfo, markAsRead, delete bool) (chan Response, error) {
	return generateMail(info, "ALL", nil, markAsRead, delete)
}

// GetUnread will find all unread emails in the folder and return them as a list.
func GetUnread(info MailboxInfo, markAsRead, delete bool) ([]Email, error) {
	// call chan, put 'em in a list, return
	var emails []Email

	responses, err := GenerateUnread(info, markAsRead, delete)
	if err != nil {
		return emails, err
	}

	for resp := range responses {
		if resp.Err != nil {
			return emails, resp.Err
		}
		emails = append(emails, resp.Email)
	}

	return emails, nil
}

// GenerateUnread will find all unread emails in the folder and pass them along to the responses channel.
func GenerateUnread(info MailboxInfo, markAsRead, delete bool) (chan Response, error) {
	return generateMail(info, "UNSEEN", nil, markAsRead, delete)
}

// GetSince will pull all emails that have an internal date after the given time.
func GetSince(info MailboxInfo, since time.Time, markAsRead, delete bool) ([]Email, error) {
	var emails []Email
	responses, err := GenerateSince(info, since, markAsRead, delete)
	if err != nil {
		return emails, err
	}

	for resp := range responses {
		if resp.Err != nil {
			return emails, resp.Err
		}
		emails = append(emails, resp.Email)
	}

	return emails, nil
}

// GenerateSince will find all emails that have an internal date after the given time and pass them along to the
// responses channel.
func GenerateSince(info MailboxInfo, since time.Time, markAsRead, delete bool) (chan Response, error) {
	return generateMail(info, "", &since, markAsRead, delete)
}

// MarkAsUnread will set the UNSEEN flag on a supplied slice of UIDs
func MarkAsUnread(info MailboxInfo, uids []uint32) error {

	client, err := newIMAPClient(info)
	if err != nil {
		return err
	}
	defer func() {
		client.Close(true)
		client.Logout(30 * time.Second)
	}()
	for _, u := range uids {
		err := alterEmail(client, u, "\\SEEN", false)
		if err != nil {
			return err //return on first failure
		}
	}
	return nil

}

// DeleteEmails will delete emails from the supplied slice of UIDs
func DeleteEmails(info MailboxInfo, uids []uint32) error {

	client, err := newIMAPClient(info)
	if err != nil {
		return err
	}
	defer func() {
		client.Close(true)
		client.Logout(30 * time.Second)
	}()
	for _, u := range uids {
		err := deleteEmail(client, u)
		if err != nil {
			return err //return on first failure
		}
	}
	return nil

}

// ValidateIMAP validates supplied IMAP model by connecting to the server
func ValidateIMAP(s *models.IMAP) error {

	err := s.Validate()
	if err != nil {
		log.Error(err)
		return err
	}

	s.Host = s.Host + ":" + strconv.Itoa(int(s.Port)) // Append port
	mailSettings := MailboxInfo{
		Host:   s.Host,
		TLS:    s.TLS,
		User:   s.Username,
		Pwd:    s.Password,
		Folder: s.Folder}

	client, err := newIMAPClient(mailSettings)
	if err != nil {
		log.Error(err.Error())
	} else {
		client.Close(true)
		client.Logout(30 * time.Second)
	}
	return err
}

// Email is a simplified email struct containing the basic pieces of an email. If you want more info,
// it should all be available within the Message attribute.
type Email struct {
	Message *mail.Message

	From         *mail.Address   `json:"from"`
	To           []*mail.Address `json:"to"`
	InternalDate time.Time       `json:"internal_date"`
	Precedence   string          `json:"precedence"`
	Subject      string          `json:"subject"`
	HTML         []byte          `json:"html"`
	Text         []byte          `json:"text"`
	IsMultiPart  bool            `json:"is_multipart"`
	UID          uint32          `json:"uid"`
}

var (
	styleTag       = []byte("style")
	scriptTag      = []byte("script")
	headTag        = []byte("head")
	metaTag        = []byte("meta")
	doctypeTag     = []byte("doctype")
	shapeTag       = []byte("v:shape")
	imageDataTag   = []byte("v:imagedata")
	commentTag     = []byte("!")
	nonVisibleTags = [][]byte{
		styleTag,
		scriptTag,
		headTag,
		metaTag,
		doctypeTag,
		shapeTag,
		imageDataTag,
		commentTag,
	}
)

func VisibleText(body io.Reader) ([][]byte, error) {
	var (
		text [][]byte
		skip bool
		err  error
	)
	z := html.NewTokenizer(body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if err = z.Err(); err == io.EOF {
				return text, nil
			}
			return text, err
		case html.TextToken:
			if !skip {
				tmp := bytes.TrimSpace(z.Text())
				if len(tmp) == 0 {
					continue
				}
				tagText := make([]byte, len(tmp))
				copy(tagText, tmp)
				text = append(text, tagText)
			}
		case html.StartTagToken, html.EndTagToken:
			tn, _ := z.TagName()
			for _, nvTag := range nonVisibleTags {
				if bytes.Equal(tn, nvTag) {
					skip = (tt == html.StartTagToken)
					break
				}
			}
		}
	}
	return text, nil
}

// VisibleText will return any visible text from an HTML
// email body.
func (e *Email) VisibleText() ([][]byte, error) {
	// if theres no HTML, just return text
	if len(e.HTML) == 0 {
		return [][]byte{e.Text}, nil
	}
	return VisibleText(bytes.NewReader(e.HTML))
}

// String is to spit out a somewhat pretty version of the email.
func (e *Email) String() string {
	return fmt.Sprintf(`
----------------------------
From:           %s
To:             %s
Internal Date:  %s
Precedence:     %s
Subject:        %s
HTML:           %s

Text:           %s
----------------------------

`,
		e.From,
		e.To,
		e.InternalDate,
		e.Precedence,
		e.Subject,
		string(e.HTML),
		string(e.Text),
	)
}

// Response is a helper struct to wrap the email responses and possible errors.
type Response struct {
	Email Email
	Err   error
}

// newIMAPClient will initiate a new IMAP connection with the given creds.
func newIMAPClient(info MailboxInfo) (*imap.Client, error) {
	var client *imap.Client
	var err error
	if info.TLS {
		client, err = imap.DialTLS(info.Host, new(tls.Config))
		if err != nil {
			return client, err
		}
	} else {
		client, err = imap.Dial(info.Host)
		if err != nil {
			return client, err
		}
	}

	_, err = client.Login(info.User, info.Pwd)
	if err != nil {
		return client, err
	}

	_, err = imap.Wait(client.Select(info.Folder, info.ReadOnly))
	if err != nil {
		return client, err
	}

	return client, nil
}

const dateFormat = "02-Jan-2006"

// findEmails will run a find the UIDs of any emails that match the search.:
func findEmails(client *imap.Client, search string, since *time.Time) (*imap.Command, error) {
	var specs []imap.Field
	if len(search) > 0 {
		specs = append(specs, search)
	}

	if since != nil {
		sinceStr := since.Format(dateFormat)
		specs = append(specs, "SINCE", sinceStr)
	}

	// get headers and UID for UnSeen message in src inbox...
	cmd, err := imap.Wait(client.UIDSearch(specs...))
	if err != nil {
		return &imap.Command{}, fmt.Errorf("uid search failed: %s", err)
	}
	return cmd, nil
}

var GenerateBufferSize = 100

func generateMail(info MailboxInfo, search string, since *time.Time, markAsRead, delete bool) (chan Response, error) {
	responses := make(chan Response, GenerateBufferSize)
	client, err := newIMAPClient(info)
	if err != nil {
		close(responses)
		return responses, fmt.Errorf("failed to create IMAP connection: %s", err)
	}

	go func() {
		defer func() {
			client.Close(true)
			client.Logout(30 * time.Second)
			close(responses)
		}()

		var cmd *imap.Command
		// find all the UIDs
		cmd, err = findEmails(client, search, since)
		if err != nil {
			responses <- Response{Err: err}
			return
		}
		// gotta fetch 'em all
		getEmails(client, cmd, markAsRead, delete, responses)
	}()

	return responses, nil
}

func getEmails(client *imap.Client, cmd *imap.Command, markAsRead, delete bool, responses chan Response) {
	seq := &imap.SeqSet{}
	msgCount := 0
	for _, rsp := range cmd.Data {
		for _, uid := range rsp.SearchResults() {
			msgCount++
			seq.AddNum(uid)
		}
	}

	// nothing to request?! why you even callin me, foolio?
	if seq.Empty() {
		return
	}

	fCmd, err := imap.Wait(client.UIDFetch(seq, "INTERNALDATE", "BODY[]", "UID", "RFC822.HEADER"))
	if err != nil {
		responses <- Response{Err: fmt.Errorf("unable to perform uid fetch: %s", err)}
		return
	}

	var email Email
	for _, msgData := range fCmd.Data {
		msgFields := msgData.MessageInfo().Attrs

		// make sure is a legit response before we attempt to parse it
		// deal with unsolicited FETCH responses containing only flags
		// I'm lookin' at YOU, Gmail!
		// http://mailman13.u.washington.edu/pipermail/imap-protocol/2014-October/002355.html
		// http://stackoverflow.com/questions/26262472/gmail-imap-is-sometimes-returning-bad-results-for-fetch
		if _, ok := msgFields["RFC822.HEADER"]; !ok {
			continue
		}

		email, err = NewEmail(msgFields)
		if err != nil {
			responses <- Response{Err: fmt.Errorf("unable to parse email: %s", err)}
			return
		}

		responses <- Response{Email: email}

		if !markAsRead {
			err = removeSeen(client, imap.AsNumber(msgFields["UID"]))
			if err != nil {
				responses <- Response{Err: fmt.Errorf("unable to remove seen flag: %s", err)}
				return
			}
		}

		if delete {
			err = deleteEmail(client, imap.AsNumber(msgFields["UID"]))
			if err != nil {
				responses <- Response{Err: fmt.Errorf("unable to delete email: %s", err)}
				return
			}
		}
	}
	return
}

func deleteEmail(client *imap.Client, UID uint32) error {
	return alterEmail(client, UID, "\\DELETED", true)
}

func removeSeen(client *imap.Client, UID uint32) error {
	return alterEmail(client, UID, "\\SEEN", false)
}

func alterEmail(client *imap.Client, UID uint32, flag string, plus bool) error {
	flg := "-FLAGS"
	if plus {
		flg = "+FLAGS"
	}
	fSeq := &imap.SeqSet{}
	fSeq.AddNum(UID)
	_, err := imap.Wait(client.UIDStore(fSeq, flg, flag))
	if err != nil {
		return err
	}

	return nil
}

func hasEncoding(word string) bool {
	return strings.Contains(word, "=?") && strings.Contains(word, "?=")
}

func isEncodedWord(word string) bool {
	return strings.HasPrefix(word, "=?") && strings.HasSuffix(word, "?=") && strings.Count(word, "?") == 4
}

func parseSubject(subject string) string {
	if !hasEncoding(subject) {
		return subject
	}

	dec := mime.WordDecoder{}
	sub, _ := dec.DecodeHeader(subject)
	return sub
}

// NewEmail will parse an imap.FieldMap into an Email. This
// will expect the message to container the internaldate and the body with
// all headers included.
func NewEmail(msgFields imap.FieldMap) (Email, error) {
	var email Email
	// parse the header
	var message bytes.Buffer
	message.Write(imap.AsBytes(msgFields["RFC822.HEADER"]))
	message.Write([]byte("\n\n"))
	rawBody := imap.AsBytes(msgFields["BODY[]"])
	message.Write(rawBody)
	msg, err := mail.ReadMessage(&message)
	if err != nil {
		return email, fmt.Errorf("unable to read header: %s", err)
	}

	from, err := mail.ParseAddress(msg.Header.Get("From"))
	if err != nil {
		return email, fmt.Errorf("unable to parse from address: %s", err)
	}

	to, err := mail.ParseAddressList(msg.Header.Get("To"))
	if err != nil {
		to = []*mail.Address{}
	}

	email = Email{
		Message:      msg,
		InternalDate: imap.AsDateTime(msgFields["INTERNALDATE"]),
		Precedence:   msg.Header.Get("Precedence"),
		From:         from,
		To:           to,
		Subject:      parseSubject(msg.Header.Get("Subject")),
		UID:          imap.AsNumber(msgFields["UID"]),
	}

	// chunk the body up into simple chunks
	email.HTML, email.Text, email.IsMultiPart, err = parseBody(msg.Header, rawBody)
	return email, err
}

var headerSplitter = []byte("\r\n\r\n")

// parseBody will accept a a raw body, break it into all its parts and then convert the
// message to UTF-8 from whatever charset it may have.
func parseBody(header mail.Header, body []byte) (html []byte, text []byte, isMultipart bool, err error) {
	var mediaType string
	var params map[string]string
	mediaType, params, err = mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
		return
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		isMultipart = true
		mr := multipart.NewReader(bytes.NewReader(body), params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}

			slurp, err := ioutil.ReadAll(p)
			if err != nil {
				// error and no results to use
				if len(slurp) == 0 {
					break
				}
			}

			partMediaType, partParams, err := mime.ParseMediaType(p.Header.Get("Content-Type"))
			if err != nil {
				break
			}

			var htmlT, textT []byte
			htmlT, textT, err = parsePart(partMediaType, partParams["charset"], p.Header.Get("Content-Transfer-Encoding"), slurp)
			if len(htmlT) > 0 {
				html = htmlT
			} else {
				text = textT
			}
		}
	} else {

		splitBody := bytes.SplitN(body, headerSplitter, 2)
		if len(splitBody) < 2 {
			err = errors.New("unexpected email format. (single part and no \\r\\n\\r\\n separating headers/body")
			return
		}

		body = splitBody[1]
		html, text, err = parsePart(mediaType, params["charset"], header.Get("Content-Transfer-Encoding"), body)
	}
	return
}

func parsePart(mediaType, charsetStr, encoding string, part []byte) (html, text []byte, err error) {
	// deal with charset
	if strings.ToLower(charsetStr) == "iso-8859-1" {
		var cr io.Reader
		cr, err = charset.NewReader("latin1", bytes.NewReader(part))
		if err != nil {
			return
		}

		part, err = ioutil.ReadAll(cr)
		if err != nil {
			return
		}
	}

	// deal with encoding
	var body []byte
	switch strings.ToLower(encoding) {
	case "quoted-printable":
		dec := qprintable.NewDecoder(qprintable.WindowsTextEncoding, bytes.NewReader(part))
		body, err = ioutil.ReadAll(dec)
		if err != nil {
			return
		}
	case "base64":
		decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(part))
		body, err = ioutil.ReadAll(decoder)
		if err != nil {
			return
		}
	default:
		body = part
	}

	// deal with media type
	mediaType = strings.ToLower(mediaType)
	switch {
	case strings.Contains(mediaType, "text/html"):
		html = body
	case strings.Contains(mediaType, "text/plain"):
		text = body
	}
	return
}
