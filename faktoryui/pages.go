package faktoryui

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
)

func statsHandler(w http.ResponseWriter, r *http.Request) {
	hash, err := ctx(r).Server().CurrentState(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Cache-Control", "no-cache")
	_, _ = w.Write(data)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if ctx(r).Server() == nil {
		http.Error(w, "Server not booted", http.StatusInternalServerError)
		return
	}
	ego_index(w, r)
}

func queuesHandler(w http.ResponseWriter, r *http.Request) {
	ego_listQueues(w, r)
}

func queueHandler(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if name == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	q, ok := ctx(r).Store().ExistingQueue(r.Context(), name)
	if !ok {
		Redirect(w, r, "/queues", http.StatusFound)
		return
	}

	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		keys := r.Form["bkey"]
		if len(keys) > 0 {
			// delete specific entries
			bkeys := make([][]byte, len(keys))
			for idx := range keys {
				bindata, err := base64.RawURLEncoding.DecodeString(keys[idx])
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				bkeys[idx] = bindata
			}
			err := q.Delete(r.Context(), bkeys)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			action := r.FormValue("action")
			if action == "delete" {
				// clear entire queue
				_, err := q.Clear(r.Context())
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else if action == "pause" {
				err := ctx(r).webui.Server.Manager().PauseQueue(r.Context(), q.Name())
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else if action == "resume" {
				err := ctx(r).webui.Server.Manager().ResumeQueue(r.Context(), q.Name())
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			Redirect(w, r, "/queues", http.StatusFound)
			return
		}
		Redirect(w, r, "/queues/"+name, http.StatusFound)
		return
	}

	currentPage := uint64(1)
	p := r.URL.Query()["page"]
	if p != nil {
		val, err := strconv.Atoi(p[0])
		if err != nil {
			http.Error(w, "Invalid parameter", http.StatusBadRequest)
			return
		}
		currentPage = uint64(val)
	}
	count := uint64(25)

	ego_queue(w, r, q, count, currentPage)
}

func retriesHandler(w http.ResponseWriter, r *http.Request) {
	set := ctx(r).Store().Retries()

	if r.Method == "POST" {
		action := r.FormValue("action")
		keys := r.Form["key"]
		err := actOn(r, set, action, keys)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			Redirect(w, r, "/retries", http.StatusFound)
		}
		return
	}

	currentPage := uint64(1)
	p := r.URL.Query()["page"]
	if p != nil {
		val, err := strconv.Atoi(p[0])
		if err != nil {
			http.Error(w, "Invalid parameter", http.StatusBadRequest)
			return
		}
		currentPage = uint64(val)
	}
	count := uint64(25)

	ego_listRetries(w, r, set, count, currentPage)
}

func retryHandler(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if name == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	key, err := url.QueryUnescape(name)
	if err != nil {
		http.Error(w, "Invalid URL input", http.StatusBadRequest)
		return
	}

	set := ctx(r).Store().Retries()
	if r.Method == "POST" {
		action := r.FormValue("action")
		keys := []string{key}
		err := actOn(r, set, action, keys)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			Redirect(w, r, "/retries", http.StatusFound)
		}
		return
	}

	data, err := set.Get(r.Context(), []byte(key))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if data == nil {
		// retry has disappeared?  possibly requeued while the user was sitting on the /retries page
		Redirect(w, r, "/retries", http.StatusTemporaryRedirect)
		return
	}

	job, err := data.Job()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if job.Failure == nil {
		http.Error(w, fmt.Sprintf("Job %s is not a retry", job.Jid), http.StatusInternalServerError)
		return
	}
	ego_retry(w, r, key, job)
}

func scheduledHandler(w http.ResponseWriter, r *http.Request) {
	set := ctx(r).Store().Scheduled()

	if r.Method == "POST" {
		action := r.FormValue("action")
		keys := r.Form["key"]
		err := actOn(r, set, action, keys)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			Redirect(w, r, "/scheduled", http.StatusFound)
		}
		return
	}

	currentPage := uint64(1)
	p := r.URL.Query()["page"]
	if p != nil {
		val, err := strconv.Atoi(p[0])
		if err != nil {
			http.Error(w, "Invalid parameter", http.StatusBadRequest)
			return
		}
		currentPage = uint64(val)
	}
	count := uint64(25)

	ego_listScheduled(w, r, set, count, currentPage)
}

func scheduledJobHandler(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if name == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	key, err := url.QueryUnescape(name)
	if err != nil {
		http.Error(w, "Invalid URL input", http.StatusBadRequest)
		return
	}

	set := ctx(r).Store().Scheduled()
	if r.Method == "POST" {
		action := r.FormValue("action")
		keys := []string{key}
		err := actOn(r, set, action, keys)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			Redirect(w, r, "/scheduled", http.StatusFound)
		}
		return
	}

	data, err := set.Get(r.Context(), []byte(key))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if data == nil {
		// retry has disappeared?  possibly requeued while the user was sitting on the /retries page
		Redirect(w, r, "/scheduled", http.StatusTemporaryRedirect)
		return
	}

	job, err := data.Job()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ego_scheduled_job(w, r, key, job)
}

func morgueHandler(w http.ResponseWriter, r *http.Request) {
	set := ctx(r).Store().Dead()

	if r.Method == "POST" {
		action := r.FormValue("action")
		keys := r.Form["key"]
		err := actOn(r, set, action, keys)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			Redirect(w, r, "/morgue", http.StatusFound)
		}
		return
	}

	currentPage := uint64(1)
	p := r.URL.Query()["page"]
	if p != nil {
		val, err := strconv.Atoi(p[0])
		if err != nil {
			http.Error(w, "Invalid parameter", http.StatusBadRequest)
			return
		}
		currentPage = uint64(val)
	}
	count := uint64(25)

	ego_listDead(w, r, set, count, currentPage)
}

func deadHandler(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if name == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	key, err := url.QueryUnescape(name)
	if err != nil {
		http.Error(w, "Invalid URL input", http.StatusBadRequest)
		return
	}
	data, err := ctx(r).Store().Dead().Get(r.Context(), []byte(key))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if data == nil {
		// retry has disappeared?  possibly requeued while the user was sitting on the listing page
		Redirect(w, r, "/morgue", http.StatusTemporaryRedirect)
		return
	}

	job, err := data.Job()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ego_dead(w, r, key, job)
}

func busyHandler(w http.ResponseWriter, r *http.Request) {
	ego_busy(w, r)
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	ego_debug(w, r)
}

func Redirect(w http.ResponseWriter, r *http.Request, path string, code int) {
	http.Redirect(w, r, relative(r, path), code)
}
