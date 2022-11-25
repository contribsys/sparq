package public

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/contribsys/sparq/db"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestPublicWeb(t *testing.T) {
	stopper, err := db.TestDB("web")
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
	withQuery("/", func(w *httptest.ResponseRecorder, req *http.Request) {
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Timeline")
	})
}

func withQuery(query string, fn func(w *httptest.ResponseRecorder, req *http.Request)) {
	req := httptest.NewRequest("GET", "http://localhost.dev:9494"+query, nil)
	w := httptest.NewRecorder()
	r := mux.NewRouter()
	AddPublicEndpoints(r)
	r.ServeHTTP(w, req)
	fn(w, req)
}
