package models

import (
	"errors"
	"log"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/jordan-wright/gophish/config"
	_ "github.com/mattn/go-sqlite3" // Blank import needed to import sqlite3
)

var db gorm.DB
var err error

// ErrUsernameTaken is thrown when a user attempts to register a username that is taken.
var ErrUsernameTaken = errors.New("username already taken")

// Logger is a global logger used to show informational, warning, and error messages
var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

const (
	CAMPAIGN_IN_PROGRESS string = "In progress"
	CAMPAIGN_QUEUED      string = "Queued"
	CAMPAIGN_EMAILS_SENT string = "Emails Sent"
	CAMPAIGN_COMPLETE    string = "Completed"
	EVENT_SENT           string = "Email Sent"
	EVENT_OPENED         string = "Email Opened"
	EVENT_CLICKED        string = "Clicked Link"
	STATUS_SUCCESS       string = "Success"
	STATUS_UNKNOWN       string = "Unknown"
	ERROR                string = "Error"
)

// Flash is used to hold flash information for use in templates.
type Flash struct {
	Type    string
	Message string
}

// Response contains the attributes found in an API response
type Response struct {
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// Setup initializes the Conn object
// It also populates the Gophish Config object
func Setup() error {
	create_db := false
	if _, err = os.Stat(config.Conf.DBPath); err != nil || config.Conf.DBPath == ":memory:" {
		create_db = true
	}
	db, err = gorm.Open("sqlite3", config.Conf.DBPath)
	db.LogMode(false)
	db.SetLogger(Logger)
	if err != nil {
		Logger.Println(err)
		return err
	}
	//If the file already exists, delete it and recreate it
	if create_db {
		Logger.Printf("Database not found... creating db at %s\n", config.Conf.DBPath)
		db.CreateTable(User{})
		db.CreateTable(Target{})
		db.CreateTable(Result{})
		db.CreateTable(Group{})
		db.CreateTable(GroupTarget{})
		db.CreateTable(Template{})
		db.CreateTable(Attachment{})
		db.CreateTable(Page{})
		db.CreateTable(SMTP{})
		db.CreateTable(Event{})
		db.CreateTable(Campaign{})
		//Create the default user
		initUser := User{
			Username: "admin",
			Hash:     "$2a$10$IYkPp0.QsM81lYYPrQx6W.U6oQGw7wMpozrKhKAHUBVL4mkm/EvAS", //gophish
			ApiKey:   "12345678901234567890123456789012",
		}
		err = db.Save(&initUser).Error
		if err != nil {
			Logger.Println(err)
		}
	}
	return nil
}
