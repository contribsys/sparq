package faktory

import (
	"context"
	"reflect"

	"github.com/contribsys/faktory/client"
	"github.com/contribsys/faktory/manager"
)

type Pusher interface {
	Push(*client.Job) error
}

type Manager interface {
	Push(*client.Job) error
	Fetch(ctx context.Context, wid string, queues ...string) (*client.Job, error)

	Fail(fail *manager.FailPayload) error
	Acknowledge(jid string) (*client.Job, error)
}

func ToFailure(jid string, err error) *manager.FailPayload {
	return &manager.FailPayload{
		Jid:          jid,
		ErrorMessage: err.Error(),
		ErrorType:    reflect.TypeOf(err).String(),
	}
}
