package models

import (
	"encoding/json"
	"errors"

	"github.com/jinzhu/gorm"
)

// Task contains the fields used for a Task model
// Currently, the following tasks are supported:
// - LANDING_PAGE - Point users to a landing page
// - SEND_EMAIL - Send an email to users
//
// Tasks are stored in a list format in the database.
// Each task points to both its previous task and its next task.
type Task struct {
	Id         int64  `json:"id" gorm:"column:id; primary_key:yes"`
	UserId     int64  `json:"-" gorm:"column:user_id"`
	CampaignId int64  `json:"campaign_id" gorm:"column:campaign_id"`
	Type       string `json:"type"`
	PreviousId int64  `json:"previous_id" gorm:"column:previous_id"`
	NextId     int64  `json:"next_id" gorm:"column:next_id"`
	Metadata   string `json:"metadata" gorm:column:metadata"`
}

// ErrTaskTypeNotSpecified occurs when a type is not provided in a task
var ErrTaskTypeNotSpecfied = errors.New("No type specified for task")

// ErrInvalidTaskType occurs when an invalid task type is specified
var ErrInvalidTaskType = errors.New("Invalid task type")

// PageMetadata contains the attributes for the metadata on a LANDING_PAGE
// task
type PageMetadata struct {
	URL    string `json:"url"`
	PageId int64  `json:"page_id"`
	UserId int64  `json:"-"`
}

// ErrUrlNotSpecified occurs when a URL is not provided in a LANDING_PAGE
// task
var ErrUrlNotSpecified = errors.New("No URL specfied")

// ErrPageIdNotSpecified occurs when a page id is not provided in a LANDING_PAGE
// task
var ErrPageIdNotSpecified = errors.New("Page Id not specified")

// Validate validates that there exists a URL and a
// PageId in the metadata
// We also validate that the PageId is valid for the
// given UserId
func (p *PageMetadata) Validate() error {
	switch {
	case p.URL == "":
		return ErrUrlNotSpecified
	case p.PageId == 0:
		return ErrPageIdNotSpecified
	}
	_, err := GetPage(p.PageId, p.UserId)
	if err == gorm.ErrRecordNotFound {
		return ErrPageNotFound
	}
	return err
}

// SMTPMetadata contains the attributes for the metadata of a SEND_EMAIL
// task
type SMTPMetadata struct {
	SMTPId     int64 `json:"smtp_id"`
	TemplateId int64 `json:"template_id"`
	UserId     int64 `json:"-"`
}

// ErrSMTPIdNotSpecified occurs when an SMTP Id is not specified in
// a SEND_EMAIL task
var ErrSMTPIdNotSpecified = errors.New("SMTP Id not specified")

// ErrTemplateIdNotSpecified occurs when a template id is not specified in
// a SEND_EMAIL task
var ErrTemplateIdNotSpecified = errors.New("Template Id not specified")

// Validate validates that there exists an SMTPId and a
// TemplateId in the task metadata
// We also validate that the SMTPId and TemplateId are
// valid for the given UserId
func (s *SMTPMetadata) Validate() error {
	// Check that the values are provided
	switch {
	case s.SMTPId == 0:
		return ErrSMTPIdNotSpecified
	case s.TemplateId == 0:
		return ErrTemplateIdNotSpecified
	}
	// Check that the template and smtp are valid
	_, err := GetTemplate(s.TemplateId, s.UserId)
	if err == gorm.ErrRecordNotFound {
		return ErrTemplateNotFound
	}
	if err != nil {
		return err
	}
	_, err = GetSMTP(s.SMTPId, s.UserId)
	if err == gorm.ErrRecordNotFound {
		return ErrSMTPNotFound
	}
	return err
}

// Validate validates that the required metadata and core information
// is present in a Task
func (t *Task) Validate() error {
	switch {
	case t.Type == "LANDING_PAGE":
		p := PageMetadata{UserId: t.UserId}
		err := json.Unmarshal([]byte(t.Metadata), &p)
		if err != nil {
			return err
		}
		return p.Validate()
	case t.Type == "SEND_EMAIL":
		s := SMTPMetadata{UserId: t.UserId}
		err := json.Unmarshal([]byte(t.Metadata), &s)
		if err != nil {
			return err
		}
		return s.Validate()
	}
	return ErrInvalidTaskType
}

// Next returns the next task in the flow
func (t *Task) Next() (Task, error) {
	n := Task{}
	err := db.Debug().Where("id=?", t.NextId).Find(&n).Error
	if err != nil {
		Logger.Println(err)
	}
	return n, err
}

// Previous returns the previous task in the flow
func (t *Task) Previous() (Task, error) {
	p := Task{}
	err := db.Debug().Where("id=?", t.PreviousId).Find(&p).Error
	if err != nil {
		Logger.Println(err)
	}
	return p, err
}

// GetTasks returns all the tasks in the campaign flow
func GetTasks(uid int64, cid int64) ([]Task, error) {
	ts := []Task{}
	t := Task{}
	// Get the campaign to find the starting task ID
	c := Campaign{}
	err := db.Where("id=? and user_id=?", cid, uid).Find(&c).Error
	if err != nil {
		Logger.Println(err)
		return ts, err
	}
	// Get the first task
	err = db.Debug().Where("id=? and user_id=?", c.TaskId, uid).Find(&t).Error
	if err != nil {
		Logger.Println(err)
		return ts, err
	}
	ts = append(ts, t)
	// Enumerate through all the rest of the tasks, appending them to our list
	for t.NextId != 0 && err != nil {
		t, err = t.Next()
		ts = append(ts, t)
	}
	// Return the results
	return ts, err
}

// GetTask returns the task, if it exists, specified by the given id and user_id.
func GetTask(id int64, uid int64) (Task, error) {
	t := Task{}
	err := db.Where("user_id=? and id=?", uid, id).Find(&t).Error
	if err != nil {
		Logger.Println(err)
	}
	return t, err
}

// PostTask creates a new task and saves it to the database
// Additionally, if there is a previous id the task points to,
// it will update the previous task's "NextId" to point to itself.
func PostTask(t *Task) error {
	err := t.Validate()
	if err != nil {
		Logger.Println(err)
		return err
	}
	err = db.Save(t).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	if t.PreviousId == 0 {
		return nil
	}
	p := Task{}
	err = db.Where("user_id=? and id=?", t.UserId, t.PreviousId).Find(&p).Error
	if err != nil {
		return err
	}
	p.NextId = t.Id
	err = db.Save(&p).Error
	return err
}

// PostTasks is a helper to automatically handle the setting
// of task PreviousId and NextId. It validates tasks before
// saving them to the database.
func PostTasks(ts []*Task) error {
	// Validate all the tasks
	for _, t := range ts {
		if err := t.Validate(); err != nil {
			Logger.Println(err)
			return err
		}
	}
	// Now, we can insert all the tasks
	for i, t := range ts {
		// The first element does not have a PreviousId
		if i > 0 {
			ts[i].PreviousId = ts[i-1].Id
		}
		// Insert the task
		err := PostTask(t)
		if err != nil {
			return err
		}
		// Finally, we have to update the previous task with the
		// NextId that was just automatically set

		if t.PreviousId != 0 {
			err = db.Where("user_id=? and id=?", t.UserId, t.PreviousId).Find(&ts[i-1]).Error
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteTask deletes an existing task in the database.
// An error is returned if a page with the given user id and task id is not found.
func DeleteTask(id int64, uid int64) error {
	err = db.Where("user_id=?", uid).Delete(Task{Id: id}).Error
	if err != nil {
		Logger.Println(err)
	}
	return err
}
