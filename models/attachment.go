package models

type Attachment struct {
	Id         int64  `json:"-"`
	TemplateId int64  `json:"-"`
	Content    string `json:"content"`
	Type       string `json:"type"`
	Name       string `json:"name"`
}
