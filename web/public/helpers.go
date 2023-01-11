package public

import (
	"embed"
	"net/http"

	"github.com/contribsys/sparq/web"
)

var (
	//go:embed static/*.css static/*.js static/*.png static/*.jpg
	staticFiles embed.FS
)

func httpError(w http.ResponseWriter, err error, code int) {
	web.HttpError(w, err, code)
}
