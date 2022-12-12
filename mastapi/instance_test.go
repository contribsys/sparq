package mastapi

import (
	"encoding/json"
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
	ts := &TestSvr{}

	t.Run("instance", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://localhost.dev:9494/api/v1/instance", nil)
		w := httptest.NewRecorder()
		instanceHandler(ts)(w, req)
		assert.Equal(t, w.Code, 200)
		assert.Contains(t, w.Body.String(), `"domain": "localhost.dev",`)

		var testy map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &testy)
		assert.NoError(t, err)
	})

	t.Run("apps", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "http://localhost.dev:9494/api/v1/apps", nil)
		w := httptest.NewRecorder()
		appsHandler(ts)(w, req)
		assert.Equal(t, w.Code, 204)
		assert.Contains(t, w.Body.String(), "")

		// GET not allowed
		req = httptest.NewRequest("GET", "http://localhost.dev:9494/api/v1/apps", nil)
		w = httptest.NewRecorder()
		appsHandler(ts)(w, req)
		assert.Equal(t, 400, w.Code)
		assert.Contains(t, w.Body.String(), "Bad method")

		// no body
		req = httptest.NewRequest("POST", "http://localhost.dev:9494/api/v1/apps", nil)
		w = httptest.NewRecorder()
		appsHandler(ts)(w, req)
		assert.Equal(t, 400, w.Code)
		assert.Contains(t, w.Body.String(), "")

		j := `{"client_name":"Pinafore",
		 "redirect_uris":"https://pinafore.social/settings/instances/add",
		 "scopes":"read write follow push",
		 "website":"https://pinafore.social"}`
		br := strings.NewReader(j)
		req = httptest.NewRequest("POST", "http://localhost.dev:9494/api/v1/apps", br)
		req.Header.Add("Content-Type", "application/json")
		w = httptest.NewRecorder()
		h := appsHandler(ts)
		h(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "client_secret")
		assert.Contains(t, w.Body.String(), "Pinafore")

		var count int
		err := db.Database().QueryRow("select count(*) from oauth_apps where ClientName = 'Pinafore'").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

type TestSvr struct {
}

func (ts *TestSvr) DB() *sqlx.DB {
	return db.Database()
}

func (ts *TestSvr) Hostname() string {
	return "localhost.dev"
}
