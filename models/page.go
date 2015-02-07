package models

import (
	"errors"
	"time"
)

// Page contains the fields used for a Page model
type Page struct {
	Id           int64     `json:"id" gorm:"column:id; primary_key:yes"`
	UserId       int64     `json:"-" gorm:"column:user_id"`
	Name         string    `json:"name"`
	HTML         string    `json:"html" gorm:"column:html"`
	ModifiedDate time.Time `json:"modified_date"`
}

// ErrPageNameNotSpecified is thrown if the name of the landing page is blank.
var ErrPageNameNotSpecified = errors.New("Page Name not specified")

// Validate ensures that a page contains the appropriate details
func (p *Page) Validate() error {
	if p.Name == "" {
		return ErrPageNameNotSpecified
	}
	return nil
}

// GetPages returns the pages owned by the given user.
func GetPages(uid int64) ([]Page, error) {
	ps := []Page{}
	err := db.Where("user_id=?", uid).Find(&ps).Error
	if err != nil {
		Logger.Println(err)
		return ps, err
	}
	return ps, err
}

// GetPage returns the page, if it exists, specified by the given id and user_id.
func GetPage(id int64, uid int64) (Page, error) {
	p := Page{}
	err := db.Where("user_id=? and id=?", uid, id).Find(&p).Error
	if err != nil {
		Logger.Println(err)
	}
	return p, err
}

// GetPageByName returns the page, if it exists, specified by the given name and user_id.
func GetPageByName(n string, uid int64) (Page, error) {
	p := Page{}
	err := db.Where("user_id=? and name=?", uid, n).Find(&p).Error
	if err != nil {
		Logger.Println(err)
	}
	return p, err
}

// PostPage creates a new page in the database.
func PostPage(p *Page) error {
	err := p.Validate()
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Insert into the DB
	err = db.Save(p).Error
	if err != nil {
		Logger.Println(err)
	}
	return err
}

// PutPage edits an existing Page in the database.
// Per the PUT Method RFC, it presumes all data for a page is provided.
func PutPage(p *Page) error {
	err := db.Debug().Where("id=?", p.Id).Save(p).Error
	if err != nil {
		Logger.Println(err)
	}
	return err
}

// DeletePage deletes an existing page in the database.
// An error is returned if a page with the given user id and page id is not found.
func DeletePage(id int64, uid int64) error {
	err = db.Where("user_id=?", uid).Delete(Page{Id: id}).Error
	if err != nil {
		Logger.Println(err)
	}
	return err
}
