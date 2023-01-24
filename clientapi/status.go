package clientapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/model"
	"github.com/contribsys/sparq/web"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var (
	dupeDetector = map[string]time.Time{}
	dupeMu       sync.Mutex
	submitCount  int32
)

type Poll struct {
	ExpiresIn      int
	HideTotals     bool
	MultipleChoice bool
	O1             string
	O2             string
	O3             string
	O4             string
	O5             string
	O6             string
}
type Toot struct {
	AuthorId           uint64
	Content            string
	MediaIds           []string
	InReplyTo          *string
	InReplyToAccountId *string
	Sensitive          bool
	Summary            string
	Visibility         string
	LanguageCode       string
	ScheduledAt        string
	*Poll
}

func getTootHandler(svr sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			httpError(w, errors.New("GET only"), http.StatusBadRequest)
			return
		}

		sid := mux.Vars(r)["id"]
		attrs, err := TootMap(svr.DB(), sid)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				httpError(w, err, http.StatusNotFound)
				return
			}
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		err = enc.Encode(attrs)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
	}
}

func PostTootHandler(svr sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			httpError(w, errors.New("POST only"), http.StatusBadRequest)
			return
		}
		err := r.ParseForm()
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}

		ctx := web.Ctx(r)
		if ctx.CurrentUserID == web.Anonymous {
			httpError(w, errors.New("Unauthorized"), http.StatusUnauthorized)
			return
		}
		aid, err := strconv.ParseUint(ctx.CurrentUserID, 10, 64)
		if err != nil {
			httpError(w, err, 503)
			return
		}
		fmt.Printf("%d %+v\n", aid, r.Form)

		toot := &Toot{
			AuthorId:     aid,
			Content:      r.Form.Get("status"),
			Sensitive:    r.Form.Get("sensitive") == "true",
			Summary:      r.Form.Get("spoiler_text"),
			Visibility:   r.Form.Get("visibility"),
			LanguageCode: r.Form.Get("language"),
			ScheduledAt:  r.Form.Get("scheduled_at"),
		}
		rto := r.Form.Get("in_reply_to_id")
		if rto != "" {
			toot.InReplyTo = &rto
		}
		medias := r.Form["media_ids[]"]
		if toot.Content == "" && len(medias) == 0 {
			httpError(w, errors.New("Please enter a message"), 400)
			return
		}

		if len(medias) > 0 {
			// media and poll are mutually exclusive
		} else if r.Form.Get("poll[expires_in]") != "" {
			expy, err := strconv.Atoi(r.Form.Get("poll[expires_in]"))
			if err != nil {
				httpError(w, err, 401)
				return
			}
			p := &Poll{}
			p.ExpiresIn = expy
			p.HideTotals = r.Form.Get("poll[hide_totals]") == "true"
			p.MultipleChoice = r.Form.Get("poll[multiple]") == "true"
			opts := r.Form["poll[options][]"]
			if len(opts) < 2 || len(opts) > 6 {
				httpError(w, errors.New("Polls must have between 2 and 6 options"), 401)
				return
			}
			// ugh this is horrible, is there a cleaner way to convert from an
			// array to named fields?
			p.O1 = opts[0]
			p.O2 = opts[1]
			if len(opts) > 2 {
				p.O3 = opts[2]
			}
			if len(opts) > 3 {
				p.O4 = opts[3]
			}
			if len(opts) > 4 {
				p.O5 = opts[4]
			}
			if len(opts) > 5 {
				p.O6 = opts[5]
			}
			toot.Poll = p
		}

		// verify we're not getting a duplicate submission
		// since we are a single process, we don't need Redis
		ikey := r.Header.Get("Idempotency-Key")
		dupeMu.Lock()
		defer dupeMu.Unlock()
		_, ok := dupeDetector[ikey]
		if ok {
			httpError(w, errors.New("Duplicate toot, ignoring"), 401)
			return
		}
		dupeDetector[ikey] = time.Now()
		defer dupeCleaner()

		post, err := saveToot(svr, r, toot, medias)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}

		sid := post.Sid
		attrs, err := TootMap(svr.DB(), sid)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		err = enc.Encode(attrs)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		fmt.Printf("Created toot: %s\n", post.Uri)
	}
}

func dupeCleaner() {
	// housekeeping
	// every 20th status we'll clear out old idempotency keys
	val := atomic.AddInt32(&submitCount, 1)
	if val%20 == 0 {
		cleanDupeMap()
	}
}

func cleanDupeMap() {
	toDelete := []string{}
	now := time.Now()
	for key, submissionTime := range dupeDetector {
		if now.Sub(submissionTime).Hours() > 1.0 {
			toDelete = append(toDelete, key)
		}
	}
	if len(toDelete) == 0 {
		return
	}
	dupeMu.Lock()
	defer dupeMu.Unlock()
	for idx := range toDelete {
		delete(dupeDetector, toDelete[idx])
	}
}

func saveToot(svr sparq.Server, r *http.Request, toot *Toot, medias []string) (*model.Toot, error) {
	sid := model.Snowflakes.NextSID()
	p := &model.Toot{
		Sid:        sid,
		Uri:        fmt.Sprintf("https://%s/@%s/%s", svr.Hostname(), "admin", sid),
		AccountId:  toot.AuthorId,
		Summary:    toot.Summary,
		Content:    toot.Content,
		Visibility: model.ToVis(toot.Visibility),
		InReplyTo:  toot.InReplyTo,
		CreatedAt:  time.Now(),
	}
	x := web.Ctx(r).ClientApp()
	if x != nil {
		p.AppId = &x.Id
	}
	tx, err := svr.DB().Begin()
	if err != nil {
		return nil, err
	}
	if toot.Poll != nil {
		po := toot.Poll
		res, err := tx.ExecContext(r.Context(), `
			insert into polls (expires_in, multiple, hide, o1, o2, o3, o4, o5, o6) values (?, ?, ?, ?, ?, ?, ?)`,
			po.ExpiresIn, po.MultipleChoice, po.HideTotals,
			po.O1, po.O2, po.O3, po.O4, po.O5, po.O6)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		x, _ := res.LastInsertId()
		y := uint64(x)
		p.PollId = &y
	}
	_, err = tx.ExecContext(r.Context(), `
	  insert into toots (Sid, Uri, InReplyTo, AuthorId, ActorId, PollId, Summary, Content, Lang, Visibility, AppId) values
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Sid, p.Uri, p.InReplyTo, p.AuthorId, p.AuthorId, p.PollId, p.Summary, p.Content, p.Lang, p.Visibility, p.AppId)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if len(medias) > 0 {
		query, args, err := sqlx.In(`update toot_medias set sid = ? where id in (?)`, p.Sid, medias)
		if err != nil {
			_ = tx.Rollback()
			return nil, errors.Wrap(err, "medias")
		}
		query = svr.DB().Rebind(query)
		_, err = tx.ExecContext(r.Context(), query, args...)
		if err != nil {
			_ = tx.Rollback()
			return nil, errors.Wrap(err, "media exec")
		}
	}
	err = saveTags(r.Context(), tx, p)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func saveTags(ctx context.Context, tx *sql.Tx, p *model.Toot) error {
	tags := extractTags(p.Content)
	for _, tag := range tags {
		fmt.Printf("Saving tag for %s: %s\n", p.Sid, tag)
		_, err := tx.ExecContext(ctx, `insert into toot_tags (sid, tag) values (?, ?)`, p.Sid, strings.ToLower(tag))
		if err != nil {
			return err
		}
	}
	return nil
}

var (
	tagRegexp = regexp.MustCompile(`\s*#([[:alpha:]][[:word:]]{1,20})(?:\s|\z)`)
	empty     = []string{}
)

func extractTags(content string) []string {
	result := tagRegexp.FindAllStringSubmatch(content, -1)
	if result == nil {
		return empty
	}
	rc := []string{}
	for _, matches := range result {
		rc = append(rc, matches[1])
	}
	return rc
}

func TootMap(db *sqlx.DB, sid string) (map[string]interface{}, error) {
	attrs := map[string]interface{}{}
	base := `select t.sid as id, t.AuthorId as authorId, t.CreatedAt as created_at, t.Summary as spoiler_text, t.Visibility as viz, t.Lang as language,
	        t.URI as uri, t.URI as url, 0 as replies_count, 0 as reblogs_count, 0 as favourites_count, false as favourited,
					false as reblogged, false as muted, false as bookmarked, t.Content as content, null as reblog,
					null as media_attachments, null as mentions, null as tags, null as emojis, null as card, null as poll,
					oc.name as app_name, oc.website as app_website
					from toots t
					left outer join oauth_clients oc on t.appid = oc.id
					where t.sid = ?`
	err := db.QueryRowx(base, sid).MapScan(attrs)
	if err != nil {
		return nil, errors.Wrap(err, "Error with toot "+sid)
	}

	attrs["visibility"] = model.FromVis(model.PostVisibility(attrs["viz"].(int64)))
	delete(attrs, "viz")

	medias, err := fetchTootMedias(db, sid)
	if err != nil {
		return nil, err
	}
	attrs["media_attachments"] = medias

	tags, err := fetchTootTags(attrs["content"].(string))
	if err != nil {
		return nil, err
	}
	attrs["tags"] = tags

	if attrs["app_name"] != nil {
		attrs["application"] = map[string]any{
			"name":    attrs["app_name"],
			"website": attrs["app_website"],
		}
		delete(attrs, "app_name")
		delete(attrs, "app_website")
	} else {
		attrs["application"] = nil
	}
	return attrs, nil
}

func fetchTootTags(content string) ([]map[string]any, error) {
	tags := extractTags(content)
	results := []map[string]any{}
	for _, tag := range tags {
		tagm := map[string]any{
			"name": tag,
		}
		results = append(results, tagm)
	}
	return results, nil
}

func fetchTootMedias(db *sqlx.DB, sid string) ([]map[string]any, error) {
	results := make([]map[string]any, 0)
	media := `select tm.* from toot_medias tm where tm.sid = ?`
	rows, err := db.Queryx(media, sid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		media := model.TootMedia{}
		err := rows.StructScan(&media)
		if err != nil {
			return nil, errors.Wrap(err, "Toot media query")
		}
		results = append(results, toAttachmentMap(&media))
	}
	return results, nil
}
