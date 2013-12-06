package main

/*
gophish - Open-Source Phishing Framework

The MIT License (MIT)

Copyright (c) 2013 Jordan Wright

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type SMTPServer struct {
	Server   string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type Config struct {
	URL  string     `json:"url"`
	SMTP SMTPServer `json:"smtp"`
}

var config Config

func main() {
	// Get the config file
	config_file, e := ioutil.ReadFile("./config.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	json.Unmarshal(config_file, &config)
	fmt.Printf("Gophish server started at http://%s\n", config.URL)
	http.Handle("/", createRouter())
	http.ListenAndServe(config.URL, nil)
}
