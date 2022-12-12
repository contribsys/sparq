package public

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/activitystreams"
	"github.com/contribsys/sparq/db"
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

func LoggedInHandler(w http.ResponseWriter, r *http.Request) {
	// this handler is called before all resources requiring a logged in user
	// verify we have a user OR we'll redirect to /login
	session, _ := sessionStore.Get(r, "sparq-session")
	uid := session.Values["uid"]
	if uid == nil {
		session.Values["returnTo"] = r.Form
		err := sessionStore.Save(r, w, session)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
}

var (
	DefaultTabs = []Tab{
		{"Home", "/"},
		{"Local", "/local"},
		{"Federated", "/federated"},
	}
	ErrNotFound   = errors.New("User not found")
	staticHandler = http.FileServer(http.FS(staticFiles))
)

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
	root.HandleFunc("/", indexHandler)
	root.HandleFunc("/login", loginHandler(s))
	// mux.HandleFunc("/home", homeHandler)
	// mux.HandleFunc("/public/local", localHandler)
	// mux.HandleFunc("/public", publicHandler)
	// mux.HandleFunc("/auth/sign_up", signupHandler)
	// mux.HandleFunc("/auth/sign_in", signinHandler)
}

func loginHandler(s sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := sessionStore.Get(r, "sparq-session")
		if r.Method == "POST" {
			if r.Form == nil {
				_ = r.ParseForm()
			}
			username := r.Form.Get("username")
			password := r.Form.Get("password")
			var uid int64
			var hash []byte
			err := s.DB().QueryRowxContext(r.Context(), `
			  select us.UserId, us.PasswordHash
				from users u join user_securities us
				on u.Id = us.UserId
				where u.Nick = ?`, strings.ToLower(username)).Scan(&uid, &hash)
			if err != nil {
				if err == sql.ErrNoRows {
					session.AddFlash("Invalid username or password")
					ego_login(w, r)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			pwdhash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if uid > 0 && subtle.ConstantTimeCompare(hash, pwdhash) == 1 {
				session.Values["uid"] = uid
				redir, ok := session.Values["redirectTo"].(string)
				delete(session.Values, "redirectTo")
				_ = session.Save(r, w)
				if !ok {
					redir = "/"
				}
				http.Redirect(w, r, redir, http.StatusFound)
				return
			}
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
