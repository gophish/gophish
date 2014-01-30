package models

import (
	"net/mail"

	// SMTPServer is used to provide a default SMTP server preference.
	"time"
)

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
	Id       int    `json:"id"`
	Username string `json:"username"`
	Hash     string `json:"-"`
	APIKey   string `json:"apikey"`
}

// Flash is used to hold flash information for use in templates.
type Flash struct {
	Type    string
	Message string
}

//Campaign is a struct representing a created campaign
type Campaign struct {
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	CreatedDate   time.Time `json:"created_date" db:"created_date"`
	CompletedDate time.Time `json:"completed_date" db:"completed_date"`
	Template      string    `json:"template"` //This may change
	Status        string    `json:"status"`
	Uid           int
}

type UserCampaigns struct {
	CampaignId int
	UserId     int
}

type Result struct {
	Id       int
	TargetId int
	Status   string `json:"status"`
}

type CampaignResults struct {
	CampaignId int
	TargetId   int
}

type Target struct {
	Id    int          `json:"-"`
	Email mail.Address `json:"email"`
}
