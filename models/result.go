package models

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/mail"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/oschwald/maxminddb-golang"
)

type mmCity struct {
	GeoPoint mmGeoPoint `maxminddb:"location"`
}

type mmGeoPoint struct {
	Latitude  float64 `maxminddb:"latitude"`
	Longitude float64 `maxminddb:"longitude"`
}

// Result contains the fields for a result object,
// which is a representation of a target in a campaign.
type Result struct {
	Id         int64     `json:"-"`
	CampaignId int64     `json:"-"`
	UserId     int64     `json:"-"`
	RId        string    `json:"id"`
	Email      string    `json:"email"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Position   string    `json:"position"`
	Status     string    `json:"status" sql:"not null"`
	IP         string    `json:"ip"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	SendDate   time.Time `json:"send_date"`
}

// UpdateStatus updates the status of the result in the database
func (r *Result) UpdateStatus(s string) error {
	return db.Table("results").Where("id=?", r.Id).Update("status", s).Error
}

// UpdateGeo updates the latitude and longitude of the result in
// the database given an IP address
func (r *Result) UpdateGeo(addr string) error {
	// Open a connection to the maxmind db
	mmdb, err := maxminddb.Open("static/db/geolite2-city.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer mmdb.Close()
	ip := net.ParseIP(addr)
	var city mmCity
	// Get the record
	err = mmdb.Lookup(ip, &city)
	if err != nil {
		return err
	}
	// Update the database with the record information
	return db.Table("results").Where("id=?", r.Id).Updates(map[string]interface{}{
		"ip":        addr,
		"latitude":  city.GeoPoint.Latitude,
		"longitude": city.GeoPoint.Longitude,
	}).Error
}

// GenerateId generates a unique key to represent the result
// in the database
func (r *Result) GenerateId() error {
	// Keep trying until we generate a unique key (shouldn't take more than one or two iterations)
	const alphaNum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	k := make([]byte, 7)
	for {
		for i := range k {
			idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphaNum))))
			if err != nil {
				return err
			}
			k[i] = alphaNum[idx.Int64()]
		}
		r.RId = string(k)
		err := db.Table("results").Where("r_id=?", r.RId).First(&Result{}).Error
		if err == gorm.ErrRecordNotFound {
			break
		}
	}
	return nil
}

// Returns the email address to use in the "To" header of the email
func (r *Result) FormatAddress() string {
	addr := r.Email
	if r.FirstName != "" && r.LastName != "" {
		a := &mail.Address{
			Name:    fmt.Sprintf("%s %s", r.FirstName, r.LastName),
			Address: r.Email,
		}
		addr = a.String()
	}
	return addr
}

// GetResult returns the Result object from the database
// given the ResultId
func GetResult(rid string) (Result, error) {
	r := Result{}
	err := db.Where("r_id=?", rid).First(&r).Error
	return r, err
}
