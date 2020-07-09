package models

import (
	"time"

	log "github.com/gophish/gophish/logger"
)

// ReportedEmail contains the attributes for non-campaign emails reported by users
type ReportedEmail struct {
	//Id           int64     `json:"id" gorm:"column:id; primary_key:yes;AUTO_INCREMENT"`
	Id              int64     `json:"id"`
	UserId          int64     `json:"user_id"`           // ID of the user account
	ReportedByName  string    `json:"reported_by_name"`  // Email of the user reporting the email
	ReportedByEmail string    `json:"reported_by_email"` // Email of the user reporting the email
	ReportedTime    time.Time `json:"reported_time"`     // When the email was reported
	ReportedHTML    string    `json:"reported_html"`
	ReportedText    string    `json:"reported_text"`
	ReportedSubject string    `json:"reported_subject"`

	/*EmailFrom    string `json:"email_from"`
	EmailTo      string `json:"email_to"`
	EmailCC      string `json:"email_cc"`
	EmailSubject string `json:"email_subject"`
	EmailTime    string `json:"email_time"`
	EmailText    string `json:"email_text"`
	EmailHTML    string `json:"email_html"`
	EmailHeaders string `json:"email_headers"`
	EmailBlob string `json:"email_blob"`*/

	IMAPUID int64  `json:"imap_uid"`
	Status  string `json:"status"`
	Notes   string `json:"notes"` // Free form notes for operator to give additional info

	Attachments []*ReportedAttachment `json:"attachments" gorm:"foreignkey:Rid"`
}

//Todo: Add Enabled boolean, and attachments option

// ReportedEmailAttachment contains email attachments
type ReportedAttachment struct {
	Rid      int64  `json:"-"` // Foreign key
	Id       int64  `json:"id"`
	Filename string `json:"filename"`
	Header   string `json:"header"`
	Size     int    `json:"size"` // File size in bytes
	Content  string `json:"content,omitempty"`
}

// TableName specifies the database tablename for Gorm to use
func (em ReportedEmail) TableName() string {
	return "reported"
}

// GetReportedEmailAttachment gets an attachment
func GetReportedEmailAttachment(uid, id int64) (ReportedAttachment, error) {

	att := ReportedAttachment{}

	err := db.Debug().Table("reported_attachments").Select("reported_attachments.filename, reported_attachments.header, reported_attachments.content").Joins("left join reported on reported.id = reported_attachments.rid").Where("reported.user_id=? AND reported_attachments.id=?", uid, id).Take(&att).Error

	return att, err
}

// GetReportedEmails gets reported emails
func GetReportedEmails(uid, emailid, limit, offset int64) ([]*ReportedEmail, error) {

	ems := []*ReportedEmail{}
	var err error

	// We have three conditions; fetch all email, fetch one email by id, or fetch a subset of emails by limit and offset
	if emailid == -1 {
		if offset == -1 {
			err = db.Debug().Preload("Attachments").Where("user_id=?", uid).Find(&ems).Error
		} else {
			err = db.Preload("Attachments").Where("user_id=?", uid).Order("ReportedTime", true).Offset(offset).Limit(limit).Find(&ems).Error
		}
	} else {
		err = db.Preload("Attachments").Where("user_id=? AND id=?", uid, emailid).Find(&ems).Error
	}

	if err != nil {
		log.Error(err)
	}

	// Remove attachmetns and HTML/plaintext content for bulk requests. TODO: Don't retrieve these in the first place. Maybe with joins.
	if emailid == -1 {
		for _, e := range ems {
			e.ReportedHTML = ""
			e.ReportedText = ""
			if len(e.Attachments) > 0 { // Remove attachment content, but leave other details (filename, size, header)
				for _, a := range e.Attachments {
					a.Content = ""
				}
			}
		}
	}

	// Reverse order so newest emails are first. Could not figure out how to add ORDER to the Preload() in the queries
	for i, j := 0, len(ems)-1; i < j; i, j = i+1, j-1 {
		ems[i], ems[j] = ems[j], ems[i]
	}
	return ems, err
}

// GetReportedEmail gets a single reported emails
func GetReportedEmail(uid, emailid int64) ([]*ReportedEmail, error) {

	ems := []*ReportedEmail{}
	err := db.Preload("Attachments").Where("user_id=? AND id=?", uid, emailid).Find(&ems).Error
	if err != nil {
		log.Error(err)
	}

	return ems, err
}

// SaveReportedEmail updates IMAP settings for a user in the database.
func SaveReportedEmail(em *ReportedEmail) error {

	// Insert into the DB
	err := db.Save(em).Error
	if err != nil {
		log.Error("Unable to save to database: ", err.Error())
	}
	return err
}

// DeleteReportedEmail deletes
func DeleteReportedEmail(id int64) error {
	err := db.Where("id=?", id).Delete(&ReportedEmail{}).Error
	if err != nil {
		log.Error(err)
	}
	return err
}
