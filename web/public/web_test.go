package public

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/db"
	"github.com/contribsys/sparq/web"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestPublicStatic(t *testing.T) {
	stopper, err := db.TestDB("public")
	assert.NoError(t, err)
	defer stopper()

	r := mux.NewRouter()
	svr := &testSvr{}
	AddPublicEndpoints(svr, r)

	req := httptest.NewRequest("GET", "http://localhost.dev:9494/static/logo-sm.png", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "image/png")

	req = httptest.NewRequest("GET", "http://localhost.dev:9494/users/nosuch", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
	assert.Contains(t, w.Body.String(), "Not found")

	req = httptest.NewRequest("GET", "http://localhost.dev:9494/users/admin", nil)
	req.Header.Set("Accept", "application/activity+json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "END PUBLIC KEY")

	req = httptest.NewRequest("GET", "http://localhost.dev:9494/home", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 302, w.Code)
	assert.Contains(t, w.Body.String(), "/login")
}

func TestPublicLogin(t *testing.T) {
	stopper, err := db.TestDB("public")
	assert.NoError(t, err)
	defer stopper()
	ts := &testSvr{}
	root := rootRouter(ts)
	AddPublicEndpoints(ts, root)

	t.Run("login", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://localhost.dev:9494/login", nil)
		w := httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)

		req = httptest.NewRequest("POST", "http://localhost.dev:9494/login", nil)
		w = httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid username")

		payload := url.Values{
			"username": {"admin"},
			"password": {"sparq123"},
		}
		req = httptest.NewRequest("POST", "http://localhost.dev:9494/login", strings.NewReader(payload.Encode()))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w = httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 302, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "/home")
	})
}

func withQuery(query string, fn func(w *httptest.ResponseRecorder, req *http.Request)) {
	req := httptest.NewRequest("GET", "http://localhost.dev:9494"+query, nil)
	w := httptest.NewRecorder()
	r := mux.NewRouter()
	s := &testSvr{}
	AddPublicEndpoints(s, r)
	r.ServeHTTP(w, req)
	fn(w, req)
}

func rootRouter(s sparq.Server) *mux.Router {
	root := mux.NewRouter()
	store := &web.SqliteOauthStore{DB: s.DB()}
	root.Use(web.Auth(store))
	return root
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

func (ts *testSvr) Context() context.Context {
	return context.Background()
}
