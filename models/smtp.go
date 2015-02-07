package models

// SMTP contains the attributes needed to handle the sending of campaign emails
type SMTP struct {
	SMTPId      int64  `json:"-" gorm:"column:smtp_id; primary_key:yes"`
	CampaignId  int64  `json:"-" gorm:"column:campaign_id"`
	Host        string `json:"host"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty" sql:"-"`
	FromAddress string `json:"from_address"`
}

// TableName specifies the database tablename for Gorm to use
func (s SMTP) TableName() string {
	return "smtp"
}

// Validate ensures that SMTP configs/connections are valid
func (s *SMTP) Validate() (string, bool) {
	switch {
	case s.FromAddress == "":
		return "No from address specified", false
	case s.Host == "":
		return "No hostname specified", false
	}
	return "", true
}
