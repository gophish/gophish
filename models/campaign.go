package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

//Campaign is a struct representing a created campaign
type Campaign struct {
	Id            int64     `json:"id"`
	UserId        int64     `json:"-"`
	Name          string    `json:"name" sql:"not null"`
	CreatedDate   time.Time `json:"created_date"`
	CompletedDate time.Time `json:"completed_date"`
	TemplateId    int64     `json:"-"`
	Template      Template  `json:"template"` //This may change
	Status        string    `json:"status"`
	EmailsSent    string    `json:"emails_sent"`
	Results       []Result  `json:"results,omitempty"`
	Groups        []Group   `json:"groups,omitempty"`
	Events        []Event   `json:"timeline,omitemtpy"`
	SMTP          SMTP      `json:"smtp"`
}

func (c *Campaign) Validate() (string, bool) {
	switch {
	case c.Name == "":
		return "Must specify campaign name", false
	case len(c.Groups) == 0:
		return "No groups specified", false
	case c.Template.Name == "":
		return "No template specified", false
	}
	return "", true
}

func (c *Campaign) UpdateStatus(s string) error {
	// This could be made simpler, but I think there's a bug in gorm
	return db.Table("campaigns").Where("id=?", c.Id).Update("status", s).Error
}

func (c *Campaign) AddEvent(e Event) error {
	e.CampaignId = c.Id
	e.Time = time.Now()
	return db.Debug().Save(&e).Error
}

type Event struct {
	Id         int64     `json:"-"`
	CampaignId int64     `json:"-"`
	Email      string    `json:"email"`
	Time       time.Time `json:"time"`
	Message    string    `json:"message"`
}

// GetCampaigns returns the campaigns owned by the given user.
func GetCampaigns(uid int64) ([]Campaign, error) {
	cs := []Campaign{}
	err := db.Model(&User{Id: uid}).Related(&cs).Error
	if err != nil {
		fmt.Println(err)
	}
	for i, _ := range cs {
		err := db.Model(&cs[i]).Related(&cs[i].Results).Error
		if err != nil {
			Logger.Println(err)
		}
		err = db.Model(&cs[i]).Related(&cs[i].Events).Error
		if err != nil {
			Logger.Println(err)
		}
		err = db.Table("templates").Where("id=?", cs[i].TemplateId).Find(&cs[i].Template).Error
		if err != nil {
			Logger.Println(err)
		}
	}
	return cs, err
}

// GetCampaign returns the campaign, if it exists, specified by the given id and user_id.
func GetCampaign(id int64, uid int64) (Campaign, error) {
	c := Campaign{}
	err := db.Where("id = ?", id).Where("user_id = ?", uid).Find(&c).Error
	if err != nil {
		return c, err
	}
	err = db.Model(&c).Related(&c.Results).Error
	if err != nil {
		return c, err
	}
	err = db.Model(&c).Related(&c.Events).Error
	if err != nil {
		return c, err
	}
	err = db.Table("templates").Where("id=?", c.TemplateId).Find(&c.Template).Error
	return c, err
}

// PostCampaign inserts a campaign and all associated records into the database.
func PostCampaign(c *Campaign, uid int64) error {
	// Fill in the details
	c.UserId = uid
	c.CreatedDate = time.Now()
	c.CompletedDate = time.Time{}
	c.Status = CAMPAIGN_QUEUED
	// Check to make sure all the groups already exist
	for i, g := range c.Groups {
		c.Groups[i], err = GetGroupByName(g.Name, uid)
		if err == gorm.RecordNotFound {
			Logger.Printf("Error - Group %s does not exist", g.Name)
			return err
		} else if err != nil {
			Logger.Println(err)
			return err
		}
	}
	// Check to make sure the template exists
	t, err := GetTemplateByName(c.Template.Name, uid)
	if err == gorm.RecordNotFound {
		Logger.Printf("Error - Template %s does not exist", t.Name)
		return err
	} else if err != nil {
		Logger.Println(err)
		return err
	}
	c.Template = t
	c.TemplateId = t.Id
	// Insert into the DB
	err = db.Save(c).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	err = c.AddEvent(Event{Message: "Campaign Created"})
	if err != nil {
		Logger.Println(err)
	}
	// Insert all the results
	for _, g := range c.Groups {
		// Insert a result for each target in the group
		for _, t := range g.Targets {
			r := Result{Email: t.Email, Status: STATUS_UNKNOWN, CampaignId: c.Id, UserId: c.UserId}
			r.GenerateId()
			err = db.Save(&r).Error
			if err != nil {
				Logger.Printf("Error adding result record for target %s\n", t.Email)
				Logger.Println(err)
			}
			c.Results = append(c.Results, r)
		}
	}
	return nil
}

//DeleteCampaign deletes the specified campaign
func DeleteCampaign(id int64) error {
	// Delete all the campaign results
	err := db.Where("campaign_id=?", id).Delete(&Result{}).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	err = db.Where("campaign_id=?", id).Delete(&Event{}).Error
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
