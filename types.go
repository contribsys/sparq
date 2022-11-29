package sparq

import "github.com/jmoiron/sqlx"

type Server interface {
	DB() *sqlx.DB
	Hostname() string
}
