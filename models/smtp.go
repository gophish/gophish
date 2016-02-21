package models

import 	(
	"errors"
	"time"

//	"github.com/jinzhu/gorm"
)

// SMTP contains the attributes needed to handle the sending of campaign emails
type SMTP struct {
	Id			int64		`json:"-" gorm:"column:id; primary_key:yes"`
	Name			string		`json:"-" gorm:"column:name"`
	UserId			int64		`json:"-" gorm:"column:user_id"`
	Host			string		`json:"host"`
	Username		string		`json:"username,omitempty"`
	Password		string		`json:"password,omitempty" sql:"-"`
	FromAddress		string		`json:"from_address"`
	IgnoreCertErrors	bool		`json:"ignore_cert_errors"`
	ModifiedDate		time.Time	`json:"modified_date"`
}

// ErrFromAddressNotSpecified is thrown when there is no "From" address
// specified in the SMTP configuration
var ErrFromAddressNotSpecified = errors.New("No From Address specified")

// ErrHostNotSpecified is thrown when there is no Host specified
// in the SMTP configuration
var ErrHostNotSpecified = errors.New("No SMTP Host specified")

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
	}
	return nil
}

// GetSMTPs returns the SMTPs owned by the given user.
func GetSMTPs(uid int64) ([]SMTP, error) {
	ss := []SMTP{}
	err := db.Where("user_id=?", uid).Find(&ss).Error
	if err != nil {
		Logger.Println(err)
		return ss, err
	}
	return ss, err
}

// GetSMTP returns the SMTP, if it exists, specified by the given id and user_id.
func GetSMTP(id int64, uid int64) (SMTP, error) {
	s := SMTP{}
	err := db.Where("user_id=? and id=?", uid, id).Find(&s).Error
	if err != nil {
		Logger.Println(err)
		return s, err
	}
	return s, err
}

// GetSMTPByName returns the SMTP, if it exists, specified by the given name and user_id.
func GetSMTPByName(n string, uid int64) (SMTP, error) {
	s := SMTP{}
	err := db.Where("user_id=? and name=?", uid, n).Find(&s).Error
	if err != nil {
		Logger.Println(err)
	}
	return s, err
}

// PostSMTP creates a new SMTP in the database.
func PostSMTP(s *SMTP) error {
	err := s.Validate()
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Insert into the DB
	err = db.Save(s).Error
	if err != nil {
		Logger.Println(err)
	}
	return err
}

// PutSMTP edits an existing SMTP in the database.
// Per the PUT Method RFC, it presumes all data for a SMTP is provided.
func PutSMTP(s *SMTP) error {
	err := db.Where("id=?", s.Id).Save(s).Error
	if err != nil {
		Logger.Println(err)
	}
	return err
}

// DeleteSMTP deletes an existing SMTP in the database.
// An error is returned if a SMTP with the given user id and SMTP id is not found.
func DeleteSMTP(id int64, uid int64) error {
	err = db.Where("user_id=?", uid).Delete(SMTP{Id: id}).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	return err
}
