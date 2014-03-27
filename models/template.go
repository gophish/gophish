package models

import "time"

type Template struct {
	Id           int64     `json:"id"`
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
	err := db.Table("templates t").Select("t.*").Joins("left join user_templates ut ON t.id = ut.template_id").Where("ut.user_id=?", uid).Scan(&ts).Error
	return ts, err
}

// GetTemplate returns the template, if it exists, specified by the given id and user_id.
func GetTemplate(id int64, uid int64) (Template, error) {
	t := Template{}
	err := db.Table("templates t").Select("t.*").Joins("left join user_templates ut ON t.id = ut.template_id").Where("ut.user_id=? and t.id=?", uid, id).Scan(&t).Error
	return t, err
}

// PostTemplate creates a new template in the database.
func PostTemplate(t *Template, uid int64) error {
	// Insert into the DB
	err := db.Save(t).Error
	if err != nil {
		return err
	}
	// Now, let's add the user->user_templates->template mapping
	err = db.Exec("INSERT OR IGNORE INTO user_templates VALUES (?,?)", uid, t.Id).Error
	if err != nil {
		Logger.Printf("Error adding many-many mapping for template %s\n", t.Name)
	}
	return nil
}

func PutTemplate(t *Template, uid int64) error {
	return nil
}
