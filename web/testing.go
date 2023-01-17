package web

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/db"
	"github.com/jmoiron/sqlx"
)

func NewTestServer(t *testing.T, name string) (sparq.Server, func()) {
	dbx, stopper, err := db.TestDB(name)
	if err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp("", "sparq-test-*")
	if err != nil {
		panic(err)
	}
	svr := &testSvr{
		db:   dbx,
		root: dir,
	}
	return svr, func() {
		os.RemoveAll(svr.root)
		stopper()
	}
}

type testSvr struct {
	db   *sqlx.DB
	root string
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

func (ts *testSvr) Root() string {
	return ts.root
}

func (ts *testSvr) MediaRoot() string {
	return fmt.Sprintf("%s/media", ts.root)
}

func (ts *testSvr) Context() context.Context {
	return context.Background()
}
