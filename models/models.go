package models

import (
	"crypto/rand"
	"fmt"
	"io"

	"bitbucket.org/liamstask/goose/lib/goose"

	"github.com/binodlamsal/gophish/config"
	log "github.com/binodlamsal/gophish/logger"
	_ "github.com/go-sql-driver/mysql" // Blank import needed to import mysql
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3" // Blank import needed to import sqlite3
)

var db *gorm.DB
var err error

const (
	CAMPAIGN_IN_PROGRESS string = "In progress"
	CAMPAIGN_QUEUED      string = "Queued"
	CAMPAIGN_CREATED     string = "Created"
	CAMPAIGN_EMAILS_SENT string = "Emails Sent"
	CAMPAIGN_COMPLETE    string = "Completed"
	EVENT_SENT           string = "Email Sent"
	EVENT_SENDING_ERROR  string = "Error Sending Email"
	EVENT_OPENED         string = "Email Opened"
	EVENT_CLICKED        string = "Clicked Link"
	EVENT_DATA_SUBMIT    string = "Submitted Data"
	EVENT_REPORTED       string = "Email Reported"
	EVENT_PROXY_REQUEST  string = "Proxied request"
	STATUS_SUCCESS       string = "Success"
	STATUS_QUEUED        string = "Queued"
	STATUS_SENDING       string = "Sending"
	STATUS_UNKNOWN       string = "Unknown"
	STATUS_SCHEDULED     string = "Scheduled"
	STATUS_RETRY         string = "Retrying"
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

func chooseDBDriver(name, openStr string) goose.DBDriver {
	d := goose.DBDriver{Name: name, OpenStr: openStr}

	switch name {
	case "mysql":
		d.Import = "github.com/go-sql-driver/mysql"
		d.Dialect = &goose.MySqlDialect{}

	// Default database is sqlite3
	default:
		d.Import = "github.com/mattn/go-sqlite3"
		d.Dialect = &goose.Sqlite3Dialect{}
	}

	return d
}

// Setup initializes the Conn object
// It also populates the Gophish Config object
func Setup() error {
	// Setup the goose configuration
	migrateConf := &goose.DBConf{
		MigrationsDir: config.Conf.MigrationsPath,
		Env:           "production",
		Driver:        chooseDBDriver(config.Conf.DBName, config.Conf.DBPath),
	}
	// Get the latest possible migration
	latest, err := goose.GetMostRecentDBVersion(migrateConf.MigrationsDir)
	if err != nil {
		log.Error(err)
		return err
	}
	// Open our database connection
	db, err = gorm.Open(config.Conf.DBName, config.Conf.DBPath)
	db.LogMode(false)
	db.SetLogger(log.Logger)
	db.DB().SetMaxOpenConns(1)
	if err != nil {
		log.Error(err)
		return err
	}
	// Migrate up to the latest version
	err = goose.RunMigrationsOnDb(migrateConf, migrateConf.MigrationsDir, latest, db.DB())
	if err != nil {
		log.Error(err)
		return err
	}
	// Create the admin user if it doesn't exist
	var userCount int64
	db.Model(&User{}).Count(&userCount)
	if userCount == 0 {
		initUser := User{
			Username: "admin",
			Hash:     "$2a$10$IYkPp0.QsM81lYYPrQx6W.U6oQGw7wMpozrKhKAHUBVL4mkm/EvAS", //gophish
		}
		initUser.ApiKey = generateSecureKey()
		err = db.Save(&initUser).Error
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}
