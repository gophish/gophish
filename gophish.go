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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gophish/gophish/auth"
	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/controllers"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/handlers"
)

var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

func main() {
	// Setup the global variables and settings
	err := models.Setup()
	if err != nil {
		fmt.Println(err)
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	// Start the web servers
	go func() {
		defer wg.Done()
		adminHandler := gziphandler.GzipHandler(controllers.CreateAdminRouter())
		auth.Store.Options.Secure = config.Conf.AdminConf.UseTLS
		if config.Conf.AdminConf.UseTLS { // use TLS for Admin web server if available
			Logger.Printf("Starting admin server at https://%s\n", config.Conf.AdminConf.ListenURL)
			err := CheckAndCreateSSL(config.Conf.AdminConf.ListenURL, config.Conf.AdminConf.CertPath, config.Conf.AdminConf.KeyPath, Logger)
			if err != nil {
				Logger.Fatal(err)
			}
			Logger.Fatal(http.ListenAndServeTLS(config.Conf.AdminConf.ListenURL, config.Conf.AdminConf.CertPath, config.Conf.AdminConf.KeyPath,
				handlers.CombinedLoggingHandler(os.Stdout, adminHandler)))
		} else {
			Logger.Printf("Starting admin server at http://%s\n", config.Conf.AdminConf.ListenURL)
			Logger.Fatal(http.ListenAndServe(config.Conf.AdminConf.ListenURL, handlers.CombinedLoggingHandler(os.Stdout, adminHandler)))
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		phishHandler := gziphandler.GzipHandler(controllers.CreatePhishingRouter())
		if config.Conf.PhishConf.UseTLS { // use TLS for Phish web server if available
			Logger.Printf("Starting phishing server at https://%s\n", config.Conf.PhishConf.ListenURL)
			Logger.Fatal(http.ListenAndServeTLS(config.Conf.PhishConf.ListenURL, config.Conf.PhishConf.CertPath, config.Conf.PhishConf.KeyPath,
				handlers.CombinedLoggingHandler(os.Stdout, phishHandler)))
		} else {
			Logger.Printf("Starting phishing server at http://%s\n", config.Conf.PhishConf.ListenURL)
			Logger.Fatal(http.ListenAndServe(config.Conf.PhishConf.ListenURL, handlers.CombinedLoggingHandler(os.Stdout, phishHandler)))
		}
	}()
	wg.Wait()
}

func CheckAndCreateSSL(ListenURL string, CertPath string, KeyPath string, Logger *log.Logger) error {
	// Check whether there is an existing SSL certificate and/or key, and if so, abort execution of this function
	if _, err := os.Stat(CertPath); !os.IsNotExist(err) {
		return nil
	}
	if _, err := os.Stat(KeyPath); !os.IsNotExist(err) {
		return nil
	}

	Logger.Printf("Creating new self-signed certificates for administration interface...\n")

	// RSA would be a perfectly fine alternative
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

	notBefore := time.Now()
	// Generate a certificate that lasts for 10 years
	notAfter := notBefore.Add(10 * 365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	if err != nil {
		return errors.New(fmt.Sprintf("TLS Certificate Generation: Failed to generate a random serial number: %s", err))
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, priv.Public(), priv)
	if err != nil {
		return errors.New(fmt.Sprintf("TLS Certificate Generation: Failed to create certificate: %s", err))
	}

	certOut, err := os.Create(CertPath)
	if err != nil {
		return errors.New(fmt.Sprintf("TLS Certificate Generation: Failed to open %s for writing: %s", CertPath, err))
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(KeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.New(fmt.Sprintf("TLS Certificate Generation: Failed to open %s for writing", KeyPath))
	}

	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return errors.New(fmt.Sprintf("TLS Certificate Generation: Unable to marshal ECDSA private key: %v", err))
	}

	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	keyOut.Close()

	Logger.Println("TLS Certificate Generation complete")
	return nil
}
