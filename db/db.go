package db

import (
	"database/sql"
	"errors"
	"log"
	"net/mail"
	"os"
	"time"

	"github.com/coopernurse/gorp"
	"github.com/jordan-wright/gophish/config"
	"github.com/jordan-wright/gophish/models"
	_ "github.com/mattn/go-sqlite3"
)

var Conn *gorp.DbMap
var DB *sql.DB
var err error
var ErrUsernameTaken = errors.New("Username already taken")
var Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

// Setup initializes the Conn object
// It also populates the Gophish Config object
func Setup() error {
	DB, err := sql.Open("sqlite3", config.Conf.DBPath)
	Conn = &gorp.DbMap{Db: DB, Dialect: gorp.SqliteDialect{}}
	//If the file already exists, delete it and recreate it
	_, err = os.Stat(config.Conf.DBPath)
	Conn.AddTableWithName(models.User{}, "users").SetKeys(true, "Id")
	Conn.AddTableWithName(models.Campaign{}, "campaigns").SetKeys(true, "Id")
	Conn.AddTableWithName(models.Group{}, "groups").SetKeys(true, "Id")
	if err != nil {
		Logger.Println("Database not found, recreating...")
		createTablesSQL := []string{
			//Create tables
			`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT NOT NULL, hash VARCHAR(60) NOT NULL, api_key VARCHAR(32), UNIQUE(username), UNIQUE(api_key));`,
			`CREATE TABLE campaigns (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, created_date TIMESTAMP NOT NULL, completed_date TIMESTAMP, template TEXT, status TEXT NOT NULL, uid INTEGER, FOREIGN KEY (uid) REFERENCES users(id));`,
			`CREATE TABLE targets (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT NOT NULL, UNIQUE(email));`,
			`CREATE TABLE groups (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, modified_date TIMESTAMP NOT NULL, UNIQUE(name));`,
			`CREATE TABLE user_groups (uid INTEGER NOT NULL, gid INTEGER NOT NULL, FOREIGN KEY (uid) REFERENCES users(id), FOREIGN KEY (gid) REFERENCES groups(id), UNIQUE(uid, gid))`,
			`CREATE TABLE group_targets (gid INTEGER NOT NULL, tid INTEGER NOT NULL, FOREIGN KEY (gid) REFERENCES groups(id), FOREIGN KEY (tid) REFERENCES targets(id), UNIQUE(gid, tid));`,
		}
		Logger.Printf("Creating db at %s\n", config.Conf.DBPath)
		//Create the tables needed
		for _, stmt := range createTablesSQL {
			_, err = DB.Exec(stmt)
			if err != nil {
				return err
			}
		}
		//Create the default user
		init_user := models.User{
			Username: "admin",
			Hash:     "$2a$10$IYkPp0.QsM81lYYPrQx6W.U6oQGw7wMpozrKhKAHUBVL4mkm/EvAS",
			APIKey:   "12345678901234567890123456789012",
		}
		Conn.Insert(&init_user)
		if err != nil {
			Logger.Println(err)
		}
		c := models.Campaign{
			Name:          "Test Campaigns",
			CreatedDate:   time.Now().UTC(),
			CompletedDate: time.Now().UTC(),
			Template:      "test template",
			Status:        "In progress",
			Uid:           init_user.Id,
		}
		Conn.Insert(&c)
	}
	return nil
}

// API Functions (GET, POST, PUT, DELETE)

// GetUser returns the user that the given id corresponds to. If no user is found, an
// error is thrown.
func GetUser(id int64) (models.User, error) {
	u := models.User{}
	err := Conn.SelectOne(&u, "SELECT * FROM Users WHERE id=?", id)
	if err != nil {
		return u, err
	}
	return u, nil
}

// GetUserByAPIKey returns the user that the given API Key corresponds to. If no user is found, an
// error is thrown.
func GetUserByAPIKey(key []byte) (models.User, error) {
	u := models.User{}
	err := Conn.SelectOne(&u, "SELECT id, username, api_key FROM Users WHERE apikey=?", key)
	if err != nil {
		return u, err
	}
	return u, nil
}

// GetUserByAPIKey returns the user that the given API Key corresponds to. If no user is found, an
// error is thrown.
func GetUserByUsername(username string) (models.User, error) {
	u := models.User{}
	err := Conn.SelectOne(&u, "SELECT * FROM Users WHERE username=?", username)
	if err != sql.ErrNoRows {
		return u, ErrUsernameTaken
	} else if err != nil {
		return u, err
	}
	return u, nil
}

func PutUser(u *models.User) error {
	_, err := Conn.Update(u)
	return err
}

func GetCampaigns(key interface{}) ([]models.Campaign, error) {
	cs := []models.Campaign{}
	_, err := Conn.Select(&cs, "SELECT c.id, name, created_date, completed_date, status, template FROM campaigns c, users u WHERE c.uid=u.id AND u.api_key=?", key)
	return cs, err
}

func GetCampaign(id int64, key interface{}) (models.Campaign, error) {
	c := models.Campaign{}
	err := Conn.SelectOne(&c, "SELECT campaigns.id, name, created_date, completed_date, status, template FROM campaigns, users WHERE campaigns.uid=users.id AND campaigns.id =? AND users.api_key=?", id, key)
	return c, err
}

func PutCampaign(c *models.Campaign) error {
	_, err := Conn.Update(c)
	return err
}

func GetGroups(key interface{}) ([]models.Group, error) {
	gs := []models.Group{}
	_, err := Conn.Select(&gs, "SELECT g.id, g.name, g.modified_date FROM groups g, user_groups ug, users u WHERE ug.uid=u.id AND ug.gid=g.id AND u.api_key=?", key)
	if err != nil {
		Logger.Println(err)
		return gs, err
	}
	for i, _ := range gs {
		_, err := Conn.Select(&gs[i].Targets, "SELECT t.id, t.email FROM targets t, group_targets gt WHERE gt.gid=? AND gt.tid=t.id", gs[i].Id)
		if err != nil {
			Logger.Println(err)
		}
	}
	return gs, nil
}

func PostGroup(g *models.Group, uid int64) error {
	// Insert into the DB
	err = Conn.Insert(g)
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Let's start a transaction to handle the bulk inserting
	trans, err := Conn.Begin()
	if err != nil {
		Logger.Println(err)
		return err
	}
	// Now, let's add the user->user_groups->group mapping
	_, err = Conn.Exec("INSERT OR IGNORE INTO user_groups VALUES (?,?)", uid, g.Id)
	if err != nil {
		Logger.Printf("Error adding many-many mapping for group %s\n", g.Name)
	}
	// TODO
	for _, t := range g.Targets {
		if _, err = mail.ParseAddress(t.Email); err != nil {
			Logger.Printf("Invalid email %s\n", t.Email)
			continue
		}
		_, err := Conn.Exec("INSERT OR IGNORE INTO targets VALUES (null, ?)", t.Email)
		if err != nil {
			Logger.Printf("Error adding email: %s\n", t.Email)
		}
		// Bug: res.LastInsertId() does not work for this, so we need to select it manually (how frustrating.)
		t.Id, err = Conn.SelectInt("SELECT id FROM targets WHERE email=?", t.Email)
		if err != nil {
			Logger.Printf("Error getting id for email: %s\n", t.Email)
		}
		_, err = Conn.Exec("INSERT OR IGNORE INTO group_targets VALUES (?,?)", g.Id, t.Id)
		if err != nil {
			Logger.Printf("Error adding many-many mapping for %s\n", t.Email)
		}
	}
	err = trans.Commit()
	if err != nil {
		Logger.Println(err)
		return err
	}
	return nil
}

func DeleteGroup(id int) error {
	return nil
}
