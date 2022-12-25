package public

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/model"
	"github.com/contribsys/sparq/util"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type Tab struct {
	Name string
	Path string
}

// func LoggedInHandler(w http.ResponseWriter, r *http.Request) {
// 	// this handler is called before all resources requiring a logged in user
// 	// verify we have a user OR we'll redirect to /login
// 	session, _ := sessionStore.Get(r, "sparq-session")
// 	uid := session.Values["uid"]
// 	if uid == nil {
// 		session.Values["returnTo"] = r.Form
// 		err := sessionStore.Save(r, w, session)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		http.Redirect(w, r, "/login", http.StatusFound)
// 		return
// 	}
// }

var (
	DefaultTabs = []Tab{
		{"Home", "/"},
		{"Local", "/local"},
		{"Federated", "/federated"},
	}
	ErrNotFound   = errors.New("User not found")
	staticHandler = cacheControl(http.FileServer(http.FS(staticFiles)))
)

func cacheControl(pass http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "public, max-age=300")
		pass.ServeHTTP(w, r)
	})
}

func setCtx(pass http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := r.WithContext(context.WithValue(r.Context(), ctxKey, newCtx(w, r)))
		pass.ServeHTTP(w, req)
	})
}

func AddPublicEndpoints(s sparq.Server, root *mux.Router) {
	root.Use(setCtx)
	root.PathPrefix("/static").Handler(staticHandler)
	root.HandleFunc("/users/{nick:[a-z0-9]{4,20}}", getUser)
	root.HandleFunc("/@{nick:[a-z0-9]{4,20}}", getUser)
	root.HandleFunc("/home", requireLogin(homeHandler))
	root.HandleFunc("/", indexHandler)
	root.HandleFunc("/login", loginHandler(s))
	root.HandleFunc("/logout", logoutHandler(s))
	// mux.HandleFunc("/home", homeHandler)
	// mux.HandleFunc("/public/local", localHandler)
	// mux.HandleFunc("/public", publicHandler)
	// mux.HandleFunc("/auth/sign_up", signupHandler)
	// mux.HandleFunc("/auth/sign_in", signinHandler)
}

func requireLogin(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := sparq.SessionStore.Get(r, "sparq-session")
		uid, ok := session.Values["uid"]
		if !ok {
			if r.Form == nil {
				_ = r.ParseForm()
			}
			util.Debugf("Anonymous, %s requires /login", r.URL.Path)
			session.Values["returnForm"] = r.Form
			session.AddFlash("Please sign in to continue")
			_ = session.Save(r, w)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		util.Debugf("Current UID: %d", uid)
		fn(w, r)
	}
}

func logoutHandler(s sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := sparq.SessionStore.Get(r, "sparq-session")
		delete(session.Values, "uid")
		_ = session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func loginHandler(s sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := sparq.SessionStore.Get(r, "sparq-session")
		if session.Values["uid"] != nil {
			util.Debugf("User %d is already logged in", session.Values["uid"])
			http.Redirect(w, r, "/home", http.StatusFound)
			return
		}
		if r.Method == "POST" {
			if r.Form == nil {
				_ = r.ParseForm()
			}
			username := strings.ToLower(r.Form.Get("username"))
			password := r.Form.Get("password")
			var uid int64
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
					render(w, r, "login", nil)
					return
				}
				httpError(w, err, http.StatusInternalServerError)
				return
			}
			util.Debugf("Login %s (uid %d)", username, uid)
			err = bcrypt.CompareHashAndPassword(hash, []byte(password))
			if err == nil {
				session.Values["uid"] = uid
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
		render(w, r, "login", nil)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "home", []model.Post{})
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "index", nil)
}
