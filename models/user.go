package models

// User represents the user model for gophish.
type User struct {
	Id       int64  `json:"id"`
	Username string `json:"username" sql:"not null;unique"`
	Email    string `json:"email" sql:"not null;unique"`
	Hash     string `json:"-"`
	ApiKey   string `json:"api_key" sql:"not null;unique"`
}

// Roles represents the role model for gophish.
type Roles struct {
	Rid    int64  `json:"rid"`
	Name   string `json:"name" sql:"not null;unique"`
	Weight string `json:"weight" sql:"not null;unique"`
}

// Roles represents the role model for gophish.
type UsersRole struct {
	Uid int64 `json:"uid"`
	Rid int64 `json:"rid" sql:"not"`
}

// TableName specifies the database tablename for Gorm to use
func (s UsersRole) TableName() string {
	return "users_role"
}

// GetUser returns the user that the given id corresponds to. If no user is found, an
// error is thrown.
func GetUser(id int64) (User, error) {
	u := User{}
	err := db.Where("id=?", id).First(&u).Error
	return u, err
}

// GetUserByAPIKey returns the user that the given API Key corresponds to. If no user is found, an
// error is thrown.
func GetUserByAPIKey(key string) (User, error) {
	u := User{}
	err := db.Where("api_key = ?", key).First(&u).Error
	return u, err
}

// GetUserByUsername returns the user that the given username corresponds to. If no user is found, an
// error is thrown.
func GetUserByUsername(username string) (User, error) {
	u := User{}
	err := db.Where("username = ?", username).First(&u).Error
	return u, err
}

// PutUser updates the given user
func PutUser(u *User) error {
	err := db.Save(u).Error
	return err
}

// PutRoles updates the given user
func PutRoles(r *Roles) error {
	err := db.Save(r).Error
	return err
}

// PutRoles updates the given user
func PutUserRoles(ur *UsersRole) error {
	err := db.Save(ur).Error
	return err
}

// GetUsers returns the campaigns owned by the given user.
func GetUsers(uid int64) ([]User, error) {
	u := []User{}
	err := db.Order("id asc").Find(&u).Error
	return u, err
}

// GetRoles returns the roles set in the site.
func GetRoles(uid int64) ([]Roles, error) {
	r := []Roles{}
	err := db.Order("rid asc").Find(&r).Error
	return r, err
}

// GetRoles returns the roles set in the site.
func GetUserRoles(uid int64) ([]UsersRole, error) {
	r := []UsersRole{}
	err := db.Where("uid = ?", uid).First(&r).Error
	return r, err
}

// GetRoles returns the roles set in the site.
func DeleteUserRoles(uid int64) error {
	err = db.Delete(UsersRole{}, "uid = ?", uid).Error
	return err
}

//Delete  the specified user
func DeleteUser(id int64) error {
	// Delete the campaign
	err = db.Delete(&User{Id: id}).Error
	return err
}
