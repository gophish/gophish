package main

import (
	"github.com/jordan-wright/gophish/db"
	"testing"
)

func TestDBSetup(t *testing.T) {
	err := db.Setup()
	if err != nil {
		t.Fatalf("Failed creating database: %v", err)
	}
}
