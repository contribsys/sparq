package mastapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/db"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// POST https://mastodon.example/api/v1/accounts
// GET https://mastodon.example/api/v1/accounts/verify_credentials
// PATCH https://mastodon.example/api/v1/accounts/update_credentials
// GET https://mastodon.example/api/v1/accounts/:id
// GET https://mastodon.example/api/v1/accounts/:id/statuses
// GET https://mastodon.example/api/v1/accounts/:id/followers
// GET https://mastodon.example/api/v1/accounts/:id/following
// POST https://mastodon.example/api/v1/accounts/:id/follow
// POST https://mastodon.example/api/v1/accounts/:id/unfollow
// POST https://mastodon.example/api/v1/accounts/:id/block
// POST https://mastodon.example/api/v1/accounts/:id/unblock
// POST https://mastodon.example/api/v1/accounts/:id/mute
// POST https://mastodon.example/api/v1/accounts/:id/unmute
// GET https://mastodon.example/api/v1/accounts/relationships
// GET https://mastodon.example/api/v1/accounts/search

func AddPublicEndpoints(s sparq.Server, mux *mux.Router) {
	mux.HandleFunc("/custom_emojis", emptyHandler(s))
	mux.HandleFunc("/lists", emptyHandler(s))
	mux.HandleFunc("/filters", emptyHandler(s))
	mux.HandleFunc("/instance", instanceHandler(s))
	mux.HandleFunc("/timelines/{type}", timelineHandler(s))
	mux.HandleFunc("/statuses", statusHandler(s))
	mux.HandleFunc("/accounts/verify_credentials", verifyCredentialsHandler(s))
	mux.HandleFunc("/accounts/{sfid:[0-9]+}", getAccount)
	mux.HandleFunc("/accounts/{sfid:[0-9]+}/statuses", getAccountStatuses)
	// mux.HandleFunc("/accounts/{sfid:[0-9]+}/followers", getAccountFollowers)
	// mux.HandleFunc("/accounts/{sfid:[0-9]+}/following", getAccountFollowing)
}

// openssl rand -hex 32
// ruby -rsecurerandom -e "puts SecureRandom.hex(32)"
var sessionStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

func verifyCredentialsHandler(s sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := requireLogin(w, r)
		if uid == 0 {
			return
		}
		// verifyToken
		// lookupUserAccount
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write([]byte(fmt.Sprintf(`{"id": "%d"}`, uid)))
	}
}

func requireLogin(w http.ResponseWriter, r *http.Request) int {
	session, _ := sessionStore.Get(r, "sparq-session")
	val, ok := session.Values["uid"]
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return 0
	}
	uid, err := strconv.Atoi(val.(string))
	if err != nil {
		return 0
	}
	return uid
}

func getAccountStatuses(w http.ResponseWriter, r *http.Request) {
	sfid := mux.Vars(r)["sfid"]
	rows, err := db.Database().Queryx(`
	  select posts.* from posts
		inner join actors on posts.authorid = actors.id
		inner join users on users.id = actors.userid
		where users.sfid = ?
		order by posts.uri DESC
		limit 50
		`, sfid)
	if err == sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	results := []map[string]any{}
	for rows.Next() {
		rowdata := map[string]interface{}{}
		err = rows.MapScan(rowdata)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, rowdata)
	}

	data, err := json.Marshal(results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func getAccount(w http.ResponseWriter, r *http.Request) {
	sfid := mux.Vars(r)["sfid"]
	// fmt.Printf("Hello %s\n", sfid)

	userdata := map[string]interface{}{}
	err := db.Database().QueryRowx("select * from users where sfid = ?", sfid).MapScan(userdata)
	if err == sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(userdata)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(data)
}
