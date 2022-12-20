package clientapi

import (
	"context"
	"net/http"
	"net/http/httputil"
	"os"
	"testing"
	"time"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/model"
	"github.com/contribsys/sparq/oauth2"
	"github.com/contribsys/sparq/public"
	"github.com/contribsys/sparq/util"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func httpError(w http.ResponseWriter, err error, code int) {
	er := errors.Wrap(err, "Unexpected HTTP error")
	util.Infof(er.Error())
	for _, f := range er.(stackTracer).StackTrace() {
		util.Infof("%+s:%d\n", f, f)
	}
	http.Error(w, err.Error(), code)
}

func rootRouter(s sparq.Server) *mux.Router {
	root := mux.NewRouter()
	root.NotFoundHandler = DebugLog(http.NotFoundHandler())
	root.Use(DebugLog)
	root.Use(Cors)
	store := &public.SqliteOauthStore{DB: s.DB()}
	root.Use(public.BearerAuth(store))
	return root
}

// returns the access token or error
func registerToken(t *testing.T, s sparq.Server) (string, error) {
	clientHash := map[string]string{
		"client_name":   "Pinafore",
		"redirect_uris": "https://pinafore.social/settings/instances/add",
		"scopes":        "read write follow push",
		"website":       "https://pinafore.social"}
	result, err := createOauthClient(s, clientHash)
	assert.NoError(t, err)
	cid := result["client_id"].(string)

	ag := oauth2.NewAccessGenerate()
	createdAt := time.Now()
	token, _, err := ag.Token(context.Background(), cid, "1", createdAt, false)
	assert.NoError(t, err)
	ti := &model.OauthToken{
		ClientId:        cid,
		UserId:          1,
		RedirectUri:     "https://example.com/oauth-client/add",
		Scope:           "read write follow push",
		Access:          token,
		AccessCreatedAt: createdAt,
		AccessExpiresIn: 2 * time.Hour,
		CreatedAt:       createdAt,
	}
	store := &public.SqliteOauthStore{DB: s.DB()}
	err = store.Create(context.Background(), ti)
	assert.NoError(t, err)
	if err != nil {
		return "", err
	}
	return token, nil
}

func Cors(pass http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Header().Add("Access-Control-Allow-Methods", "POST, PUT, DELETE, GET, PATCH, OPTIONS")
			w.Header().Add("Access-Control-Allow-Headers", "*")
			w.Header().Add("Cache-Control", "public, max-age=3600")
			w.WriteHeader(204)
			return
		}
		pass.ServeHTTP(w, r)
	})
}

func DebugLog(pass http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			err := r.ParseForm()
			if err != nil {
				httpError(w, err, http.StatusBadRequest)
				return
			}
		}
		data, _ := httputil.DumpRequest(r, r.Method == "POST")
		os.Stdout.Write(data)
		if r.Method == "POST" {
			os.Stdout.WriteString("\n\n")
		}
		w.Header().Add("Server", sparq.ServerHeader)

		pass.ServeHTTP(w, r)
	})
}
