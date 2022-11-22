package finger

import (
	"context"
	"database/sql"
	"html/template"
	"net/http"
	"strings"

	"github.com/contribsys/sparq/db"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var (
	ErrNotFound = errors.New("User not found")
)

type result struct {
	Nick string `db:"Nick"`
	Acct string
}

func (r *result) URI() string {
	return "https://" + db.InstanceHostname + "/users/" + r.Nick
}

func fingerLookup(ctx context.Context, db *sqlx.DB, username, host string) (*result, error) {
	// user := map[string]any{}
	var r result
	err := db.Get(&r, `select Nick from users where lower(Nick) = ?`,
		strings.ToLower(username))
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "sql")
	}
	r.Acct = username + "@" + host

	return &r, nil
}

var (
	fingerResponseText = `{
	"subject": "acct:{{.Acct}}",
	"links": [{
		"rel": "self",
		"type": "application/activity+json",
		"href": "{{.URI}}"
	}]
}`

	/*
	   {
	   	"subject":"acct:getajobmike@ruby.social",
	   	"aliases":["https://ruby.social/@getajobmike",
	   	           "https://ruby.social/users/getajobmike"],
	   	"links":[
	   		{"rel":"http://webfinger.net/rel/profile-page","type":"text/html","href":"https://ruby.social/@getajobmike"},
	   		{"rel":"self","type":"application/activity+json","href":"https://ruby.social/users/getajobmike"},
	   		{"rel":"http://ostatus.org/schema/1.0/subscribe","template":"https://ruby.social/authorize_interaction?uri={uri}"}
	   	]
	   }
	*/
	fingerResponseTemplate = template.Must(template.New("fingerResponse").Parse(fingerResponseText))
)

func webfingerHandler(dbx *sqlx.DB) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		username := req.URL.Query().Get("resource")
		if username == "" ||
			!strings.HasPrefix(username, "acct:") ||
			!strings.HasSuffix(username, "@"+db.InstanceHostname) {
			http.Error(resp, "Invalid input", http.StatusBadRequest)
			return
		}

		parts := strings.Split(username[5:], "@")

		result, err := fingerLookup(req.Context(), dbx, parts[0], parts[1])
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}

		resp.Header().Add("Content-Type", "application/jrd+json")
		err = fingerResponseTemplate.Execute(resp, result)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func AddPublicEndpoints(mux *mux.Router) {
	mux.HandleFunc("/.well-known/webfinger", webfingerHandler(db.Database()))
}
