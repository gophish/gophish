package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/jinzhu/gorm"
	"github.com/jordan-wright/gophish/config"
	_ "github.com/mattn/go-sqlite3"
)

var Conn *gorp.DbMap
var db gorm.DB
var DB *sql.DB
var err error
var ErrUsernameTaken = errors.New("username already taken")
var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

// Setup initializes the Conn object
// It also populates the Gophish Config object
func Setup() error {
	DB, err := sql.Open("sqlite3", config.Conf.DBPath)
	if err != nil {
		fmt.Printf("%v", err)
	}
	db, err = gorm.Open("sqlite3", "gophish_gorm.db")
	db.SetLogger(Logger)
	db.CreateTable(User{})
	db.CreateTable(Target{})
	db.CreateTable(Result{})
	db.CreateTable(Group{})
	db.CreateTable(GroupTarget{})
	db.CreateTable(UserGroup{})
	db.CreateTable(Template{})
	db.CreateTable(Campaign{})
	Conn = &gorp.DbMap{Db: DB, Dialect: gorp.SqliteDialect{}}
	//If the file already exists, delete it and recreate it
	_, err = os.Stat(config.Conf.DBPath)
	Conn.AddTableWithName(User{}, "users").SetKeys(true, "Id")
	Conn.AddTableWithName(Campaign{}, "campaigns").SetKeys(true, "Id")
	Conn.AddTableWithName(Group{}, "groups").SetKeys(true, "Id")
	Conn.AddTableWithName(Template{}, "templates").SetKeys(true, "Id")
	if err != nil {
		Logger.Println("Database not found, recreating...")
		createTablesSQL := []string{
			//Create tables
			`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT NOT NULL, hash VARCHAR(60) NOT NULL, api_key VARCHAR(32), UNIQUE(username), UNIQUE(api_key));`,
			`CREATE TABLE campaigns (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, created_date TIMESTAMP NOT NULL, completed_date TIMESTAMP, template TEXT, status TEXT NOT NULL);`,
			`CREATE TABLE targets (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT NOT NULL, UNIQUE(email));`,
			`CREATE TABLE groups (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, modified_date TIMESTAMP NOT NULL);`,
			`CREATE TABLE campaign_results (cid INTEGER NOT NULL, email TEXT NOT NULL, status TEXT NOT NULL, FOREIGN KEY (cid) REFERENCES campaigns(id), UNIQUE(cid, email, status))`,
			`CREATE TABLE templates (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, modified_date TIMESTAMP NOT NULL, html TEXT NOT NULL, text TEXT NOT NULL);`,
			`CREATE TABLE files (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, path TEXT NOT NULL);`,
			`CREATE TABLE user_campaigns (uid INTEGER NOT NULL, cid INTEGER NOT NULL, FOREIGN KEY (uid) REFERENCES users(id), FOREIGN KEY (cid) REFERENCES campaigns(id), UNIQUE(uid, cid))`,
			`CREATE TABLE user_groups (uid INTEGER NOT NULL, gid INTEGER NOT NULL, FOREIGN KEY (uid) REFERENCES users(id), FOREIGN KEY (gid) REFERENCES groups(id), UNIQUE(uid, gid))`,
			`CREATE TABLE group_targets (gid INTEGER NOT NULL, tid INTEGER NOT NULL, FOREIGN KEY (gid) REFERENCES groups(id), FOREIGN KEY (tid) REFERENCES targets(id), UNIQUE(gid, tid));`,
			`CREATE TABLE user_templates (uid INTEGER NOT NULL, tid INTEGER NOT NULL, FOREIGN KEY (uid) REFERENCES users(id), FOREIGN KEY (tid) REFERENCES templates(id), UNIQUE(uid, tid));`,
			`CREATE TABLE template_files (tid INTEGER NOT NULL, fid INTEGER NOT NULL, FOREIGN KEY (tid) REFERENCES templates(id), FOREIGN KEY(fid) REFERENCES files(id), UNIQUE(tid, fid));`,
		}
		Logger.Printf("Creating db at %s\n", config.Conf.DBPath)
		//Create the tables needed
		for _, stmt := range createTablesSQL {
			_, err = DB.Exec(stmt)
			if err != nil {
				return err
			}
		}
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
		init_campaign := Campaign{
			UserId:        1,
			Name:          "First Campaign",
			CreatedDate:   time.Now(), //gophish
			CompletedDate: time.Now(),
			Template:      "",
			Status:        "In Progress",
			Results:       []Result{},
			Groups:        []Group{},
		}
		err = db.Debug().Save(&init_campaign).Error
		if err != nil {
			Logger.Println(err)
		}
		init_result := Result{
			CampaignId: 1,
			Email:      "test@example.com",
			Status:     "Unsuccessful",
		}
		err = db.Debug().Save(&init_result).Error
		if err != nil {
			Logger.Println(err)
		}
		init_group := Group{
			Name:         "New Group",
			ModifiedDate: time.Now(),
		}
		err = db.Debug().Save(&init_group).Error
		if err != nil {
			Logger.Println(err)
		}
		init_ug := UserGroup{
			UserId:  1,
			GroupId: 1,
		}
		err = db.Debug().Save(&init_ug).Error
		if err != nil {
			Logger.Println(err)
		}
		init_target := Target{
			Email: "test@example.com",
		}
		err = db.Debug().Save(&init_target).Error
		if err != nil {
			Logger.Println(err)
		}
		init_gt := GroupTarget{
			GroupId:  1,
			TargetId: 1,
		}
		err = db.Debug().Save(&init_gt).Error
		if err != nil {
			Logger.Println(err)
		}
	}
	return nil
}

// Flash is used to hold flash information for use in templates.
type Flash struct {
	Type    string
	Message string
}
