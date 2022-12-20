package clientapi

import (
	"net/http"

	"github.com/contribsys/sparq"
)

func statusHandler(svr sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			err := r.ParseForm()
			if err != nil {
				httpError(w, err, http.StatusBadRequest)
				return
			}
		}
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}
}
