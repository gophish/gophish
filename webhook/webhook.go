package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	log "github.com/gophish/gophish/logger"
)

const (

	// DefaultTimeoutSeconds is the number of seconds before a timeout occurs
	// when sending a webhook
	DefaultTimeoutSeconds = 10

	// MinHTTPStatusErrorCode is the lower bound of HTTP status codes which
	// indicate an error occurred
	MinHTTPStatusErrorCode = 400

	// SignatureHeader is the name of the HTTP header which contains the
	// webhook signature
	SignatureHeader = "X-Gophish-Signature"

	// Sha256Prefix is the prefix that specifies the hashing algorithm used
	// for the signature
	Sha256Prefix = "sha256"
)

// Sender represents a type which can send webhooks to an EndPoint
type Sender interface {
	Send(endPoint EndPoint, data interface{}) error
}

type defaultSender struct {
	client *http.Client
}

var senderInstance = &defaultSender{
	client: &http.Client{
		Timeout: time.Second * DefaultTimeoutSeconds,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	},
}

// SetTransport sets the underlying transport for the default webhook client.
func SetTransport(tr *http.Transport) {
	senderInstance.client.Transport = tr
}

// EndPoint represents a URL to send the webhook to, as well as a secret used
// to sign the event
type EndPoint struct {
	URL    string
	Secret string
}

// Send sends data to a single EndPoint
func Send(endPoint EndPoint, data interface{}) error {
	return senderInstance.Send(endPoint, data)
}

// SendAll sends data to multiple EndPoints
func SendAll(endPoints []EndPoint, data interface{}) {
	for _, e := range endPoints {
		go func(e EndPoint) {
			senderInstance.Send(e, data)
		}(e)
	}
}

// Send contains the implementation of sending webhook to an EndPoint
func (ds defaultSender) Send(endPoint EndPoint, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		return err
	}

	req, err := http.NewRequest("POST", endPoint.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error(err)
		return err
	}
	signat, err := sign(endPoint.Secret, jsonData)
	if err != nil {
		log.Error(err)
		return err
	}
	req.Header.Set(SignatureHeader, fmt.Sprintf("%s=%s", Sha256Prefix, signat))
	req.Header.Set("Content-Type", "application/json")
	resp, err := ds.client.Do(req)
	if err != nil {
		log.Error(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= MinHTTPStatusErrorCode {
		errMsg := fmt.Sprintf("http status of response: %s", resp.Status)
		log.Error(errMsg)
		return errors.New(errMsg)
	}
	return nil
}

func sign(secret string, data []byte) (string, error) {
	hash1 := hmac.New(sha256.New, []byte(secret))
	_, err := hash1.Write(data)
	if err != nil {
		return "", err
	}
	hexStr := hex.EncodeToString(hash1.Sum(nil))
	return hexStr, nil
}
