package faktory

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/contribsys/faktory/manager"
)

type reservationReaper struct {
	m     manager.Manager
	count int64
}

func (r *reservationReaper) Name() string {
	return "Busy"
}

func (r *reservationReaper) Execute(ctx context.Context) error {
	count, err := r.m.ReapExpiredJobs(ctx, time.Now())
	if err != nil {
		return err
	}

	atomic.AddInt64(&r.count, int64(count))
	return nil
}

func (r *reservationReaper) Stats(context.Context) map[string]interface{} {
	return map[string]interface{}{
		"size":   r.m.WorkingCount(),
		"reaped": atomic.LoadInt64(&r.count),
	}
}
