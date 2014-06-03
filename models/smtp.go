package models

type SMTP struct {
	SMTPId      int64  `json:"-"`
	CampaignId  int64  `json:"-"`
	Hostname    string `json:"hostname"`
	Port        int    `json:"port"`
	UseAuth     bool   `json:"use_auth"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty" sql:"-"`
	FromAddress string `json:"from_address"`
}

func (s *SMTP) Validate() (string, bool) {
	switch {
	case s.UseAuth == false && (s.Username == "" && s.Password == ""):
		return "Auth requested, but username or password blank", false
	case s.FromAddress == "":
		return "No from address specified", false
	case s.Hostname == "":
		return "No hostname specified", false
	case s.Port == 0:
		return "No port specified", false
	}
	return "", true
}
