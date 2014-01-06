package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

//Setup creates and returns the database needed by Gophish
func Setup() (*sql.DB, error) {
	//If the file already exists, delete it and recreate it
	if _, err := os.Stat(config.DBPath); err == nil {
		os.Remove(config.DBPath)
	}
	fmt.Println("Creating db at " + config.DBPath)
	db, err := sql.Open("sqlite3", config.DBPath)
	if err != nil {
		return nil, err
	}
	//Create the tables needed
	_, err = db.Exec(
		`CREATE TABLE Users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, hash VARCHAR(32), apikey VARCHAR(32));`)
	if err != nil {
		return nil, err
	}
	//Create the default user
	stmt, err := db.Prepare(`INSERT INTO Users (username, hash, apikey) VALUES (?, ?, ?);`)
	defer stmt.Close()
	if err != nil {
		return nil, err
	}
	_, err = stmt.Exec("jordan", "12345678901234567890123456789012", "12345678901234567890123456789012")
	if err != nil {
		return nil, err
	}
	return db, nil
}
