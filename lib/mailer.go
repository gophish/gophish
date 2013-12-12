package main

import (
	"net/smtp"
)

func Send(email Email, server Server) {
	auth = smtp.PlainAuth("", server.User, server.Password, server.Server)
}
