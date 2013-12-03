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

func API_Campaigns_Id_Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	fmt.Fprintf(w, "{\"id\" : "+vars["id"]+"}")
}
func API_Campaigns_Id_Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	fmt.Fprintf(w, "{\"id\" : "+vars["id"]+"}")
}
