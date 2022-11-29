package core

import (
	"net/http"
	"time"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/mastapi"
	"github.com/contribsys/sparq/util"
	"github.com/contribsys/sparq/wellknown"
	"github.com/gorilla/mux"
)

func buildServer(s *Service) *http.Server {
	root := mux.NewRouter()
	if s.Options.LogLevel == "debug" {
		root.NotFoundHandler = DebugLog(http.NotFoundHandler())
	}
	root.Use(Log)
	// finger.AddPublicEndpoints(root)
	apiv1 := root.PathPrefix("/api/v1").Subrouter()
	apiv1.Use(Cors)
	mastapi.AddPublicEndpoints(s, apiv1)
	// public.AddPublicEndpoints(root)
	s.FaktoryUI.Embed(root, "/faktory")
	s.AdminUI.Embed(root, "/admin")
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

/*
headers: :any,
credentials: false,
expose: ['Link', 'X-RateLimit-Reset', 'X-RateLimit-Limit', 'X-RateLimit-Remaining', 'X-Request-Id']
*/
func Cors(pass http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "POST, PUT, DELETE, GET, PATCH, OPTIONS")
		w.Header().Add("Access-Control-Allow-Headers", "*")
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
		w.Header().Add("Server", sparq.ServerHeader)
		pass.ServeHTTP(w, r)
		util.DumpRequest(r)
	})
}
