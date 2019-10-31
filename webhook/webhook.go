package webhook

type Webhook struct {
  client *http.Client
}

//TODO


func (whook *Webhook) Send(server string, secret []byte, data interface{}) error {
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


//TODO
func (as *Server) Webhook(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)

  switch {
  case r.Method == "GET":
    JSONResponse(w, c, http.StatusOK)

  case r.Method == "POST":
    JSONResponse(w, c, http.StatusOK)

  case r.Method == "DELETE":
    JSONResponse(w, c, http.StatusOK)



  }
}

func (as *Server) Ping(w http.ResponseWriter, r *http.Request) {
  switch {
  case r.Method == "POST":
    //TODO


  }
}

func sign(data interface{}) {
  return "TODO"
}