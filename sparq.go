package sparq

import (
	"context"
	"embed"
	"fmt"

	"github.com/contribsys/faktory/client"
)

const (
	Name    = "Sparq⚡️"
	Version = "0.0.1"
)

//go:embed db/migrate/*.sql
var Migrations embed.FS

var (
	UserAgent = fmt.Sprintf("%s v%s", Name, Version)
)

type PerformFunc func(ctx context.Context, args ...interface{}) error

type Pusher interface {
	Push(context.Context, *client.Job) error
}

type JobService interface {
	Pusher
	Register(jobtype string, fn PerformFunc)
}
