package db

import (
	"database/sql"
	"fmt"

	"github.com/contribsys/sparq/util"
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
		insert into account_profiles (accountid) values (?)`
	newPostInsert = `
		insert into posts (URI, AuthorId, InReplyTo, Summary, Content)
		values (?, ?, ?, ?, ?)`
)

func Seed() error {
	if err := createAdmin(); err != nil {
		return err
	}
	if err := createPosts(); err != nil {
		return err
	}
	return nil
}

func createPosts() error {
	if noRows("select * from posts limit 1") {
		fmt.Println("Creating fake posts")
		uri := fmt.Sprintf("https://%s/@admin", InstanceHostname)

		_, _ = dbx.Exec(newPostInsert, uri+"/status/116672815607840768", uri, nil, "CW: Hello World",
			"This is a test toot!\nAnother line.")
		_, err := dbx.Exec(newPostInsert, uri+"/status/116672815607840769", uri,
			uri+"/status/116672815607840768", "CW: Part 2", "This is a test reply!\nAnother line.")
		if err != nil {
			return err
		}
	}
	return nil
}

func noRows(query string, args ...any) bool {
	result := map[string]interface{}{}
	row := dbx.QueryRowx(query, args...)
	err := row.MapScan(result)
	return err == sql.ErrNoRows
}

func createAdmin() error {
	if noRows("select * from accounts where id = ?", 1) {
		fmt.Println("Creating admin user")
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
		_, err = dbx.Exec("insert into actors (Id,AccountId) values (?, ?)", "https://"+InstanceHostname+"/@admin", 1)
		if err != nil {
			return errors.Wrap(err, "new actor")
		}
	}
	return nil
}
