package runtime

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/contribsys/sparq/faktory"
	"github.com/contribsys/sparq/faktoryui"
	"github.com/contribsys/sparq/finger"
	"github.com/contribsys/sparq/jobrunner"
	"github.com/contribsys/sparq/util"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
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
	Database  *gorm.DB
	JobServer *faktory.Server
	FaktoryUI *faktoryui.WebUI
	JobRunner *jobrunner.JobRunner

	https  *http.Server
	cancel context.CancelFunc
	ctx    context.Context
}

func NewService(opts Options) (*Service, error) {
	dbx, err := gorm.Open(sqlite.Open("sparq.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &Service{
		Database: dbx,
		ctx:      ctx,
		cancel:   cancel,
		Options:  opts,
	}

	js, _ := faktory.NewServer(faktory.Options{
		StorageDirectory: opts.StorageDirectory,
		RedisSock:        "sparq.redis.sock",
	})
	err = js.Run(ctx) // does not block
	if err != nil {
		return nil, err
	}
	s.JobServer = js
	s.FaktoryUI = faktoryui.NewWeb(js, opts.Binding)
	s.JobRunner = jobrunner.NewJobRunner(js.Manager())
	return s, nil
}

func (s *Service) Close() error {
	s.cancel()
	return nil
}

func (s *Service) Run() error {
	// This is the context which signals that we are starting
	// the shutdown process

	root := http.NewServeMux()
	root.HandleFunc("/.well-known/webfinger", finger.HttpHandler(s.Database, s.Binding))
	root.Handle("/faktory/", http.StripPrefix("/faktory", s.FaktoryUI.App))

	ht := &http.Server{
		Addr:           s.Binding,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 16,
		Handler:        root,
	}
	s.https = ht

	go func() {
		util.Infof("Web now running at %s", s.Binding)
		err := ht.ListenAndServe()
		if err != http.ErrServerClosed {
			util.Error("web server crashed", err)
		}
	}()

	err := s.JobRunner.Run(s.ctx)
	if err != nil {
		return err
	}

	<-s.ctx.Done()
	s.shutdown(10 * time.Second)
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

	// this shuts down Redis, can't call until JobRunner is dead
	s.JobServer.Close()
}
