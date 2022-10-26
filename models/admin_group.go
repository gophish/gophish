package models

import (
	log "github.com/gophish/gophish/logger"
)

type AdminGroup struct {
    Id      int64       `json:"id,string"`
    Name    string      `json:"name"`
    Users   []User      `json:"users" gorm:"association_autoupdate:false;association_autocreate:false;many2many:users_admin_groups;"`
}

func GetAdminGroups() ([]AdminGroup, error) {
    admin_groups := []AdminGroup{}
    err := db.Preload("Users").Find(&admin_groups).Error
    if err != nil {
        log.Error(err)
        return admin_groups, err
    }

    return admin_groups, nil
}

func PutAdminGroup(ag *AdminGroup) error {
    db.Model(&ag).Association("Users").Replace(ag.Users)
    err := db.Save(ag).Error
    return err
}

func GetAdminGroup(id int64) (AdminGroup, error) {
    admin_group := AdminGroup{}
    err := db.Preload("Users").First(&admin_group, "id = ?", id).Error
    return admin_group, err
}

func GetAdminGroupsUsersIsPartOf(uid int64) ([]AdminGroup, error) {
    admin_groups := []AdminGroup{}

    err := db.Preload("Users").
        Where("id IN (SELECT user_id FROM users_admin_groups WHERE user_id = ?)", uid).
        Find(&admin_groups).Error
    if err != nil {
        log.Error(err)
        return admin_groups, err
    }

    return admin_groups, nil
}
