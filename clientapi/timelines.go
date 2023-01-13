package clientapi

import (
	"fmt"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/model"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type TimelineQuery struct {
	min_id     string
	max_id     string
	since_id   string
	limit      uint64
	list_id    uint64
	local      bool
	remote     bool
	only_media bool

	db *sqlx.DB
}

func TQ(db *sqlx.DB) *TimelineQuery {
	return &TimelineQuery{
		limit: 20,
		db:    db,
	}
}

func (tq *TimelineQuery) Execute() ([]*model.Toot, error) {
	if tq.limit > 50 {
		tq.limit = 50
	}
	base := sq.Select(`t.*, oc.*`).From("toots t").
		JoinClause("LEFT OUTER JOIN oauth_clients oc on t.appid = oc.id").
		Limit(tq.limit)
	if tq.min_id != "" && tq.max_id != "" {
		base.Where("t.sid between ? and ?", tq.min_id)
	} else if tq.min_id != "" {
		base.Where("t.sid > ?", tq.min_id)
	} else if tq.max_id != "" {
		base.Where("t.sid <= ?", tq.max_id)
	} else if tq.since_id != "" {
		base.Where("t.sid > ?", tq.since_id)
	}
	if tq.only_media {
		base.LeftJoin("toot_medias tm on t.sid = tm.sid")
	}
	if tq.local || tq.remote {
		if tq.local {
			base.Where("t.account_id is not null")
		} else {
			base.Where("t.account_id is null")
		}
	}
	// TODO: list_id

	base.OrderBy("t.CreatedAt DESC")
	sql, args, err := base.ToSql()
	fmt.Println(sql)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid timeline query")
	}
	results := make([]*model.Toot, 0)
	rows, err := tq.db.Query(sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Bad timeline query")
	}
	for rows.Next() {
		toot := &model.Toot{}
		err := rows.Scan(&toot)
		if err != nil {
			return nil, errors.Wrap(err, "Toot")
		}
		results = append(results, toot)
	}
	return results, nil
}

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
