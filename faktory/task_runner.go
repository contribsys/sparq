package faktory

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/contribsys/sparq/util"
)

/*
 * The task runner allows us to run internal tasks on
 * a recurring schedule, e.g. "reap old heartbeats every 30 seconds".
 *
 * tr = newTaskRunner()
 * tr.AddTask("heartbeat reaper", reapHeartbeats, 30)
 * ts.Run(...)
 */
type taskRunner struct {
	tasks []*task

	walltimeNs int64
	cycles     int64
	executions int64
	mutex      sync.RWMutex
}

type task struct {
	runner     Taskable
	every      int64
	runs       int64
	walltimeNs int64
}

type Taskable interface {
	Name() string
	Execute() error
	Stats() map[string]interface{}
}

func newTaskRunner() *taskRunner {
	return &taskRunner{
		tasks: make([]*task, 0),
	}
}

func (ts *taskRunner) AddTask(sec int64, thing Taskable) {
	var tsk task
	tsk.runner = thing
	tsk.every = sec
	ts.mutex.Lock()
	ts.tasks = append(ts.tasks, &tsk)
	ts.mutex.Unlock()
}

func (ts *taskRunner) Run(ctx context.Context) {
	go func() {
		// add random jitter so the runner goroutine doesn't fire at 000ms
		time.Sleep(time.Duration(rand.Float64()) * time.Second) //nolint:gosec
		timer := time.NewTicker(1 * time.Second)
		defer timer.Stop()

		for {
			ts.cycle()
			select {
			case <-timer.C:
			case <-ctx.Done():
				util.Debug("Stopping scheduled tasks")
				return
			}
		}
	}()
}

func (ts *taskRunner) Stats() map[string]map[string]interface{} {
	data := map[string]map[string]interface{}{}

	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	for _, task := range ts.tasks {
		data[task.runner.Name()] = task.runner.Stats()
	}
	return data
}

func (ts *taskRunner) cycle() {
	count := int64(0)
	start := time.Now()
	sec := start.Unix()
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	for _, t := range ts.tasks {
		if sec%t.every != 0 {
			continue
		}
		tstart := time.Now()
		// util.Debugf("Running task %s", t.runner.Name())
		err := t.runner.Execute()
		tend := time.Now()
		if err != nil {
			util.Warnf("Error running task %s: %v", t.runner.Name(), err)
		}
		atomic.AddInt64(&t.runs, 1)
		atomic.AddInt64(&t.walltimeNs, tend.Sub(tstart).Nanoseconds())
		count++
	}
	end := time.Now()
	atomic.AddInt64(&ts.cycles, 1)
	atomic.AddInt64(&ts.executions, count)
	atomic.AddInt64(&ts.walltimeNs, end.Sub(start).Nanoseconds())
}

func (s *Server) startTasks() {
	ts := newTaskRunner()
	// scan the various sets, looking for things to do
	ts.AddTask(5, &scanner{name: "Scheduled", set: s.store.Scheduled(), task: s.mgr.EnqueueScheduledJobs})
	ts.AddTask(5, &scanner{name: "Retries", set: s.store.Retries(), task: s.mgr.RetryJobs})
	ts.AddTask(60, &scanner{name: "Dead", set: s.store.Dead(), task: s.mgr.Purge})

	// reaps job reservations which have expired
	ts.AddTask(15, &reservationReaper{s.mgr, 0})

	ts.Run(s.ctx)
	s.taskRunner = ts
}