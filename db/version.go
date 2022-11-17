package db

import (
	"fmt"

	"github.com/contribsys/sparq/util"
	"github.com/jmoiron/sqlx"
)

var (
	db     *sqlx.DB
	dbVer  int64
	migVer int64
)

func Database() *sqlx.DB {
	return db
}

func BootDB() error {
	var err error
	db, err = sqlx.Open("sqlite", "./sparq.db")
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
	return versionCheck()
}

func CloseDatabase() error {
	return db.Close()
}

func SqliteVersion() string {
	var ver string
	_ = db.QueryRow("select sqlite_version()").Scan(&ver)
	return ver
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
