package core

import (
	"net/http"
	"time"

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
	// apiv1 := root.PathPrefix("/api/v1").Subrouter()
	// mastapi.AddPublicEndpoints(apiv1)
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

func Log(pass http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		pass.ServeHTTP(w, r)
		util.Infof("%s %s %v", r.Method, r.RequestURI, time.Since(start))
	})
}

func DebugLog(pass http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass.ServeHTTP(w, r)
		util.Infof("\033[31m404\033[0m %s %s", r.Method, r.RequestURI)
		// util.Infof("[404] %s %s %+v", r.Method, r.RequestURI, r.Header)
	})
}
