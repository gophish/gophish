package models

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/jinzhu/gorm"
)

type Result struct {
	Id         int64  `json:"-"`
	CampaignId int64  `json:"-"`
	UserId     int64  `json:"-"`
	RId        string `json:"id"`
	Email      string `json:"email"`
	Status     string `json:"status" sql:"not null"`
}

func (r *Result) UpdateStatus(s string) error {
	return db.Table("results").Where("id=?", r.Id).Update("status", s).Error
}

func (r *Result) GenerateId() {
	// Keep trying until we generate a unique key (shouldn't take more than one or two iterations)
	k := make([]byte, 32)
	for {
		io.ReadFull(rand.Reader, k)
		r.RId = fmt.Sprintf("%x", k)
		err := db.Table("results").Where("id=?", r.RId).First(&Result{}).Error
		if err == gorm.RecordNotFound {
			break
		}
	}
}

func GetResult(rid string) (Result, error) {
	r := Result{}
	err := db.Where("r_id=?", rid).First(&r).Error
	return r, err
}
