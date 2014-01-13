package db

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/jordan-wright/gophish/config"
	_ "github.com/mattn/go-sqlite3"
)

var Conn *sql.DB

// Setup initializes the Conn object
// It also populates the Gophish Config object
func Setup() error {
	createTablesSQL := []string{
		//Create tables
		`CREATE TABLE Users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT NOT NULL, hash VARCHAR(60) NOT NULL, apikey VARCHAR(32));`,
		`CREATE TABLE Campaigns (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, created_date TEXT NOT NULL, completed_date TEXT, status TEXT NOT NULL);`,
	}
	//If the file already exists, delete it and recreate it
	_, err := os.Stat(config.Conf.DBPath)
	if err == nil {
		os.Remove(config.Conf.DBPath)
	}
	fmt.Println("Creating db at " + config.Conf.DBPath)
	Conn, err = sql.Open("sqlite3", config.Conf.DBPath)
	if err != nil {
		return err
	}
	//Create the tables needed
	for _, stmt := range createTablesSQL {
		_, err = Conn.Exec(stmt)
		if err != nil {
			return err
		}
	}
	//Create the default user
	stmt, err := Conn.Prepare(`INSERT INTO Users (username, hash, apikey) VALUES (?, ?, ?);`)
	defer stmt.Close()
	if err != nil {
		return err
	}
	_, err = stmt.Exec("jordan", "$2a$10$d4OtT.RkEOQn.iruVWIQ5u8CeV/85ZYF41y8wKeUwsAPqPNFvTccW", "12345678901234567890123456789012")
	if err != nil {
		return err
	}
	return nil
}
