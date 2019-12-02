package webhook

//TODO

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"log"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	// "io/ioutil"
	"bytes"

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

func (s *WebhookSuite) TestSendMocked() {
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

func (s *WebhookSuite) TestSendReal() {
	expectedSign := "4775314ed81be378b2b14f18ac29a6db0eb83b44ed464a000400d43100c8a01e"
	successfulHttpResponseCode := 200
	secret := "secret456"

	hClient := &http.Client{}

	//TODO
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("[test] running the server")

		realSign := r.Header.Get(webhook.SignatureHeader)
		assert.Equal(s.T(), expectedSign, realSign)

		neHeader := r.Header.Get("not-existing-header")
		assert.Equal(s.T(), neHeader, "")

		contTypeJsonHeader := r.Header.Get("Content-Type")
		assert.Equal(s.T(), contTypeJsonHeader, "application/json")
	}))
	defer ts.Close()

	d1 := map[string]interface{} {
		"key11": "val1",
		"key22": "val22",
		"key33": map[string]string {
			"key4": "val444",
		},
	}

	jsonData, err := json.Marshal(d1)
	s.Nil(err)

	req, err := http.NewRequest("POST", ts.URL, bytes.NewBuffer(jsonData))
	hash1 := hmac.New(sha256.New, []byte(secret))
	_, err = hash1.Write(jsonData)
	s.Nil(err)

	sign1 := hex.EncodeToString(hash1.Sum(nil))

	req.Header.Set(webhook.SignatureHeader, sign1)
	req.Header.Set("Content-Type", "application/json")
	resp, err := hClient.Do(req)
	s.Nil(err)
	defer resp.Body.Close()

	assert.Equal(s.T(), resp.StatusCode, successfulHttpResponseCode)
	assert.NotEqual(s.T(), resp.StatusCode, webhook.MinHTTPStatusErrorCode)
	assert.True(s.T(), resp.StatusCode < webhook.MinHTTPStatusErrorCode)
}

func (s *WebhookSuite) TestSignature() {
	expectedSign := "167c12505cebb59eeb4170306e863e8f9d59d2a652c8e73673afc62a50ce32fa"
	d1 := map[string]string {
		"key1": "val1",
		"key2": "val22",
		"key3": "val333",
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
