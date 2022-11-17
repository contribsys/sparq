package db

import (
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed migrate/*.sql
var Migrations embed.FS

func init() {
	goose.SetBaseFS(Migrations)
}

// the latest migration version in the sqlite database on disk
func getDatabaseVersion() (int64, error) {
	return goose.GetDBVersion(db.DB)
}

// the latest migration version packed into this binary
func getMigrationsVersion() (int64, error) {
	maxInt := int64((1 << 63) - 1)
	migs, err := goose.CollectMigrations("migrate", 0, maxInt)
	if err != nil {
		return 0, err
	}

	mig, err := migs.Last()
	if err != nil {
		return 0, err
	}
	if mig != nil {
		return mig.Version, nil
	}
	return 0, nil
}
