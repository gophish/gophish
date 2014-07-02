package models

import (
	"errors"
	"log"
	"os"

	"github.com/coopernurse/gorp"
	"github.com/jinzhu/gorm"
	"github.com/jordan-wright/gophish/config"
	_ "github.com/mattn/go-sqlite3"
)

var Conn *gorp.DbMap
var db gorm.DB
var err error
var ErrUsernameTaken = errors.New("username already taken")
var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

const (
	CAMPAIGN_IN_PROGRESS string = "In progress"
	CAMPAIGN_QUEUED      string = "Queued"
	CAMPAIGN_COMPLETE    string = "Completed"
	STATUS_SENT          string = "Email Sent"
	STATUS_OPENED        string = "Email Opened"
	STATUS_CLICKED       string = "Clicked Link"
	ERROR                string = "Error"
)

// Flash is used to hold flash information for use in templates.
type Flash struct {
	Type    string
	Message string
}

type Response struct {
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// Setup initializes the Conn object
// It also populates the Gophish Config object
func Setup() error {
	db, err = gorm.Open("sqlite3", config.Conf.DBPath)
	db.LogMode(false)
	db.SetLogger(Logger)
	if err != nil {
		Logger.Println(err)
		return err
	}
	//If the file already exists, delete it and recreate it
	_, err = os.Stat(config.Conf.DBPath)
	if err != nil {
		Logger.Printf("Database not found... creating db at %s\n", config.Conf.DBPath)
		db.CreateTable(User{})
		db.CreateTable(Target{})
		db.CreateTable(Result{})
		db.CreateTable(Group{})
		db.CreateTable(GroupTarget{})
		db.CreateTable(Template{})
		db.CreateTable(SMTP{})
		db.CreateTable(Event{})
		db.CreateTable(Campaign{})
		//Create the default user
		init_user := User{
			Username: "admin",
			Hash:     "$2a$10$IYkPp0.QsM81lYYPrQx6W.U6oQGw7wMpozrKhKAHUBVL4mkm/EvAS", //gophish
			ApiKey:   "12345678901234567890123456789012",
		}
		err = db.Save(&init_user).Error
		if err != nil {
			Logger.Println(err)
		}
	}
	return nil
}
