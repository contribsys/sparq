package web

import (
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/util"
	"github.com/gorilla/mux"
)

func RootRouter(s sparq.Server) *mux.Router {
	root := mux.NewRouter()
	root.NotFoundHandler = DebugLog(http.NotFoundHandler())
	root.Use(DebugLog)
	root.Use(Cors)
	root.Use(EstablishContext(s))
	return root
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
		start := time.Now()
		if util.LogDebug {
			if r.Method == "POST" {
				err := r.ParseForm()
				if err != nil {
					util.Warnf("Unable to parse POST: %v", err)
					http.Error(w, err.Error(), 400)
					return
				}
			}
			data, _ := httputil.DumpRequest(r, false) // r.Method == "POST")
			os.Stdout.Write(data)
			if r.Method == "POST" {
				os.Stdout.WriteString("\n\n")
			}
		}
		w.Header().Add("Server", sparq.ServerHeader)

		pass.ServeHTTP(w, r)
		util.Infof("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}
