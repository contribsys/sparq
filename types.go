package sparq

import (
	"github.com/jmoiron/sqlx"
)

type Server interface {
	DB() *sqlx.DB
	Hostname() string
	LogLevel() string
}

var (
	ContextKey int = 7
)

type WebContext struct {
	Bearer            string
	Locale            string
	LoggedInNick      string
	LoggedInAccountID int
}
