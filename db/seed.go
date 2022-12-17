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

		_, _ = db.Exec(newPostInsert,
			uri+"/status/116672815607840768",
			uri, nil, "CW: Hello World",
			"This is a test toot!\nAnother line.")
		_, err := db.Exec(newPostInsert,
			uri+"/status/116672815607840769",
			uri,
			uri+"/status/116672815607840768",
			"CW: Part 2",
			"This is a test reply!\nAnother line.")
		if err != nil {
			return err
		}
	}
	return nil
}

func noRows(query string, args ...any) bool {
	result := map[string]interface{}{}
	row := db.QueryRowx(query, args...)
	err := row.MapScan(result)
	return err == sql.ErrNoRows
}

func createAdmin() error {
	if noRows("select * from accounts where id = ?", 1) {
		fmt.Println("Creating admin user")
		_, err := db.Exec(newUserInsert, 1, "116672815607840768", "admin", "admin@"+InstanceHostname, "Sparq Admin", -1)
		if err != nil {
			return errors.Wrap(err, "create admin")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte("sparq123"), 12)
		if err != nil {
			return err
		}
		pub, priv := util.GenerateKeys()
		_, err = db.Exec(newSecurityInsert, 1, hash, pub, priv)
		if err != nil {
			return errors.Wrap(err, "new account security")
		}
		_, err = db.Exec("insert into actors (Id,AccountId) values (?, ?)",
			"https://"+InstanceHostname+"/@admin", 1)
		if err != nil {
			return errors.Wrap(err, "new actor")
		}
	}
	return nil
}
