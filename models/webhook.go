package models

import (
  // "errors"

  log "github.com/gophish/gophish/logger"
)

type Webhook struct {
  Id     int64  `json:"id"`
  Title  string `json:"title"`
  Url    string `json:"url"`
  Secret string `json:"secret"`
}

func GetWebhook(id int64, uid int64) (Webhook, error) {
  wh := Webhook{}
  err := db.Where("user_id=? and id=?", uid, id).Find(&wh).Error
  if err != nil {
    log.Error(err)
  }
  return wh, err
}

func GetWebhooks(uid int64) ([]Webhook, error) {
  whs := []Webhook{}
  err := db.Where("user_id=?", uid).Find(&whs).Error
  if err != nil {
    log.Error(err)
  }
  return whs, nil
}

func UpdateWebhook(wh Webhook) error {
  return nil
}