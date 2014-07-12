package models

type Attachment struct {
	TemplateId string `json:"-"`
	Content    string `json:"content"`
	Type       string `json:"type"`
	Name       string `json:"name"`
}
