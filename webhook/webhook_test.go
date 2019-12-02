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
	"io/ioutil"

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

func (s *WebhookSuite) TestSend1() {
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

		//TODO
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello, client333")
		}))
		defer ts.Close()

		res, err := http.Get(ts.URL)
		if err != nil {
			log.Fatal(err)
		}
		_, err = ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
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
