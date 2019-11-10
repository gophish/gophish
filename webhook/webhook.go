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

type Transport struct {
  Client *http.Client
}


type Sender interface {
  Send(url string, secret string, data interface{}) error
}



func (whTr *Transport) Send(url string, secret string, data interface{}) error {
  if whTr.Client == nil {
    errMsg := "Client must be initialized"
    log.Error(errMsg)
    panic(errMsg)
  }
  jsonData, err := json.Marshal(data)
  if err != nil {
    log.Error(err)
    return err
  }
  req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
  signat, err := sign(secret, data)
  req.Header.Set(SignatureHeader, signat)
  req.Header.Set("Content-Type", "application/json")
  resp, err := whTr.Client.Do(req)
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
