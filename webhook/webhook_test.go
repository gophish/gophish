package webhook

//TODO

import (
	"testing"
	"net/http"
	"log"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/stretchr/testify/suite"
	"github.com/gophish/gophish/webhook"
	"github.com/stretchr/testify/assert"
)

type WebhookSuite struct {
	suite.Suite
}

type mockSender struct {
	client *http.Client
}

func newMockSender() *mockSender {
	ms := &mockSender {
		client: &http.Client{},
	}
	return ms
}

func (ms mockSender) Send(endPoint webhook.EndPoint, data interface{}) error {
	log.Println("[test] mocked 'Send' function")
	return nil
}

func (s *WebhookSuite) TestSend() {
	mcSnd := newMockSender()
	endp1 := webhook.EndPoint{URL: "http://example.com/a1", Secret: "s1"}
	d1 := map[string]string {
		"a1": "a11",
		"a2": "a22",
		"a3": "a33",
	}
	err := mcSnd.Send(endp1, d1)
	s.Nil(err)
}

func (s *WebhookSuite) TestSend2() {

}

func (s *WebhookSuite) TestSignature() {
	expectedSign := "8c5f1d7b39c52b7a68cd6ff4340dc7907e2681d683b32c1e40e81bfad1d65f80"
	d1 := map[string]string {
		"key1": "a11",
		"key2": "a22",
		"key3": "a33",
	};

	jsonData, err := json.Marshal(d1)
	s.Nil(err)

	secret := "secret123"
	hash1 := hmac.New(sha256.New, []byte(secret))
	_, err = hash1.Write(jsonData)
	s.Nil(err)

	realSign := hex.EncodeToString(hash1.Sum(nil))
	assert.Equal(s.T(), expectedSign, realSign)
}

func TestWebhookSuite(t *testing.T) {
	suite.Run(t, new(WebhookSuite))
}
