package webhook

type Webhook struct {
  client *http.Client
}

func (whook *Webhook) Send(server string, secret []byte, data interface{}) error {
  resp, err := http.Post(server, ???)
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

func (as *Server) PingWebhook(w http.ResponseWriter, r *http.Request) {
  switch {
  case r.Method == "POST":
    //TODO


  }
}