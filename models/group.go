package models

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	log "github.com/gophish/gophish/logger"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

// Group contains the fields needed for a user -> group mapping
// Groups contain 1..* Targets
type Group struct {
	Id           int64     `json:"id"`
	UserId       int64     `json:"-"`
	Name         string    `json:"name"`
	ModifiedDate time.Time `json:"modified_date"`
	Targets      []Target  `json:"targets" sql:"-"`
}

// GroupSummaries is a struct representing the overview of Groups.
type GroupSummaries struct {
	Total  int64          `json:"total"`
	Groups []GroupSummary `json:"groups"`
}

// GroupSummary represents a summary of the Group model. The only
// difference is that, instead of listing the Targets (which could be expensive
// for large groups), it lists the target count.
type GroupSummary struct {
	Id           int64     `json:"id"`
	Name         string    `json:"name"`
	ModifiedDate time.Time `json:"modified_date"`
	NumTargets   int64     `json:"num_targets"`
}

// GroupTarget is used for a many-to-many relationship between 1..* Groups and 1..* Targets
type GroupTarget struct {
	GroupId  int64 `json:"-"`
	TargetId int64 `json:"-"`
}

// Target contains the fields needed for individual targets specified by the user
// Groups contain 1..* Targets, but 1 Target may belong to 1..* Groups
type Target struct {
	Id int64 `json:"id"`
	BaseRecipient
}

// BaseRecipient contains the fields for a single recipient. This is the base
// struct used in members of groups and campaign results.
type BaseRecipient struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Position  string `json:"position"`
}

// DataTable is used to return a JSON object suitable for consumption by DataTables
// when using pagination
type DataTable struct {
	Draw            int64         `json:"draw"`
	RecordsTotal    int64         `json:"recordsTotal"`
	RecordsFiltered int64         `json:"recordsFiltered"`
	Data            []interface{} `json:"data"`
}

// FormatAddress returns the email address to use in the "To" header of the email
func (r *BaseRecipient) FormatAddress() string {
	addr := r.Email
	if r.FirstName != "" && r.LastName != "" {
		a := &mail.Address{
			Name:    fmt.Sprintf("%s %s", r.FirstName, r.LastName),
			Address: r.Email,
		}
		addr = a.String()
	}
	return addr
}

// FormatAddress returns the email address to use in the "To" header of the email
func (t *Target) FormatAddress() string {
	addr := t.Email
	if t.FirstName != "" && t.LastName != "" {
		a := &mail.Address{
			Name:    fmt.Sprintf("%s %s", t.FirstName, t.LastName),
			Address: t.Email,
		}
		addr = a.String()
	}
	return addr
}

// ErrEmailNotSpecified is thrown when no email is specified for the Target
var ErrEmailNotSpecified = errors.New("No email address specified")

// ErrGroupNameNotSpecified is thrown when a group name is not specified
var ErrGroupNameNotSpecified = errors.New("Group name not specified")

// ErrNoTargetsSpecified is thrown when no targets are specified by the user
var ErrNoTargetsSpecified = errors.New("No targets specified")

// Validate performs validation on a group given by the user
func (g *Group) Validate() error {
	switch {
	case g.Name == "":
		return ErrGroupNameNotSpecified
	case len(g.Targets) == 0:
		return ErrNoTargetsSpecified
	}
	return nil
}

// GetGroups returns the groups owned by the given user.
func GetGroups(uid int64) ([]Group, error) {
	gs := []Group{}
	err := db.Where("user_id=?", uid).Find(&gs).Error
	if err != nil {
		log.Error(err)
		return gs, err
	}
	for i := range gs {
		gs[i].Targets, err = GetTargets(gs[i].Id)
		if err != nil {
			log.Error(err)
		}
	}
	return gs, nil
}

// GetGroupSummaries returns the summaries for the groups
// created by the given uid.
func GetGroupSummaries(uid int64) (GroupSummaries, error) {
	gs := GroupSummaries{}
	query := db.Table("groups").Where("user_id=?", uid)
	err := query.Select("id, name, modified_date").Scan(&gs.Groups).Error
	if err != nil {
		log.Error(err)
		return gs, err
	}
	for i := range gs.Groups {
		query = db.Table("group_targets").Where("group_id=?", gs.Groups[i].Id)
		err = query.Count(&gs.Groups[i].NumTargets).Error
		if err != nil {
			return gs, err
		}
	}
	gs.Total = int64(len(gs.Groups))
	return gs, nil
}

// GetGroup returns the group, if it exists, specified by the given id and user_id.
func GetGroup(id int64, uid int64) (Group, error) {
	g := Group{}
	err := db.Where("user_id=? and id=?", uid, id).Find(&g).Error
	if err != nil {
		log.Error(err)
		return g, err
	}
	g.Targets, err = GetTargets(g.Id)
	if err != nil {
		log.Error(err)
	}
	return g, nil
}

// GetGroupSummary returns the summary for the requested group
func GetGroupSummary(id int64, uid int64) (GroupSummary, error) {
	g := GroupSummary{}
	query := db.Table("groups").Where("user_id=? and id=?", uid, id)
	err := query.Select("id, name, modified_date").Scan(&g).Error
	if err != nil {
		log.Error(err)
		return g, err
	}
	query = db.Table("group_targets").Where("group_id=?", id)
	err = query.Count(&g.NumTargets).Error
	if err != nil {
		return g, err
	}
	return g, nil
}

// GetGroupByName returns the group, if it exists, specified by the given name and user_id.
func GetGroupByName(n string, uid int64) (Group, error) {
	g := Group{}
	err := db.Where("user_id=? and name=?", uid, n).Find(&g).Error
	if err != nil {
		log.Error(err)
		return g, err
	}
	g.Targets, err = GetTargets(g.Id)
	if err != nil {
		log.Error(err)
	}
	return g, err
}

// PostGroup creates a new group in the database.
func PostGroup(g *Group) error {
	if err := g.Validate(); err != nil {
		return err
	}
	// Insert the group into the DB
	tx := db.Begin()
	err := tx.Save(g).Error
	if err != nil {
		tx.Rollback()
		log.Error(err)
		return err
	}
	for _, t := range g.Targets {
		err = insertTargetIntoGroup(tx, t, g.Id)
		if err != nil {
			tx.Rollback()
			log.Error(err)
			return err
		}
	}
	err = tx.Commit().Error
	if err != nil {
		log.Error(err)
		tx.Rollback()
		return err
	}
	return nil
}

// PutGroup updates the given group if found in the database.
func PutGroup(g *Group) error {
	if err := g.Validate(); err != nil {
		return err
	}
	// Fetch group's existing targets from database.
	ts, err := GetTargets(g.Id)
	if err != nil {
		log.WithFields(logrus.Fields{
			"group_id": g.Id,
		}).Error("Error getting targets from group")
		return err
	}
	// Preload the caches
	cacheNew := make(map[string]int64, len(g.Targets))
	for _, t := range g.Targets {
		cacheNew[t.Email] = t.Id
	}

	cacheExisting := make(map[string]int64, len(ts))
	for _, t := range ts {
		cacheExisting[t.Email] = t.Id
	}

	tx := db.Begin()
	// Check existing targets, removing any that are no longer in the group.
	for _, t := range ts {
		if _, ok := cacheNew[t.Email]; ok {
			continue
		}

		// If the target does not exist in the group any longer, we delete it
		err := tx.Where("group_id=? and target_id=?", g.Id, t.Id).Delete(&GroupTarget{}).Error
		if err != nil {
			tx.Rollback()
			log.WithFields(logrus.Fields{
				"email": t.Email,
			}).Error("Error deleting email")
		}
	}
	// Add any targets that are not in the database yet.
	for _, nt := range g.Targets {
		// If the target already exists in the database, we should just update
		// the record with the latest information.
		if id, ok := cacheExisting[nt.Email]; ok {
			nt.Id = id
			err = UpdateTarget(tx, nt)
			if err != nil {
				log.Error(err)
				tx.Rollback()
				return err
			}
			continue
		}
		// Otherwise, add target if not in database
		err = insertTargetIntoGroup(tx, nt, g.Id)
		if err != nil {
			log.Error(err)
			tx.Rollback()
			return err
		}
	}
	err = tx.Save(g).Error
	if err != nil {
		log.Error(err)
		return err
	}
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// DeleteGroup deletes a given group by group ID and user ID
func DeleteGroup(g *Group) error {
	// Delete all the group_targets entries for this group
	err := db.Where("group_id=?", g.Id).Delete(&GroupTarget{}).Error
	if err != nil {
		log.Error(err)
		return err
	}
	// Delete the group itself
	err = db.Delete(g).Error
	if err != nil {
		log.Error(err)
		return err
	}
	return err
}

// DeleteTarget deletes a single target from a group given by target ID
func DeleteTarget(t *Target, gid int64, uid int64) error {

	targetOwner, err := GetTargetOwner(t.Id)
	if err != nil {
		return err
	}
	if targetOwner != uid {
		return errors.New("No such target id (wrong owner)")
	}

	err = db.Delete(t).Error
	if err != nil {
		return err
	}
	err = db.Where("target_id=?", t.Id).Delete(&GroupTarget{}).Error
	if err != nil {
		return err
	}
	// Update group modification date
	err = db.Model(&Group{}).Where("id=?", gid).Update("ModifiedDate", time.Now().UTC()).Error
	return err
}

// UpdateGroup updates a given group (without updating the targets)
// Note: I thought about putting this in the Group() function, but we'd have to skip the validation and have a boolean
//    	 indicating we just want to rename the group.
func UpdateGroup(g *Group) error {
	if g.Name == "" {
		return ErrGroupNameNotSpecified
	}
	err := db.Save(g).Error
	return err
}

// AddTargetsToGroup adds targets to a group, updating on duplicate email
func AddTargetsToGroup(nts []Target, gid int64) error {

	// Fetch group's existing targets from database.
	ets, err := GetTargets(gid)
	if err != nil {
		return err
	}
	// Load email to target id cache
	existingTargetCache := make(map[string]int64, len(ets))
	for _, t := range ets {
		existingTargetCache[t.Email] = t.Id
	}

	// Step over each new target and see if it exists in the cache map.
	tx := db.Begin()
	for _, nt := range nts {
		if _, ok := existingTargetCache[nt.Email]; ok {
			// Update
			nt.Id = existingTargetCache[nt.Email]
			err = UpdateTarget(tx, nt)
			if err != nil {
				log.Error(err)
				tx.Rollback()
				return err
			}
		} else {
			// Otherwise, add target if not in database
			err = insertTargetIntoGroup(tx, nt, gid)
			if err != nil {
				log.Error(err)
				tx.Rollback()
				return err
			}
		}
	} // for each new target

	err = tx.Model(&Group{}).Where("id=?", gid).Update("ModifiedDate", time.Now().UTC()).Error // put this in the tx too TODO
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
	}
	return err
}

func insertTargetIntoGroup(tx *gorm.DB, t Target, gid int64) error {
	if _, err := mail.ParseAddress(t.Email); err != nil {
		log.WithFields(logrus.Fields{
			"email": t.Email,
		}).Error("Invalid email")
		return err
	}
	err := tx.Where(t).FirstOrCreate(&t).Error
	if err != nil {
		log.WithFields(logrus.Fields{
			"email": t.Email,
		}).Error(err)
		return err
	}
	err = tx.Save(&GroupTarget{GroupId: gid, TargetId: t.Id}).Error
	if err != nil {
		log.Error(err)
		return err
	}
	if err != nil {
		log.WithFields(logrus.Fields{
			"email": t.Email,
		}).Error("Error adding many-many mapping")
		return err
	}
	return nil
}

// UpdateTarget updates the given target information in the database.
func UpdateTarget(tx *gorm.DB, target Target) error {
	targetInfo := map[string]interface{}{
		"first_name": target.FirstName,
		"last_name":  target.LastName,
		"position":   target.Position,
	}
	err := tx.Model(&target).Where("id = ?", target.Id).Updates(targetInfo).Error
	if err != nil {
		log.WithFields(logrus.Fields{
			"email": target.Email,
		}).Error("Error updating target information")
	}
	return err
}

// GetTargets performs a many-to-many select to get all the Targets for a Group
func GetTargets(gid int64) ([]Target, error) {
	ts := []Target{}
	err := db.Table("targets").Select("targets.id, targets.email, targets.first_name, targets.last_name, targets.position").Joins("left join group_targets gt ON targets.id = gt.target_id").Where("gt.group_id=?", gid).Scan(&ts).Error
	return ts, err
}

// GetDataTable performs a many-to-many select to get all the Targets for a Group with supplied filters
// start, length, and search, order can be supplied, or -1, -1, "", "" to ignore
func GetDataTable(gid int64, start int64, length int64, search string, order string) (DataTable, error) {

	dt := DataTable{}
	ts := []Target{}
	order = strings.TrimSpace(order)
	search = strings.TrimSpace(search)
	if order == "" {
		order = "targets.first_name asc"
	} else {
		order = "targets." + order
	}

	// 1. Get the total number of targets in group:
	err := db.Table("group_targets").Where("group_id=?", gid).Count(&dt.RecordsTotal).Error
	if err != nil {
		return dt, err
	}

	// 2. Fetch targets, applying relevant start, length, search, and order paramters.
	// TODO: Rather than having two queries create a partial query and include the search options. Haven't been able to figure out how yet.
	if search != "" {
		var count int64
		search = "%" + search + "%"

		// 2.1 Apply search filter
		err = db.Order(order).Table("targets").Select("targets.id, targets.email, targets.first_name, targets.last_name, targets.position").Joins("left join group_targets gt ON targets.id = gt.target_id").Where("gt.group_id=?", gid).Where("targets.first_name LIKE ? OR targets.last_name LIKE ? OR targets.email LIKE ? or targets.position LIKE ?", search, search, search, search).Count(&count).Offset(start).Limit(length).Scan(&ts).Error

		dt.RecordsFiltered = count // The number of results from applying the search filter (calculated before trimming down the results with offset and limit)

	} else {
		err = db.Order(order).Table("targets").Select("targets.id, targets.email, targets.first_name, targets.last_name, targets.position").Joins("left join group_targets gt ON targets.id = gt.target_id").Where("gt.group_id=?", gid).Offset(start).Limit(length).Scan(&ts).Error
		dt.RecordsFiltered = dt.RecordsTotal
	}

	// 3. Insert targes into datatable struct
	dt.Data = make([]interface{}, len(ts)) // Pseudocode of 'dT.Data = g.Targets'. https://golang.org/doc/faq#convert_slice_of_interface
	for i, v := range ts {
		dt.Data[i] = v
	}

	return dt, err
}

// GetTargetByEmail gets a single target from a group by email address and group id
func GetTargetByEmail(gid int64, email string) ([]Target, error) {
	ts := []Target{}
	err := db.Table("targets").Select("targets.id, targets.email, targets.first_name, targets.last_name, targets.position").Joins("left join group_targets gt ON targets.id = gt.target_id").Where("gt.group_id=?", gid).Where("targets.email=?", email).First(&ts).Error
	return ts, err
}

// GetTargetOwner returns the user id owner of a given target id
func GetTargetOwner(tid int64) (int64, error) {
	g := Group{}
	err := db.Table("groups").Select("groups.user_id").Joins("left join group_targets on group_targets.group_id = groups.id").Where("group_targets.target_id = ?", tid).Scan(&g).Error
	return g.UserId, err
}
