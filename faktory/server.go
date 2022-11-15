package faktory

import (
	"context"
	"fmt"
	rt "runtime"
	"sync"
	"time"

	"github.com/contribsys/faktory/client"
	"github.com/contribsys/faktory/manager"
	"github.com/contribsys/faktory/storage"
	"github.com/contribsys/sparq/util"
	"github.com/go-redis/redis/v9"
)

type Options struct {
	StorageDirectory string
	RedisSock        string
}

type Server struct {
	Options
	StartedAt time.Time

	mgr          manager.Manager
	store        storage.Store
	taskRunner   *taskRunner
	mu           sync.Mutex
	closed       bool
	redisStopper func()
}

func NewServer(opts Options) (*Server, error) {
	if opts.StorageDirectory == "" {
		opts.StorageDirectory = "."
	}
	if opts.RedisSock == "" {
		opts.RedisSock = fmt.Sprintf("%s/%s", opts.StorageDirectory, "sparq.redis.sock")
	}
	// this runs Redis as a child process
	stopper, err := storage.Boot(opts.StorageDirectory, opts.RedisSock)
	if err != nil {
		return nil, err
	}

	s := &Server{
		Options:      opts,
		StartedAt:    time.Now(),
		closed:       false,
		redisStopper: stopper,
	}

	return s, nil
}

func (s *Server) Store() storage.Store {
	return s.store
}

func (s *Server) Manager() manager.Manager {
	return s.mgr
}

func (s *Server) Run(ctx context.Context) error {
	store, err := storage.Open(s.Options.RedisSock, rt.NumCPU()*20)
	if err != nil {
		return fmt.Errorf("cannot open redis database: %w", err)
	}

	s.mu.Lock()
	s.store = store
	s.mgr = manager.NewManager(store)
	s.startTasks(ctx)
	s.mu.Unlock()

	util.Infof("Faktory %s booted", client.Version)
	return nil
}

func (s *Server) Close() {
	s.store.Close()
	s.redisStopper()
}

func (s *Server) uptimeInSeconds() int {
	return int(time.Since(s.StartedAt).Seconds())
}

func (s *Server) RuntimeStats() map[string]interface{} {
	return map[string]interface{}{
		"description":     client.Name,
		"faktory_version": client.Version,
		"uptime":          s.uptimeInSeconds(),
		"used_memory_mb":  util.MemoryUsageMB(),
	}
}

func (s *Server) CurrentState(ctx context.Context) (map[string]interface{}, error) {
	queueCmd := map[string]*redis.IntCmd{}
	_, err := s.store.Redis().Pipelined(ctx, func(pipe redis.Pipeliner) error {
		s.store.EachQueue(ctx, func(q storage.Queue) {
			queueCmd[q.Name()] = pipe.LLen(ctx, q.Name())
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	queues := map[string]int64{}
	totalQueued := int64(0)
	totalQueues := len(queueCmd)
	for name, cmd := range queueCmd {
		qsize := cmd.Val()
		totalQueued += qsize
		queues[name] = qsize
	}

	return map[string]interface{}{
		"now":             util.Nows(),
		"server_utc_time": time.Now().UTC().Format("15:04:05 UTC"),
		"faktory": map[string]interface{}{
			"total_failures":  s.store.TotalFailures(ctx),
			"total_processed": s.store.TotalProcessed(ctx),
			"total_enqueued":  totalQueued,
			"total_queues":    totalQueues,
			"queues":          queues,
			"tasks":           s.taskRunner.Stats(ctx),
		},
		"server": map[string]interface{}{
			"description":     client.Name,
			"faktory_version": client.Version,
			"uptime":          s.uptimeInSeconds(),
			"used_memory_mb":  util.MemoryUsageMB(),
		},
	}, nil
}
