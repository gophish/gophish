package util

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"testing"

	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/models"
	"github.com/stretchr/testify/suite"
)

type UtilSuite struct {
	suite.Suite
}

func (s *UtilSuite) SetupSuite() {
	config.Conf.DBName = "sqlite3"
	config.Conf.DBPath = ":memory:"
	config.Conf.MigrationsPath = "../db/db_sqlite3/migrations/"
	err := models.Setup()
	if err != nil {
		s.T().Fatalf("Failed creating database: %v", err)
	}
	s.Nil(err)
}

func buildCSVRequest(csvPayload string) (*http.Request, error) {
	csvHeader := "First Name,Last Name,Email\n"
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files[]", "example.csv")
	if err != nil {
		return nil, err
	}
	part.Write([]byte(csvHeader))
	part.Write([]byte(csvPayload))
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	r, err := http.NewRequest("POST", "http://127.0.0.1", body)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", writer.FormDataContentType())
	return r, nil
}

func (s *UtilSuite) TestParseCSVEmail() {
	expected := models.Target{
		BaseRecipient: models.BaseRecipient{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "johndoe@example.com",
		},
	}

	csvPayload := fmt.Sprintf("%s,%s,<%s>", expected.FirstName, expected.LastName, expected.Email)
	r, err := buildCSVRequest(csvPayload)
	s.Nil(err)

	got, err := ParseCSV(r)
	s.Nil(err)
	s.Equal(len(got), 1)
	if !reflect.DeepEqual(expected, got[0]) {
		s.T().Fatalf("Incorrect targets received. Expected: %#v\nGot: %#v", expected, got)
	}
}

func TestUtilSuite(t *testing.T) {
	suite.Run(t, new(UtilSuite))
}
