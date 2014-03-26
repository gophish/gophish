package models

import "time"

type Template struct {
	Id           int64     `json:"id"`
	Name         string    `json:"name"`
	Text         string    `json:"text"`
	Html         string    `json:"html"`
	ModifiedDate time.Time `json:"modified_date"`
}

// GetTemplates returns the templates owned by the given user.
func GetTemplates(uid int64) ([]Template, error) {
	ts := []Template{}
	_, err := Conn.Select(&ts, "SELECT t.id, t.name, t.modified_date, t.text, t.html FROM templates t, user_templates ut, users u WHERE ut.uid=u.id AND ut.tid=t.id AND u.id=?", uid)
	return ts, err
}

// GetTemplate returns the template, if it exists, specified by the given id and user_id.
func GetTemplate(id int64, uid int64) (Template, error) {
	t := Template{}
	err := Conn.SelectOne(&t, "SELECT t.id, t.name, t.modified_date, t.text, t.html FROM templates t, user_templates ut, users u WHERE ut.uid=u.id AND ut.tid=t.id AND t.id=? AND u.id=?", id, uid)
	if err != nil {
		return t, err
	}
	return t, err
}

// PostTemplate creates a new template in the database.
func PostTemplate(t *Template, uid int64) error {
	// Insert into the DB
	err = Conn.Insert(t)
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Now, let's add the user->user_templates->template mapping
	_, err = Conn.Exec("INSERT OR IGNORE INTO user_templates VALUES (?,?)", uid, t.Id)
	if err != nil {
		Logger.Printf("Error adding many-many mapping for template %s\n", t.Name)
	}
	return nil
}

func PutTemplate(t *Template, uid int64) error {
	return nil
}
