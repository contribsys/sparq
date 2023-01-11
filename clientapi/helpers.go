package clientapi

import (
	"net/http"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/web"
	"github.com/gorilla/mux"
)

func rootRouter(s sparq.Server) *mux.Router {
	root := web.RootRouter(s)
	store := &web.SqliteOauthStore{DB: s.DB()}
	root.Use(web.Auth(store))
	return root
}

func httpError(w http.ResponseWriter, err error, code int) {
	web.HttpError(w, err, code)
}
