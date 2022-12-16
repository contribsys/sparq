package wellknown

import (
	"context"
	"database/sql"
	"fmt"
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
	err := db.Get(&r, `select Nick from accounts where lower(Nick) = ?`,
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

		resp.Header().Add("Access-Control-Allow-Origin", "*")
		resp.Header().Add("Access-Control-Allow-Headers", "*")
		resp.Header().Add("Access-Control-Allow-Methods", "GET")
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
	mux.HandleFunc("/.well-known/nodeinfo", func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Add("Content-Type", "application/json")
		resp.Header().Add("Access-Control-Allow-Origin", "*")
		resp.Header().Add("Access-Control-Allow-Headers", "*")
		resp.Header().Add("Access-Control-Allow-Methods", "GET")
		resp.WriteHeader(200)
		_, _ = resp.Write([]byte(`{"links":[{"rel":"http://nodeinfo.diaspora.software/ns/schema/2.1","href":"https://` + db.InstanceHostname + `/nodeinfo/2.1"}]}`))
	})
	mux.HandleFunc("/nodeinfo/2.1", func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Add("Content-Type", "application/json")
		resp.Header().Add("Cache-Control", "public, max-age=600")

		var userCount int
		err := db.Database().QueryRow("select count(*) from accounts").Scan(&userCount)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = resp.Write([]byte(fmt.Sprintf(`{
			"version":"2.1",
			"software": {
				"name":"sparq","version":"0.0.1",
				"repository":"tbd","homepage":"tbd",
			},
			"protocols": ["activitypub"],
			"services": {"outbound":[],"inbound":[]},
			"usage": {
				"users": {"total": %d,"activeMonth":0,"activeHalfyear":0},
				"localPosts": 0,
				"localComments": 0
			},
			"openRegistrations": true,
			"metadata": {}}`, userCount)))
	})
}
