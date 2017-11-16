package models

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
        "time"
	"bitbucket.org/liamstask/goose/lib/goose"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3" // Blank import needed to import sqlite3
	"github.com/teamnsrg/gophish/config"
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
	CAMPAIGN_CREATED     string = "Created"
	CAMPAIGN_EMAILS_SENT string = "Emails Sent"
	CAMPAIGN_COMPLETE    string = "Completed"
	EVENT_SENT           string = "Email Sent"
	EVENT_SENDING_ERROR  string = "Error Sending Email"
	EVENT_OPENED         string = "Email Opened"
	EVENT_CLICKED        string = "Clicked Link"
	EVENT_DATA_SUBMIT    string = "Submitted Data"
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

func chooseDBDriver(name, openStr string) goose.DBDriver {
	d := goose.DBDriver{Name: name, OpenStr: openStr}

	switch name {
	case "mysql":
		d.Import = "github.com/go-sql-driver/mysql"
		d.Dialect = &goose.MySqlDialect{}
	case "postgres":
		d.Import = "github.com/lib/pq"
		d.Dialect = &goose.PostgresDialect{}
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
	create_db := false
	if _, err = os.Stat(config.Conf.DBPath); err != nil || config.Conf.DBPath == ":memory:" {
		create_db = true
	}
	// Setup the goose configuration
	migrateConf := &goose.DBConf{
		MigrationsDir: config.Conf.MigrationsPath,
		Env:           "production",
		Driver:        chooseDBDriver(config.Conf.DBName, config.Conf.DBPath),
	}
	// Get the latest possible migration
	latest, err := goose.GetMostRecentDBVersion(migrateConf.MigrationsDir)
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Open our database connection
	db, err = gorm.Open(config.Conf.DBName, config.Conf.DBPath)
	db.LogMode(false)
	db.SetLogger(Logger)
	db.DB().SetMaxOpenConns(1)
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

		//Populate the Database
		if config.Conf.Environment == "production" {
			Logger.Printf("Production Database seed not Implemented")
		} else {
			Logger.Printf("Setting up Dev Environment Database with demo campaign...")

                        //Construct Demo Targets
                        target1 := Target{
                            Id: 1,
                            FirstName: "Keepak",
                            LastName: "Dumar",
                            Email: "dpk@km.ar",
                            Position: "Chief Solutions Officer",
                        }
                        target2 := Target{
                            Id: 2,
                            FirstName: "Mane",
                            LastName: "Za",
                            Email: "zzma@km.ar",
                            Position: "Go Guru Recruiter",
                        }
                        target3 := Target{
                            Id: 3,
                            FirstName: "Roshua",
                            LastName: "Jeynolds",
                            Email: "reyjey@km.ar",
                            Position: "Director of Good Movies and/or Snacks",
                        }
                        target4 := Target{
                            Id: 4,
                            FirstName: "Robert",
                            LastName: "La'Bla",
                            Email: "barkbark@km.ar",
                            Position: "Fifth Wheel",
                        }
                        target5 := Target{
                            Id: 5,
                            FirstName: "Michael",
                            LastName: "Bailey, M.D.",
                            Email: "thatkindofdoctor@km.ar",
                            Position: "Trauma Accupuncture Specialist",
                        }

                        //Save Targets
		        err = db.Save(&target1).Error
		        if err != nil {
                            Logger.Printf("Error saving Target 1")
			    Logger.Println(err)
			    return err
		        }
		        err = db.Save(&target2).Error
		        if err != nil {
                            Logger.Printf("Error saving Target 2")
			    Logger.Println(err)
			    return err
		        }
		        err = db.Save(&target3).Error
		        if err != nil {
                            Logger.Printf("Error saving Target 3")
			    Logger.Println(err)
			    return err
		        }
		        err = db.Save(&target4).Error
		        if err != nil {
                            Logger.Printf("Error saving Target 4")
			    Logger.Println(err)
			    return err
		        }
		        err = db.Save(&target5).Error
		        if err != nil {
                            Logger.Printf("Error saving Target 5")
			    Logger.Println(err)
			    return err
		        }

                        //Build Groups
                        group1 := Group{
                            Id : 1,
                            UserId : 1,
                            Name : "IT Dept.",
                            ModifiedDate : time.Now(),
                            Targets : []Target{target1, target2, target3,},
                        }
                        group2 := Group{
                            Id : 2,
                            UserId : 1,
                            Name : "Admin Dept.",
                            ModifiedDate : time.Now(),
                            Targets : []Target{target4, target4,},
                        }

                        //Save Groups
		        err = db.Save(&group1).Error
		        if err != nil {
                            Logger.Printf("Error saving Group 1")
			    Logger.Println(err)
			    return err
		        }
		        err = db.Save(&group2).Error
		        if err != nil {
                            Logger.Printf("Error saving Group 2")
			    Logger.Println(err)
			    return err
		        }
                        Logger.Printf("Done Prepopulating")

                        //Build Group -> Target Links
                        g1_t1 := GroupTarget {
                            GroupId  : 1,
                            TargetId : 1,
                        }
                        g1_t2 := GroupTarget {
                            GroupId  : 1,
                            TargetId : 2,
                        }
                        g1_t3 := GroupTarget {
                            GroupId  : 1,
                            TargetId : 3,
                        }
                        g2_t3 := GroupTarget {
                            GroupId  : 2,
                            TargetId : 3,
                        }
                        g2_t4 := GroupTarget {
                            GroupId  : 2,
                            TargetId : 4,
                        }
                        g2_t5 := GroupTarget {
                            GroupId  : 2,
                            TargetId : 5,
                        }

                        //Save Group->Target Links
		        err = db.Save(&g1_t1).Error
		        if err != nil {
                            Logger.Printf("Error saving Group-Target 1")
			    Logger.Println(err)
			    return err
		        }
		        err = db.Save(&g1_t2).Error
		        if err != nil {
                            Logger.Printf("Error saving Group-Target 2")
			    Logger.Println(err)
			    return err
		        }
		        err = db.Save(&g1_t3).Error
		        if err != nil {
                            Logger.Printf("Error saving Group-Target 3")
			    Logger.Println(err)
			    return err
		        }
		        err = db.Save(&g2_t3).Error
		        if err != nil {
                            Logger.Printf("Error saving Group-Target 4")
			    Logger.Println(err)
			    return err
		        }
		        err = db.Save(&g2_t4).Error
		        if err != nil {
                            Logger.Printf("Error saving Group-Target 5")
			    Logger.Println(err)
			    return err
		        }
		        err = db.Save(&g2_t5).Error
		        if err != nil {
                            Logger.Printf("Error saving Group-Target 6")
			    Logger.Println(err)
			    return err
		        }

                        //Make a Sending Profile
                        sending_profile := SMTP{
                            Id          : 1,
                            UserId      : 1,
                            Interface   : "SMTP",
                            Name        : "Default Sending Profile",
                            Host        : "example.com",
                            Username    : "test",
                            Password    : "12345",
                            FromAddress : "example.com",
                            IgnoreCertErrors : true,
                            Headers     : []Header{},
                            ModifiedDate: time.Now(),
                        }

                        //Save Sending Profile
		        err = db.Save(&sending_profile).Error
		        if err != nil {
                            Logger.Printf("Error saving Sending Profile")
			    Logger.Println(err)
			    return err
		        }

                        //Make an Email Template
                        email_template := Template{
                            Id       : 1,
                            UserId   : 1,
                            Name     : "Default Email Template",
                            Subject  : "You should click this link",
                            Text     : "Hi\nYou should go to this link {LINK}.\nSincerely,\nYour Boss",
                            HTML     : "<html></html>",
                            ModifiedDate: time.Now(),
                            Attachments : []Attachment{},
                        }

                        //Save Email Template
		        err = db.Save(&email_template).Error
		        if err != nil {
                            Logger.Printf("Error saving Landing Page")
			    Logger.Println(err)
			    return err
		        }

                        //TODO Figure out why this wont work 
                        //Make a landing page
                        landing_page := Page{
	                    Id           : 1,
	                    UserId       : 1,
                            Name         : "Default Phish Page",
	                    HTML         : "<html><head></head><body><form><p>Please type sensitive information here</p><input></input><form></body></html>",
	                    CaptureCredentials : false,
	                    CapturePasswords : false,
	                    RedirectURL  : "http://example.com",
	                    ModifiedDate : time.Now(),
                        }

                        //Save Landing Page
		        err = db.Save(&landing_page).Error
		        if err != nil {
                            Logger.Printf("Error saving Landing Page")
			    Logger.Println(err)
			    return err
		        }

                        //Make Campaign
                        campaign := Campaign{
	                    Id             : 1,
	                    UserId         : 1,
	                    Name           : "Demo Campaign",
	                    CreatedDate    : time.Now(),
	                    LaunchDate     : time.Now(),
	                    CompletedDate  : time.Now(),
	                    TemplateId     : 1,
	                    Template       : email_template,
	                    PageId         : 1,
	                    Page           : landing_page,
	                    Status         : "In Development",
	                    Results        : []Result{},
	                    Groups         : []Group{group1, group2},
	                    Events         : []Event{},
	                    SMTPId         : 1,
	                    SMTP           : sending_profile,
	                    URL            : "www.example.com",
                        }

                        //Save Campaign 
		        err = db.Save(&campaign).Error
		        if err != nil {
                            Logger.Printf("Error saving Campaign")
			    Logger.Println(err)
			    return err
		        }
                        Logger.Printf("Successfully initialized data")
		}

	}
	return nil
}
