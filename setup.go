package main

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"os"

	"github.com/jordan-wright/gophish/config"
	"github.com/jordan-wright/gophish/models"
	_ "github.com/mattn/go-sqlite3"
)

var Db sql.DB

//init registers the necessary models to be saved in the session later
func init() {
	gob.Register(&models.User{})
}

// Setup creates and returns the database needed by Gophish.
// It also populates the Gophish Config object
func Setup() error {
	//If the file already exists, delete it and recreate it
	if _, err := os.Stat(config.Conf.DBPath); err == nil {
		os.Remove(Conf.DBPath)
	}
	fmt.Println("Creating db at " + config.Conf.DBPath)
	db, err := sql.Open("sqlite3", config.Conf.DBPath)
	defer db.Close()
	if err != nil {
		return err
	}
	//Create the tables needed
	_, err = db.Exec(
		`CREATE TABLE Users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, hash VARCHAR(32), apikey VARCHAR(32));`)
	if err != nil {
		return err
	}
	//Create the default user
	stmt, err := db.Prepare(`INSERT INTO Users (username, hash, apikey) VALUES (?, ?, ?);`)
	defer stmt.Close()
	if err != nil {
		return err
	}
	_, err = stmt.Exec("jordan", "12345678901234567890123456789012", "12345678901234567890123456789012")
	if err != nil {
		return err
	}
	return nil
}
