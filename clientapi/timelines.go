package clientapi

import (
	"net/http"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/model"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func listHandler(svr sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// listname := mux.Vars(r)["name"]
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}
}

func homeHandler(svr sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}
}

func publicHandler(svr sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}
}

func TootsForHome(db *sqlx.DB) (map[string]interface{}, error) {
	attrs := map[string]interface{}{}
	base := `select t.sid as id, t.CreatedAt as created_at, t.Summary as spoiler_text, t.Visibility as viz, t.Lang as language,
	        t.URI as uri, t.URI as url, 0 as replies_count, 0 as reblogs_count, 0 as favourites_count, false as favourited,
					false as reblogged, false as muted, false as bookmarked, t.Content as content, null as reblog,
					null as media_attachments, null as mentions, null as tags, null as emojis, null as card, null as poll,
					oc.name as app_name, oc.website as app_website
					from toots t
					left outer join oauth_clients oc on t.appid = oc.id
					where t.sid = ?`
	err := db.QueryRowx(base, nil).MapScan(attrs)
	if err != nil {
		return nil, errors.Wrap(err, "Error with toot "+"")
	}

	attrs["visibility"] = model.FromVis(model.PostVisibility(attrs["viz"].(int64)))
	delete(attrs, "viz")

	attrs["application"] = map[string]any{
		"name":    attrs["app_name"],
		"website": attrs["app_website"],
	}
	delete(attrs, "app_name")
	delete(attrs, "app_website")
	return attrs, nil
}
