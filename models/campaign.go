package models

import (
	"database/sql"
	"fmt"
	"time"
)

//Campaign is a struct representing a created campaign
type Campaign struct {
	Id            int64     `json:"id"`
	UserId        int64     `json:"-"`
	Name          string    `json:"name" sql:"not null"`
	CreatedDate   time.Time `json:"created_date"`
	CompletedDate time.Time `json:"completed_date"`
	Template      string    `json:"template"` //This may change
	Status        string    `json:"status"`
	Results       []Result  `json:"results,omitempty"`
	Groups        []Group   `json:"groups,omitempty"`
}

type Result struct {
	Id         int64  `json:"-"`
	CampaignId int64  `json:"-"`
	Email      string `json:"email"`
	Status     string `json:"status" sql:"not null"`
}

// GetCampaigns returns the campaigns owned by the given user.
func GetCampaigns(uid int64) ([]Campaign, error) {
	cs := []Campaign{}
	err := db.Model(&User{Id: uid}).Related(&cs).Error
	if err != nil {
		fmt.Println(err)
	}
	 */for i, _ := range cs {
		err := db.Model(&cs[i]).Related(&cs[i].Results).Error
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Printf("%v", cs)
	return cs, err
}

// GetCampaign returns the campaign, if it exists, specified by the given id and user_id.
func GetCampaign(id int64, uid int64) (Campaign, error) {
	c := Campaign{}
	err := db.Where("id = ?", id).Where("user_id = ?", uid).Find(&c).Error
	 */if err != nil {
		return c, err
	}
	err = db.Model(&c).Related(&c.Results).Error
	 */return c, err
}

// PostCampaign inserts a campaign and all associated records into the database.
func PostCampaign(c *Campaign, uid int64) error {
	// Check to make sure all the groups already exist
	for i, g := range c.Groups {
		c.Groups[i], err = GetGroupByName(g.Name, uid)
		if err == sql.ErrNoRows {
			Logger.Printf("Error - Group %s does not exist", g.Name)
			return err
		} else if err != nil {
			Logger.Println(err)
			return err
		}
	}
	// Insert into the DB
	err = db.Save(c).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Insert all the results
	for _, g := range c.Groups {
		// Insert a result for each target in the group
		for _, t := range g.Targets {
			r := Result{Email: t.Email, Status: "Unknown", CampaignId: c.Id}
			c.Results = append(c.Results, r)
			fmt.Printf("%v", c.Results)
			err := db.Save(&r).Error
			if err != nil {
				Logger.Printf("Error adding result record for target %s\n", t.Email)
				Logger.Println(err)
			}
		}
	}
	return nil
}

//DeleteCampaign deletes the specified campaign
func DeleteCampaign(id int64) error {
	// Delete all the campaign results
	err := db.Delete(&Result{CampaignId: id}).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Delete the campaign
	err = db.Delete(&Campaign{Id: id}).Error
	if err != nil {
		Logger.Panicln(err)
		return err
	}
	return err
}
