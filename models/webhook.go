package models

import (
  "errors"

  log "github.com/gophish/gophish/logger"
)

type Webhook struct {
  Id    int64  `json:"id"`
  Title string `json:"url"`
  Url   string `json:"url"`
}

func GetWebhook(id int64, uid int64) (Webhook, error) {
  wh := Webhook{}
  err := db.Where("user_id=? and id=?", uid, id).Find(&wh).Error
  if err != nil {
    log.Error(err)
    return t, err
  }
  return wh, nil
}

// GetUsers returns the users registered in Gophish
func GetWebhooks(uid int64) ([]Webhook, error) {
  wh_s := []Webhook{}
  err := db.Where("user_id=?", uid).Find(&wh_s).Error
  if err != nil {
  log.Error(err)
    return gs, err
  }
  return wh_s, nil
}


