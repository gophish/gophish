package main

import (
	"testing"

	"github.com/jordan-wright/gophish/models"
)

func TestDBSetup(t *testing.T) {
	err := models.Setup()
	if err != nil {
		t.Fatalf("Failed creating database: %v", err)
	}
}
