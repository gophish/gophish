package models

import "time"

type Template struct {
	Id           int64     `json:"id"`
	UserId       int64     `json:"-"`
	Name         string    `json:"name"`
	Text         string    `json:"text"`
	Html         string    `json:"html"`
	ModifiedDate time.Time `json:"modified_date"`
}

type UserTemplate struct {
	UserId     int64 `json:"-"`
	TemplateId int64 `json:"-"`
}

// GetTemplates returns the templates owned by the given user.
func GetTemplates(uid int64) ([]Template, error) {
	ts := []Template{}
	err := db.Where("user_id=?", uid).Find(&ts).Error
	if err != nil {
		Logger.Println(err)
		return ts, err
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
	return nil
}

// PutTemplate edits an existing template in the database.
// Per the PUT Method RFC, it presumes all data for a template is provided.
func PutTemplate(t *Template, uid int64) error {
	return nil
	//err :=
}

// DeleteTemplate deletes an existing template in the database.
// An error is returned if a template with the given user id and template id is not found.
func DeleteTemplate(id int64, uid int64) error {
	err := db.Where("user_id=?", uid).Delete(Template{Id: id}).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	return nil
}
