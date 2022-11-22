package adminui

import (
	"context"
	"fmt"

	"github.com/contribsys/faktory/client"
	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/jobrunner"
)

func NewJob(jobtype string, queue string, args ...interface{}) *client.Job {
	job := client.NewJob(jobtype, args...)
	job.Queue = queue
	return job
}

func Register(js sparq.JobService) {
	js.Register("atype", func(ctx context.Context, args ...interface{}) error {
		return AType(ctx, args[0].(string))
	})
	js.Register("btype", func(ctx context.Context, args ...interface{}) error {
		return BType(ctx, args[0].(string))
	})
}

func AType(ctx context.Context, name string) error {
	fmt.Println("Hello!", name)
	helper := jobrunner.HelperFor(ctx)
	return helper.With(func(p sparq.Pusher) error {
		return p.Push(ctx, NewJob("btype", "low", name))
	})
}

func BType(ctx context.Context, name string) error {
	fmt.Println("Goodbye", name)
	return nil
}
