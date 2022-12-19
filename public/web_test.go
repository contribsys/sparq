package public

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/db"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestPublicStatic(t *testing.T) {
	stopper, err := db.TestDB("public")
	assert.NoError(t, err)
	defer stopper()

	req := httptest.NewRequest("GET", "http://localhost.dev:9494/static/logo-sm.png", nil)
	w := httptest.NewRecorder()
	r := mux.NewRouter()
	svr := &testSvr{}
	AddPublicEndpoints(svr, r)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "image/png")
}

func TestPublicWeb(t *testing.T) {
	stopper, err := db.TestDB("public")
	assert.NoError(t, err)
	defer stopper()

	withQuery("/users/nosuch", func(w *httptest.ResponseRecorder, req *http.Request) {
		assert.Equal(t, 404, w.Code)
		assert.Contains(t, w.Body.String(), "Not found")
	})
	withQuery("/users/admin", func(w *httptest.ResponseRecorder, req *http.Request) {
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "END PUBLIC KEY")
	})
	withQuery("/home", func(w *httptest.ResponseRecorder, req *http.Request) {
		assert.Equal(t, 302, w.Code)
		assert.Contains(t, w.Body.String(), "/login")
	})
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
	store := &SqliteOauthStore{DB: s.DB()}
	root.Use(BearerAuth(store))
	return root
}
