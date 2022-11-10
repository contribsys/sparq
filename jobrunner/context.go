package jobrunner

import (
	"context"
	"log"

	"github.com/contribsys/faktory/client"
	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/faktory"
)

// internal keys for context value storage
type valueKey int

const (
	mgrKey valueKey = 2
	jobKey valueKey = 3
)

// The Helper provides access to valuable data and APIs
// within an executing job.
//
// We're pretty strict about what's exposed in the Helper
// because execution should be orthogonal to
// most of the Job payload contents.
//
//		func myJob(ctx context.Context, args ...interface{}) error {
//		  helper := worker.HelperFor(ctx)
//		  jid := helper.Jid()
//
//		  helper.With(func(mgr faktory.Manager) error {
//	      job := client.NewJob("JobType", 1, 2, 3)
//		    mgr.Push(job)
//			})
type Helper interface {
	Jid() string
	JobType() string

	// Custom provides access to the job custom hash.
	// Returns the value and `ok=true` if the key was found.
	// If not, returns `nil` and `ok=false`.
	//
	// No type checking is performed, please use with caution.
	Custom(key string) (value interface{}, ok bool)

	// allows direct access to the Faktory server from the job
	With(func(sparq.Pusher) error) error
}

type jobHelper struct {
	job *client.Job
	mgr sparq.Pusher
}

// ensure type compatibility
var _ Helper = &jobHelper{}

func (h *jobHelper) Jid() string {
	return h.job.Jid
}
func (h *jobHelper) JobType() string {
	return h.job.Type
}
func (h *jobHelper) Custom(key string) (value interface{}, ok bool) {
	return h.job.GetCustom(key)
}

// Caution: this method must only be called within the
// context of an executing job. It will panic if it cannot
// create a Helper due to missing context values.
func HelperFor(ctx context.Context) Helper {
	if j := ctx.Value(jobKey); j != nil {
		job := j.(*client.Job)
		if p := ctx.Value(mgrKey); p != nil {
			mgr := p.(faktory.Manager)
			return &jobHelper{
				job: job,
				mgr: mgr,
			}
		}
	}
	log.Panic("Invalid job context, cannot create job helper")
	return nil
}

func jobContext(mgr *Runner, job *client.Job) context.Context {
	ctx := mgr.ctx
	ctx = context.WithValue(ctx, mgrKey, mgr.mgr)
	ctx = context.WithValue(ctx, jobKey, job)
	return ctx
}

func (h *jobHelper) With(fn func(sparq.Pusher) error) error {
	return fn(h.mgr)
}
