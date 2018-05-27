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
	"compress/gzip"
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/NYTimes/gziphandler"
	"github.com/gophish/gophish/auth"
	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/controllers"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/mailer"
	"github.com/gophish/gophish/models"
	"github.com/gophish/gophish/util"
	"github.com/gorilla/handlers"
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
	config.LoadConfig(*configPath)
	config.Version = string(version)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Provide the option to disable the built-in mailer
	if !*disableMailer {
		go mailer.Mailer.Start(ctx)
	}
	// Setup the global variables and settings
	err = models.Setup()
	if err != nil {
		log.Fatal(err)
	}
	// Unlock any maillogs that may have been locked for processing
	// when Gophish was last shutdown.
	err = models.UnlockAllMailLogs()
	if err != nil {
		log.Fatal(err)
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	// Start the web servers
	go func() {
		defer wg.Done()
		gzipWrapper, _ := gziphandler.NewGzipLevelHandler(gzip.BestCompression)
		adminHandler := gzipWrapper(controllers.CreateAdminRouter())
		auth.Store.Options.Secure = config.Conf.AdminConf.UseTLS
		if config.Conf.AdminConf.UseTLS { // use TLS for Admin web server if available
			err := util.CheckAndCreateSSL(config.Conf.AdminConf.CertPath, config.Conf.AdminConf.KeyPath)
			if err != nil {
				log.Fatal(err)
			}
			log.Infof("Starting admin server at https://%s", config.Conf.AdminConf.ListenURL)
			log.Info(http.ListenAndServeTLS(config.Conf.AdminConf.ListenURL, config.Conf.AdminConf.CertPath, config.Conf.AdminConf.KeyPath,
				handlers.CombinedLoggingHandler(log.Writer(), adminHandler)))
		} else {
			log.Infof("Starting admin server at http://%s", config.Conf.AdminConf.ListenURL)
			log.Info(http.ListenAndServe(config.Conf.AdminConf.ListenURL, handlers.CombinedLoggingHandler(os.Stdout, adminHandler)))
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		phishHandler := gziphandler.GzipHandler(controllers.CreatePhishingRouter())
		if config.Conf.PhishConf.UseTLS { // use TLS for Phish web server if available
			log.Infof("Starting phishing server at https://%s", config.Conf.PhishConf.ListenURL)
			log.Info(http.ListenAndServeTLS(config.Conf.PhishConf.ListenURL, config.Conf.PhishConf.CertPath, config.Conf.PhishConf.KeyPath,
				handlers.CombinedLoggingHandler(log.Writer(), phishHandler)))
		} else {
			log.Infof("Starting phishing server at http://%s", config.Conf.PhishConf.ListenURL)
			log.Fatal(http.ListenAndServe(config.Conf.PhishConf.ListenURL, handlers.CombinedLoggingHandler(os.Stdout, phishHandler)))
		}
	}()
	wg.Wait()
}
