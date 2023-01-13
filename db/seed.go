package db

import (
	"database/sql"
	"fmt"

	"github.com/contribsys/sparq/util"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	newUserInsert = `
		insert into accounts (Id, Sfid, Nick, Email, FullName, RoleMask)
		values (?, ?, ?, ?, ?, ?)`
	newSecurityInsert = `
		insert into account_securities (AccountId, PasswordHash, PublicKey, PrivateKey)
		values (?, ?, ?, ?)`
	newProfileInsert = `
		insert into account_profiles (AccountId) values (?)`
	newTootInsert = `
		insert into toots (SID, URI, AuthorId, ActorId, InReplyTo, InReplyToAccountId, Summary, Content)
		values (?, ?, ?, ?, ?, ?, ?, ?)`
)

func Seed(dbx *sqlx.DB) error {
	if err := createAdmin(dbx); err != nil {
		return err
	}
	if err := createToots(dbx); err != nil {
		return err
	}
	return nil
}

func createToots(dbx *sqlx.DB) error {
	if noRows(dbx, "select * from posts limit 1") {
		uri := fmt.Sprintf("https://%s/@admin", InstanceHostname)

		_, err := dbx.Exec(newTootInsert, "AABA", uri+"/status/AABA", 1, 1, nil, nil, "CW: Hello World",
			"This is a test toot!\nAnother line.")
		if err != nil {
			return errors.Wrap(err, "1")
		}
		_, err = dbx.Exec(newTootInsert, "AABB", uri+"/status/AABB", 1, 1, uri+"/status/AABA", 1,
			"CW: Part 2", "This is a test reply!\nAnother line.")
		if err != nil {
			return errors.Wrap(err, "2")
		}
	}
	return nil
}

func noRows(dbx *sqlx.DB, query string, args ...any) bool {
	result := map[string]interface{}{}
	row := dbx.QueryRowx(query, args...)
	err := row.MapScan(result)
	return err == sql.ErrNoRows
}

func createAdmin(dbx *sqlx.DB) error {
	if noRows(dbx, "select * from accounts where id = ?", 1) {
		_, err := dbx.Exec(newUserInsert, 1, "116672815607840768", "admin", "admin@"+InstanceHostname, "Sparq Admin", -1)
		if err != nil {
			return errors.Wrap(err, "create admin")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte("sparq123"), 12)
		if err != nil {
			return err
		}
		pub, priv := util.GenerateKeys()
		_, err = dbx.Exec(newSecurityInsert, 1, hash, pub, priv)
		if err != nil {
			return errors.Wrap(err, "new account security")
		}
		_, err = dbx.Exec(newProfileInsert, 1)
		if err != nil {
			return errors.Wrap(err, "new profile")
		}
		_, err = dbx.Exec(fmt.Sprintf("insert into account_fields (accountid, name, value, verifiedat) values (1, 'Website', 'https://%s', current_timestamp)", InstanceHostname))
		if err != nil {
			return errors.Wrap(err, "account fields")
		}
		_, err = dbx.Exec("insert into account_fields (accountid, name, value) values (1, 'Wu Tang', 'Forever')")
		if err != nil {
			return errors.Wrap(err, "account fields")
		}
		_, err = dbx.Exec("insert into actors (AccountId) values (?)", 1)
		if err != nil {
			return errors.Wrap(err, "new actor")
		}
	}
	return nil
}
