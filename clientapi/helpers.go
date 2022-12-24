package clientapi

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/public"
	"github.com/contribsys/sparq/util"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func httpError(w http.ResponseWriter, err error, code int) {
	er := errors.Wrap(err, "Unexpected HTTP error")
	var build strings.Builder
	build.WriteString(er.Error())
	for _, f := range er.(stackTracer).StackTrace() {
		build.WriteString(fmt.Sprintf("\n%+v", f))
	}
	util.Infof(build.String())
	http.Error(w, err.Error(), code)
}

func rootRouter(s sparq.Server) *mux.Router {
	root := mux.NewRouter()
	root.NotFoundHandler = DebugLog(http.NotFoundHandler())
	root.Use(DebugLog)
	root.Use(Cors)
	store := &public.SqliteOauthStore{DB: s.DB()}
	root.Use(public.Auth(store))
	return root
}

func Cors(pass http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST, PUT, DELETE, GET, PATCH, OPTIONS")
		w.Header().Add("Access-Control-Allow-Headers", "*")
		w.Header().Add("Cache-Control", "public, max-age=3600")
		if r.Method == "OPTIONS" {
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
