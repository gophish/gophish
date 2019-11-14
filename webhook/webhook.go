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
  "encoding/gob"

  log "github.com/gophish/gophish/logger"
)

const (
  DefaultTimeoutSeconds = 10
  MinHttpStatusErrorCode = 400
  SignatureHeader = "X-Gophish-Signature"
)


//TODO

type Sender interface {
  Send(url string, secret string, data interface{}) error
}

type DefaultSender struct {
  client *http.Client
} 

var mainSenderInstance Sender
func newDefaultSender() Sender {
  sn := &DefaultSender{}
  sn.client = &http.Client{
      Timeout: DefaultTimeoutSeconds,
  }
  return sn
}

func SendAll(whsInfo map[string]string, data interface{}) {
  if mainSenderInstance == nil {
    mainSenderInstance = newDefaultSender()
  }

  for k, v := range whsInfo {
    go func(url string, secret string) {
          mainSenderInstance.Send(url, secret, data)
       }(k, v)
  }
}

func (ds DefaultSender) Send(url string, secret string, data interface{}) error {
  jsonData, err := json.Marshal(data)
  if err != nil {
    log.Error(err)
    return err
  }
  req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
  data2, err := interfaceToBytes(data)
  if err != nil {
    log.Error(err)
    return err
  }
  signat, err := sign(secret, data2)
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

func interfaceToBytes(data interface{}) ([]byte, error) {
  var buf bytes.Buffer
  enc := gob.NewEncoder(&buf)
  err := enc.Encode(data)
  if err != nil {
    return nil, err
  }
  return buf.Bytes(), nil
}
