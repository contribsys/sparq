package wellknown

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/contribsys/sparq/web"
	"github.com/stretchr/testify/assert"
)

func TestFinger(t *testing.T) {
	ts, stopper := web.NewTestServer(t, "status")
	defer stopper()

	fn := webfingerHandler(ts.DB())

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
