package main

type SMTPServer struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type Email struct {
	Subject string
	Body    string
	To      []string
	Bcc     []string
	Cc      []string
	From    string
}

type Config struct {
	URL  string     `json:"url"`
	SMTP SMTPServer `json:"smtp"`
}

type User struct {
	Id       string
	Username string
	Hash     string
	APIKey   string
}
