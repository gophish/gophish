package util

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"

	"github.com/jordan-wright/gophish/models"
)

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
