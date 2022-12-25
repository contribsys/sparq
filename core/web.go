package core

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/clientapi"
	"github.com/contribsys/sparq/public"
	"github.com/contribsys/sparq/util"
	"github.com/contribsys/sparq/wellknown"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func rootRouter(s sparq.Server) *mux.Router {
	root := mux.NewRouter()
	root.NotFoundHandler = DebugLog(http.NotFoundHandler())
	root.Use(DebugLog)
	root.Use(Cors)
	return root
}

func BuildWeb(s *Service) *http.Server {
	root := rootRouter(s)
	public.IntegrateOauth(s, root)
	apiv1 := root.PathPrefix("/api/v1").Subrouter()
	clientapi.AddPublicEndpoints(s, apiv1)
	public.AddPublicEndpoints(s, root)
	// s.FaktoryUI.Embed(root, "/faktory")
	// s.AdminUI.Embed(root, "/admin")
	wellknown.AddPublicEndpoints(root)

	ht := &http.Server{
		Addr:        s.Binding,
		ReadTimeout: 5 * time.Second,

		// this timeout affects streaming sockets,
		// will need to reconnect every 5 minutes
		WriteTimeout:   300 * time.Second,
		MaxHeaderBytes: 1 << 16,
		Handler:        root,
	}
	return ht
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
			data, _ := httputil.DumpRequest(r, r.Method == "POST")
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
