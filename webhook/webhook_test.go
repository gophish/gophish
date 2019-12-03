package webhook

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

func (ms mockSender) Send(endPoint EndPoint, data interface{}) error {
	log.Println("[test] mocked 'Send' function")
	return nil
}

func (s *WebhookSuite) TestSendMocked() {
	mcSnd := newMockSender()
	endp1 := EndPoint{URL: "http://example.com/a1", Secret: "s1"}
	d1 := map[string]string {
		"a1": "a11",
		"a2": "a22",
		"a3": "a33",
	}
	err := mcSnd.Send(endp1, d1)
	s.Nil(err)
}


func (s *WebhookSuite) TestSendReal() {
	expectedSign := "004b36ca3fcbc01a08b17bf5d4a7e1aa0b10e14f55f3f8bd9acac0c7e8d2635d"
	secret := "secret456"
	d1 := map[string]interface{} {
		"key1": "val1",
		"key2": "val2",
		"key3": "val3",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("[test] running the server...")

		realSign := r.Header.Get(SignatureHeader)
		assert.Equal(s.T(), expectedSign, realSign)

		contTypeJsonHeader := r.Header.Get("Content-Type")
		assert.Equal(s.T(), contTypeJsonHeader, "application/json")

		body, err := ioutil.ReadAll(r.Body)
		s.Nil(err)

		var d2 map[string]interface{}
		err = json.Unmarshal(body, &d2)
		s.Nil(err)
		assert.Equal(s.T(), d1, d2)
	}))

	defer ts.Close()
	endp1 := EndPoint{URL: ts.URL, Secret: secret}
	err := Send(endp1, d1)
	s.Nil(err)
}

func (s *WebhookSuite) TestSignature() {
	expectedSign := "167c12505cebb59eeb4170306e863e8f9d59d2a652c8e73673afc62a50ce32fa"
	secret := "secret123"
	d1 := map[string]string {
		"key1": "val1",
		"key2": "val22",
		"key3": "val333",
	};

	jsonData, err := json.Marshal(d1)
	s.Nil(err)

	hash1 := hmac.New(sha256.New, []byte(secret))
	_, err = hash1.Write(jsonData)
	s.Nil(err)

	realSign := hex.EncodeToString(hash1.Sum(nil))
	assert.Equal(s.T(), expectedSign, realSign)
}

//TODO
func (s *WebhookSuite) TestSignature2() {
	secret := "secret123"
	payload := []byte("some payload456")
	expectedSign := "ab7844c1e9149f8dc976c4188a72163c005930f3c2266a163ffe434230bdf761"
	realSign, err := sign(secret, payload)
	s.Nil(err)
	assert.Equal(s.T(), expectedSign, realSign)
}

func TestWebhookSuite(t *testing.T) {
	suite.Run(t, new(WebhookSuite))
}
