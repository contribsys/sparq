package adminui

import (
	"embed"
	"net/http"

	"strings"
	"time"

	"github.com/contribsys/sparq"
	"github.com/gorilla/mux"
	"github.com/justinas/nosurf"
)

//go:generate ego .

type Tab struct {
	Name string
	Path string
}

var (
	DefaultTabs = []Tab{
		{"Home", "/"},
		{"Busy", "/busy"},
		{"Queues", "/queues"},
		{"Retries", "/retries"},
		{"Scheduled", "/scheduled"},
		{"Dead", "/morgue"},
	}

	//go:embed static/*.css static/*.js static/img/*
	staticFiles embed.FS

	//go:embed static/locales/*
	localeFiles embed.FS

	staticHandler = cache(http.FileServer(http.FS(staticFiles)))
)

type localeMap map[string]map[string]string
type assetLookup func(string) ([]byte, error)

var (
	AssetLookups = []assetLookup{
		localeFiles.ReadFile,
	}
	locales = localeMap{}
)

func init() {
	files, err := localeFiles.ReadDir("static/locales")
	if err != nil {
		panic(err)
	}
	for idx := range files {
		name := strings.Split(files[idx].Name(), ".")[0]
		locales[name] = nil
	}
	// util.Debugf("Initialized %d locales", len(files))
}

type WebUI struct {
	sparq.Pusher
	StartedAt   time.Time
	Binding     string
	enabledCSRF bool
}

func NewWeb(p sparq.Pusher, binding string) *WebUI {
	ui := &WebUI{
		Pusher:      p,
		Binding:     binding,
		StartedAt:   time.Now(),
		enabledCSRF: true,
	}
	return ui
}

func (ui *WebUI) Embed(root *mux.Router, prefix string) *mux.Router {
	app := root
	if prefix != "" {
		app = root.PathPrefix(prefix).Subrouter()
	}
	app.PathPrefix("/static/").Handler(http.StripPrefix(prefix, staticHandler))
	app.HandleFunc("/", Log(ui, func(w http.ResponseWriter, r *http.Request) {
		job := NewJob("atype", "high", "Bob")
		err := ctx(r).Pusher().Push(r.Context(), job)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte("Enqueued job!"))
	}))

	// app.HandleFunc("/", Log(ui, GetOnly(indexHandler)))
	return root
}

// func Layout(w io.Writer, req *http.Request, yield func()) {
// ego_layout(w, req, yield)
// }

/////////////////////////////////////

// The stats handler is hit a lot and adds much noise to the log,
// quiet it down.
func DebugLog(ui *WebUI, pass http.HandlerFunc) http.HandlerFunc {
	return setup(ui, pass, true)
}

func Log(ui *WebUI, pass http.HandlerFunc) http.HandlerFunc {
	return protect(ui.enabledCSRF, setup(ui, pass, false))
}

func setup(ui *WebUI, pass http.HandlerFunc, debug bool) http.HandlerFunc {
	genericSetup := func(w http.ResponseWriter, r *http.Request) {
		dctx := NewContext(ui, r, w)
		pass(w, r.WithContext(dctx))
	}
	return genericSetup
}

/*
func basicAuth(pwd string, pass http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Faktory"`)
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}
		if subtle.ConstantTimeCompare([]byte(password), []byte(pwd)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Faktory"`)
			http.Error(w, "Authorization failed", http.StatusUnauthorized)
			return
		}
		pass(w, r)
	}
}
*/

func GetOnly(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			h(w, r)
			return
		}
		http.Error(w, "get only", http.StatusMethodNotAllowed)
	}
}

func PostOnly(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			h(w, r)
			return
		}
		http.Error(w, "post only", http.StatusMethodNotAllowed)
	}
}

func cache(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "public, max-age=3600")
		h.ServeHTTP(w, r)
	}
}

func protect(enabled bool, h http.HandlerFunc) http.HandlerFunc {
	hndlr := nosurf.New(h)
	hndlr.ExemptFunc(func(r *http.Request) bool {
		return !enabled
	})
	return func(w http.ResponseWriter, r *http.Request) {
		hndlr.ServeHTTP(w, r)
	}
}
