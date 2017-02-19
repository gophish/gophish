package models

// CustomHeader contains the fields and methods for
// an email templates to have custom headers
type CustomHeader struct {
	Id         int64  `json:"-"`
	TemplateId int64  `json:"-"`
	Key        string `json:"key"`
	Value      string `json:"value"`
}
