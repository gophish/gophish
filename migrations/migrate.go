package migrations

import (
	"bitbucket.org/liamstask/goose/lib/goose"
	"github.com/gophish/gophish/config"
	log "github.com/gophish/gophish/logger"
	"github.com/jinzhu/gorm"
)

// Migration is an interface that defines the needed operations for a custom
// database migration.
type Migration interface {
	Up(db *gorm.DB) error
	Down(db *gorm.DB) error
}

// CustomMigrations are the list of migrations we need to run that include
// custom logic.
// Any migrations in this list need a corresponding SQL migration in the
// db/db_*/migrations/ directories. The corresponding SQL migration may include
// any setup instructions that are then used in these custom migrations.
var CustomMigrations = map[int64]Migration{
	2018092312000000: Migration2018092312000000{},
}

func chooseDBDriver(name, openStr string) goose.DBDriver {
	d := goose.DBDriver{Name: name, OpenStr: openStr}

	switch name {
	case "mysql":
		d.Import = "github.com/go-sql-driver/mysql"
		d.Dialect = &goose.MySqlDialect{}

	// Default database is sqlite3
	default:
		d.Import = "github.com/mattn/go-sqlite3"
		d.Dialect = &goose.Sqlite3Dialect{}
	}
	return d
}

// Migrate executes the database migrations, resulting in an up-to-date
// instance of the database schema.
func Migrate() error {
	// Open a database connection for our migrations
	db, err := gorm.Open(config.Conf.DBName, config.Conf.DBPath)
	db.LogMode(false)
	db.SetLogger(log.Logger)
	db.DB().SetMaxOpenConns(1)
	if err != nil {
		log.Error(err)
		return err
	}
	defer db.Close()
	// Setup the goose configuration
	migrateConf := &goose.DBConf{
		MigrationsDir: config.Conf.MigrationsPath,
		Env:           "production",
		Driver:        chooseDBDriver(config.Conf.DBName, config.Conf.DBPath),
	}
	// Get the latest possible migration
	latest, err := goose.GetMostRecentDBVersion(migrateConf.MigrationsDir)
	if err != nil {
		log.Error(err)
		return err
	}
	currentVersion, err := goose.GetDBVersion(migrateConf)
	if err != nil {
		log.Error(err)
		return err
	}
	// Collect all the outstanding migrations that need to be executed
	ms, err := goose.CollectMigrations(migrateConf.MigrationsDir, currentVersion, latest)
	if err != nil {
		log.Errorf("Error collecting migrations: %s\n", err)
		return err
	}
	for _, m := range ms {
		if migration, ok := CustomMigrations[m.Version]; ok {
			// Run all the migrations up to and including this point
			log.Infof("Found custom migration %d. Running previous migrations\n", m.Version)
			err = goose.RunMigrationsOnDb(migrateConf, migrateConf.MigrationsDir, m.Version, db.DB())
			// After the setup migration runs, we can run our custom logic
			err = migration.Up(db)
			if err != nil {
				log.Errorf("Error applying migration %d: %s\n", m.Version, err)
				return err
			}
		}
	}
	// Finally, do one last pass to ensure that all the migrations up to the
	// latest one are executed
	err = goose.RunMigrationsOnDb(migrateConf, migrateConf.MigrationsDir, latest, db.DB())
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
