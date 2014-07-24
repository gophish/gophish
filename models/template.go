package models

import (
	"errors"
	"time"
)

type Template struct {
	Id           int64        `json:"id"`
	UserId       int64        `json:"-"`
	Name         string       `json:"name"`
	Subject      string       `json:"subject"`
	Text         string       `json:"text"`
	HTML         string       `json:"html"`
	ModifiedDate time.Time    `json:"modified_date"`
	Attachments  []Attachment `json:"attachments"`
}

var ErrTemplateNameNotSpecified = errors.New("Template Name not specified")
var ErrTemplateMissingParameter = errors.New("Need to specify at least plaintext or HTML format")

func (t *Template) Validate() error {
	switch {
	case t.Name == "":
		return ErrTemplateNameNotSpecified
	case t.Text == "" && t.HTML == "":
		return ErrTemplateMissingParameter
	}
	return nil
}

// GetTemplates returns the templates owned by the given user.
func GetTemplates(uid int64) ([]Template, error) {
	ts := []Template{}
	err := db.Where("user_id=?", uid).Find(&ts).Error
	if err != nil {
		Logger.Println(err)
		return ts, err
	}
	for i, _ := range ts {
		err = db.Where("template_id=?", ts[i].Id).Find(&ts[i].Attachments).Error
		if err != nil {
			Logger.Println(err)
			return ts, err
		}
	}
	return ts, err
}

// GetTemplate returns the template, if it exists, specified by the given id and user_id.
func GetTemplate(id int64, uid int64) (Template, error) {
	t := Template{}
	err := db.Where("user_id=? and id=?", uid, id).Find(&t).Error
	if err != nil {
		Logger.Println(err)
		return t, err
	}
	err = db.Where("template_id=?", t.Id).Find(&t.Attachments).Error
	if err != nil {
		Logger.Println(err)
		return t, err
	}
	return t, err
}

// GetTemplateByName returns the template, if it exists, specified by the given name and user_id.
func GetTemplateByName(n string, uid int64) (Template, error) {
	t := Template{}
	err := db.Where("user_id=? and name=?", uid, n).Find(&t).Error
	if err != nil {
		Logger.Println(err)
		return t, err
	}
	return t, nil
}

// PostTemplate creates a new template in the database.
func PostTemplate(t *Template) error {
	// Insert into the DB
	err := db.Save(t).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	for i, _ := range t.Attachments {
		Logger.Println(t.Attachments[i].Name)
		t.Attachments[i].TemplateId = t.Id
		err := db.Save(&t.Attachments[i]).Error
		if err != nil {
			Logger.Println(err)
			return err
		}
	}
	return nil
}

// PutTemplate edits an existing template in the database.
// Per the PUT Method RFC, it presumes all data for a template is provided.
func PutTemplate(t *Template) error {
	// Delete all attachments, and replace with new ones
	err := db.Where("template_id=?", t.Id).Delete(&Attachment{}).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	for i, _ := range t.Attachments {
		t.Attachments[i].TemplateId = t.Id
		err := db.Save(&t.Attachments[i]).Error
		if err != nil {
			Logger.Println(err)
			return err
		}
	}
	err = db.Where("id=?", t.Id).Save(t).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	return nil
}

// DeleteTemplate deletes an existing template in the database.
// An error is returned if a template with the given user id and template id is not found.
func DeleteTemplate(id int64, uid int64) error {
	err := db.Where("template_id=?", id).Delete(&Attachment{}).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	err = db.Where("user_id=?", uid).Delete(Template{Id: id}).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	return nil
}
