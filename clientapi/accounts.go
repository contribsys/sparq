package clientapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"text/template"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/model"
	"github.com/contribsys/sparq/util"
	"github.com/contribsys/sparq/web"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
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

func verifyCredentialsHandler(s sparq.Server) http.HandlerFunc {
	x := template.New("accountCredential")
	x.Funcs(map[string]any{"rfc3339": util.Thens})
	accountCredentialTemplate := template.Must(x.Parse(accountCredentialText))

	return func(w http.ResponseWriter, r *http.Request) {
		webctx := web.Ctx(r)
		token := webctx.BearerCode
		if token == "" {
			httpError(w, errors.New("No access token supplied"), 401)
			return
		}

		var acct model.Account
		err := s.DB().Get(&acct, `
		  select a.*, ap.* from accounts a
		  join account_profiles ap on a.id = ap.accountid 
		  join oauth_tokens ot on a.id = ot.accountid 
			where ot.access = ?`, token)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				httpJsonResponse(w, map[string]interface{}{"error": "Token not found, please re-authenticate"}, http.StatusUnauthorized)
				return
			}
			httpError(w, err, http.StatusInternalServerError)
			return
		}

		// render JSON template
		w.Header().Add("Content-Type", "application/json")
		err = accountCredentialTemplate.Execute(w, &acct)
		if err != nil {
			util.Error("Unable to execute template", err)
			httpError(w, err, http.StatusInternalServerError)
			return
		}
	}
}

var (
	accountCredentialText = `
{
  "id": "{{.Id}}",
  "username": "{{.Nick}}",
  "acct": "{{.Nick}}",
  "display_name": "{{.FullName}}",
  "locked": false,
  "bot": false,
  "created_at": "{{.Created}}",
  "note": "",
  "url": "{{.URI}}",
  "avatar": "{{.Avatar}}",
  "avatar_static": "{{.Avatar}}",
  "header": "{{.Header}}",
  "header_static": "{{.Header}}",
  "followers_count": 0,
  "following_count": 0,
  "statuses_count": 0,
  "last_status_at": "{{.Created}}",
  "source": {
    "privacy": "public",
    "sensitive": false,
    "language": "en",
    "note": "",
    "fields": [],
    "follow_requests_count": 0
  },
  "emojis": [],
  "fields": []
}
`
)

func getAccountStatuses(s sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sfid := mux.Vars(r)["sfid"]
		rows, err := s.DB().Queryx(`
			select posts.* from posts
			inner join actors on posts.authorid = actors.id
			inner join accounts on accounts.id = actors.userid
			where users.sfid = ?
			order by posts.uri DESC
			limit 50
			`, sfid)
		if err == sql.ErrNoRows {
			httpError(w, err, http.StatusNotFound)
			return
		}
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}

		results := []map[string]any{}
		for rows.Next() {
			rowdata := map[string]interface{}{}
			err = rows.MapScan(rowdata)
			if err != nil {
				httpError(w, err, http.StatusInternalServerError)
				return
			}
			results = append(results, rowdata)
		}

		data, err := json.Marshal(results)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write(data)
	}
}

func getAccount(s sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sfid := mux.Vars(r)["sfid"]
		// fmt.Printf("Hello %s\n", sfid)

		userdata := map[string]interface{}{}
		err := s.DB().QueryRowx("select * from users where sfid = ?", sfid).MapScan(userdata)
		if err == sql.ErrNoRows {
			httpError(w, err, http.StatusNotFound)
			return
		}
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(userdata)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write(data)
	}
}
