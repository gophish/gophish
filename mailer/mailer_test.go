package mailer

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/textproto"
	"reflect"
	"testing"
)

func generateMessages(dialer Dialer) []Mail {
	to := []string{"to@example.com"}

	messageContents := []io.WriterTo{
		bytes.NewBuffer([]byte("First email")),
		bytes.NewBuffer([]byte("Second email")),
	}

	m1 := newMockMessage("first@example.com", to, messageContents[0])
	m2 := newMockMessage("second@example.com", to, messageContents[1])

	m1.setDialer(func() (Dialer, error) { return dialer, nil })

	messages := []Mail{m1, m2}
	return messages
}

func newMockErrorSender(err error) *mockSender {
	sender := newMockSender()
	// The sending function will send a temporary error to emulate
	// a backoff.
	sender.setSend(func(mm *mockMessage) error {
		if len(sender.messages) == 1 {
			return err
		}
		sender.messageChan <- mm
		return nil
	})
	return sender
}

func TestDialHost(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	md := newMockDialer()
	md.setDial(md.unreachableDial)
	_, err := dialHost(ctx, md)
	if _, ok := err.(*ErrMaxConnectAttempts); !ok {
		t.Fatalf("Didn't receive expected ErrMaxConnectAttempts. Got: %s", err)
	}
	e := err.(*ErrMaxConnectAttempts)
	if e.underlyingError != errHostUnreachable {
		t.Fatalf("Got invalid underlying error. Expected %s Got %s\n", e.underlyingError, errHostUnreachable)
	}
	if md.dialCount != MaxReconnectAttempts {
		t.Fatalf("Unexpected number of reconnect attempts. Expected %d, Got %d", MaxReconnectAttempts, md.dialCount)
	}
	md.setDial(md.defaultDial)
	_, err = dialHost(ctx, md)
	if err != nil {
		t.Fatalf("Unexpected error when dialing the mock host: %s", err)
	}
}

func TestMailWorkerStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mw := NewMailWorker()
	go func(ctx context.Context) {
		mw.Start(ctx)
	}(ctx)

	sender := newMockSender()
	dialer := newMockDialer()
	dialer.setDial(func() (Sender, error) {
		return sender, nil
	})

	messages := generateMessages(dialer)

	// Send the campaign
	mw.Queue(messages)

	got := []*mockMessage{}

	idx := 0
	for message := range sender.messageChan {
		got = append(got, message)
		original := messages[idx].(*mockMessage)
		if original.from != message.from {
			t.Fatalf("Invalid message received. Expected %s, Got %s", original.from, message.from)
		}
		idx++
	}
	if len(got) != len(messages) {
		t.Fatalf("Unexpected number of messages received. Expected %d Got %d", len(got), len(messages))
	}
}

func TestBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mw := NewMailWorker()
	go func(ctx context.Context) {
		mw.Start(ctx)
	}(ctx)

	expectedError := &textproto.Error{
		Code: 400,
		Msg:  "Temporary error",
	}

	sender := newMockErrorSender(expectedError)
	dialer := newMockDialer()
	dialer.setDial(func() (Sender, error) {
		return sender, nil
	})

	messages := generateMessages(dialer)

	// Send the campaign
	mw.Queue(messages)

	got := []*mockMessage{}

	for message := range sender.messageChan {
		got = append(got, message)
	}
	// Check that we only sent one message
	expectedCount := 1
	if len(got) != expectedCount {
		t.Fatalf("Unexpected number of messages received. Expected %d Got %d", len(got), expectedCount)
	}

	// Check that it's the correct message
	originalFrom := messages[1].(*mockMessage).from
	if got[0].from != originalFrom {
		t.Fatalf("Invalid message received. Expected %s, Got %s", originalFrom, got[0].from)
	}

	// Check that the first message performed a backoff
	backoffCount := messages[0].(*mockMessage).backoffCount
	if backoffCount != expectedCount {
		t.Fatalf("Did not receive expected backoff. Got backoffCount %d, Expected %d", backoffCount, expectedCount)
	}

	// Check that there was a reset performed on the sender
	if sender.resetCount != expectedCount {
		t.Fatalf("Did not receive expected reset. Got resetCount %d, expected %d", sender.resetCount, expectedCount)
	}
}

func TestPermError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mw := NewMailWorker()
	go func(ctx context.Context) {
		mw.Start(ctx)
	}(ctx)

	expectedError := &textproto.Error{
		Code: 500,
		Msg:  "Permanent error",
	}

	sender := newMockErrorSender(expectedError)
	dialer := newMockDialer()
	dialer.setDial(func() (Sender, error) {
		return sender, nil
	})

	messages := generateMessages(dialer)

	// Send the campaign
	mw.Queue(messages)

	got := []*mockMessage{}

	for message := range sender.messageChan {
		got = append(got, message)
	}
	// Check that we only sent one message
	expectedCount := 1
	if len(got) != expectedCount {
		t.Fatalf("Unexpected number of messages received. Expected %d Got %d", len(got), expectedCount)
	}

	// Check that it's the correct message
	originalFrom := messages[1].(*mockMessage).from
	if got[0].from != originalFrom {
		t.Fatalf("Invalid message received. Expected %s, Got %s", originalFrom, got[0].from)
	}

	message := messages[0].(*mockMessage)

	// Check that the first message did not perform a backoff
	expectedBackoffCount := 0
	backoffCount := message.backoffCount
	if backoffCount != expectedBackoffCount {
		t.Fatalf("Did not receive expected backoff. Got backoffCount %d, Expected %d", backoffCount, expectedCount)
	}

	// Check that there was a reset performed on the sender
	if sender.resetCount != expectedCount {
		t.Fatalf("Did not receive expected reset. Got resetCount %d, expected %d", sender.resetCount, expectedCount)
	}

	// Check that the email errored out appropriately
	if !reflect.DeepEqual(message.err, expectedError) {
		t.Fatalf("Did not received expected error. Got %#v\nExpected %#v", message.err, expectedError)
	}
}

func TestUnknownError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mw := NewMailWorker()
	go func(ctx context.Context) {
		mw.Start(ctx)
	}(ctx)

	expectedError := errors.New("Unexpected error")

	sender := newMockErrorSender(expectedError)
	dialer := newMockDialer()
	dialer.setDial(func() (Sender, error) {
		return sender, nil
	})

	messages := generateMessages(dialer)

	// Send the campaign
	mw.Queue(messages)

	got := []*mockMessage{}

	for message := range sender.messageChan {
		got = append(got, message)
	}
	// Check that we only sent one message
	expectedCount := 1
	if len(got) != expectedCount {
		t.Fatalf("Unexpected number of messages received. Expected %d Got %d", len(got), expectedCount)
	}

	// Check that it's the correct message
	originalFrom := messages[1].(*mockMessage).from
	if got[0].from != originalFrom {
		t.Fatalf("Invalid message received. Expected %s, Got %s", originalFrom, got[0].from)
	}

	message := messages[0].(*mockMessage)

	// If we get an unexpected error, this means that it's likely the
	// underlying connection dropped. When this happens, we expect the
	// connection to be re-established (see #997).
	// In this case, we're successfully reestablishing the connection
	// so we expect the backoff to occur.
	expectedBackoffCount := 1
	backoffCount := message.backoffCount
	if backoffCount != expectedBackoffCount {
		t.Fatalf("Did not receive expected backoff. Got backoffCount %d, Expected %d", backoffCount, expectedBackoffCount)
	}

	// Check that the underlying connection was reestablished
	expectedDialCount := 2
	if dialer.dialCount != expectedDialCount {
		t.Fatalf("Did not receive expected dial count. Got %d expected %d", dialer.dialCount, expectedDialCount)
	}

	// Check that the email errored out appropriately
	if !reflect.DeepEqual(message.err, expectedError) {
		t.Fatalf("Did not received expected error. Got %#v\nExpected %#v", message.err, expectedError)
	}
}
