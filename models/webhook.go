package models

import (
  // "errors"

)

type Webhook struct {
  Id     int64  `json:"id"`
  Title  string `json:"title"`
  Url    string `json:"url"`
  Secret string `json:"secret"`

  //TODO
  //add UserId ?
}

//TODO sort out "user_id"


func GetWebhooks(uid int64) ([]Webhook, error) {
  whs := []Webhook{}
  err := db.Where("user_id=?", uid).Find(&whs).Error
  return whs, err
}

func GetWebhook(id int64) (Webhook, error) {
  wh := Webhook{}
  err := db.Where("id=?", id).First(&wh).Error
  return wh, err
}

func UpdateWebhook(wh *Webhook) error {
  err := db.Save(wh).Error
  return err
}

func DeleteWebhook(id int64) error {
  err := db.Where("id=?", id).Delete(&Webhook{}).Error
  return err
}