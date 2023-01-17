package clientapi

import (
	"context"
	"testing"

	"github.com/contribsys/sparq/web"
	"github.com/stretchr/testify/assert"
)

func TestStreaming(t *testing.T) {
	ts, stopper := web.NewTestServer(t, "streaming")
	defer stopper()
	s := NewStreamer(ts)
	assert.Contains(t, s.Metrics(), "streams")

	ctx, cancel := context.WithCancel(context.Background())
	s.Run(ctx)

	handler := s.Handler(ts)
	assert.NotNil(t, handler)
	// TODO How to test?
	// r := httptest.NewRequest("GET", "http://localhost:9494/api/v1/streaming/user", nil)
	// w := httptest.NewRecorder()
	// handler(w, r)
	// assert.Equal(t, 401, w.Code)

	cancel()
}
