package models

import (
	"database/sql"
	"fmt"
	"time"
)

//Campaign is a struct representing a created campaign
type Campaign struct {
	Id            int64     `json:"id"`
	Name          string    `json:"name"`
	CreatedDate   time.Time `json:"created_date" db:"created_date"`
	CompletedDate time.Time `json:"completed_date" db:"completed_date"`
	Template      string    `json:"template"` //This may change
	Status        string    `json:"status"`
	Results       []Result  `json:"results,omitempty" db:"-"`
	Groups        []Group   `json:"groups,omitempty" db:"-"`
}

type Result struct {
	Target
	Status string `json:"status"`
}

// GetCampaigns returns the campaigns owned by the given user.
func GetCampaigns(uid int64) ([]Campaign, error) {
	cs := []Campaign{}
	_, err := Conn.Select(&cs, "SELECT c.id, name, created_date, completed_date, status, template FROM campaigns c, user_campaigns uc, users u WHERE uc.uid=u.id AND uc.cid=c.id AND u.id=?", uid)
	for i, _ := range cs {
		_, err = Conn.Select(&cs[i].Results, "SELECT r.email, r.status FROM campaign_results r WHERE r.cid=?", cs[i].Id)
	}
	return cs, err
}

// GetCampaign returns the campaign, if it exists, specified by the given id and user_id.
func GetCampaign(id int64, uid int64) (Campaign, error) {
	c := Campaign{}
	err := Conn.SelectOne(&c, "SELECT c.id, name, created_date, completed_date, status, template FROM campaigns c, user_campaigns uc, users u WHERE uc.uid=u.id AND uc.cid=c.id AND c.id=? AND u.id=?", id, uid)
	if err != nil {
		return c, err
	}
	_, err = Conn.Select(&c.Results, "SELECT r.email, r.status FROM campaign_results r WHERE r.cid=?", c.Id)
	return c, err
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
	err = Conn.Insert(c)
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Insert all the results
	for _, g := range c.Groups {
		// Insert a result for each target in the group
		for _, t := range g.Targets {
			r := Result{Target: t, Status: "Unknown"}
			c.Results = append(c.Results, r)
			fmt.Printf("%v", c.Results)
			_, err = Conn.Exec("INSERT INTO campaign_results VALUES (?,?,?)", c.Id, r.Email, r.Status)
			if err != nil {
				Logger.Printf("Error adding result record for target %s\n", t.Email)
				Logger.Println(err)
			}
		}
	}
	_, err = Conn.Exec("INSERT OR IGNORE INTO user_campaigns VALUES (?,?)", uid, c.Id)
	if err != nil {
		Logger.Printf("Error adding many-many mapping for campaign %s\n", c.Name)
	}
	return nil
}

func DeleteCampaign(id int64) error {
	// Delete all the campaign_results entries for this group
	_, err := Conn.Exec("DELETE FROM campaign_results WHERE cid=?", id)
	if err != nil {
		return err
	}
	// Delete the reference to the campaign in the user_campaigns table
	_, err = Conn.Exec("DELETE FROM user_campaigns WHERE cid=?", id)
	if err != nil {
		return err
	}
	// Delete the campaign itself
	_, err = Conn.Exec("DELETE FROM campaigns WHERE id=?", id)
	return err
}
