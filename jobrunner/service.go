package jobrunner

import (
	"context"

	"github.com/contribsys/faktory/client"
	"github.com/contribsys/faktory/manager"
	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/util"
)

type JobRunner struct {
	mgr  manager.Manager
	exec *Runner
}

type Options struct {
	Concurrency int
	Queues      []string
}

func NewJobRunner(mgr manager.Manager, opts Options) *JobRunner {
	exec := NewRunner(mgr)
	exec.Concurrency = opts.Concurrency
	exec.Queues = opts.Queues
	return &JobRunner{mgr, exec}
}

func (jr *JobRunner) Run(ctx context.Context) error {
	util.Infof("Starting Faktory job runner with %d concurrency", jr.exec.Concurrency)
	return jr.exec.Run(ctx)
}

func (jr *JobRunner) Shutdown(ctx context.Context) error {
	// TODO Fix this to use context, not Wait forever
	jr.exec.Terminate()
	return nil
}

func (jr *JobRunner) Push(job *client.Job) error {
	return jr.mgr.Push(job)
}

func (jr *JobRunner) Register(jobtype string, fn sparq.PerformFunc) {
	jr.exec.Register(jobtype, fn)
}
