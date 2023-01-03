package clientapi

import (
	"net/http"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/public"
	"github.com/contribsys/sparq/webutil"
	"github.com/gorilla/mux"
)

func rootRouter(s sparq.Server) *mux.Router {
	root := webutil.RootRouter(s)
	store := &public.SqliteOauthStore{DB: s.DB()}
	root.Use(public.Auth(store))
	return root
}

func httpError(w http.ResponseWriter, err error, code int) {
	webutil.HttpError(w, err, code)
}
