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
	InstanceHostname string = "localhost.dev"
)

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

func TestDB(filename string) (*sqlx.DB, func(), error) {
	return InitDB(filename, true)
}

// Used to initialize a new database for tests
func InitDB(filename string, remove bool) (*sqlx.DB, func(), error) {
	fname := fmt.Sprintf("./sparq.%s.db", filename)
	dbx, err := OpenDB(DatabaseOptions{
		Filename:         fname,
		SkipVersionCheck: true,
	})
	if err != nil {
		return nil, nil, err
	}

	if err := goose.Up(dbx.DB, "migrate"); err != nil {
		return nil, nil, errors.Wrap(err, "Unable to migrate database")
	}
	if err := Seed(dbx); err != nil {
		return nil, nil, errors.Wrap(err, "Unable to seed database")
	}

	return dbx, func() {
		dbx.Close()
		if remove {
			_ = os.Remove(fname)
		}
	}, nil
}

func OpenDB(opts DatabaseOptions) (*sqlx.DB, error) {
	var err error
	dbx, err := sqlx.Open("sqlite", opts.Filename)
	if err != nil {
		return nil, err
	}
	err = goose.SetDialect("sqlite3")
	if err != nil {
		return nil, err
	}
	dbVer, err := getDatabaseVersion(dbx)
	if err != nil {
		return nil, err
	}
	migVer, err := getMigrationsVersion(dbx)
	if err != nil {
		return nil, err
	}
	if opts.SkipVersionCheck {
		return dbx, nil
	}
	util.Debugf("Database: %d, Migrations: %d", dbVer, migVer)

	if dbVer > migVer {
		return nil, errors.New(fmt.Sprintf("Your sparq database version %d is too new, expecting <= %d. Are you accidentally running an old binary?", dbVer, migVer))
	}

	if dbVer < migVer {
		return nil, errors.New("Please migrate your sparq database, run `sparq migrate`")
	}
	return dbx, nil
}

func SqliteVersion(dbx *sqlx.DB) string {
	var ver string
	_ = dbx.QueryRow("select sqlite_version()").Scan(&ver)
	return ver
}

// func LoadHostname(dbx *sqlx.DB) error {
// 	return dbx.QueryRow("select hostname from instance;").Scan(&InstanceHostname)
// }
