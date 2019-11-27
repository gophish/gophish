package models

import (
	"errors"

	log "github.com/gophish/gophish/logger"
)

type Webhook struct {
	Id       int64  `json:"id" gorm:"column:id; primary_key:yes"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Secret   string `json:"secret"`
	IsActive bool   `json:"is_active"`
}

func GetWebhooks() ([]Webhook, error) {
	whs := []Webhook{}
	err := db.Find(&whs).Error
	return whs, err
}

func GetActiveWebhooks() ([]Webhook, error) {
	whs := []Webhook{}
	err := db.Where("is_active=?", true).Find(&whs).Error
	return whs, err
}

func GetWebhook(id int64) (Webhook, error) {
	wh := Webhook{}
	err := db.Where("id=?", id).First(&wh).Error
	return wh, err
}

func PostWebhook(wh *Webhook) error {
	err := validate(wh)
	if err != nil {
		log.Error(err)
		return err
	}
	err = db.Save(wh).Error
	if err != nil {
		log.Error(err)
	}
	return err
}

func PutWebhook(wh *Webhook) error {
	err := validate(wh)
	if err != nil {
		log.Error(err)
		return err
	}
	err = db.Save(wh).Error
	return err
}

func DeleteWebhook(id int64) error {
	err := db.Where("id=?", id).Delete(&Webhook{}).Error
	return err
}

func validate(wh *Webhook) error {
	if wh.URL == "" {
		return errors.New("url can't be empty")
	}
	if wh.Title == "" {
		return errors.New("title can't be empty")
	}
	return nil
}
