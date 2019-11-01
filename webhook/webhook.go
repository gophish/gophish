package webhook

import (
  "crypto/hmac"
  "crypto/sha256"
  "encoding/hex"
  "time"
)

const DefaultTimeoutSeconds = 10

type Webhook struct {
  client *http.Client
}

//TODO


func (wh *Webhook) Send(server string, secret []byte, data interface{}) error {
  jsonData, err := json.Marshal(data)
  if err != nil {
    http.Error(w, "Error converting data parameter to JSON", http.StatusInternalServerError)
    log.Error(err)
  }




  req, err := http.NewRequest("POST", wh.Url, bytes.NewBuffer(jsonData))

  ts := int32(time.Now().Unix())
  signat := sign(data, ts)
  req.Header.Set("X-Gophish-Signature", signat)
  req.Header.Set("Content-Type", "application/json")

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    http.Error(w, "Error sending request", http.StatusInternalServerError)
    log.Error(err)
  }
  defer resp.Body.Close()

  if resp.Status >= 400 {
    //error
  }

}


//TODO
func (wh *Webhook) sign(data interface{}, ts int32) {
  //TODO: add timestamp
  // data2 := fmt.Sprintf("%s__%s", data, ts) 
  data2 := data

  hash1 := hmac.New(sha256.New, []byte(wh.Secret))
  hash1.Write(byte[](data2))
  return hex.EncodeToString(hash1.Sum(nil))
}