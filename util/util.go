package util

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/mail"

	"github.com/jordan-wright/email"
	"github.com/jordan-wright/gophish/models"
)

// ParseMail takes in an HTTP Request and returns an Email object
// TODO: This function will likely be changed to take in a []byte
func ParseMail(r *http.Request) (email.Email, error) {
	e := email.Email{}
	m, err := mail.ReadMessage(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	body, err := ioutil.ReadAll(m.Body)
	e.HTML = body
	return e, err
}

// ParseCSV contains the logic to parse the user provided csv file containing Target entries
func ParseCSV(r *http.Request) ([]models.Target, error) {
	mr, err := r.MultipartReader()
	ts := []models.Target{}
	if err != nil {
		return ts, err
	}
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		// Skip the "submit" part
		if part.FileName() == "" {
			continue
		}
		defer part.Close()
		reader := csv.NewReader(part)
		reader.TrimLeadingSpace = true
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		fi := -1
		li := -1
		ei := -1
		pi := -1
		fn := ""
		ln := ""
		ea := ""
		ps := ""
		for i, v := range record {
			switch {
			case v == "First Name":
				fi = i
			case v == "Last Name":
				li = i
			case v == "Email":
				ei = i
			case v == "Position":
				pi = i
			}
		}
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if fi != -1 {
				fn = record[fi]
			}
			if li != -1 {
				ln = record[li]
			}
			if ei != -1 {
				ea = record[ei]
			}
			if pi != -1 {
				ps = record[pi]
			}
			t := models.Target{
				FirstName: fn,
				LastName:  ln,
				Email:     ea,
				Position:  ps,
			}
			ts = append(ts, t)
		}
	}
	return ts, nil
}
