package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/db"
	"github.com/pressly/goose/v3"
)

func die(msg any) {
	log.Printf("%+v", msg)
	os.Exit(0)
}

func getDatabaseVersion(db *sql.DB) (int64, error) {
	return goose.GetDBVersion(db)
}

func getMigrationsVersion(db *sql.DB) (int64, error) {
	maxInt := int64((1 << 63) - 1)
	migs, err := goose.CollectMigrations("migrate", 0, maxInt)
	if err != nil {
		return 0, err
	}

	mig, err := migs.Last()
	if err != nil {
		return 0, err
	}
	return mig.Version, nil
}

func migrateExec(args []string) {
	dbx, err := sql.Open("sqlite", "./sparq.db")
	if err != nil {
		log.Printf("Unable to open database: %v\n", err)
		return
	}

	goose.SetBaseFS(db.Migrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Printf("Unable to migrate: %v\n", err)
		return
	}

	// get the current database schema version
	var dbver int64
	if dbver, err = getDatabaseVersion(dbx); err != nil {
		log.Printf("Unable to determine version: %v\n", err)
		return
	}
	var migver int64
	if migver, err = getMigrationsVersion(dbx); err != nil {
		log.Printf("Unable to determine version: %v\n", err)
		return
	}

	if dbver > migver {
		die(fmt.Sprintf("Your sparq %s database version %d is newer than this binary %d, are you using the wrong version?", sparq.Version, dbver, migver))
	}

	cmd := "up"
	if len(args) == 1 {
		cmd = args[0]
	}

	if cmd == "redo" {
		if err := goose.Redo(dbx, "migrate"); err != nil {
			log.Printf("Unable to migrate database: %v\n", err)
			return
		}
	} else {
		if dbver == migver {
			die("Your sparq database version is current.")
		}
		if err := goose.Up(dbx, "migrate"); err != nil {
			log.Printf("Unable to migrate database: %v\n", err)
			return
		}
	}
}
