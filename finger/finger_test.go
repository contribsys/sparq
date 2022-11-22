package finger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/contribsys/sparq/db"
	"github.com/stretchr/testify/assert"
)

func TestFinger(t *testing.T) {
	stopper, err := db.TestDB("finger")
	assert.NoError(t, err)
	defer stopper()

	fn := webfingerHandler(db.Database())

	withQuery("", func(w *httptest.ResponseRecorder, req *http.Request) {
		fn(w, req)
		assert.Equal(t, 400, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid input")
	})
	withQuery("?resource=acct:admin@localhost.dev", func(w *httptest.ResponseRecorder, req *http.Request) {
		fn(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "https://localhost.dev/users/admin")
	})
}

func withQuery(query string, fn func(w *httptest.ResponseRecorder, req *http.Request)) {
	req := httptest.NewRequest("GET", "http://localhost.dev:9494/.well-known/webfinger"+query, nil)
	w := httptest.NewRecorder()
	fn(w, req)
}
