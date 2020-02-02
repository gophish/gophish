package util

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"testing"

	"github.com/gophish/gophish/models"
)

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

func TestParseCSVEmail(t *testing.T) {
	expected := models.Target{
		BaseRecipient: models.BaseRecipient{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "johndoe@example.com",
		},
	}

	csvPayload := fmt.Sprintf("%s,%s,<%s>", expected.FirstName, expected.LastName, expected.Email)
	r, err := buildCSVRequest(csvPayload)
	if err != nil {
		t.Fatalf("error building CSV request: %v", err)
	}

	got, err := ParseCSV(r)
	if err != nil {
		t.Fatalf("error parsing CSV: %v", err)
	}
	expectedLength := 1
	if len(got) != expectedLength {
		t.Fatalf("invalid number of results received from CSV. expected %d got %d", expectedLength, len(got))
	}
	if !reflect.DeepEqual(expected, got[0]) {
		t.Fatalf("Incorrect targets received. Expected: %#v\nGot: %#v", expected, got)
	}
}
