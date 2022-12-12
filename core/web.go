package core

import (
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/mastapi"
	"github.com/contribsys/sparq/public"
	"github.com/contribsys/sparq/util"
	"github.com/contribsys/sparq/wellknown"
	"github.com/gorilla/mux"
)

func buildServer(s *Service) *http.Server {
	root := mux.NewRouter()
	if s.Options.LogLevel == "debug" {
		root.NotFoundHandler = DebugLog(http.NotFoundHandler())
	}
	root.Use(DebugLog)
	root.Use(Cors)
	apiv1 := root.PathPrefix("/api/v1").Subrouter()
	apiv1.Use(Cors)
	mastapi.AddPublicEndpoints(s, apiv1)

	public.AddPublicEndpoints(s, root)

	public.IntegrateOauth(s, root)
	// s.FaktoryUI.Embed(root, "/faktory")
	// s.AdminUI.Embed(root, "/admin")
	wellknown.AddPublicEndpoints(root)

	ht := &http.Server{
		Addr:           s.Binding,
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   10 * time.Second,
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
		}
		pass.ServeHTTP(w, r)
	})
}

func Log(pass http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w.Header().Add("Server", sparq.ServerHeader)
		pass.ServeHTTP(w, r)
		util.Infof("%s %s %v", r.Method, r.RequestURI, time.Since(start))
	})
}

func DebugLog(pass http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := httputil.DumpRequest(r, r.Method == "POST")
		os.Stdout.Write(data)
		if r.Method == "POST" {
			os.Stdout.WriteString("\n\n")
		}
		w.Header().Add("Server", sparq.ServerHeader)

		pass.ServeHTTP(w, r)
	})
}
