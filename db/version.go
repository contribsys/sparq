package db

import (
	"fmt"
	"os"

	"github.com/contribsys/sparq/util"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

var (
	dbx    *sqlx.DB
	dbVer  int64
	migVer int64

	InstanceHostname string = "localhost.dev"
)

func Database() *sqlx.DB {
	return dbx
}

type DatabaseOptions struct {
	Filename         string
	SkipVersionCheck bool
}

var (
	Defaults = DatabaseOptions{
		Filename:         "./sparq.db",
		SkipVersionCheck: false,
	}
)

func TestDB(filename string) (func(), error) {
	return InitDB(filename, true)
}

// Used to initialize a new database for tests
func InitDB(filename string, remove bool) (func(), error) {
	fname := fmt.Sprintf("./sparq.%s.db", filename)
	err := OpenDB(DatabaseOptions{
		Filename:         fname,
		SkipVersionCheck: true,
	})
	if err != nil {
		return nil, err
	}

	if err := goose.Up(dbx.DB, "migrate"); err != nil {
		return nil, errors.Wrap(err, "Unable to migrate database")
	}
	if err := Seed(); err != nil {
		return nil, errors.Wrap(err, "Unable to seed database")
	}

	return func() {
		dbx.Close()
		dbx = nil
		if remove {
			_ = os.Remove(fname)
		}
	}, nil
}

func OpenDB(opts DatabaseOptions) error {
	var err error
	dbx, err = sqlx.Open("sqlite", opts.Filename)
	if err != nil {
		return err
	}
	err = goose.SetDialect("sqlite3")
	if err != nil {
		return err
	}
	dbVer, err = getDatabaseVersion()
	if err != nil {
		return err
	}
	migVer, err = getMigrationsVersion()
	if err != nil {
		return err
	}
	if opts.SkipVersionCheck {
		return nil
	}
	return versionCheck()
}

func CloseDatabase() error {
	return dbx.Close()
}

func SqliteVersion() string {
	var ver string
	_ = dbx.QueryRow("select sqlite_version()").Scan(&ver)
	return ver
}

func LoadHostname() error {
	return dbx.QueryRow("select hostname from instance;").Scan(&InstanceHostname)
}

func versionCheck() error {
	util.Debugf("Database: %d, Migrations: %d", dbVer, migVer)

	if dbVer > migVer {
		return fmt.Errorf("Your sparq database version %d is too new, expecting <= %d. Are you accidentally running an old binary?", dbVer, migVer)
	}

	if dbVer < migVer {
		return fmt.Errorf("Please migrate your sparq database, run `sparq migrate`")
	}
	return nil
}
