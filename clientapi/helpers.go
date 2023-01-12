package clientapi

import (
	"encoding/json"
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

func httpJsonResponse(w http.ResponseWriter, body map[string]interface{}, code int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	_ = enc.Encode(body)
}
