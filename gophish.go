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
	"io/ioutil"
	"os"
	"os/signal"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/controllers"
	"github.com/gophish/gophish/imap"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/middleware"
	"github.com/gophish/gophish/models"
)

var (
	configPath    = kingpin.Flag("config", "Location of config.json.").Default("./config.json").String()
	disableMailer = kingpin.Flag("disable-mailer", "Disable the mailer (for use with multi-system deployments)").Bool()
)

func main() {
	// Load the version

	version, err := ioutil.ReadFile("./VERSION")
	if err != nil {
		log.Fatal(err)
	}
	kingpin.Version(string(version))

	// Parse the CLI flags and load the config
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	// Load the config
	conf, err := config.LoadConfig(*configPath)
	// Just warn if a contact address hasn't been configured
	if err != nil {
		log.Fatal(err)
	}
	if conf.ContactAddress == "" {
		log.Warnf("No contact address has been configured.")
		log.Warnf("Please consider adding a contact_address entry in your config.json")
	}
	config.Version = string(version)

	err = log.Setup(conf.Logging)
	if err != nil {
		log.Fatal(err)
	}

	// Provide the option to disable the built-in mailer
	// Setup the global variables and settings
	err = models.Setup(conf)
	if err != nil {
		log.Fatal(err)
	}

	// Unlock any maillogs that may have been locked for processing
	// when Gophish was last shutdown.
	err = models.UnlockAllMailLogs()
	if err != nil {
		log.Fatal(err)
	}

	// Create our servers
	adminOptions := []controllers.AdminServerOption{}
	if *disableMailer {
		adminOptions = append(adminOptions, controllers.WithWorker(nil))
	}
	adminConfig := conf.AdminConf
	adminServer := controllers.NewAdminServer(adminConfig, adminOptions...)
	middleware.Store.Options.Secure = adminConfig.UseTLS

	phishConfig := conf.PhishConf
	phishServer := controllers.NewPhishingServer(phishConfig)

	imapMonitor := imap.NewMonitor()

	go adminServer.Start()
	go phishServer.Start()
	go imapMonitor.Start()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Info("CTRL+C Received... Gracefully shutting down servers")
	adminServer.Shutdown()
	phishServer.Shutdown()
	imapMonitor.Shutdown()

}
