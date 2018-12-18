package models

// Roles
const (
	Administrator = 1
	Partner       = 2
	Customer      = 3
	ChildUser     = 4
	LMSUser       = 5
)

// User represents the user model for gophish.
type User struct {
	Id       int64  `json:"id"`
	Username string `json:"username" sql:"not null;unique"`
	Email    string `json:"email" sql:"not null;unique"`
	Partner  int64  `json:"partner" sql:"not null"`
	Hash     string `json:"-"`
	ApiKey   string `json:"api_key" sql:"not null;unique"`
}

// Role represents the role model for gophish.
type Role struct {
	Rid    int64  `json:"rid"`
	Name   string `json:"name" sql:"not null;unique"`
	Weight string `json:"weight" sql:"not null;unique"`
}

// UserRole represents the user role model for gophish.
type UserRole struct {
	Uid int64 `json:"uid"`
	Rid int64 `json:"rid" sql:"not"`
}

// Roles is a list of roles
type Roles []Role

// UserRoles is a list of user roles
type UserRoles []UserRole

// AvailableFor returns roles which a user with the given role can create users with
func (roles Roles) AvailableFor(role UserRole) Roles {
	if role.Is(Administrator) {
		return roles
	}

	if role.Is(Partner) {
		return Roles{
			Role{Rid: 3, Name: "Customer", Weight: "2"},
			Role{Rid: 4, Name: "Child User", Weight: "3"},
		}
	}

	if role.Is(ChildUser) {
		return Roles{
			Role{Rid: 3, Name: "Customer", Weight: "2"},
		}
	}

	return Roles{}
}

// TableName specifies the database tablename for Gorm to use
func (ur UserRole) TableName() string {
	return "users_role"
}

// Name returns this role's name
func (ur UserRole) Name() string {
	role := Role{}
	err := db.Where("rid = ?", ur.Rid).First(&role).Error

	if err != nil {
		return "Unknown"
	}

	return role.Name
}

// Is tells if this role id matches the given one
func (r Role) Is(rid int64) bool {
	return r.Rid == rid
}

// Is tells if this user role id matches the given one
func (ur UserRole) Is(rid int64) bool {
	return ur.Rid == rid
}

// IsOneOf tells if this user role id is among the given role ids
func (ur UserRole) IsOneOf(rids []int64) bool {
	for _, rid := range rids {
		if ur.Rid == rid {
			return true
		}
	}

	return false
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

// PutRole updates role
func PutRole(r *Role) error {
	err := db.Save(r).Error
	return err
}

// PutUserRole updates role of the given user
func PutUserRole(ur *UserRole) error {
	err := db.Save(ur).Error
	return err
}

// GetUsers returns the users owned by the given user.
func GetUsers(uid int64) ([]User, error) {
	users := []User{}
	role, err := GetUserRole(uid)

	if err != nil {
		return users, err
	}

	if role.Is(Administrator) {
		err = db.Order("id asc").Find(&users).Error
	} else if role.Is(Partner) {
		err = db.Where("partner = ?", uid).Order("id asc").Find(&users).Error
	} else if role.Is(ChildUser) {
		user, err := GetUser(uid)

		if err != nil {
			return users, err
		}

		err = db.Where("partner = ? and id <> ?", user.Partner, uid).Order("id asc").Find(&users).Error
	}

	return users, err
}

// GetRoles returns all available roles
func GetRoles() (Roles, error) {
	r := Roles{}
	err := db.Order("rid asc").Find(&r).Error
	return r, err
}

// GetUserRole returns a role assigned to the given uid
func GetUserRole(uid int64) (UserRole, error) {
	r := UserRole{}
	err := db.Where("uid = ?", uid).First(&r).Error
	return r, err
}

// DeleteUserRoles deletes all roles of a given uid
func DeleteUserRoles(uid int64) error {
	err = db.Delete(UserRole{}, "uid = ?", uid).Error
	return err
}

// DeleteUser deletes the specified user
func DeleteUser(uid int64) error {
	if err := db.Delete(&User{Id: uid}).Error; err != nil {
		return err
	}

	return DeleteUserRoles(uid)
}

// GetUserPartners returns all the partners from the database
func GetUserPartners() ([]User, error) {
	u := []User{}
	err = db.Raw("SELECT * FROM users u LEFT JOIN users_role ur ON (u.id = ur.uid) where ur.rid in (?, ?)", 1, 2).Scan(&u).Error
	return u, err
}
