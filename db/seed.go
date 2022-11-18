package db

import (
	"database/sql"
	"fmt"

	"github.com/contribsys/sparq/util"
	"golang.org/x/crypto/bcrypt"
)

var (
	newUserInsert = `
		insert into users (
			Id, Sfid, Nick, Email, FullName, RoleMask
		) values (?, ?, ?, ?, ?, ?)`
	newSecurityInsert = `
		insert into user_securities (
			UserId, PasswordHash, PublicKey, PrivateKey
		) values (?, ?, ?, ?)`
)

func Seed() error {
	fmt.Println("Seeding...")

	admin := map[string]interface{}{}
	row := db.QueryRowx("select * from users where id = ?", 1)
	err := row.MapScan(admin)
	if err == sql.ErrNoRows {
		sf := util.NewSnowflake()
		_, err := db.Exec(newUserInsert, 1, sf.NextID(), "admin", "admin@localhost.dev", "Sparq Admin", -1)
		if err != nil {
			return err
		}
		hash, err := bcrypt.GenerateFromPassword([]byte("sparq123"), 12)
		if err != nil {
			return err
		}
		pub, priv := util.GenerateKeys()
		_, err = db.Exec(newSecurityInsert, 1, hash, pub, priv)
		if err != nil {
			return err
		}

		return nil
	}
	return err
}
