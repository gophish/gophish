package models

// SMTPServer is used to provide a default SMTP server preference.
type SMTPServer struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// Config represents the configuration information.
type Config struct {
	URL    string     `json:"url"`
	SMTP   SMTPServer `json:"smtp"`
	DBPath string     `json:"dbpath"`
}

// User represents the user model for gophish.
type User struct {
	Id       int
	Username string
	Hash     string
	APIKey   string
}

// Flash is used to hold flash information for use in templates.
type Flash struct {
	Type    string
	Message string
}
