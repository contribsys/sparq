package sparq

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Server interface {
	DB() *sqlx.DB
	Hostname() string
	LogLevel() string
	Context() context.Context
}
