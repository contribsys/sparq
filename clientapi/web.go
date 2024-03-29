package clientapi

import (
	"os"

	"github.com/contribsys/sparq"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var (
	// openssl rand -hex 32
	// ruby -rsecurerandom -e "puts SecureRandom.hex(32)"
	sessionStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
)

func AddPublicEndpoints(s sparq.Server, mux *mux.Router) {
	mux.HandleFunc("/media", postMediaHandler(s))
	mux.HandleFunc("/media/{id:[0-9]+}", getMediaAttachmentHandler(s))
	mux.HandleFunc("/statuses", PostTootHandler(s))
	mux.HandleFunc("/statuses/{id}", getTootHandler(s))
	mux.HandleFunc("/custom_emojis", emptyHandler(s))
	mux.HandleFunc("/lists", emptyHandler(s))
	mux.HandleFunc("/filters", emptyHandler(s))
	mux.HandleFunc("/notifications", emptyHandler(s))
	mux.HandleFunc("/instance", instanceHandler(s))
	mux.HandleFunc("/timelines/public", publicHandler(s))
	mux.HandleFunc("/timelines/home", homeHandler(s))
	mux.HandleFunc("/timelines/{name}", listHandler(s))
	mux.HandleFunc("/apps/verify_credentials", appsVerifyHandler(s))
	mux.HandleFunc("/apps", appsHandler(s))
	mux.HandleFunc("/accounts/verify_credentials", verifyCredentialsHandler(s))
	mux.HandleFunc("/accounts/{sfid:[0-9]+}", getAccount(s))
	mux.HandleFunc("/accounts/{sfid:[0-9]+}/statuses", getAccountToots(s))

	st := NewStreamer(s)
	r := mux.PathPrefix("/streaming").Subrouter()
	r.HandleFunc("/{key}", st.Handler(s))

	// mux.HandleFunc("/accounts/{sfid:[0-9]+}/followers", getAccountFollowers)
	// mux.HandleFunc("/accounts/{sfid:[0-9]+}/following", getAccountFollowing)
}
