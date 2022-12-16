package mastapi

import (
	"net/http"

	"github.com/contribsys/sparq"
)

func timelineHandler(svr sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}
}
