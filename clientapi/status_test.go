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

	t.Run("PostMinimal", func(t *testing.T) {
		form := strings.NewReader(url.Values{"status": []string{"<p>A <b>bold</b> text to post, so brave...</p>"}}.Encode())
		req := httptest.NewRequest("POST", "http://localhost.dev:9494/api/v1/statuses", form)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Idempotency-Key", "mike-rules")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, w.Code, 200)
		assert.Equal(t, w.Header().Get("Content-Type"), "application/json")
		assert.Contains(t, w.Body.String(), `brave`)
		assert.Contains(t, w.Body.String(), `Pinafore`)

		var testy map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &testy)
		assert.NoError(t, err)
		assert.NotNil(t, testy)

		form = strings.NewReader(url.Values{"status": []string{"<p>A <b>bold</b> text to post, so brave...</p>"}}.Encode())
		req = httptest.NewRequest("POST", "http://localhost.dev:9494/api/v1/statuses", form)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Idempotency-Key", "mike-rules")
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 401, w.Code)
		assert.Contains(t, w.Body.String(), "Duplicate")
	})
	t.Run("PostErrors", func(t *testing.T) {
		form := strings.NewReader(url.Values{"status": []string{""}}.Encode())
		req := httptest.NewRequest("POST", "http://localhost.dev:9494/api/v1/statuses", form)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Idempotency-Key", "mike-rulez")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 400, w.Code)
		assert.Contains(t, w.Body.String(), `No content`)
	})

	t.Run("GetStatus", func(t *testing.T) {
		form := strings.NewReader(url.Values{"status": []string{"<p>A strong text to post, so brave...</p>"}}.Encode())
		req := httptest.NewRequest("POST", "http://localhost.dev:9494/api/v1/statuses", form)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, w.Code, 200)

		var testy map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &testy)
		assert.NoError(t, err)
		assert.NotNil(t, testy)
		sid := testy["id"].(string)

		req = httptest.NewRequest("GET", "http://localhost.dev:9494/api/v1/statuses/"+sid, nil)
		w = httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, w.Code, 200)

		err = json.Unmarshal(w.Body.Bytes(), &testy)
		assert.NoError(t, err)
		assert.NotNil(t, testy)
	})
}
