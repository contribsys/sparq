package public

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/contribsys/sparq/activitystreams"
	"github.com/contribsys/sparq/db"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// go:generate ego .

type Tab struct {
	Name string
	Path string
}

var (
	DefaultTabs = []Tab{
		{"Home", "/"},
		{"Local", "/local"},
		{"Federated", "/federated"},
	}
	ErrNotFound   = errors.New("User not found")
	staticHandler = http.FileServer(http.FS(staticFiles))
)

func setCtx(pass http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := r.WithContext(context.WithValue(r.Context(), ctxKey, newCtx(w, r)))
		pass.ServeHTTP(w, req)
	})
}

func AddPublicEndpoints(mux *mux.Router) {
	mux.Use(setCtx)
	mux.Handle("/static", staticHandler)
	mux.HandleFunc("/users/{nick:[a-z0-9]{4,16}}", getUser)
	mux.HandleFunc("/", indexHandler)
	// mux.HandleFunc("/home", homeHandler)
	// mux.HandleFunc("/public/local", localHandler)
	// mux.HandleFunc("/public", publicHandler)
	// mux.HandleFunc("/auth/sign_up", signupHandler)
	// mux.HandleFunc("/auth/sign_in", signinHandler)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	ego_index(w, r)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	nick := mux.Vars(r)["nick"]

	userdata := map[string]interface{}{}
	err := db.Database().QueryRowx(`
	select *
	from users
	inner join user_securities
	on users.id = user_securities.userid
	where users.nick = ?`, nick).MapScan(userdata)
	if err == sql.ErrNoRows {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url := "https://" + db.InstanceHostname + "/users/" + nick
	me := activitystreams.NewPerson(url)
	me.URL = url
	me.Name = userdata["FullName"].(string)
	me.PreferredUsername = userdata["Nick"].(string)
	me.AddPubKey(string(userdata["PublicKey"].([]uint8)))

	data, err := json.Marshal(me)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/activity+json")
	_, _ = w.Write(data)
}
