package migrations

import (
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/jinzhu/gorm"
)

// Migration2018092312000000 backfills models.Result objects with the correctly
// parsed URLs for use in USB drop campaign
type Migration2018092312000000 struct{}

func (m Migration2018092312000000) generateURL(campaign *models.Campaign, result *models.Result) error {
	pctx, err := models.NewPhishingTemplateContext(campaign, result.BaseRecipient, result.RId)
	if err != nil {
		return err
	}
	result.URL = pctx.URL
	return nil
}

func (m Migration2018092312000000) updateResults(db *gorm.DB, campaign models.Campaign) error {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return tx.Error
	}
	// Gather all of the results for this campaign
	results, err := tx.Table("results").Where("campaign_id=?", campaign.Id).Rows()
	if err != nil {
		tx.Rollback()
		return err
	}
	for results.Next() {
		var result models.Result
		if err = tx.ScanRows(results, &result); err != nil {
			tx.Rollback()
			return err
		}
		if err = m.generateURL(&campaign, &result); err != nil {
			tx.Rollback()
			return err
		}
		log.Infof("Campaign ID: %d Result: %s URL: %s\n", campaign.Id, result.RId, result.URL)
		if err = tx.Save(&result).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	err = results.Close()
	if err != nil {
		log.Error(err)
	}
	log.Info("committing")
	return tx.Commit().Error
}

// Up backfills previous models.Result objects with the correct parsed URLs
func (m Migration2018092312000000) Up(db *gorm.DB) error {
	campaigns := []models.Campaign{}
	err := db.Table("campaigns").Find(&campaigns).Error
	if err != nil {
		log.Error(err)
		return err
	}
	// For each campaign, iterate over the results and parse the correct URL,
	// storing it back in the database.
	for _, campaign := range campaigns {
		log.Infof("Getting results for %d\n", campaign.Id)
		err = m.updateResults(db, campaign)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

func (m Migration2018092312000000) Down(db *gorm.DB) error {
	return nil
}
