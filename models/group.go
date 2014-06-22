package models

import (
	"net/mail"
	"time"

	"github.com/jinzhu/gorm"
)

type Group struct {
	Id           int64     `json:"id"`
	UserId       int64     `json:"-"`
	Name         string    `json:"name"`
	ModifiedDate time.Time `json:"modified_date"`
	Targets      []Target  `json:"targets" sql:"-"`
}

type GroupTarget struct {
	GroupId  int64 `json:"-"`
	TargetId int64 `json:"-"`
}

type Target struct {
	Id        int64  `json:"-"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// GetGroups returns the groups owned by the given user.
func GetGroups(uid int64) ([]Group, error) {
	gs := []Group{}
	err := db.Where("user_id=?", uid).Find(&gs).Error
	if err != nil {
		Logger.Println(err)
		return gs, err
	}
	for i, _ := range gs {
		gs[i].Targets, err = GetTargets(gs[i].Id)
		if err != nil {
			Logger.Println(err)
		}
	}
	return gs, nil
}

// GetGroup returns the group, if it exists, specified by the given id and user_id.
func GetGroup(id int64, uid int64) (Group, error) {
	g := Group{}
	err := db.Where("user_id=? and id=?", uid, id).Find(&g).Error
	if err != nil {
		Logger.Println(err)
		return g, err
	}
	g.Targets, err = GetTargets(g.Id)
	if err != nil {
		Logger.Println(err)
	}
	return g, nil
}

// GetGroupByName returns the group, if it exists, specified by the given name and user_id.
func GetGroupByName(n string, uid int64) (Group, error) {
	g := Group{}
	err := db.Where("user_id=? and name=?", uid, n).Find(&g).Error
	if err != nil {
		Logger.Println(err)
		return g, err
	}
	g.Targets, err = GetTargets(g.Id)
	if err != nil {
		Logger.Println(err)
	}
	return g, nil
}

// PostGroup creates a new group in the database.
func PostGroup(g *Group) error {
	// Insert into the DB
	err = db.Save(g).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	for _, t := range g.Targets {
		insertTargetIntoGroup(t, g.Id)
	}
	return nil
}

// PutGroup updates the given group if found in the database.
func PutGroup(g *Group) error {
	ts := []Target{}
	ts, err = GetTargets(g.Id)
	if err != nil {
		Logger.Printf("Error getting targets from group ID: %d", g.Id)
		return err
	}
	// Enumerate through, removing any entries that are no longer in the group
	// For every target in the database
	tExists := false
	for _, t := range ts {
		tExists = false
		// Is the target still in the group?
		for _, nt := range g.Targets {
			if t.Email == nt.Email {
				tExists = true
				break
			}
		}
		// If the target does not exist in the group any longer, we delete it
		if !tExists {
			err = db.Where("group_id=? and target_id=?", g.Id, t.Id).Delete(&GroupTarget{}).Error
			if err != nil {
				Logger.Printf("Error deleting email %s\n", t.Email)
			}
		}
	}
	// Insert any entries that are not in the database
	// For every target in the new group
	for _, nt := range g.Targets {
		// Check and see if the target already exists in the db
		tExists = false
		for _, t := range ts {
			if t.Email == nt.Email {
				tExists = true
				break
			}
		}
		// If the target is not in the db, we add it
		if !tExists {
			insertTargetIntoGroup(nt, g.Id)
		}
	}
	err = db.Save(g).Error
	/*_, err = Conn.Update(g)*/
	if err != nil {
		Logger.Println(err)
		return err
	}
	return nil
}

// DeleteGroup deletes a given group by group ID and user ID
func DeleteGroup(g *Group) error {
	// Delete all the group_targets entries for this group
	err := db.Where("group_id=?", g.Id).Delete(&GroupTarget{}).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Delete the group itself
	err = db.Delete(g).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	return err
}

func insertTargetIntoGroup(t Target, gid int64) error {
	if _, err = mail.ParseAddress(t.Email); err != nil {
		Logger.Printf("Invalid email %s\n", t.Email)
		return err
	}
	trans := db.Begin()
	trans.Where(t).FirstOrCreate(&t)
	if err != nil {
		Logger.Printf("Error adding target: %s\n", t.Email)
		return err
	}
	err = trans.Where("group_id=? and target_id=?", gid, t.Id).Find(&GroupTarget{}).Error
	if err == gorm.RecordNotFound {
		err = trans.Save(&GroupTarget{GroupId: gid, TargetId: t.Id}).Error
		if err != nil {
			Logger.Println(err)
			return err
		}
	}
	if err != nil {
		Logger.Printf("Error adding many-many mapping for %s\n", t.Email)
		return err
	}
	err = trans.Commit().Error
	if err != nil {
		Logger.Printf("Error committing db changes\n")
		return err
	}
	return nil
}

func GetTargets(gid int64) ([]Target, error) {
	ts := []Target{}
	err := db.Table("targets").Select("targets.id, targets.email").Joins("left join group_targets gt ON targets.id = gt.target_id").Where("gt.group_id=?", gid).Scan(&ts).Error
	return ts, err
}
