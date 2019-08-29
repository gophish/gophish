package models

import (
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/glennzw/eazye"
	log "github.com/gophish/gophish/logger"
)

// IMAP contains the attributes needed to handle logging into an IMAP server to check
// for reported emails
type IMAP struct {
	UserId       int64     `json:"-" gorm:"column:user_id"`
	Enabled      bool      `json:"enabled"`
	Host         string    `json:"host"`
	Port         uint16    `json:"port,string"`
	Username     string    `json:"username"`
	Password     string    `json:"password"`
	TLS          bool      `json:"tls"`
	ModifiedDate time.Time `json:"modified_date"`
}

// ErrIMAPHostNotSpecified is thrown when there is no Host specified
// in the IMAP configuration
var ErrIMAPHostNotSpecified = errors.New("No IMAP Host specified")

// ErrIMAPPortNotSpecified is thrown when there is no Port specified
// in the IMAP configuration
var ErrIMAPPortNotSpecified = errors.New("No IMAP Port specified")

// ErrInvalidIMAPHost indicates that the IMAP server string is invalid
var ErrInvalidIMAPHost = errors.New("Invalid IMAP server address")

// ErrInvalidIMAPPort indicates that the IMAP Port is invalid
var ErrInvalidIMAPPort = errors.New("Invalid IMAP Port")

// ErrIMAPUsernameNotSpecified is thrown when there is no Username specified
// in the IMAP configuration
var ErrIMAPUsernameNotSpecified = errors.New("No Username specified")

// ErrIMAPPasswordNotSpecified is thrown when there is no Password specified
// in the IMAP configuration
var ErrIMAPPasswordNotSpecified = errors.New("No Password specified")

// TableName specifies the database tablename for Gorm to use
func (s IMAP) TableName() string {
	return "imap"
}

// Validate ensures that IMAP configs/connections are valid
func (s *IMAP) Validate() error {
	switch {
	case s.Host == "":
		return ErrIMAPHostNotSpecified
	case s.Port == 0:
		return ErrIMAPPortNotSpecified
	case s.Username == "":
		return ErrIMAPUsernameNotSpecified
	case s.Password == "":
		return ErrIMAPPasswordNotSpecified
	}

	// Make sure s.Host is an IP or hostname. NB will fail if unable to resolve the hostname.s
	ip := net.ParseIP(s.Host)
	_, err := net.LookupHost(s.Host)
	if ip == nil && err != nil {
		return ErrInvalidIMAPHost
	}

	// Make sure 1 >= port <= 65535
	if s.Port < 1 || s.Port > 65535 {
		return ErrInvalidIMAPPort
	}
	return nil
}

// GetIMAP returns the IMAP server owned by the given user.
func GetIMAP(uid int64) (IMAP, error) {
	ss := IMAP{}
	err := db.Where("user_id=?", uid).Find(&ss).Error
	if err != nil {
		log.Error(err)
		return ss, err
	}
	return ss, nil
}

// PostIMAP creates a new IMAP in the database.
func PostIMAP(s *IMAP, uid int64) error {
	err := s.Validate()
	if err != nil {
		log.Error(err)
		return err
	}
	//Delete old entry. TODO: Save settings and if fails to Save below replace with original
	err = DeleteIMAP(uid)
	if err != nil {
		log.Error(err)
		return err
	}

	// Insert new settings into the DB
	err = db.Save(s).Error
	if err != nil {
		log.Error(err)
	}
	return err
}

// DeleteIMAP deletes the existing IMAP in the database.
func DeleteIMAP(uid int64) error {
	err := db.Where("user_id=?", uid).Delete(&IMAP{}).Error
	if err != nil {
		log.Error(err)
	}
	return err
}

// TestIMAP tests supplied IMAP settings by connecting to the server
func TestIMAP(s *IMAP) error {
	s.Host = s.Host + ":" + strconv.Itoa(int(s.Port)) //Append port
	mailSettings := eazye.MailboxInfo{
		Host:   s.Host,
		TLS:    s.TLS,
		User:   s.Username,
		Pwd:    s.Password,
		Folder: "INBOX"}

	err := eazye.ValidateSettings(mailSettings)
	if err != nil {
		log.Error(err.Error())
	}
	return err
}
