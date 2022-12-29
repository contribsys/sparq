package clientapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/model"
	"github.com/contribsys/sparq/public"
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
type Status struct {
	AuthorID           int
	Content            string
	MediaIds           []string
	InReplyTo          string
	InReplyToAccountId string
	Sensitive          bool
	WarningText        string
	Visibility         string
	LanguageCode       string
	ScheduledAt        string
	*Poll
}

func statusHandler(svr sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			httpError(w, errors.New("POST only"), http.StatusBadRequest)
			return
		}
		fmt.Printf("Form: %+v %s\n", r.Form, r.Form.Get("status"))
		err := r.ParseForm()
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		fmt.Printf("Form: %+v %s\n", r.Form, r.Form.Get("status"))

		aid := public.CurrentAccountID(r)
		if aid == public.Anonymous {
			httpError(w, errors.New("Unauthorized"), http.StatusUnauthorized)
			return
		}

		status := &Status{
			AuthorID:     aid,
			Content:      r.Form.Get("status"),
			InReplyTo:    r.Form.Get("in_reply_to_id"),
			Sensitive:    r.Form.Get("sensitive") == "true",
			WarningText:  r.Form.Get("spoiler_text"),
			Visibility:   r.Form.Get("visibility"),
			LanguageCode: r.Form.Get("language"),
			ScheduledAt:  r.Form.Get("scheduled_at"),
		}
		if status.Content == "" {
			httpError(w, errors.New("No content!"), 400)
			return
		}

		if r.Form.Get("poll[expires_in]") != "" {
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
			status.Poll = p
		}

		// verify we're not getting a duplicate submission
		// since we are a single process, we don't need Redis
		ikey := r.Header.Get("Idempotency-Key")
		dupeMu.Lock()
		defer dupeMu.Unlock()
		_, ok := dupeDetector[ikey]
		if ok {
			httpError(w, errors.New("Duplicate status, ignoring"), 401)
			return
		}
		dupeDetector[ikey] = time.Now()
		defer dupeCleaner()

		post, err := saveStatus(svr, r, status)
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		err = enc.Encode(post)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
	}
}

func UIDFromBearer(r *http.Request) {
	panic("unimplemented")
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

func saveStatus(svr sparq.Server, r *http.Request, status *Status) (*model.Post, error) {
	sid := model.Snowflakes.NextID()
	p := &model.Post{
		URI:         fmt.Sprintf("https://%s/@%s/statuses/%d", svr.Hostname(), "admin", sid),
		AuthorID:    int64(status.AuthorID),
		WarningText: status.WarningText,
		Content:     status.Content,
		Visibility:  model.ToVis(status.Visibility),
		InReplyTo:   status.InReplyTo,
	}
	tx, err := svr.DB().Begin()
	if err != nil {
		return nil, err
	}
	if status.Poll != nil {
		po := status.Poll
		res, err := tx.ExecContext(r.Context(), `
			insert into polls (expires_in, multiple, hide, o1, o2, o3, o4, o5, o6) values (?, ?, ?, ?, ?, ?, ?)`,
			po.ExpiresIn, po.MultipleChoice, po.HideTotals,
			po.O1, po.O2, po.O3, po.O4, po.O5, po.O6)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		p.PollID, _ = res.LastInsertId()
	}
	_, err = tx.ExecContext(r.Context(), `
	  insert into posts (uri, inreplyto, authorid, pollid, summary, content, lang, visibility) values
		(?, ?, ?, ?, ?, ?, ?, ?)`,
		p.URI, p.InReplyTo, p.AuthorID, p.PollID, p.WarningText, p.Content, p.Lang, p.Visibility)
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
