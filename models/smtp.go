package models

import "errors"

// SMTP contains the attributes needed to handle the sending of campaign emails
type SMTP struct {
	SMTPId      int64  `json:"-" gorm:"column:smtp_id; primary_key:yes"`
	CampaignId  int64  `json:"-" gorm:"column:campaign_id"`
	Host        string `json:"host"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty" sql:"-"`
	FromAddress string `json:"from_address"`
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
