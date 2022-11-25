package faktoryui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/contribsys/faktory/client"
	"github.com/contribsys/faktory/storage"
	"github.com/contribsys/faktory/util"
	"github.com/contribsys/sparq/faktory"
	"github.com/stretchr/testify/assert"
)

func TestPages(t *testing.T) {
	bootRuntime(t, "pages", func(ui *WebUI, s *faktory.Server, t *testing.T, dispatch http.HandlerFunc) {

		t.Run("Index", func(t *testing.T) {
			req, err := ui.NewRequest("GET", "http://localhost:7420/", nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			dispatch(w, req)
			assert.Equal(t, 200, w.Code)
			assert.True(t, strings.Contains(w.Body.String(), "uptime_in_days"), w.Body.String())
			assert.True(t, strings.Contains(w.Body.String(), "idle"), w.Body.String())
		})

		t.Run("Stats", func(t *testing.T) {
			req, err := ui.NewRequest("GET", "http://localhost:7420/stats", nil)
			assert.NoError(t, err)

			s.StartedAt = time.Now().Add(-1234567 * time.Second)
			str := s.Store()
			_, err = str.GetQueue(req.Context(), "default")
			assert.NoError(t, err)
			q, err := str.GetQueue(req.Context(), "foobar")
			assert.NoError(t, err)
			_, err = q.Clear(req.Context())
			assert.NoError(t, err)
			args := []string{"faktory", "rocks", "!!", ":)"}
			for _, v := range args {
				err = q.Push(req.Context(), []byte(v))
				assert.NoError(t, err)
			}

			w := httptest.NewRecorder()
			dispatch(w, req)
			assert.Equal(t, 200, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var content map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &content)
			assert.NoError(t, err)

			s := content["server"].(map[string]interface{})
			uid := s["uptime"].(float64)
			assert.Equal(t, float64(1234567), uid)

			queues := content["faktory"].(map[string]interface{})["queues"].(map[string]interface{})
			defaultQ := queues["default"].(float64)
			assert.Equal(t, 0.0, defaultQ)
			foobarQ := queues["foobar"]
			assert.Nil(t, foobarQ)
		})

		t.Run("Queues", func(t *testing.T) {
			req, err := ui.NewRequest("GET", "http://localhost:7420/queues", nil)
			assert.NoError(t, err)

			str := s.Store()
			_, err = str.GetQueue(req.Context(), "default")
			assert.NoError(t, err)
			q, err := str.GetQueue(req.Context(), "foobar")
			assert.NoError(t, err)
			_, err = q.Clear(req.Context())
			assert.NoError(t, err)
			err = q.Push(req.Context(), []byte("1l23j12l3"))
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			dispatch(w, req)
			assert.Equal(t, 200, w.Code)
			assert.True(t, strings.Contains(w.Body.String(), "default"), w.Body.String())
			assert.False(t, strings.Contains(w.Body.String(), "foobar"), w.Body.String())
		})

		t.Run("Queue", func(t *testing.T) {
			req, err := ui.NewRequest("GET", "http://localhost:7420/queues/foobar", nil)
			assert.NoError(t, err)
			s.Store().Flush(req.Context())

			str := s.Store()
			q, err := str.GetQueue(req.Context(), "foobar")
			assert.NoError(t, err)
			_, err = q.Clear(req.Context())
			assert.NoError(t, err)

			job := client.NewJob("SomeWorker", "1l23j12l3")
			job.Queue = "foobar"

			err = s.Manager().Push(req.Context(), job)
			assert.NoError(t, err)
			assert.EqualValues(t, 1, q.Size(req.Context()))

			w := httptest.NewRecorder()
			dispatch(w, req)
			assert.Equal(t, 200, w.Code)
			assert.True(t, strings.Contains(w.Body.String(), "1l23j12l3"), w.Body.String())
			assert.True(t, strings.Contains(w.Body.String(), "foobar"), w.Body.String())

			payload := url.Values{
				"action": {"delete"},
			}
			req, err = ui.NewRequest("POST", "http://localhost:7420/queues/"+q.Name(), strings.NewReader(payload.Encode()))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w = httptest.NewRecorder()
			dispatch(w, req)

			assert.Equal(t, w.Body.String(), "")
			assert.Equal(t, 302, w.Code)
			assert.EqualValues(t, 0, q.Size(req.Context()))
		})

		t.Run("Retries", func(t *testing.T) {
			req, err := ui.NewRequest("GET", "http://localhost:7420/retries", nil)
			assert.NoError(t, err)
			s.Store().Flush(req.Context())

			str := s.Store()
			retries := str.Retries()
			err = retries.Clear(req.Context())
			assert.NoError(t, err)
			jid1, data := fakeJob()
			err = retries.AddElement(req.Context(), util.Nows(), jid1, data)
			assert.NoError(t, err)

			jid2, data := fakeJob()
			err = retries.AddElement(req.Context(), util.Nows(), jid2, data)
			assert.NoError(t, err)

			jid3, data := fakeJob()
			err = retries.AddElement(req.Context(), util.Nows(), jid3, data)
			assert.NoError(t, err)

			var key []byte
			err = retries.Each(req.Context(), func(idx int, entry storage.SortedEntry) error {
				key, err = entry.Key()
				return err
			})
			assert.NoError(t, err)
			keys := string(key)

			w := httptest.NewRecorder()
			dispatch(w, req)
			assert.Equal(t, 200, w.Code)
			assert.True(t, strings.Contains(w.Body.String(), jid1), w.Body.String())
			assert.True(t, strings.Contains(w.Body.String(), jid2), w.Body.String())

			def, err := str.GetQueue(req.Context(), "default")
			assert.NoError(t, err)
			sz := def.Size(req.Context())
			cnt, err := def.Clear(req.Context())
			assert.NoError(t, err)
			assert.Equal(t, sz, cnt)

			assert.EqualValues(t, 0, def.Size(req.Context()))
			assert.EqualValues(t, 3, retries.Size(req.Context()))
			payload := url.Values{
				"key":    {keys},
				"action": {"retry"},
			}
			req, err = ui.NewRequest("POST", "http://localhost:7420/retries", strings.NewReader(payload.Encode()))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w = httptest.NewRecorder()
			dispatch(w, req)

			assert.Equal(t, "", w.Body.String())
			assert.Equal(t, 302, w.Code)
			assert.EqualValues(t, 2, retries.Size(req.Context()))
			assert.EqualValues(t, 1, def.Size(req.Context()))

			err = retries.Each(req.Context(), func(idx int, entry storage.SortedEntry) error {
				key, err = entry.Key()
				return err
			})
			assert.NoError(t, err)

			keys = string(key)
			payload = url.Values{
				"key":    {keys},
				"action": {"kill"},
			}
			assert.EqualValues(t, 0, str.Dead().Size(req.Context()))
			req, err = ui.NewRequest("POST", "http://localhost:7420/retries", strings.NewReader(payload.Encode()))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w = httptest.NewRecorder()
			dispatch(w, req)

			assert.Equal(t, "", w.Body.String())
			assert.Equal(t, 302, w.Code)
			assert.EqualValues(t, 1, retries.Size(req.Context()))
			assert.EqualValues(t, 1, str.Dead().Size(req.Context()))

			// Now try to operate on an element which disappears by clearing the sset
			// manually before submitting
			err = retries.Each(req.Context(), func(idx int, entry storage.SortedEntry) error {
				key, err = entry.Key()
				return err
			})
			assert.NoError(t, err)

			keys = string(key)
			payload = url.Values{
				"key":    {keys},
				"action": {"kill"},
			}
			err = retries.Clear(req.Context()) // clear it under us!
			assert.NoError(t, err)
			req, err = ui.NewRequest("POST", "http://localhost:7420/retries", strings.NewReader(payload.Encode()))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w = httptest.NewRecorder()
			dispatch(w, req)

			assert.Equal(t, "", w.Body.String())
			assert.Equal(t, 302, w.Code)
			assert.EqualValues(t, 0, retries.Size(req.Context()))
			assert.EqualValues(t, 1, str.Dead().Size(req.Context()))
		})

		t.Run("Retry", func(t *testing.T) {
			jid, data := fakeJob()
			ts := util.Nows()
			req, err := ui.NewRequest("GET", fmt.Sprintf("http://localhost:7420/retries/%s|%s", ts, jid), nil)
			assert.NoError(t, err)

			str := s.Store()
			q := str.Retries()
			err = q.Clear(req.Context())
			assert.NoError(t, err)

			err = q.AddElement(req.Context(), ts, jid, data)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			dispatch(w, req)
			assert.Equal(t, 200, w.Code)
			assert.True(t, strings.Contains(w.Body.String(), jid), w.Body.String())
		})

		t.Run("Scheduled", func(t *testing.T) {
			req, err := ui.NewRequest("GET", "http://localhost:7420/scheduled", nil)
			assert.NoError(t, err)

			str := s.Store()
			q := str.Scheduled()
			err = q.Clear(req.Context())
			assert.NoError(t, err)
			jid, data := fakeJob()

			err = q.AddElement(req.Context(), util.Nows(), jid, data)
			assert.NoError(t, err)

			var key []byte
			err = q.Each(req.Context(), func(idx int, entry storage.SortedEntry) error {
				key, err = entry.Key()
				return err
			})
			assert.NoError(t, err)
			keys := string(key)

			w := httptest.NewRecorder()
			dispatch(w, req)
			assert.Equal(t, 200, w.Code)
			assert.True(t, strings.Contains(w.Body.String(), "SomeWorker"), w.Body.String())
			assert.True(t, strings.Contains(w.Body.String(), keys), w.Body.String())

			assert.EqualValues(t, 1, q.Size(req.Context()))
			payload := url.Values{
				"key":    {keys},
				"action": {"add_to_queue"},
			}
			req, err = ui.NewRequest("POST", "http://localhost:7420/scheduled", strings.NewReader(payload.Encode()))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w = httptest.NewRecorder()
			dispatch(w, req)

			assert.Equal(t, 302, w.Code)
			assert.EqualValues(t, 0, q.Size(req.Context()))
			assert.False(t, strings.Contains(w.Body.String(), keys), w.Body.String())
		})

		t.Run("ScheduledJob", func(t *testing.T) {
			jid, data := fakeJob()
			ts := util.Thens(time.Now().Add(1e6 * time.Second))
			req, err := ui.NewRequest("GET", fmt.Sprintf("http://localhost:7420/scheduled/%s|%s", ts, jid), nil)
			assert.NoError(t, err)

			str := s.Store()
			q := str.Scheduled()
			err = q.Clear(req.Context())
			assert.NoError(t, err)

			err = q.AddElement(req.Context(), ts, jid, data)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			dispatch(w, req)
			assert.Equal(t, 200, w.Code)
			assert.True(t, strings.Contains(w.Body.String(), jid), w.Body.String())
		})

		t.Run("Morgue", func(t *testing.T) {
			req, err := ui.NewRequest("GET", "http://localhost:7420/morgue", nil)
			assert.NoError(t, err)

			str := s.Store()
			q := str.Dead()
			err = q.Clear(req.Context())
			assert.NoError(t, err)
			jid, data := fakeJob()

			err = q.AddElement(req.Context(), util.Nows(), jid, data)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			dispatch(w, req)
			assert.Equal(t, 200, w.Code)
			assert.True(t, strings.Contains(w.Body.String(), jid), w.Body.String())
		})

		t.Run("Dead", func(t *testing.T) {
			jid, data := fakeJob()
			ts := util.Nows()
			req, err := ui.NewRequest("GET", fmt.Sprintf("http://localhost:7420/morgue/%s|%s", ts, jid), nil)
			assert.NoError(t, err)

			str := s.Store()
			q := str.Dead()
			err = q.Clear(req.Context())
			assert.NoError(t, err)

			err = q.AddElement(req.Context(), ts, jid, data)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			dispatch(w, req)
			assert.Equal(t, 200, w.Code)
			assert.True(t, strings.Contains(w.Body.String(), jid), w.Body.String())

			assert.EqualValues(t, 1, q.Size(req.Context()))
			payload := url.Values{
				"key":    {"all"},
				"action": {"delete"},
			}
			req, err = ui.NewRequest("POST", "http://localhost:7420/morgue", strings.NewReader(payload.Encode()))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w = httptest.NewRecorder()
			dispatch(w, req)

			assert.Equal(t, 302, w.Code)
			assert.Equal(t, "", w.Body.String())
			assert.EqualValues(t, 0, q.Size(req.Context()))
		})

		t.Run("Busy", func(t *testing.T) {
			req, err := ui.NewRequest("GET", "http://localhost:7420/busy", nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			dispatch(w, req)
			assert.Equal(t, 200, w.Code)
		})

	})
}

func (ui *WebUI) NewRequest(method string, urlstr string, body io.Reader) (*http.Request, error) {
	r := httptest.NewRequest(method, urlstr, body)
	dctx := &DefaultContext{
		Context: r.Context(),
		webui:   ui,
		request: r,
		locale:  "en",
		strings: translations("en"),
	}
	return r.WithContext(dctx), nil
}
