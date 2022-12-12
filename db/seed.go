package db

import (
	"database/sql"
	"fmt"

	"github.com/contribsys/sparq/util"
	"golang.org/x/crypto/bcrypt"
)

var (
	newUserInsert = `
		insert into users (Id, Sfid, Nick, Email, FullName, RoleMask)
		values (?, ?, ?, ?, ?, ?)`
	newSecurityInsert = `
		insert into user_securities (UserId, PasswordHash, PublicKey, PrivateKey)
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
	if noRows("select * from users where id = ?", 1) {
		fmt.Println("Creating admin user")
		_, err := db.Exec(newUserInsert, 1, "116672815607840768", "admin", "admin@"+InstanceHostname, "Sparq Admin", -1)
		if err != nil {
			return err
		}
		hash, err := bcrypt.GenerateFromPassword([]byte("sparq123"), 12)
		if err != nil {
			return err
		}
		util.Infof("Admin password hash: %s", hash)
		pub, priv := util.GenerateKeys()
		_, err = db.Exec(newSecurityInsert, 1, hash, pub, priv)
		if err != nil {
			return err
		}
		_, err = db.Exec("insert into actors (Id,UserId) values (?, ?)",
			"https://"+InstanceHostname+"/@admin", 1)
		if err != nil {
			return err
		}
	}
	return nil
}
