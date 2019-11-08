package webhook

import (
  "crypto/hmac"
  "crypto/sha256"
  "encoding/hex"
  "time"
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
)

//TODO rename to Sender because "http" contains Transport too
type Transport struct {
  Client *http.Client
}

//TODO


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
  ts := int32(time.Now().Unix())
  signat, err := sign(secret, data, ts)
  req.Header.Set("X-Gophish-Signature", signat)
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


//TODO
func sign(secret string, data interface{}, ts int32) (string, error) {
  // data2 := fmt.Sprintf("%s__%s", data, ts) //TODO: add timestamp
  data2, err := interfaceToBytes(data)

  if err != nil {
    return "", err
  }
  hash1 := hmac.New(sha256.New, []byte(secret))
  _, err = hash1.Write(data2)
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