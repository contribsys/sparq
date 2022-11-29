package db

import (
	"embed"
	"fmt"
	"log"

	"github.com/contribsys/sparq"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
)

//go:embed migrate/*.sql
var Migrations embed.FS

func init() {
	goose.SetBaseFS(Migrations)
	_ = goose.SetDialect("sqlite3")
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

func MigrateExec(args []string) error {
	err := OpenDB(DatabaseOptions{
		Filename:         "./sparq.db",
		SkipVersionCheck: true,
	})
	if err != nil {
		return errors.Wrap(err, "Unable to open database")
	}

	if dbVer > migVer {
		return errors.New(fmt.Sprintf("Your sparq %s database version %d is newer than this binary %d, are you using the wrong version?", sparq.Version, dbVer, migVer))
	}

	cmd := "up"
	if len(args) == 1 {
		cmd = args[0]
	}

	if cmd == "redo" {
		if err := goose.Redo(db.DB, "migrate"); err != nil {
			return errors.Wrap(err, "Unable to migrate database")
		}
		if err := Seed(); err != nil {
			return errors.Wrap(err, "Unable to seed database")
		}
	} else {
		if dbVer == migVer {
			log.Printf("Your sparq database version is current: %d\n", dbVer)
			return nil
		}
		if err := goose.Up(db.DB, "migrate"); err != nil {
			return errors.Wrap(err, "Unable to migrate database")
		}
	}
	return nil
}
