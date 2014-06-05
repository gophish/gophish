package models

type SMTP struct {
	SMTPId      int64  `json:"-"`
	CampaignId  int64  `json:"-"`
	Host        string `json:"host"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty" sql:"-"`
	FromAddress string `json:"from_address"`
}

func (s *SMTP) Validate() (string, bool) {
	switch {
	case s.FromAddress == "":
		return "No from address specified", false
	case s.Host == "":
		return "No hostname specified", false
	}
	return "", true
}
