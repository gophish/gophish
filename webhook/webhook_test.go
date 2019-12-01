package webhook

//TODO

import (
	// "encoding/json"
	"fmt"
	"testing"
	"net/http"

	"github.com/stretchr/testify/suite"
	"github.com/gophish/gophish/webhook"
)

type WebhookSuite struct {
	suite.Suite
}

type mockSender struct {
	client *http.Client
}

func newMockSender() *mockSender {
	ms := &mockSender{
		client: &http.Client{},
	}
	return ms
}


func (mcs mockSender) Send(endPoint webhook.EndPoint, data interface{}) error {
	fmt.Println("Mocked Send function")
	// _, err := json.Marshal(data)
	// s.Nil(err)

	return nil
}

func (s *WebhookSuite) TestSend() {
	snd1 := newMockSender()
	endp1 := webhook.EndPoint{URL: "http://example.com/a1", Secret: "s1"}
	d1 := map[string]string {
		"a1": "a11",
		"a2": "a22",
		"a3": "a33",
	}
	err := snd1.Send(endp1, d1)
	s.Nil(err)
}

func TestWebhookSuite(t *testing.T) {
	suite.Run(t, new(WebhookSuite))
}
