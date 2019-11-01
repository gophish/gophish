package webhook

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




  req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

  signat := sign(data)
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

func (wh *Webhook) sign(data interface{}) {
  return "TODO"
}