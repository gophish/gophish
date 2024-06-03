package mailer

import (
	"bytes"
	"errors"
	"io"
	"time"

	"github.com/gophish/gomail"
)

// errHostUnreachable is a mock error to represent a host
// being unreachable
var errHostUnreachable = errors.New("host unreachable")

// mockDialer keeps track of calls to Dial
type mockDialer struct {
	dialCount int
	dial      func() (Sender, error)
}

// newMockDialer returns a new instance of the mockDialer with the default
// dialer set.
func newMockDialer() *mockDialer {
	md := &mockDialer{}
	md.dial = md.defaultDial
	return md
}

// defaultDial simply returns a mockSender
func (md *mockDialer) defaultDial() (Sender, error) {
	return newMockSender(), nil
}

// unreachableDial is to simulate network error conditions in which
// a host is unavailable.
func (md *mockDialer) unreachableDial() (Sender, error) {
	return nil, errHostUnreachable
}

// Dial increments the internal dial count. Otherwise, it's a no-op for the mock client.
func (md *mockDialer) Dial() (Sender, error) {
	md.dialCount++
	return md.dial()
}

// setDial sets the Dial function for the mockDialer
func (md *mockDialer) setDial(dial func() (Sender, error)) {
	md.dial = dial
}

// mockSender is a mock gomail.Sender used for testing.
type mockSender struct {
	messages    []*mockMessage
	status      string
	send        func(*mockMessage) error
	messageChan chan *mockMessage
	resetCount  int
}

func newMockSender() *mockSender {
	ms := &mockSender{
		status:      "ehlo",
		messageChan: make(chan *mockMessage),
	}
	ms.send = ms.defaultSend
	return ms
}

func (ms *mockSender) setSend(send func(*mockMessage) error) {
	ms.send = send
}

func (ms *mockSender) defaultSend(mm *mockMessage) error {
	ms.messageChan <- mm
	return nil
}

// Send just appends the provided message record to the internal slice
func (ms *mockSender) Send(from string, to []string, msg io.WriterTo) error {
	mm := newMockMessage(from, to, msg)
	ms.messages = append(ms.messages, mm)
	ms.status = "sent"
	return ms.send(mm)
}

// Close is a noop for the mock client
func (ms *mockSender) Close() error {
	ms.status = "closed"
	close(ms.messageChan)
	return nil
}

// Reset sets the status to "Reset". In practice, this would reset the connection
// to the same state as if the client had just sent an EHLO command.
func (ms *mockSender) Reset() error {
	ms.status = "reset"
	ms.resetCount++
	return nil
}

// mockMessage holds the information sent via a call to MockClient.Send()
type mockMessage struct {
	from         string
	to           []string
	message      []byte
	sendAt       time.Time
	backoffCount int
	getdialer    func() (Dialer, error)
	err          error
	finished     bool
}

func newMockMessage(from string, to []string, msg io.WriterTo) *mockMessage {
	buff := &bytes.Buffer{}
	msg.WriteTo(buff)
	mm := &mockMessage{
		from:    from,
		to:      to,
		message: buff.Bytes(),
		sendAt:  time.Now(),
	}
	mm.getdialer = mm.defaultDialer
	return mm
}

func (mm *mockMessage) setDialer(dialer func() (Dialer, error)) {
	mm.getdialer = dialer
}

func (mm *mockMessage) defaultDialer() (Dialer, error) {
	return newMockDialer(), nil
}

func (mm *mockMessage) GetDialer() (Dialer, error) {
	return mm.getdialer()
}

func (mm *mockMessage) Backoff(reason error) error {
	mm.backoffCount++
	mm.err = reason
	return nil
}

func (mm *mockMessage) Error(err error) error {
	mm.err = err
	mm.finished = true
	return nil
}

func (mm *mockMessage) Finish() error {
	mm.finished = true
	return nil
}

func (mm *mockMessage) Generate(message *gomail.Message) error {
	message.SetHeaders(map[string][]string{
		"From": {mm.from},
		"To":   mm.to,
	})
	message.SetBody("text/html", string(mm.message))
	return nil
}

func (mm *mockMessage) GetSmtpFrom() (string, error) {
	return mm.from, nil
}

func (mm *mockMessage) Success() error {
	mm.finished = true
	return nil
}
