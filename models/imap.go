package models

import (
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/glennzw/eazye"
	log "github.com/gophish/gophish/logger"
)

const DefaultIMAPFolder = "INBOX"

// IMAP contains the attributes needed to handle logging into an IMAP server to check
// for reported emails
type IMAP struct {
	UserId            int64     `json:"-" gorm:"column:user_id"`
	Enabled           bool      `json:"enabled"`
	Host              string    `json:"host"`
	Port              uint16    `json:"port,string,omitempty"`
	Username          string    `json:"username"`
	Password          string    `json:"password"`
	TLS               bool      `json:"tls"`
	Folder            string    `json:"folder"`
	RestrictDomain    string    `json:"restrict_domain"`
	DeleteCampaign    bool      `json:"delete_campaign"`
	LastLogin         time.Time `json:"last_login,omitempty"`
	LastLoginFriendly string    `json:"last_login_friendly,omitonempty"`
	ModifiedDate      time.Time `json:"modified_date"`
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

	// Set the default value for Folder
	if s.Folder == "" {
		s.Folder = DefaultIMAPFolder
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
func GetIMAP(uid int64) ([]IMAP, error) {
	ss := []IMAP{}
	count := 0
	err := db.Where("user_id=?", uid).Find(&ss).Count(&count).Error

	if err != nil {
		log.Error(err)
		return ss, err
	}
	return ss, nil
}

// PostIMAP updates IMAP settings for a user in the database.
func PostIMAP(s *IMAP, uid int64) error {
	err := s.Validate()
	if err != nil {
		log.Error(err)
		return err
	}

	// Delete old entry. TODO: Save settings and if fails to Save below replace with original
	err = DeleteIMAP(uid)
	if err != nil {
		log.Error(err)
		return err
	}

	// Insert new settings into the DB
	err = db.Save(s).Error
	if err != nil {
		log.Error("Bad things happened here ", err.Error())
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

// ValidateIMAP validates supplied IMAP settings by connecting to the server
func ValidateIMAP(s *IMAP) error {

	err := s.Validate()
	if err != nil {
		log.Error(err)
		return err
	}

	s.Host = s.Host + ":" + strconv.Itoa(int(s.Port)) // Append port
	mailSettings := eazye.MailboxInfo{
		Host:   s.Host,
		TLS:    s.TLS,
		User:   s.Username,
		Pwd:    s.Password,
		Folder: s.Folder}

	err = eazye.ValidateMailboxInfo(mailSettings)
	if err != nil {
		log.Error(err.Error())
	}
	return err
}

// GetEnabledIMAPs returns IMAP settings that are currently marked as enabled
// This is used by the imapMonitor service that runs on server startup.
func GetEnabledIMAPs() ([]IMAP, error) {
	ss := []IMAP{}
	err := db.Where("enabled=true").Find(&ss).Error
	if err != nil {
		log.Error(err)
		return ss, err
	}
	return ss, nil
}

func SuccessfulLogin(s *IMAP) error {
	err := db.Model(&s).Update("last_login", time.Now()).Error
	if err != nil {
		log.Error(err)
	}
	return err
}
