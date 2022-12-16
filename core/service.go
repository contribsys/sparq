package core

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/contribsys/sparq/adminui"
	"github.com/contribsys/sparq/db"
	"github.com/contribsys/sparq/faktory"
	"github.com/contribsys/sparq/faktoryui"
	"github.com/contribsys/sparq/jobrunner"
	"github.com/contribsys/sparq/util"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Options struct {
	Binding          string
	Hostname         string
	LogLevel         string
	ConfigDirectory  string
	StorageDirectory string
}

// This is the main Sparq service.
// It holds all of the child services and orchestrates them.
type Service struct {
	Options
	JobServer *faktory.Server
	FaktoryUI *faktoryui.WebUI
	AdminUI   *adminui.WebUI
	JobRunner *jobrunner.JobRunner

	https  *http.Server
	cancel context.CancelFunc
	ctx    context.Context
}

// Implement sparq.Server interface
func (s *Service) DB() *sqlx.DB {
	return db.Database()
}

func (s *Service) LogLevel() string {
	return s.Options.LogLevel
}

func (s *Service) Hostname() string {
	return s.Options.Hostname
}

func NewService(opts Options) (*Service, error) {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Service{
		ctx:     ctx,
		cancel:  cancel,
		Options: opts,
	}

	js, _ := faktory.NewServer(faktory.Options{
		StorageDirectory: opts.StorageDirectory,
		RedisSock:        fmt.Sprintf("sparq.redis.%s.sock", opts.Hostname),
	})
	err := js.Run(ctx) // does not block
	if err != nil {
		return nil, err
	}
	s.JobServer = js
	s.FaktoryUI = faktoryui.NewWeb(js, opts.Binding)
	s.AdminUI = adminui.NewWeb(js.Manager(), opts.Binding)
	s.JobRunner = jobrunner.NewJobRunner(js.Manager(), jobrunner.Options{
		Concurrency: 1,
		Queues:      []string{"high", "default", "low"},
	})
	adminui.Register(s.JobRunner)
	return s, nil
}

func (s *Service) Close() error {
	s.cancel()
	return nil
}

func (s *Service) Run() error {
	defer s.JobServer.RedisStopper()
	// This is the context which signals that we are starting
	// the shutdown process

	s.https = BuildWeb(s)

	util.Infof("Web now running at %s", s.Binding)
	go func() {
		err := s.https.ListenAndServe()
		if err != http.ErrServerClosed {
			util.Error("web server crashed", err)
		}
	}()

	err := s.JobRunner.Run(s.ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Welcome to the Fediverse, \033[32m%s\033[0m\n", s.Options.Hostname)
	<-s.ctx.Done()
	s.shutdown(20 * time.Second)
	return nil
}

func (s *Service) shutdown(timeout time.Duration) {
	hardTimeout, cancel := context.WithTimeout(context.Background(), timeout)
	s.cancel = cancel

	var grp sync.WaitGroup

	grp.Add(1)
	go func() {
		err := s.https.Shutdown(hardTimeout)
		if err != nil {
			util.Error("shutdown", err)
		}
		grp.Done()
	}()

	grp.Add(1)
	go func() {
		err := s.JobRunner.Shutdown(hardTimeout)
		if err != nil {
			util.Error("shutdown", err)
		}
		grp.Done()
	}()
	grp.Wait()

	util.Infof("Stopping job server")
	// this shuts down Redis, can't call until JobRunner is dead
	s.JobServer.Close()
}
