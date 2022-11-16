package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/contribsys/sparq"
	"github.com/pressly/goose/v3"
)

func die(msg any) {
	log.Printf("%+v", msg)
	os.Exit(0)
}

var (
	DatabaseVersion  int64
	MigrationVersion int64
)

func GetDatabaseVersion(db *sql.DB) int64 {
	if DatabaseVersion > 0 {
		return DatabaseVersion
	}
	dbver, err := goose.GetDBVersion(db)
	if err != nil {
		die(err)
	}
	DatabaseVersion = dbver
	return DatabaseVersion
}

func GetMigrationsVersion(db *sql.DB) int64 {
	if MigrationVersion > 0 {
		return MigrationVersion
	}
	maxInt := int64((1 << 63) - 1)
	migs, err := goose.CollectMigrations("db/migrate", 0, maxInt)
	if err != nil {
		die(err)
	}

	mig, _ := migs.Last()
	if mig != nil {
		MigrationVersion = mig.Version
	}
	return MigrationVersion
}

func migrateExec(args []string) {
	db, err := sql.Open("sqlite", "./sparq.db")
	if err != nil {
		log.Printf("Unable to open database: %v\n", err)
		return
	}

	goose.SetBaseFS(sparq.Migrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Printf("Unable to migrate: %v\n", err)
		return
	}

	// get the current database schema version
	dbver := GetDatabaseVersion(db)
	migver := GetMigrationsVersion(db)

	if dbver > migver {
		die(fmt.Sprintf("Your sparq %s database version %d is newer than this binary %d, are you using the wrong version?", sparq.Version, dbver, migver))
	}

	cmd := "up"
	if len(args) == 1 {
		cmd = args[0]
	}

	if cmd == "redo" {
		if err := goose.Redo(db, "db/migrate"); err != nil {
			log.Printf("Unable to migrate database: %v\n", err)
			return
		}
	} else {
		if dbver == migver {
			die("Your sparq database version is current.")
		}
		if err := goose.Up(db, "db/migrate"); err != nil {
			log.Printf("Unable to migrate database: %v\n", err)
			return
		}
	}
}
