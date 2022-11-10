package faktory

import (
	"context"
	"reflect"

	"github.com/contribsys/faktory/client"
	"github.com/contribsys/faktory/manager"
	"github.com/contribsys/sparq"
)

type Manager interface {
	sparq.Pusher
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
