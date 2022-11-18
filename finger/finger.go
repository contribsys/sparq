package finger

import (
	"context"
	"database/sql"
	"html/template"
	"net/http"
	"strings"

	"github.com/contribsys/sparq/db"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var (
	ErrNotFound = errors.New("User not found")
)

type actor struct {
	Nick      string `db:"Nick"`
	PublicKey string `db:"PublicKey"`
}

func (a *actor) Inbox() string {
	return db.InstanceHostname + "/@" + a.Nick + "/inbox"
}

func (a *actor) Domain() string {
	return db.InstanceHostname
}

func (a *actor) URI() string {
	return db.InstanceHostname + "/@" + a.Nick
}

func lookup(ctx context.Context, db *sqlx.DB, username, host string) (*actor, error) {
	// user := map[string]any{}
	var a actor
	err := db.Get(&a, `
		select Nick, PublicKey from users join user_securities on users.Id = user_securities.UserId where lower(Nick) = ?`,
		strings.ToLower(username))
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "sql")
	}
	return &a, nil
}

var (
	fingerResponseText = `
{
	"@context": ["https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"],
	"id": "https://{{.URI}}",
	"type": "Person",
	"preferredUsername": "{{.Nick}}",
	"inbox": "https://{{.Inbox}}",
	"publicKey": {
		"id": "https://{{.URI}}#main-key",
		"owner": "https://{{.URI}}",
		"publicKeyPem": "{{.PublicKey}}"
	}
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

func HttpHandler(db *sqlx.DB, hostname string) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		username := req.URL.Query().Get("resource")
		if username == "" || !strings.HasSuffix(username, "@"+hostname) {
			http.Error(resp, "Invalid input", http.StatusBadRequest)
			return
		}

		parts := strings.Split(username, "@")

		result, err := lookup(req.Context(), db, parts[0], parts[1])
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}

		// fmt.Println(result)

		err = fingerResponseTemplate.Execute(resp, result)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
