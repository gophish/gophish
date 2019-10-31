package webhook

type Webhook struct {
  client *http.Client
}

func (whook *Webhook) Send(server string, secret []byte, data interface{}) error {
  jsonData, err := json.Marshal(data)
  if err != nil {
    http.Error(w, "Error converting data parameter to JSON", http.StatusInternalServerError)
    log.Error(err)
  }


              w.Header().Set("Content-Type", "application/json")
              w.WriteHeader(c)
              fmt.Fprintf(w, "%s", dj)




  req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

  //TODO
  sign := ""
  req.Header.Set("X-Gophish-Signature", sign)
  req.Header.Set("Content-Type", "application/json")

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    panic(err)
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