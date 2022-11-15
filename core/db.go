package core

import (
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

var (
	db     *sqlx.DB
	dbVer  int64
	migVer int64
)

func init() {
	db = mustOpenDatabase()
	dbVer = getDatabaseVersion()
	migVer = getMigrationsVersion()
}

func DatabaseVersion() int64 {
	return dbVer
}

func MigrationsVersion() int64 {
	return migVer
}

func mustOpenDatabase() *sqlx.DB {
	dbo, err := sqlx.Open("sqlite", "./sparq.db")
	if err != nil {
		panic(err)
	}
	return dbo
}

func CloseDatabase() error {
	return db.Close()
}

func getDatabaseVersion() int64 {
	dbver, err := goose.GetDBVersion(db.DB)
	if err != nil {
		panic(err)
	}
	return dbver
}

func getMigrationsVersion() int64 {
	maxInt := int64((1 << 63) - 1)
	migs, err := goose.CollectMigrations("db/migrate", 0, maxInt)
	if err != nil {
		panic(err)
	}

	mig, _ := migs.Last()
	if mig != nil {
		return mig.Version
	}
	return 0
}
