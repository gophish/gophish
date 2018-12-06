package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net"
	"time"

	log "github.com/gophish/gophish/logger"
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
	Id           int64     `json:"-"`
	CampaignId   int64     `json:"-"`
	UserId       int64     `json:"-"`
	RId          string    `json:"id"`
	Status       string    `json:"status" sql:"not null"`
	IP           string    `json:"ip"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	SendDate     time.Time `json:"send_date"`
	Reported     bool      `json:"reported" sql:"not null"`
	ModifiedDate time.Time `json:"modified_date"`
	BaseRecipient
}

func (r *Result) createEvent(status string, details interface{}) (*Event, error) {
	c, err := GetCampaign(r.CampaignId, r.UserId)
	if err != nil {
		return nil, err
	}
	e := &Event{Email: r.Email, Message: status}
	if details != nil {
		dj, err := json.Marshal(details)
		if err != nil {
			return nil, err
		}

		if status == EVENT_DATA_SUBMIT && c.PublicKeyId != 0 { // Zero is unset
			//Taken from crypto/cipher CFB example
			key := make([]byte, 32)

			if _, err = rand.Read(key); err != nil { // 32 Bytes here selects for AES256
				return nil, err
			}

			blockCipher, err := aes.NewCipher(key)
			if err != nil {
				return nil, err
			}

			blockCiphertext := make([]byte, aes.BlockSize+len(dj))
			iv := blockCiphertext[:aes.BlockSize] // IV must be unique, however doesnt need to be secret
			if _, err := io.ReadFull(rand.Reader, iv); err != nil {
				return nil, err
			}

			streamCipher := cipher.NewCFBEncrypter(blockCipher, iv)
			streamCipher.XORKeyStream(blockCiphertext[aes.BlockSize:], dj) // IV:plaintext

			publcKeyStructure, err := GetPublicKey(c.PublicKeyId, r.UserId)
			if err != nil {
				return nil, err
			}

			pubKey, err := DecodePEMBlock(publcKeyStructure.PubKey)
			if err != nil {
				return nil, err
			}

			keyCipherText, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, key, []byte("key"))
			if err != nil {
				return nil, err
			}

			e.Key = base64.StdEncoding.EncodeToString(keyCipherText)
			e.Details = base64.StdEncoding.EncodeToString(blockCiphertext)
		} else {
			e.Details = string(dj)
		}
	}
	c.AddEvent(e)
	return e, nil
}

// HandleEmailSent updates a Result to indicate that the email has been
// successfully sent to the remote SMTP server
func (r *Result) HandleEmailSent() error {
	event, err := r.createEvent(EVENT_SENT, nil)
	if err != nil {
		return err
	}
	r.SendDate = event.Time
	r.Status = EVENT_SENT
	r.ModifiedDate = event.Time
	return db.Save(r).Error
}

// HandleEmailError updates a Result to indicate that there was an error when
// attempting to send the email to the remote SMTP server.
func (r *Result) HandleEmailError(err error) error {
	event, err := r.createEvent(EVENT_SENDING_ERROR, EventError{Error: err.Error()})
	if err != nil {
		return err
	}
	r.Status = ERROR
	r.ModifiedDate = event.Time
	return db.Save(r).Error
}

// HandleEmailBackoff updates a Result to indicate that the email received a
// temporary error and needs to be retried
func (r *Result) HandleEmailBackoff(err error, sendDate time.Time) error {
	event, err := r.createEvent(EVENT_SENDING_ERROR, EventError{Error: err.Error()})
	if err != nil {
		return err
	}
	r.Status = STATUS_RETRY
	r.SendDate = sendDate
	r.ModifiedDate = event.Time
	return db.Save(r).Error
}

// HandleEmailOpened updates a Result in the case where the recipient opened the
// email.
func (r *Result) HandleEmailOpened(details EventDetails) error {
	event, err := r.createEvent(EVENT_OPENED, details)
	if err != nil {
		return err
	}
	// Don't update the status if the user already clicked the link
	// or submitted data to the campaign
	if r.Status == EVENT_CLICKED || r.Status == EVENT_DATA_SUBMIT {
		return nil
	}
	r.Status = EVENT_OPENED
	r.ModifiedDate = event.Time
	return db.Save(r).Error
}

// HandleClickedLink updates a Result in the case where the recipient clicked
// the link in an email.
func (r *Result) HandleClickedLink(details EventDetails) error {
	event, err := r.createEvent(EVENT_CLICKED, details)
	if err != nil {
		return err
	}
	// Don't update the status if the user has already submitted data via the
	// landing page form.
	if r.Status == EVENT_DATA_SUBMIT {
		return nil
	}
	r.Status = EVENT_CLICKED
	r.ModifiedDate = event.Time
	return db.Save(r).Error
}

// HandleFormSubmit updates a Result in the case where the recipient submitted
// credentials to the form on a Landing Page.
func (r *Result) HandleFormSubmit(details EventDetails) error {
	event, err := r.createEvent(EVENT_DATA_SUBMIT, details)
	if err != nil {
		return err
	}
	r.Status = EVENT_DATA_SUBMIT
	r.ModifiedDate = event.Time
	return db.Save(r).Error
}

// HandleEmailReport updates a Result in the case where they report a simulated
// phishing email using the HTTP handler.
func (r *Result) HandleEmailReport(details EventDetails) error {
	event, err := r.createEvent(EVENT_REPORTED, details)
	if err != nil {
		return err
	}
	r.Reported = true
	r.ModifiedDate = event.Time
	return db.Save(r).Error
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
	r.IP = addr
	r.Latitude = city.GeoPoint.Latitude
	r.Longitude = city.GeoPoint.Longitude
	return db.Save(r).Error
}

func generateResultId() (string, error) {
	const alphaNum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	k := make([]byte, 7)
	for i := range k {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphaNum))))
		if err != nil {
			return "", err
		}
		k[i] = alphaNum[idx.Int64()]
	}
	return string(k), nil
}

// GenerateId generates a unique key to represent the result
// in the database
func (r *Result) GenerateId() error {
	// Keep trying until we generate a unique key (shouldn't take more than one or two iterations)
	for {
		rid, err := generateResultId()
		if err != nil {
			return err
		}
		r.RId = rid
		err = db.Table("results").Where("r_id=?", r.RId).First(&Result{}).Error
		if err == gorm.ErrRecordNotFound {
			break
		}
	}
	return nil
}

// GetResult returns the Result object from the database
// given the ResultId
func GetResult(rid string) (Result, error) {
	r := Result{}
	err := db.Where("r_id=?", rid).First(&r).Error
	return r, err
}
