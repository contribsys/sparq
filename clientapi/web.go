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
	mux.HandleFunc("/custom_emojis", emptyHandler(s))
	mux.HandleFunc("/lists", emptyHandler(s))
	mux.HandleFunc("/filters", emptyHandler(s))
	mux.HandleFunc("/notifications", emptyHandler(s))
	mux.HandleFunc("/instance", instanceHandler(s))
	mux.HandleFunc("/timelines/{type}", timelineHandler(s))
	mux.HandleFunc("/statuses", statusHandler(s))
	mux.HandleFunc("/apps/verify_credentials", appsVerifyHandler(s))
	mux.HandleFunc("/apps", appsHandler(s))
	mux.HandleFunc("/accounts/verify_credentials", verifyCredentialsHandler(s))
	mux.HandleFunc("/accounts/{sfid:[0-9]+}", getAccount)
	mux.HandleFunc("/accounts/{sfid:[0-9]+}/statuses", getAccountStatuses)
	// mux.HandleFunc("/accounts/{sfid:[0-9]+}/followers", getAccountFollowers)
	// mux.HandleFunc("/accounts/{sfid:[0-9]+}/following", getAccountFollowing)

}
