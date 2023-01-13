package wellknown

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/db"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestFinger(t *testing.T) {
	ts, stopper := testServer(t, "status")
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

func testServer(t *testing.T, name string) (sparq.Server, func()) {
	dbx, stopper, err := db.TestDB(name)
	if err != nil {
		t.Fatal(err)
	}
	return &testSvr{db: dbx}, stopper
}

type testSvr struct {
	db *sqlx.DB
}

func (ts *testSvr) DB() *sqlx.DB {
	return ts.db
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
