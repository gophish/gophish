package main

type SMTPServer struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type Config struct {
	URL    string     `json:"url"`
	SMTP   SMTPServer `json:"smtp"`
	DBPath string     `json:"dbpath"`
}

type User struct {
	Id       int
	Username string
	Hash     string
	APIKey   string
}
