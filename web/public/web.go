package public

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/clientapi"
	"github.com/contribsys/sparq/model"
	"github.com/contribsys/sparq/util"
	"github.com/contribsys/sparq/web"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound   = errors.New("User not found")
	staticHandler = cacheControl(http.FileServer(http.FS(staticFiles)))
)

func cacheControl(pass http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "public, max-age=300")
		pass.ServeHTTP(w, r)
	})
}

func AddPublicEndpoints(s sparq.Server, root *mux.Router) {
	root.PathPrefix("/static").Handler(staticHandler)
	root.HandleFunc("/users/{nick:[a-z0-9]{4,20}}", getUser(s))
	root.HandleFunc("/@{nick:[a-z0-9]{4,20}}/{id:[A-Z0-9]+}", showStatusHandler(s))
	root.HandleFunc("/@{nick:[a-z0-9]{4,20}}", getUser(s))
	root.Methods("POST").Path("/home").Handler(clientapi.PostTootHandler(s))
	root.HandleFunc("/home", web.RequireLogin(homeHandler))
	root.HandleFunc("/", indexHandler)
	root.HandleFunc("/login", loginHandler(s))
	root.HandleFunc("/logout", logoutHandler(s))
	root.HandleFunc("/public/local", localHandler(s))
	// mux.HandleFunc("/public", publicHandler)
	// mux.HandleFunc("/auth/sign_up", signupHandler)
	// mux.HandleFunc("/auth/sign_in", signinHandler)
}

func localHandler(svr sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tq := model.TQ(svr.DB())
		tq.Local = true
		tq.Visibility = model.VisPublic
		result, err := tq.Execute()
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		web.Render(w, r, "public/local", result)
	}
}

func showStatusHandler(svr sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sid := mux.Vars(r)["id"]
		attrs, err := clientapi.TootMap(svr.DB(), sid)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				httpError(w, err, http.StatusNotFound)
				return
			}
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		if attrs["visibility"] == "public" ||
			attrs["authorId"] == web.IsLoggedIn(r) {
			web.Render(w, r, "public/status", attrs)
		} else {
			httpError(w, err, http.StatusNotFound)
		}
	}
}

func logoutHandler(s sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := web.SessionStore.Get(r, "sparq-session")
		delete(session.Values, "uid")
		_ = session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func loginHandler(s sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := web.SessionStore.Get(r, "sparq-session")
		if session.Values["uid"] != nil {
			util.Debugf("User %s is already logged in", session.Values["uid"])
			http.Redirect(w, r, "/home", http.StatusFound)
			return
		}
		if r.Method == "POST" {
			if r.Form == nil {
				_ = r.ParseForm()
			}
			username := strings.ToLower(r.Form.Get("username"))
			password := r.Form.Get("password")
			var uid uint64
			var hash []byte
			err := s.DB().QueryRowxContext(r.Context(), `
			  select a.id, us.passwordhash
				from accounts	a join account_securities us
				on a.id = us.accountid
				where a.nick = ?`, username).Scan(&uid, &hash)
			if err != nil {
				if err == sql.ErrNoRows {
					util.Debugf("Username not found: %s", username)
					session.AddFlash("Invalid username or password")
					web.Render(w, r, "public/login", nil)
					return
				}
				httpError(w, err, http.StatusInternalServerError)
				return
			}
			util.Debugf("Login %s (uid %d)", username, uid)
			err = bcrypt.CompareHashAndPassword(hash, []byte(password))
			if err == nil {
				session.Values["uid"] = strconv.FormatUint(uid, 10)
				session.Values["username"] = username
				redir, ok := session.Values["redirectTo"].(string)
				delete(session.Values, "redirectTo")
				_ = session.Save(r, w)
				if !ok {
					redir = "/home"
				}
				http.Redirect(w, r, redir, http.StatusFound)
				return
			}
			util.Debugf("Password %q doesn't match: %s", password, hash)
			session.AddFlash("Invalid username or password")
		}
		web.Render(w, r, "public/login", nil)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	web.Render(w, r, "public/home", []model.Toot{})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if web.IsLoggedIn(r) != web.Anonymous {
		http.Redirect(w, r, "/home", http.StatusFound)
		return
	}
	web.Render(w, r, "public/index", nil)
}
