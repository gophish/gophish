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

	// DefaultTimeoutSeconds is amount of seconds of timeout used by HTTP sender
	DefaultTimeoutSeconds = 10

	// MinHTTPStatusErrorCode is the lowest number of an HTTP response which indicates an error
	MinHTTPStatusErrorCode = 400

	// SignatureHeader is the name of an HTTP header used to which contains signature of a webhook
	SignatureHeader = "X-Gophish-Signature"

	// Sha256Prefix is the prefix that specifies the hashing algorithm used for signature
	Sha256Prefix = "sha256"
)

// Sender defines behaviour of an entity by which webhook is sent
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

// EndPoint represents and end point to send a webhook to: url and secret by which payload is signed
type EndPoint struct {
	URL    string
	Secret string
}

// Send sends data to a single EndPoint
func Send(endPoint EndPoint, data interface{}) error {
	return senderInstance.Send(endPoint, data)
}

// SendAll sends data to each of the EndPoints
func SendAll(endPoints []EndPoint, data interface{}) {
	for _, ept := range endPoints {
		go func(ept1 EndPoint) {
			senderInstance.Send(ept1, data)
		}(EndPoint{URL: ept.URL, Secret: ept.Secret})
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
	signat, err := sign(endPoint.Secret, jsonData)
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
