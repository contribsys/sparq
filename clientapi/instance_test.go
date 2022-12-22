package clientapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/db"
	"github.com/contribsys/sparq/model"
	"github.com/contribsys/sparq/oauth2"
	"github.com/contribsys/sparq/public"
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

// returns the access token or error
func registerToken(t *testing.T, s sparq.Server) (string, error) {
	clientHash := map[string]string{
		"client_name":   "Pinafore",
		"redirect_uris": "https://pinafore.social/settings/instances/add",
		"scopes":        "read write follow push",
		"website":       "https://pinafore.social"}
	result, err := createOauthClient(s, clientHash)
	assert.NoError(t, err)
	cid := result["client_id"].(string)

	ag := oauth2.NewAccessGenerate()
	createdAt := time.Now()
	token, _, err := ag.Token(context.Background(), cid, "1", createdAt, false)
	assert.NoError(t, err)
	ti := &model.OauthToken{
		ClientId:        cid,
		UserId:          1,
		RedirectUri:     "https://example.com/oauth-client/add",
		Scope:           "read write follow push",
		Access:          token,
		AccessCreatedAt: createdAt,
		AccessExpiresIn: 2 * time.Hour,
		CreatedAt:       createdAt,
	}
	store := &public.SqliteOauthStore{DB: s.DB()}
	err = store.Create(context.Background(), ti)
	assert.NoError(t, err)
	if err != nil {
		return "", err
	}
	return token, nil
}
