package models

import (
  log "github.com/gophish/gophish/logger"
)

type Webhook struct {
  Id       int64  `json:"id" gorm:"column:id; primary_key:yes"`
  Title    string `json:"title"`
  Url      string `json:"url"`
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
  err := wh.Validate()
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
  err := db.Save(wh).Error
  return err
}

func DeleteWebhook(id int64) error {
  err := db.Where("id=?", id).Delete(&Webhook{}).Error
  return err
}


//TODO
func (wh *Webhook) Validate() error {
  
  return nil
}
