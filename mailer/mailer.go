package mailer

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/models"

	"gopkg.in/gomail.v2"
)

var MaxReconnectAttempts = 10

// ErrMaxConnectAttempts is thrown when the maximum number of reconnect attempts
// is reached.
var ErrMaxConnectAttempts = errors.New("max connection attempts reached")

// Logger is the logger for the worker
var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

// Dialer dials to an SMTP server and returns the SendCloser
type Dialer interface {
	Dial() (gomail.SendCloser, error)
}

// MockDialer keeps track of calls to Dial
type MockDialer struct {
	dialCount int
}

func (md *MockDialer) Dial() (gomail.SendCloser, error) {
	md.dialCount++
	return nil, nil
}

// MockMessage holds the information sent via a call to MockClient.Send()
type MockMessage struct {
	from    string
	to      string
	message io.WriterTo
}

// MockClient is a mock gomail Sender used for testing.
type MockClient struct {
	messages []MockMessage
}

// Send just appends the provided message record to the internal slice
func (m *MockClient) Send(msg *gomail.Message) error {
	m.messages = append(m.messages, MockMessage{
		from:    msg.GetHeader("From")[0],
		to:      msg.GetHeader("To")[0],
		message: msg,
	})
	return nil
}

// Close is a noop for the mock client
func (m *MockClient) Close() error {
	return nil
}

type SMTPDialer struct {
	dialer    *gomail.Dialer
	dialCount int
}

// Dial increments the internal dialCount and attempts to connect
// to the SMTP server, returning a gomail.SendCloser. Returns
// ErrMaxConnectAttempts if the maximum number of reconnect attempts is
// exceeded.
func (s *SMTPDialer) Dial() (gomail.SendCloser, error) {
	s.dialCount++
	for s.dialCount <= MaxReconnectAttempts {
		sc, err := s.dialer.Dial()
		if err != nil {
			Logger.Println(err)
			continue
		}
		return sc, err
	}
	return nil, ErrMaxConnectAttempts
}

// NewDialer returns a dialer based on the given SMTP information
func NewDialer(smtp models.SMTP) Dialer {
	if config.Conf.TestFlag {
		return &MockDialer{}
	}
	hp := strings.Split(smtp.Host, ":")
	if len(hp) < 2 {
		hp = append(hp, "25")
	}
	// Any issues should have been caught in validation, so we just log
	// errors and set a reasonable default
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		Logger.Println(err)
		port = 25
	}
	d := gomail.NewDialer(hp[0], port, smtp.Username, smtp.Password)
	d.TLSConfig = &tls.Config{
		ServerName:         smtp.Host,
		InsecureSkipVerify: smtp.IgnoreCertErrors,
	}
	return &SMTPDialer{
		dialer: d,
	}
}
