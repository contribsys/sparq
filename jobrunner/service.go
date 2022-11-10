package jobrunner

import (
	"context"

	"github.com/contribsys/faktory/client"
	"github.com/contribsys/faktory/manager"
)

type JobRunner struct {
	mgr  manager.Manager
	exec *Runner
}

func NewJobRunner(mgr manager.Manager) *JobRunner {
	return &JobRunner{mgr, NewRunner(mgr)}
}

func (jr *JobRunner) Run(ctx context.Context) error {
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
