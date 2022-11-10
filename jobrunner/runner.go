package jobrunner

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"time"

	"github.com/contribsys/faktory/client"
	"github.com/contribsys/sparq/faktory"
	"github.com/contribsys/sparq/util"
)

type lifecycleEventType int

const (
	Startup  lifecycleEventType = 1
	Quiet    lifecycleEventType = 2
	Shutdown lifecycleEventType = 3
)

type NoHandlerError struct {
	JobType string
}

func (s *NoHandlerError) Error() string {
	return fmt.Sprintf("No handler registered for job type %s", s.JobType)
}

func process(mgr *Runner, idx int) {
	mgr.shutdownWaiter.Add(1)
	defer mgr.shutdownWaiter.Done()

	// delay initial fetch randomly to prevent thundering herd.
	// this will pause between 0 and 2B nanoseconds, i.e. 0-2 seconds
	time.Sleep(time.Duration(rand.Int31()))
	sleep := 1.0
	for {
		if mgr.state != "" {
			return
		}

		// check for shutdown
		select {
		case <-mgr.ctx.Done():
			return
		default:
		}

		err := processOne(mgr)
		if err != nil {
			if err != io.EOF {
				// Faktory's Fetch from Redis doesn't use the Context under the covers
				// so we need this hack to avoid a load of errors on shutdown
				util.Error("Error running job", err)
			}
			if _, ok := err.(*NoHandlerError); !ok {
				// if we don't know how to process this jobtype,
				// we Fail it and sleep for a bit so we don't get
				// caught in an infinite loop "processing" a queue full
				// of jobs we don't understand.
				time.Sleep(50 * time.Millisecond)
			} else {
				// if we have an unknown error processing a job, use
				// exponential backoff so we don't constantly slam the
				// log with "connection refused" errors or similar.
				select {
				case <-mgr.ctx.Done():
				case <-time.After(time.Duration(sleep) * time.Second):
					sleep = math.Max(sleep*2, 30)
				}
			}
		} else {
			// success, reset sleep timer
			sleep = 1.0
		}
	}
}

func processOne(mgr *Runner) error {
	var job *client.Job

	// explicit scopes to limit variable visibility
	{
		var e error
		err := mgr.with(func(c faktory.Manager) error {
			job, e = c.Fetch(mgr.ctx, "sparq", mgr.queues...)
			if e != nil {
				return e
			}
			return nil
		})
		if err != nil {
			return err
		}
		if job == nil {
			return nil
		}
	}

	perform := mgr.jobHandlers[job.Type]

	if perform == nil {
		je := &NoHandlerError{JobType: job.Type}
		err := mgr.with(func(c faktory.Manager) error {
			return c.Fail(faktory.ToFailure(job.Jid, je))
		})
		if err != nil {
			return err
		}
		return je
	}

	joberr := dispatch(mgr.middleware, jobContext(mgr, job), job, perform)
	if joberr != nil {
		// job errors are normal and expected, we don't return early from them
		util.Error(fmt.Sprintf("Error running %s job %s", job.Type, job.Jid), joberr)
	}

	return mgr.with(func(c faktory.Manager) error {
		if joberr != nil {
			return c.Fail(faktory.ToFailure(job.Jid, joberr))
		} else {
			_, err := c.Acknowledge(job.Jid)
			return err
		}
	})
}
