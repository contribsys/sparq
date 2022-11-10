package jobrunner

import (
	"context"
	"fmt"
	"sync"

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

	ctx            context.Context
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
func (mgr *Runner) Terminate() {
	mgr.mut.Lock()
	defer mgr.mut.Unlock()

	if mgr.state == "terminate" {
		return
	}

	mgr.state = "terminate"
	_ = mgr.fireEvent(Shutdown)
	mgr.shutdownWaiter.Wait()
}

// NewManager returns a new manager with default values.
func NewRunner(mgr faktory.Manager) *Runner {
	return &Runner{
		Concurrency: 10,
		Labels:      []string{"sparq-" + sparq.Version},
		Queues:      []string{"default"},

		ctx:            nil,
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
}

// RunWithContext starts processing jobs. The method will return if an error is encountered while starting.
// If the context is present then os signals will be ignored, the context must be canceled for the method to return
// after running.
func (mgr *Runner) Run(ctx context.Context) error {
	mgr.ctx = ctx
	err := mgr.fireEvent(Startup)
	if err != nil {
		return err
	}

	for i := 0; i < mgr.Concurrency; i++ {
		go process(mgr, i)
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
