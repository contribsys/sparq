package public

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/activitystreams"
	"github.com/contribsys/sparq/db"
	"github.com/contribsys/sparq/util"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

//go:generate ego .

type Tab struct {
	Name string
	Path string
}

// openssl rand -hex 32
// ruby -rsecurerandom -e "puts SecureRandom.hex(32)"
var sessionStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

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
	root.HandleFunc("/users/{nick:[a-z0-9]{4,16}}", getUser)
	root.HandleFunc("/home", requireLogin(indexHandler))
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
		session, _ := sessionStore.Get(r, "sparq-session")
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
		session, _ := sessionStore.Get(r, "sparq-session")
		delete(session.Values, "uid")
		_ = session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func loginHandler(s sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := sessionStore.Get(r, "sparq-session")
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
			  select us.UserId, us.PasswordHash
				from users u join user_securities us
				on u.Id = us.UserId
				where u.Nick = ?`, username).Scan(&uid, &hash)
			if err != nil {
				if err == sql.ErrNoRows {
					util.Debugf("Username not found: %s", username)
					session.AddFlash("Invalid username or password")
					ego_login(w, r)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
		}
		ego_login(w, r)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	ego_index(w, r)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	nick := mux.Vars(r)["nick"]

	userdata := map[string]interface{}{}
	err := db.Database().QueryRowx(`
	select *
	from users
	inner join user_securities
	on users.id = user_securities.userid
	where users.nick = ?`, nick).MapScan(userdata)
	if err == sql.ErrNoRows {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url := "https://" + db.InstanceHostname + "/users/" + nick
	me := activitystreams.NewPerson(url)
	me.URL = url
	me.Name = userdata["FullName"].(string)
	me.PreferredUsername = userdata["Nick"].(string)
	me.AddPubKey(string(userdata["PublicKey"].([]uint8)))

	data, err := json.Marshal(me)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/activity+json")
	_, _ = w.Write(data)
}
