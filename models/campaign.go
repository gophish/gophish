package models

import (
	"errors"
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
	TaskId        int64     `json:"-"`
	Tasks         []Task    `json:"tasks"`
	Status        string    `json:"status"`
	Results       []Result  `json:"results,omitempty"`
	Groups        []Group   `json:"groups,omitempty"`
	Events        []Event   `json:"timeline,omitemtpy"`
}

// ErrCampaignNameNotSpecified indicates there was no template given by the user
var ErrCampaignNameNotSpecified = errors.New("Campaign name not specified")

// ErrGroupNotSpecified indicates there was no template given by the user
var ErrGroupNotSpecified = errors.New("No groups specified")

// ErrTemplateNotSpecified indicates there was no template given by the user
var ErrTemplateNotSpecified = errors.New("No email template specified")

// ErrPageNotSpecified indicates a landing page was not provided for the campaign
var ErrPageNotSpecified = errors.New("No landing page specified")

// ErrSMTPNotSpecified indicates a sending profile was not provided for the campaign
var ErrSMTPNotSpecified = errors.New("No sending profile specified")

// ErrTasksNotSpecified indicates there were no tasks given by the user
var ErrTasksNotSpecified = errors.New("No tasks specified")

// ErrInvalidStartTask indicates the starting task was not sending an email
var ErrInvalidStartTask = errors.New("All campaigns must start by sending an email")

// Validate checks to make sure there are no invalid fields in a submitted campaign
func (c *Campaign) Validate() error {
	switch {
	case c.Name == "":
		return ErrCampaignNameNotSpecified
	case len(c.Groups) == 0:
		return ErrGroupNotSpecified
	case len(c.Tasks) == 0:
		return ErrTasksNotSpecified
	case c.Tasks[0].Type != "SEND_EMAIL":
		return ErrInvalidStartTask
	}
	return nil
}

// SendTestEmailRequest is the structure of a request
// to send a test email to test an SMTP connection
type SendTestEmailRequest struct {
	Template    Template `json:"template"`
	Page        Page     `json:"page"`
	SMTP        SMTP     `json:"smtp"`
	URL         string   `json:"url"`
	Tracker     string   `json:"tracker"`
	TrackingURL string   `json:"tracking_url"`
	From        string   `json:"from"`
	Target
}

// Validate ensures the SendTestEmailRequest structure
// is valid.
func (s *SendTestEmailRequest) Validate() error {
	switch {
	case s.Email == "":
		return ErrEmailNotSpecified
	}
	return nil
}

// UpdateStatus changes the campaign status appropriately
func (c *Campaign) UpdateStatus(s string) error {
	// This could be made simpler, but I think there's a bug in gorm
	return db.Table("campaigns").Where("id=?", c.Id).Update("status", s).Error
}

// AddEvent creates a new campaign event in the database
func (c *Campaign) AddEvent(e Event) error {
	e.CampaignId = c.Id
	e.Time = time.Now()
	return db.Save(&e).Error
}

// getDetails retrieves the related attributes of the campaign
// from the database. If the Events and the Results are not available,
// an error is returned. Otherwise, the attribute name is set to [Deleted],
// indicating the user deleted the attribute (template, smtp, etc.)
func (c *Campaign) getDetails() error {
	err = db.Model(c).Related(&c.Results).Error
	if err != nil {
		Logger.Printf("%s: results not found for campaign\n", err)
		return err
	}
	err = db.Model(c).Related(&c.Events).Error
	if err != nil {
		Logger.Printf("%s: events not found for campaign\n", err)
		return err
	}
	c.Tasks, err = GetTasks(c.UserId, c.Id)
	if err != nil {
		Logger.Println(err)
		return err
	}
	return nil
}

// Event contains the fields for an event
// that occurs during the campaign
type Event struct {
	Id         int64     `json:"-"`
	CampaignId int64     `json:"-"`
	Email      string    `json:"email"`
	Time       time.Time `json:"time"`
	Message    string    `json:"message"`
	Details    string    `json:"details"`
}

// GetCampaigns returns the campaigns owned by the given user.
func GetCampaigns(uid int64) ([]Campaign, error) {
	cs := []Campaign{}
	err := db.Model(&User{Id: uid}).Related(&cs).Error
	if err != nil {
		fmt.Println(err)
	}
	for i, _ := range cs {
		err = cs[i].getDetails()
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
		Logger.Printf("%s: campaign not found\n", err)
		return c, err
	}
	err = c.getDetails()
	return c, err
}

// PostCampaign inserts a campaign and all associated records into the database.
func PostCampaign(c *Campaign, uid int64) error {
	if err := c.Validate(); err != nil {
		return err
	}
	// Fill in the details
	c.UserId = uid
	c.CreatedDate = time.Now()
	c.CompletedDate = time.Time{}
	c.Status = CAMPAIGN_QUEUED
	// Check to make sure all the groups already exist
	for i, g := range c.Groups {
		c.Groups[i], err = GetGroupByName(g.Name, uid)
		if err == gorm.ErrRecordNotFound {
			Logger.Printf("Error - Group %s does not exist", g.Name)
			return ErrGroupNotFound
		} else if err != nil {
			Logger.Println(err)
			return err
		}
	}
	for _, t := range c.Tasks {
		err = PostTask(&t)
		if err != nil {
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
	err = c.AddEvent(Event{Message: "Campaign Created"})
	if err != nil {
		Logger.Println(err)
	}
	// Insert all the results
	for _, g := range c.Groups {
		// Insert a result for each target in the group
		for _, t := range g.Targets {
			r := &Result{Email: t.Email, Position: t.Position, Status: STATUS_SENDING, CampaignId: c.Id, UserId: c.UserId, FirstName: t.FirstName, LastName: t.LastName}
			r.GenerateId()
			err = db.Save(r).Error
			if err != nil {
				Logger.Printf("Error adding result record for target %s\n", t.Email)
				Logger.Println(err)
			}
			c.Results = append(c.Results, *r)
		}
	}
	return nil
}

//DeleteCampaign deletes the specified campaign
func DeleteCampaign(id int64) error {
	Logger.Printf("Deleting campaign %d\n", id)
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
	// TODO Delete all the flows associated with the campaign
	// Delete the campaign
	err = db.Delete(&Campaign{Id: id}).Error
	if err != nil {
		Logger.Panicln(err)
		return err
	}
	return err
}
