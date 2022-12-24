package clientapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/util"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type StreamEvent struct {
	Name string
	Data string
}

func NewEvent(name string, data string) StreamEvent {
	return StreamEvent{name, data}
}

func NewJsonEvent(name string, data any) StreamEvent {
	datas, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return StreamEvent{name, string(datas)}
}

type Streamer struct {
	streamListeners map[any]map[int64]chan StreamEvent
	streamCount     int32
	mu              sync.Mutex
}

func NewStreamer(s sparq.Server) *Streamer {
	return &Streamer{
		streamListeners: map[any]map[int64]chan StreamEvent{},
		streamCount:     0,
		mu:              sync.Mutex{},
	}
}

func (s *Streamer) Metrics() map[string]any {
	return map[string]any{
		"streams": s.streamCount,
	}
}

func (s *Streamer) Run(ctx context.Context) {
	util.Debugf("Starting streaming ping")
	go s.ping(ctx)
}

func (s *Streamer) Fanout(key any, event StreamEvent) {
	chans := s.streamListeners[key]
	for _, chn := range chans {
		chn <- event
	}
}

func (s *Streamer) Handler(sp sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			httpError(w, http.ErrNotSupported, 400)
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			httpError(w, errors.New("http response not flushable"), 503)
			return
		}

		key := mux.Vars(r)["key"]
		if key == "user" {
			store, err := sessionStore.Get(r, "sparq-session")
			if err != nil {
				httpError(w, err, 500)
			}
			key = store.Values["uid"].(string)
		}

		chanl, dereg := s.registerStreamerFor(key)
		defer dereg()
		// util.Infof("Registered stream for %s", key)

		// Send the initial headers saying we're going to stream the response.
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		for {
			select {
			case <-sp.Context().Done():
				return
			case <-r.Context().Done():
				return
			case e := <-chanl:
				// util.Debugf("Writing stream event: %+v", e)
				_, err := io.WriteString(w, fmt.Sprintf("event: %s\n", e.Name))
				if err != nil {
					return
				}
				if e.Data != "" {
					_, err = io.WriteString(w, fmt.Sprintf("data: %s\n", e.Data))
					if err != nil {
						return
					}
				}
				_, _ = w.Write([]byte("\n"))
				flusher.Flush()
			}
		}

	}
}

func (s *Streamer) ping(ctx context.Context) {
	ping := StreamEvent{Name: ":ping"}

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(25 * time.Second):
			s.mu.Lock()
			for _, mp := range s.streamListeners {
				for _, chn := range mp {
					chn <- ping
				}
			}
			s.mu.Unlock()
		}
	}
}

func (s *Streamer) registerStreamerFor(key any) (<-chan StreamEvent, func()) {
	chn := make(chan StreamEvent, 10)
	code := rand.Int63()

	s.mu.Lock()
	if _, ok := s.streamListeners[key]; ok {
		s.streamListeners[key][code] = chn
	} else {
		mp := map[int64]chan StreamEvent{}
		mp[code] = chn
		s.streamListeners[key] = mp
	}
	s.mu.Unlock()
	atomic.AddInt32(&s.streamCount, 1)

	// chn <- StreamEvent{Name: "delete", Data: "12345"}

	return chn, func() {
		// util.Debugf("Stream %d unregistered", code)
		atomic.AddInt32(&s.streamCount, -1)
		s.mu.Lock()
		defer s.mu.Unlock()
		close(chn)
		delete(s.streamListeners[key], code)
	}
}
