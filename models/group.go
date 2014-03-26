package models

import (
	"net/mail"
	"time"
)

type Group struct {
	Id           int64     `json:"id"`
	Name         string    `json:"name"`
	ModifiedDate time.Time `json:"modified_date"`
	Targets      []Target  `json:"targets" sql:"-"`
}

type UserGroup struct {
	UserId  int64 `json:"-"`
	GroupId int64 `json:"-"`
}

type GroupTarget struct {
	GroupId  int64 `json:"-"`
	TargetId int64 `json:"-"`
}

type Target struct {
	Id    int64  `json:"-"`
	Email string `json:"email"`
}

// GetGroups returns the groups owned by the given user.
func GetGroups(uid int64) ([]Group, error) {
	gs := []Group{}
	err := db.Debug().Table("groups g").Select("g.id, g.name, g.modified_date").Joins("left join user_groups ug ON g.id = ug.group_id").Where("ug.user_id=?", uid).Scan(&gs).Error
	/*_, err := Conn.Select(&gs, "SELECT g.id, g.name, g.modified_date FROM groups g, user_groups ug, users u WHERE ug.uid=u.id AND ug.gid=g.id AND u.id=?", uid)*/
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
	err := db.Debug().Table("groups g").Select("g.id, g.name, g.modified_date").Joins("left join user_groups ug ON g.id = ug.group_id").Where("ug.user_id=? and g.id=?", uid, id).Scan(&g).Error
	/*err := Conn.SelectOne(&g, "SELECT g.id, g.name, g.modified_date FROM groups g, user_groups ug, users u WHERE ug.uid=u.id AND ug.gid=g.id AND g.id=? AND u.id=?", id, uid)*/
	if err != nil {
		Logger.Println(err)
		return g, err
	}
	g.Targets, err = GetTargets(g.Id)
	/*_, err = Conn.Select(&g.Targets, "SELECT t.id, t.email FROM targets t, group_targets gt WHERE gt.gid=? AND gt.tid=t.id", g.Id)*/
	if err != nil {
		Logger.Println(err)
	}
	return g, nil
}

// GetGroupByName returns the group, if it exists, specified by the given name and user_id.
func GetGroupByName(n string, uid int64) (Group, error) {
	g := Group{}
	err := db.Debug().Table("groups g").Select("g.id, g.name, g.modified_date").Joins("left join user_groups ug ON g.id = ug.group_id").Where("ug.user_id=? and g.name=?", uid, n).Scan(&g).Error
	/*err := Conn.SelectOne(&g, "SELECT g.id, g.name, g.modified_date FROM groups g, user_groups ug, users u WHERE ug.uid=u.id AND ug.gid=g.id AND g.name=? AND u.id=?", n, uid)*/
	if err != nil {
		Logger.Println(err)
		return g, err
	}
	g.Targets, err = GetTargets(g.Id)
	/*_, err = Conn.Select(&g.Targets, "SELECT t.id, t.email FROM targets t, group_targets gt WHERE gt.gid=? AND gt.tid=t.id", g.Id)*/
	if err != nil {
		Logger.Println(err)
	}
	return g, nil
}

// PostGroup creates a new group in the database.
func PostGroup(g *Group, uid int64) error {
	// Insert into the DB
	/*err = Conn.Insert(g)*/
	err = db.Save(&g).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Now, let's add the user->user_groups->group mapping
	err = db.Save(&UserGroup{GroupId: g.Id, UserId: uid}).Error
	if err != nil {
		Logger.Println(err)
		return err
	}
	/*_, err = Conn.Exec("INSERT OR IGNORE INTO user_groups VALUES (?,?)", uid, g.Id)
	if err != nil {
		Logger.Printf("Error adding many-many mapping for group %s\n", g.Name)
	}*/
	for _, t := range g.Targets {
		insertTargetIntoGroup(t, g.Id)
	}
	return nil
}

// PutGroup updates the given group if found in the database.
func PutGroup(g *Group, uid int64) error {
	// Update all the foreign keys, and many to many relationships
	// We will only delete the group->targets entries. We keep the actual targets
	// since they are needed by the Results table
	// Get all the targets currently in the database for the group
	ts := []Target{}
	ts, err = GetTargets(g.Id)
	/*_, err = Conn.Select(&ts, "SELECT t.id, t.email FROM targets t, group_targets gt WHERE gt.gid=? AND gt.tid=t.id", g.Id)*/
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
			_, err = Conn.Exec("DELETE FROM group_targets WHERE gid=? AND tid=?", g.Id, t.Id)
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
	// Update the group
	g.ModifiedDate = time.Now()
	err = db.Debug().Update(&g).Error
	/*_, err = Conn.Update(g)*/
	if err != nil {
		Logger.Println(err)
		return err
	}
	return nil
}

func insertTargetIntoGroup(t Target, gid int64) error {
	if _, err = mail.ParseAddress(t.Email); err != nil {
		Logger.Printf("Invalid email %s\n", t.Email)
		return err
	}
	trans, err := Conn.Begin()
	if err != nil {
		Logger.Println(err)
		return err
	}
	_, err = trans.Exec("INSERT OR IGNORE INTO targets VALUES (null, ?)", t.Email)
	if err != nil {
		Logger.Printf("Error adding email: %s\n", t.Email)
		return err
	}
	// Bug: res.LastInsertId() does not work for this, so we need to select it manually (how frustrating.)
	t.Id, err = trans.SelectInt("SELECT id FROM targets WHERE email=?", t.Email)
	if err != nil {
		Logger.Printf("Error getting id for email: %s\n", t.Email)
		return err
	}
	_, err = trans.Exec("INSERT OR IGNORE INTO group_targets VALUES (?,?)", gid, t.Id)
	if err != nil {
		Logger.Printf("Error adding many-many mapping for %s\n", t.Email)
		return err
	}
	err = trans.Commit()
	if err != nil {
		Logger.Printf("Error committing db changes\n")
		return err
	}
	return nil
}

// DeleteGroup deletes a given group by group ID and user ID
func DeleteGroup(id int64) error {
	// Delete all the group_targets entries for this group
	err := db.Debug().Where("group_id=?", id).Delete(&GroupTarget{}).Error
	/*_, err := Conn.Exec("DELETE FROM group_targets WHERE gid=?", id)*/
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Delete the reference to the group in the user_group table
	err = db.Debug().Where("group_id=?", id).Delete(&UserGroup{}).Error
	/*_, err = Conn.Exec("DELETE FROM user_groups WHERE gid=?", id)*/
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Delete the group itself
	err = db.Debug().Delete(&Group{Id: id}).Error
	/*_, err = Conn.Exec("DELETE FROM groups WHERE id=?", id)*/
	if err != nil {
		Logger.Println(err)
		return err
	}
	return err
}

func GetTargets(gid int64) ([]Target, error) {
	ts := []Target{}
	err := db.Debug().Table("targets t").Select("t.id, t.email").Joins("left join group_targets gt ON t.id = gt.target_id").Where("gt.group_id=?", gid).Scan(&ts).Error
	/*_, err := Conn.Select(&gs[i].Targets, "SELECT t.id, t.email FROM targets t, group_targets gt WHERE gt.gid=? AND gt.tid=t.id", gs[i].Id)*/
	if err != nil {
		return nil, err
	}
	return ts, nil
}
