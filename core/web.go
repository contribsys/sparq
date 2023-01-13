package core

import (
	"net/http"
	"time"

	"github.com/contribsys/sparq/clientapi"
	"github.com/contribsys/sparq/web"
	"github.com/contribsys/sparq/web/public"
	"github.com/contribsys/sparq/wellknown"
)

func BuildWeb(s *Service) *http.Server {
	root := web.RootRouter(s)
	web.IntegrateOauth(s, root)
	apiv1 := root.PathPrefix("/api/v1").Subrouter()
	clientapi.AddPublicEndpoints(s, apiv1)
	public.AddPublicEndpoints(s, root)
	// s.FaktoryUI.Embed(root, "/faktory")
	// s.AdminUI.Embed(root, "/admin")
	wellknown.AddPublicEndpoints(s, root)

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
