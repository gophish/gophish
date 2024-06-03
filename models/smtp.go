package models

import (
	"crypto/tls"
	"errors"
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gophish/gomail"
	"github.com/gophish/gophish/dialer"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/mailer"
	"github.com/jinzhu/gorm"
)

// Dialer is a wrapper around a standard gomail.Dialer in order
// to implement the mailer.Dialer interface. This allows us to better
// separate the mailer package as opposed to forcing a connection
// between mailer and gomail.
type Dialer struct {
	*gomail.Dialer
}

// Dial wraps the gomail dialer's Dial command
func (d *Dialer) Dial() (mailer.Sender, error) {
	return d.Dialer.Dial()
}

// SMTP contains the attributes needed to handle the sending of campaign emails
type SMTP struct {
	Id               int64     `json:"id" gorm:"column:id; primary_key:yes"`
	UserId           int64     `json:"-" gorm:"column:user_id"`
	Interface        string    `json:"interface_type" gorm:"column:interface_type"`
	Name             string    `json:"name"`
	Host             string    `json:"host"`
	Username         string    `json:"username,omitempty"`
	Password         string    `json:"password,omitempty"`
	FromAddress      string    `json:"from_address"`
	IgnoreCertErrors bool      `json:"ignore_cert_errors"`
	Headers          []Header  `json:"headers"`
	ModifiedDate     time.Time `json:"modified_date"`
}

// Header contains the fields and methods for a sending profile to have
// custom headers
type Header struct {
	Id     int64  `json:"-"`
	SMTPId int64  `json:"-"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

// ErrFromAddressNotSpecified is thrown when there is no "From" address
// specified in the SMTP configuration
var ErrFromAddressNotSpecified = errors.New("No From Address specified")

// ErrInvalidFromAddress is thrown when the SMTP From field in the sending
// profiles containes a value that is not an email address
var ErrInvalidFromAddress = errors.New("Invalid SMTP From address because it is not an email address")

// ErrHostNotSpecified is thrown when there is no Host specified
// in the SMTP configuration
var ErrHostNotSpecified = errors.New("No SMTP Host specified")

// ErrInvalidHost indicates that the SMTP server string is invalid
var ErrInvalidHost = errors.New("Invalid SMTP server address")

// TableName specifies the database tablename for Gorm to use
func (s SMTP) TableName() string {
	return "smtp"
}

// Validate ensures that SMTP configs/connections are valid
func (s *SMTP) Validate() error {
	switch {
	case s.FromAddress == "":
		return ErrFromAddressNotSpecified
	case s.Host == "":
		return ErrHostNotSpecified
	case !validateFromAddress(s.FromAddress):
		return ErrInvalidFromAddress
	}
	_, err := mail.ParseAddress(s.FromAddress)
	if err != nil {
		return err
	}
	// Make sure addr is in host:port format
	hp := strings.Split(s.Host, ":")
	if len(hp) > 2 {
		return ErrInvalidHost
	} else if len(hp) < 2 {
		hp = append(hp, "25")
	}
	_, err = strconv.Atoi(hp[1])
	if err != nil {
		return ErrInvalidHost
	}
	return err
}

// validateFromAddress validates
func validateFromAddress(email string) bool {
	r, _ := regexp.Compile("^([a-zA-Z0-9_\\-\\.]+)@([a-zA-Z0-9_\\-\\.]+)\\.([a-zA-Z]{2,18})$")
	return r.MatchString(email)
}

// GetDialer returns a dialer for the given SMTP profile
func (s *SMTP) GetDialer() (mailer.Dialer, error) {
	// Setup the message and dial
	hp := strings.Split(s.Host, ":")
	if len(hp) < 2 {
		hp = append(hp, "25")
	}
	host := hp[0]
	// Any issues should have been caught in validation, but we'll
	// double check here.
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		log.Error(err)
		return nil, err
	}
	dialer := dialer.Dialer()
	d := gomail.NewWithDialer(dialer, host, port, s.Username, s.Password)
	d.TLSConfig = &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: s.IgnoreCertErrors,
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Error(err)
		hostname = "localhost"
	}
	d.LocalName = hostname
	return &Dialer{d}, err
}

// GetSMTPs returns the SMTPs owned by the given user.
func GetSMTPs(uid int64) ([]SMTP, error) {
	ss := []SMTP{}
	err := db.Where("user_id=?", uid).Find(&ss).Error
	if err != nil {
		log.Error(err)
		return ss, err
	}
	for i := range ss {
		err = db.Where("smtp_id=?", ss[i].Id).Find(&ss[i].Headers).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error(err)
			return ss, err
		}
	}
	return ss, nil
}

// GetSMTP returns the SMTP, if it exists, specified by the given id and user_id.
func GetSMTP(id int64, uid int64) (SMTP, error) {
	s := SMTP{}
	err := db.Where("user_id=? and id=?", uid, id).Find(&s).Error
	if err != nil {
		log.Error(err)
		return s, err
	}
	err = db.Where("smtp_id=?", s.Id).Find(&s.Headers).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error(err)
		return s, err
	}
	return s, err
}

// GetSMTPByName returns the SMTP, if it exists, specified by the given name and user_id.
func GetSMTPByName(n string, uid int64) (SMTP, error) {
	s := SMTP{}
	err := db.Where("user_id=? and name=?", uid, n).Find(&s).Error
	if err != nil {
		log.Error(err)
		return s, err
	}
	err = db.Where("smtp_id=?", s.Id).Find(&s.Headers).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error(err)
	}
	return s, err
}

// PostSMTP creates a new SMTP in the database.
func PostSMTP(s *SMTP) error {
	err := s.Validate()
	if err != nil {
		log.Error(err)
		return err
	}
	// Insert into the DB
	err = db.Save(s).Error
	if err != nil {
		log.Error(err)
	}
	// Save custom headers
	for i := range s.Headers {
		s.Headers[i].SMTPId = s.Id
		err := db.Save(&s.Headers[i]).Error
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return err
}

// PutSMTP edits an existing SMTP in the database.
// Per the PUT Method RFC, it presumes all data for a SMTP is provided.
func PutSMTP(s *SMTP) error {
	err := s.Validate()
	if err != nil {
		log.Error(err)
		return err
	}
	err = db.Where("id=?", s.Id).Save(s).Error
	if err != nil {
		log.Error(err)
	}
	// Delete all custom headers, and replace with new ones
	err = db.Where("smtp_id=?", s.Id).Delete(&Header{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error(err)
		return err
	}
	// Save custom headers
	for i := range s.Headers {
		s.Headers[i].SMTPId = s.Id
		err := db.Save(&s.Headers[i]).Error
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return err
}

// DeleteSMTP deletes an existing SMTP in the database.
// An error is returned if a SMTP with the given user id and SMTP id is not found.
func DeleteSMTP(id int64, uid int64) error {
	// Delete all custom headers
	err := db.Where("smtp_id=?", id).Delete(&Header{}).Error
	if err != nil {
		log.Error(err)
		return err
	}
	err = db.Where("user_id=?", uid).Delete(SMTP{Id: id}).Error
	if err != nil {
		log.Error(err)
	}
	return err
}
