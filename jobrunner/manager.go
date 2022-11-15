package jobrunner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/contribsys/faktory/client"
	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/faktory"
	"github.com/contribsys/sparq/util"
)

// Runner coordinates the processes for the worker.  It is responsible for
// starting and stopping goroutines to perform work at the desired concurrency level
type Runner struct {
	mut sync.Mutex

	Concurrency int
	Labels      []string
	Queues      []string

	jobCtx         context.Context
	jobCtxCancel   context.CancelFunc
	workerCount    int32
	mgr            faktory.Manager
	state          string
	middleware     []MiddlewareFunc
	shutdownWaiter *sync.WaitGroup
	jobHandlers    map[string]Handler
	eventHandlers  map[lifecycleEventType][]LifecycleEventHandler
}

// Register a handler for the given jobtype.  It is expected that all jobtypes
// are registered upon process startup.
//
//	mgr.Register("ImportantJob", ImportantFunc)
func (mgr *Runner) Register(name string, fn sparq.PerformFunc) {
	mgr.jobHandlers[name] = func(ctx context.Context, job *client.Job) error {
		return fn(ctx, job.Args...)
	}
}

// Register a callback to be fired when a process lifecycle event occurs.
// These are useful for hooking into process startup or shutdown.
func (mgr *Runner) On(event lifecycleEventType, fn LifecycleEventHandler) {
	mgr.eventHandlers[event] = append(mgr.eventHandlers[event], fn)
}

// After calling Quiet(), no more jobs will be pulled
// from Faktory by this process.
func (mgr *Runner) Quiet() {
	mgr.mut.Lock()
	defer mgr.mut.Unlock()

	if mgr.state == "quiet" {
		return
	}

	util.Info("Quieting job runner...")
	mgr.state = "quiet"
	_ = mgr.fireEvent(Quiet)
}

// Terminate signals that the various components should shutdown.
// Blocks on the shutdownWaiter until all components have finished.
func (mgr *Runner) Terminate(shutdownCtx context.Context) {
	util.Infof("Stopping job runner...")
	mgr.Quiet()

	// We give active jobs a few seconds to finish.
	// Executing jobs use a different context so cancel()'ing the system context
	// does not immediately stop the job subsystem. This gives the
	// jobs a few seconds to finish rather than killing them when half-complete.
	poll := 100 * time.Millisecond
	timer := time.NewTimer(poll)
	defer timer.Stop()
	for {
		if mgr.workerCount == 0 {
			break
		}
		select {
		case <-shutdownCtx.Done():
			util.Debugf("%d jobs still running", mgr.workerCount)
			break
		case <-timer.C:
			timer.Reset(poll)
		}
	}

	mgr.mut.Lock()
	defer mgr.mut.Unlock()

	if mgr.state == "terminate" {
		return
	}

	util.Infof("Terminating job runner")
	mgr.state = "terminate"
	mgr.jobCtxCancel()
	_ = mgr.fireEvent(Shutdown)
	mgr.shutdownWaiter.Wait()
}

// NewManager returns a new manager with default values.
func NewRunner(mgr faktory.Manager) *Runner {
	r := &Runner{
		Concurrency: 10,
		Labels:      []string{"sparq-" + sparq.Version},
		Queues:      []string{"default"},

		workerCount:    0,
		mgr:            mgr,
		state:          "",
		shutdownWaiter: &sync.WaitGroup{},
		jobHandlers:    map[string]Handler{},
		eventHandlers: map[lifecycleEventType][]LifecycleEventHandler{
			Startup:  {},
			Quiet:    {},
			Shutdown: {},
		},
	}
	jobCtx, cancelfn := context.WithCancel(context.Background())
	r.jobCtx = jobCtx
	r.jobCtxCancel = cancelfn
	return r
}

// RunWithContext starts processing jobs. The method will return if an error is encountered while starting.
// If the context is present then os signals will be ignored, the context must be canceled for the method to return
// after running.
func (mgr *Runner) Run(ctx context.Context) error {
	err := mgr.fireEvent(Startup)
	if err != nil {
		return err
	}

	for i := 0; i < mgr.Concurrency; i++ {
		go process(mgr.jobCtx, mgr, i)
	}
	return nil
}

func (run *Runner) with(fn func(faktory.Manager) error) error {
	return fn(run.mgr)
}

func (mgr *Runner) fireEvent(event lifecycleEventType) error {
	for _, fn := range mgr.eventHandlers[event] {
		err := fn(mgr)
		if err != nil {
			return fmt.Errorf("Error running lifecycle event handler: %w", err)
		}
	}
	return nil
}
