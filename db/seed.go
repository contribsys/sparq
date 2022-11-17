package db

import (
	"database/sql"
	"fmt"

	"github.com/contribsys/sparq/model"
	"github.com/contribsys/sparq/util"
)

var (
	newUserInsert = `
		insert into users (
			id, sfid, nick, email, full_name, role_mask
		) values (?, ?, ?, ?, ?, ?)
		`
)

func Seed() error {
	admin := map[string]interface{}{}
	row := db.QueryRowx("select * from users where id = ?", 1)
	err := row.MapScan(admin)
	if err == sql.ErrNoRows {
		sf := util.NewSnowflake()
		_, err := db.Exec(newUserInsert, 1, sf.NextID(), "admin", "admin@localhost.dev", "Sparq Admin", model.RoleAll)
		if err != nil {
			return err
		}
	}
	fmt.Printf("%+v", admin)
	return err
	// if err != nil {
	// panic(err)
	// }
}
