package sparq

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
)

type Server interface {
	DB() *sqlx.DB
	Hostname() string
	LogLevel() string
	Context() context.Context
}

var (
	HelperKey int = 7

	// openssl rand -hex 32
	// ruby -rsecurerandom -e "puts SecureRandom.hex(32)"
	SessionStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
)

type WebHelper struct {
	AccessCode        string
	LoggedInAccountID string
}

func FromRequest(r *http.Request) *WebHelper {
	x := r.Context().Value(HelperKey)
	if x != nil {
		y, ok := x.(*WebHelper)
		if ok {
			return y
		}
	}
	return &WebHelper{}
}

func AccessCode(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	prefix := "Bearer "

	if auth != "" && strings.HasPrefix(auth, prefix) {
		return auth[len(prefix):]
	}
	return ""
}
