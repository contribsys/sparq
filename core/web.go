package core

import (
	"fmt"
	"net/http"
	"time"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/finger"
	"github.com/contribsys/sparq/mastapi"
	"github.com/contribsys/sparq/public"
	"github.com/contribsys/sparq/util"
	"github.com/gorilla/mux"
)

func buildServer(s *Service) *http.Server {
	root := mux.NewRouter()
	if s.Options.LogLevel == "debug" {
		root.NotFoundHandler = DebugLog(http.NotFoundHandler())
	}
	root.Use(Log)
	root.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(fmt.Sprintf("Welcome to Sparq %s!", sparq.Version)))
	})
	finger.AddPublicEndpoints(root)
	apiv1 := root.PathPrefix("/api/v1").Subrouter()
	mastapi.AddPublicEndpoints(apiv1)
	public.AddPublicEndpoints(root)
	root.Handle("/faktory/", s.FaktoryUI.App)
	root.Handle("/adminui/", s.AdminUI.App)

	ht := &http.Server{
		Addr:           s.Binding,
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 16,
		Handler:        root,
	}
	return ht
}

type ctxKeyType int

var (
	ctxKey ctxKeyType = 1
)

type WebCtx struct {
	// db *sqlx.DB
}

func Context(r *http.Request) *WebCtx {
	return r.Context().Value(ctxKey).(*WebCtx)
}

func Log(pass http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// req := r.WithContext(context.WithValue(r.Context(), ctxKey, &WebCtx{}))
		start := time.Now()
		pass.ServeHTTP(w, r)
		util.Infof("%s %s %v", r.Method, r.RequestURI, time.Since(start))
	})
}

func DebugLog(pass http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass.ServeHTTP(w, r)
		util.Infof("[404] %s %s %+v", r.Method, r.RequestURI, r.Header)
	})
}
