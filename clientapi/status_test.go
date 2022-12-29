package clientapi

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/contribsys/sparq/db"
	"github.com/stretchr/testify/assert"
)

func TestStatus(t *testing.T) {
	stopper, err := db.TestDB("status")
	assert.NoError(t, err)
	defer stopper()
	ts := &testSvr{}
	token, err := registerToken(t, ts)
	assert.NoError(t, err)
	root := rootRouter(ts)
	AddPublicEndpoints(ts, root.PathPrefix("/api/v1").Subrouter())

	t.Run("Minimal", func(t *testing.T) {
		form := strings.NewReader(url.Values{
			"status": []string{"<p>A <b>bold</b> text to post, so brave...</p>"},
		}.Encode())
		req := httptest.NewRequest("POST", "http://localhost.dev:9494/api/v1/statuses", form)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Idempotency-Key", "mike-rules")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, w.Code, 200)
		assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
		assert.Contains(t, w.Body.String(), `brave`)

		var testy map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &testy)
		assert.NoError(t, err)
		assert.Equal(t, nil, testy)
	})

}
