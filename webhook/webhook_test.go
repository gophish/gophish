package webhook

//TODO
import (

	// "bytes"
	// "context"
	// "errors"
	// "io"
	// "net/textproto"
	// "reflect"
	// "testing"

	"github.com/stretchr/testify/suite"
)

type WebhookSuite struct {
	suite.Suite
}

// mockSender is a mock gomail.Sender used for testing.
type mockSender struct {
}

func newMockSender() *mockSender {
	ms := &mockSender{}
	return ms
}
