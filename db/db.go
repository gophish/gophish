package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/jordan-wright/gophish/config"
	"github.com/jordan-wright/gophish/models"
	_ "github.com/mattn/go-sqlite3"
)

var Conn *gorp.DbMap
var DB *sql.DB
var err error

// Setup initializes the Conn object
// It also populates the Gophish Config object
func Setup() error {
	DB, err := sql.Open("sqlite3", config.Conf.DBPath)
	Conn = &gorp.DbMap{Db: DB, Dialect: gorp.SqliteDialect{}}
	//If the file already exists, delete it and recreate it
	_, err = os.Stat(config.Conf.DBPath)
	Conn.AddTableWithName(models.User{}, "users").SetKeys(true, "Id")
	Conn.AddTableWithName(models.Campaign{}, "campaigns").SetKeys(true, "Id")
	Conn.AddTableWithName(models.Group{}, "groups").SetKeys(true, "Id")
	Conn.AddTableWithName(models.GroupTarget{}, "group_target")
	if err != nil {
		fmt.Println("Database not found, recreating...")
		createTablesSQL := []string{
			//Create tables
			`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT NOT NULL, hash VARCHAR(60) NOT NULL, api_key VARCHAR(32));`,
			`CREATE TABLE campaigns (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, created_date TIMESTAMP NOT NULL, completed_date TIMESTAMP, template TEXT, status TEXT NOT NULL, uid INTEGER, FOREIGN KEY (uid) REFERENCES users(id));`,
			`CREATE TABLE targets (id INTEGER PRIMARY KEY AUTOINCREMENT, address TEXT NOT NULL);`,
			`CREATE TABLE groups (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL);`,
			`CREATE TABLE group_targets (gid INTEGER NOT NULL, tid INTEGER NOT NULL, FOREIGN KEY (gid) REFERENCES groups(id), FOREIGN KEY (tid) REFERENCES targets(id));`,
		}
		fmt.Println("Creating db at " + config.Conf.DBPath)
		//Create the tables needed
		for _, stmt := range createTablesSQL {
			_, err = DB.Exec(stmt)
			if err != nil {
				return err
			}
		}
		//Create the default user
		init_user := models.User{
			Username: "admin",
			Hash:     "$2a$10$d4OtT.RkEOQn.iruVWIQ5u8CeV/85ZYF41y8wKeUwsAPqPNFvTccW",
			APIKey:   "12345678901234567890123456789012",
		}
		Conn.Insert(&init_user)
		if err != nil {
			fmt.Println(err)
		}
		c := models.Campaign{
			Name:          "Test Campaigns",
			CreatedDate:   time.Now().UTC(),
			CompletedDate: time.Now().UTC(),
			Template:      "test template",
			Status:        "In progress",
			Uid:           init_user.Id,
		}
		Conn.Insert(&c)
	}
	return nil
}
