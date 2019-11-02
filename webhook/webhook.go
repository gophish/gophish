package webhook

import (
  "crypto/hmac"
  "crypto/sha256"
  "encoding/hex"
  "time"
)

const (
  DefaultTimeoutSeconds = 10
  MinHttpStatusErrorCode = 400
)

type Webhook struct {
  client *http.Client
}

//TODO


func (wh *Webhook) Send(data interface{}) error {
  jsonData, err := json.Marshal(data)
  if err != nil {
    log.Error(err)
    return err
  }
  req, err := http.NewRequest("POST", wh.Url, bytes.NewBuffer(jsonData))
  ts := int32(time.Now().Unix())
  signat := wh.sign(data, ts)
  req.Header.Set("X-Gophish-Signature", signat)
  req.Header.Set("Content-Type", "application/json")
  resp, err := wh.client.Do(req)
  if err != nil {
    log.Error(err)
    return err
  }
  defer resp.Body.Close()
  if resp.Status >= MinHttpStatusErrorCode {
    errMsg := fmt.Sprintf("http status of response: %d", resp.Status)
    log.Error(errMsg)
    return errors.New(errMsg)
  }
  return nil
}


//TODO
func (wh *Webhook) sign(data interface{}, ts int32) (string, error) {

  //TODO: add timestamp
  // data2 := fmt.Sprintf("%s__%s", data, ts) 
  data2 := data

  hash1 := hmac.New(sha256.New, []byte(wh.Secret))
  _, err := hash1.Write(byte[](data2))
  if err != nil {
    return "", err
  }
  hexStr := hex.EncodeToString(hash1.Sum(nil))
  return hexStr, nil
}