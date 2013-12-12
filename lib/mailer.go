package main

import (
	"net/smtp"
)

//Send sends an Email using a connection to Server.
//If a Username and Password are set for the Server, authentication will be attempted
//However, to support open-relays, authentication is optional.
func Send(email Email, server Server) {
	auth := nil
	if server.User != nil && server.Password != nil {
		auth := smtp.PlainAuth("", server.User, server.Password, server.Host)
	}
	smtp.SendMail(server.Host, auth, email.From, email.To, Email.Body)
}
