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
	"flag"
	"fmt"
	"net/http"

	"github.com/jordan-wright/gophish/config"
	"github.com/jordan-wright/gophish/controllers"
	"github.com/jordan-wright/gophish/db"
	"github.com/jordan-wright/gophish/middleware"
)

var setupFlag = flag.Bool("setup", false, "Starts the initial setup process for Gophish")

func main() {
	//Setup the global variables and settings
	flag.Parse()
	err := db.Setup(*setupFlag)
	defer db.Conn.Close()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Gophish server started at http://%s\n", config.Conf.URL)
	http.Handle("/", controllers.Use(controllers.CreateRouter().ServeHTTP, middleware.GetContext))
	http.ListenAndServe(config.Conf.URL, nil)
}
