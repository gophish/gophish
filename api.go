package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func API(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello api")
}

func API_Campaigns(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello api")
}

func API_Campaigns_Id(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	fmt.Fprintf(w, "{\"method\" : \""+r.Method+"\", \"id\" : "+vars["id"]+"}")
}

func API_Doc(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "api_doc")
}
