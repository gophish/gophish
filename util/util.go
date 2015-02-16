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
		fn := ""
		ln := ""
		ea := ""
		for i, v := range record {
			fmt.Println(v)
			switch {
			case v == "First Name":
				fi = i
			case v == "Last Name":
				li = i
			case v == "Email":
				ei = i
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
			t := models.Target{
				FirstName: fn,
				LastName:  ln,
				Email:     ea,
			}
			ts = append(ts, t)
		}
	}
	return ts, nil
}
