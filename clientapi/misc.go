package clientapi

import (
	"fmt"
	"net/http"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/util"
)

func emptyHandler(svr sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			msg := fmt.Sprintf("%s not implemented", r.URL.Path)
			util.Warnf(msg)
			http.Error(w, msg, 400)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}
}
