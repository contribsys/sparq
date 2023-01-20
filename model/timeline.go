package model

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type TimelineQuery struct {
	// these parameters match Mastodon's API parameters
	MinId     string
	MaxId     string
	SinceId   string
	Limit     uint64
	ListId    uint64
	Local     bool
	Remote    bool
	OnlyMedia bool

	Visibility PostVisibility

	db *sqlx.DB
}

type QueryResult struct {
	Toots []*Entry
}

func (qr *QueryResult) IsEmpty() bool {
	return len(qr.Toots) == 0
}

type Entry struct {
	*Toot
	db *sqlx.DB
}

func association[V any](rows *sqlx.Rows, err error, recvr V) ([]*V, error) {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*V{}, nil
		}
		return nil, err
	}
	records := make([]*V, 0)
	for rows.Next() {
		var record V
		err := rows.StructScan(&record)
		if err != nil {
			return nil, err
		}
		records = append(records, &record)
	}
	return records, nil
}

func (e *Entry) Tags() ([]*TootTag, error) {
	rows, err := e.db.Queryx("select * from toot_tags where sid = ?", e.Toot.Sid)
	return association(rows, err, TootTag{})
}

func (e *Entry) MediaAttachments() ([]*TootMedia, error) {
	rows, err := e.db.Queryx("select * from toot_medias where sid = ?", e.Toot.Sid)
	return association(rows, err, TootMedia{})
}

func TQ(db *sqlx.DB) *TimelineQuery {
	return &TimelineQuery{
		Limit: 20,
		db:    db,
	}
}

func (tq *TimelineQuery) Execute() (*QueryResult, error) {
	if tq.Limit > 50 {
		tq.Limit = 50
	}
	base := squirrel.Select(`t.*`).From("toots t").
		JoinClause("LEFT OUTER JOIN oauth_clients oc on t.appid = oc.id").
		Where("t.visibility = ?", tq.Visibility).
		Limit(tq.Limit)

	if tq.MinId != "" && tq.MaxId != "" {
		base = base.Where("t.sid between ? and ?", tq.MinId)
	} else if tq.MinId != "" {
		base = base.Where("t.sid > ?", tq.MinId)
	} else if tq.MaxId != "" {
		base = base.Where("t.sid <= ?", tq.MaxId)
	} else if tq.SinceId != "" {
		base = base.Where("t.sid > ?", tq.SinceId)
	}
	if tq.OnlyMedia {
		base = base.Join("toot_medias tm on t.sid = tm.sid")
	}
	if tq.ListId != 0 {
		// TODO
	}
	if tq.Local || tq.Remote {
		if tq.Local {
			base = base.Where("t.authorId is not null")
		} else {
			base = base.Where("t.authorId is null")
		}
	}

	base = base.OrderBy("t.CreatedAt DESC")
	sqlq, args, err := base.ToSql()
	fmt.Println(sqlq)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid timeline query")
	}
	entries := make([]*Entry, 0)
	rows, err := tq.db.Queryx(sqlq, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &QueryResult{Toots: entries}, nil
		}
		return nil, errors.Wrap(err, "Bad timeline query")
	}
	for rows.Next() {
		toot := Toot{}
		err := rows.StructScan(&toot)
		if err != nil {
			return nil, errors.Wrap(err, "Toot query")
		}
		entries = append(entries, &Entry{Toot: &toot, db: tq.db})
	}
	return &QueryResult{Toots: entries}, nil
}
