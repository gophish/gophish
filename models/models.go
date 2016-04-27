package models

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"bitbucket.org/liamstask/goose/lib/goose"

	"github.com/jli53/gophish/config"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3" // Blank import needed to import sqlite3
)

var db *gorm.DB
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
	EVENT_SENDING_ERROR  string = "Error Sending Email"
	EVENT_OPENED         string = "Email Opened"
	EVENT_CLICKED        string = "Clicked Link"
	EVENT_DATA_SUBMIT    string = "Submitted Data"
	EVENT_HTML_OPENED	 string = "HTML Opened"
	EVENT_DOC_OPENED	 string = "Doc Opened"
	STATUS_SUCCESS       string = "Success"
	STATUS_SENDING       string = "Sending"
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

// Copy of auth.GenerateSecureKey to prevent cyclic import with auth library
func generateSecureKey() string {
	k := make([]byte, 32)
	io.ReadFull(rand.Reader, k)
	return fmt.Sprintf("%x", k)
}

// Setup initializes the Conn object
// It also populates the Gophish Config object
func Setup() error {
	create_db := false
	if _, err = os.Stat(config.Conf.DBPath); err != nil || config.Conf.DBPath == ":memory:" {
		create_db = true
	}
	// Setup the goose configuration
	migrateConf := &goose.DBConf{
		MigrationsDir: config.Conf.MigrationsPath,
		Env:           "production",
		Driver: goose.DBDriver{
			Name:    "sqlite3",
			OpenStr: config.Conf.DBPath,
			Import:  "github.com/mattn/go-sqlite3",
			Dialect: &goose.Sqlite3Dialect{},
		},
	}
	// Get the latest possible migration
	latest, err := goose.GetMostRecentDBVersion(migrateConf.MigrationsDir)
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Open our database connection
	db, err = gorm.Open("sqlite3", config.Conf.DBPath)
	db.LogMode(false)
	db.SetLogger(Logger)
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Migrate up to the latest version
	err = goose.RunMigrationsOnDb(migrateConf, migrateConf.MigrationsDir, latest, db.DB())
	if err != nil {
		Logger.Println(err)
		return err
	}
	//If the database didn't exist, we need to create the admin user
	if create_db {
		//Create the default user
		initUser := User{
			Username: "admin",
			Hash:     "$2a$10$IYkPp0.QsM81lYYPrQx6W.U6oQGw7wMpozrKhKAHUBVL4mkm/EvAS", //gophish
		}
		initUser.ApiKey = generateSecureKey()
		err = db.Save(&initUser).Error
		if err != nil {
			Logger.Println(err)
			return err
		}
	}
	return nil
}
