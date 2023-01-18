package sparq

import (
	"context"
	"fmt"

	"github.com/contribsys/faktory/client"
)

const (
	Name    = "Sparq⚡️"
	Version = "0.0.1"
)

var (
	ServerHeader = fmt.Sprintf("%s v%s; github.com/contribsys/sparq", Name, Version)
)

type PerformFunc func(ctx context.Context, args ...interface{}) error

type Pusher interface {
	Push(context.Context, *client.Job) error
}

type JobService interface {
	Pusher
	Register(jobtype string, fn PerformFunc)
}
