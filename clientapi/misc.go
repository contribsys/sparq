package clientapi

import (
	"fmt"
	"net/http"

	"github.com/contribsys/sparq"
)

func emptyHandler(svr sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			err := fmt.Errorf("%s %s not implemented", r.Method, r.URL.Path)
			httpError(w, err, http.StatusBadRequest)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}
}
