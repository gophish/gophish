package webhook

//TODO

import (
	"encoding/json"
	"fmt"
	"testing"

	log "github.com/gophish/gophish/logger"
	"github.com/stretchr/testify/suite"
	"github.com/gophish/gophish/webhook"
)

type WebhookSuite struct {
	suite.Suite
}

type mockSender struct {
}

func newMockSender() *mockSender {
	ms := &mockSender{}
	return ms
}


//TODO
func (mcs mockSender) Send(endPoint webhook.EndPoint, data interface{}) error {
	fmt.Println("Mocked Send function")
	_, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}


func (s *WebhookSuite) TestSend() {
	s.Equal(1, 1)
}
func (s *WebhookSuite) TestSendAll() {
	s.Equal(2, 2)
}

func TestWebhookSuite(t *testing.T) {
	suite.Run(t, new(WebhookSuite))
}
