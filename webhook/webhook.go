package webhook

import (
  "crypto/hmac"
  "crypto/sha256"
  "encoding/hex"
  "net/http"
  "fmt"
  "errors"
  "encoding/json"
  "bytes"

  log "github.com/gophish/gophish/logger"
)

const (
  DefaultTimeoutSeconds = 10
  MinHttpStatusErrorCode = 400
  SignatureHeader = "X-Gophish-Signature"
)


//TODO

type Sender interface {
  //TODO replace with EndPoint
  Send(url string, secret string, data interface{}) error
}

type defaultSender struct {
  client *http.Client
}

var senderInstance = &defaultSender {
  client: &http.Client{
    Timeout: DefaultTimeoutSeconds,
    CheckRedirect: func(req *http.Request, via []*http.Request) error {
      return http.ErrUseLastResponse
    },
  },
}

type EndPoint struct {
  Url string
  Secret string
}

func SendSingle(endPoint EndPoint, data interface{}) error {
  return senderInstance.Send(endPoint, data)
}

func SendAll(endPoints []EndPoint, data interface{}) {
  for _, ept := range endPoints {
    //TODO handle or return errors
    go func(ept1 EndPoint) {
          senderInstance.Send(ept1, data)
       }(EndPoint{Url: ept.Url, Secret: ept.Secret})
  }
}

//TODO replace with EndPoint
func (ds defaultSender) Send(endPoint EndPoint, data interface{}) error {
  jsonData, err := json.Marshal(data)
  if err != nil {
    log.Error(err)
    return err
  }
  req, err := http.NewRequest("POST", endPoint.Url, bytes.NewBuffer(jsonData))
  signat, err := sign(endPoint.Secret, jsonData)
  req.Header.Set(SignatureHeader, signat)
  req.Header.Set("Content-Type", "application/json")
  resp, err := ds.client.Do(req)
  if err != nil {
    log.Error(err)
    return err
  }
  defer resp.Body.Close()


  //TODO
  if resp.StatusCode >= MinHttpStatusErrorCode {
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
