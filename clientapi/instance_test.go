package clientapi

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/contribsys/sparq/db"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestInstance(t *testing.T) {
	stopper, err := db.TestDB("instance")
	assert.NoError(t, err)
	defer stopper()
	ts := &testSvr{}
	root := rootRouter(ts)
	AddPublicEndpoints(ts, root.PathPrefix("/api/v1").Subrouter())

	t.Run("instance", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://localhost.dev:9494/api/v1/instance", nil)
		w := httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, w.Code, 200)
		assert.Contains(t, w.Body.String(), `"uri": "localhost.dev",`)

		var testy map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &testy)
		assert.NoError(t, err)
	})

	t.Run("apps/verify", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "http://localhost.dev:9494/api/v1/apps/verify_credentials", nil)
		w := httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 204, w.Code)
		assert.Equal(t, w.Body.String(), "")

		req = httptest.NewRequest("POST", "http://localhost.dev:9494/api/v1/apps/verify_credentials", nil)
		w = httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 401, w.Code)
		assert.Contains(t, w.Body.String(), "not found")

		req = httptest.NewRequest("GET", "http://localhost.dev:9494/api/v1/apps/verify_credentials", nil)
		w = httptest.NewRecorder()
		req.Header.Set("Authorization", "Bearer 123456")
		root.ServeHTTP(w, req)
		assert.Equal(t, 401, w.Code)
		assert.Contains(t, w.Body.String(), "invalid_token")

		assert.Equal(t, 0, oauthClientCount(t))
		token, err := registerToken(t, ts)
		assert.NoError(t, err)
		assert.Equal(t, 1, oauthClientCount(t))

		req = httptest.NewRequest("GET", "http://localhost.dev:9494/api/v1/apps/verify_credentials", nil)
		w = httptest.NewRecorder()
		req.Header.Set("Authorization", "Bearer "+token)
		root.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "pinafore.social")
		fmt.Println(w.Body.String())
	})

	t.Run("apps", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "http://localhost.dev:9494/api/v1/apps", nil)
		w := httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, w.Code, 204)
		assert.Equal(t, w.Body.String(), "")

		// GET not allowed
		req = httptest.NewRequest("GET", "http://localhost.dev:9494/api/v1/apps", nil)
		w = httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 400, w.Code)
		assert.Contains(t, w.Body.String(), "Bad method")

		// no body
		req = httptest.NewRequest("POST", "http://localhost.dev:9494/api/v1/apps", nil)
		w = httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 400, w.Code)
		assert.Contains(t, w.Body.String(), "")

		beforeCount := oauthClientCount(t)
		j := `{"client_name":"Pinafore",
		 "redirect_uris":"https://pinafore.social/settings/instances/add",
		 "scopes":"read write follow push",
		 "website":"https://pinafore.social"}`
		br := strings.NewReader(j)
		req = httptest.NewRequest("POST", "http://localhost.dev:9494/api/v1/apps", br)
		req.Header.Add("Content-Type", "application/json")
		w = httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "client_secret")
		assert.Contains(t, w.Body.String(), "Pinafore")

		var testy map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &testy)
		assert.NoError(t, err)

		afterCount := oauthClientCount(t)
		assert.NotEqual(t, beforeCount, afterCount)
	})
}

type testSvr struct {
}

func (ts *testSvr) DB() *sqlx.DB {
	return db.Database()
}

func (ts *testSvr) Hostname() string {
	return "localhost.dev"
}

func (ts *testSvr) LogLevel() string {
	return "debug"
}

func oauthClientCount(t *testing.T) int {
	var count int
	err := db.Database().QueryRow("select count(*) from oauth_clients").Scan(&count)
	assert.NoError(t, err)
	return count
}

func oauthTokenCount(t *testing.T) int {
	var count int
	err := db.Database().QueryRow("select count(*) from oauth_tokens").Scan(&count)
	assert.NoError(t, err)
	return count
}
