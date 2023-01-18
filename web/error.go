package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/contribsys/sparq/util"
	"github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func HttpError(w http.ResponseWriter, err error, code int) {
	er := errors.Wrap(err, "Unexpected HTTP error")
	var build strings.Builder
	build.WriteString(er.Error())
	for _, f := range er.(stackTracer).StackTrace() {
		build.WriteString(fmt.Sprintf("\n%+v", f))
	}
	util.Infof(build.String())

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	_ = enc.Encode(map[string]string{"error": err.Error()})
}
