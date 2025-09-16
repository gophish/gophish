// models/result_extras.go
package models

import (
	"errors"

	"github.com/jinzhu/gorm"
)

// GetResultByRID looks up a Result by its RID (tracking id).
func GetResultByRID(rid string) (*Result, error) {
	var r Result
	if err := db.Where("rid = ?", rid).First(&r).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.New("result not found")
		}
		return nil, err
	}
	return &r, nil
}

