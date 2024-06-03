package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type mockSender struct {
	client *http.Client
}

func newMockSender() *mockSender {
	ms := &mockSender{
		client: &http.Client{},
	}
	return ms
}

func (ms mockSender) Send(endPoint EndPoint, data interface{}) error {
	log.Println("[test] mocked 'Send' function")
	return nil
}

func TestSendMocked(t *testing.T) {
	ms := newMockSender()
	endpoint := EndPoint{URL: "http://example.com/a1", Secret: "s1"}
	data := map[string]string{
		"a1": "a11",
		"a2": "a22",
		"a3": "a33",
	}
	err := ms.Send(endpoint, data)
	if err != nil {
		t.Fatalf("error sending data to webhook endpoint: %v", err)
	}
}

func TestSendReal(t *testing.T) {
	expectedSig := "004b36ca3fcbc01a08b17bf5d4a7e1aa0b10e14f55f3f8bd9acac0c7e8d2635d"
	secret := "secret456"
	data := map[string]interface{}{
		"key1": "val1",
		"key2": "val2",
		"key3": "val3",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("[test] running the server...")

		signStartIdx := len(Sha256Prefix) + 1
		sigHeader := r.Header.Get(SignatureHeader)
		gotSig := sigHeader[signStartIdx:]
		if expectedSig != gotSig {
			t.Fatalf("invalid signature received. expected %s got %s", expectedSig, gotSig)
		}

		ct := r.Header.Get("Content-Type")
		expectedCT := "application/json"
		if ct != expectedCT {
			t.Fatalf("invalid content type. expected %s got %s", ct, expectedCT)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("error reading JSON body from webhook request: %v", err)
		}

		var payload map[string]interface{}
		err = json.Unmarshal(body, &payload)
		if err != nil {
			t.Fatalf("error unmarshaling webhook payload: %v", err)
		}
		if !reflect.DeepEqual(data, payload) {
			t.Fatalf("invalid payload received. expected %#v got %#v", data, payload)
		}
	}))

	defer ts.Close()
	endp1 := EndPoint{URL: ts.URL, Secret: secret}
	err := Send(endp1, data)
	if err != nil {
		t.Fatalf("error sending data to webhook endpoint: %v", err)
	}
}

func TestSignature(t *testing.T) {
	secret := "secret123"
	payload := []byte("some payload456")
	expected := "ab7844c1e9149f8dc976c4188a72163c005930f3c2266a163ffe434230bdf761"
	got, err := sign(secret, payload)
	if err != nil {
		t.Fatalf("error signing payload: %v", err)
	}
	if expected != got {
		t.Fatalf("invalid signature received. expected %s got %s", expected, got)
	}
}
