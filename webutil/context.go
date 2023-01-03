package webutil

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/contribsys/sparq"
	"github.com/gorilla/sessions"
)

type ContextType int32

var (
	Anonymous string      = ""
	HelperKey ContextType = 7

	// openssl rand -hex 32
	// ruby -rsecurerandom -e "puts SecureRandom.hex(32)"
	SessionStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
)

// this is the core data structure which is passed along with each web request
type WebCtx struct {
	BearerCode    string
	CurrentUserID string
	LangCode      string
}

func EstablishContext(svr sparq.Server) func(http.Handler) http.Handler {
	return func(pass http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			webctx := &WebCtx{
				BearerCode: bearerCode(r),
				LangCode:   langCookie(r),
			}

			pass.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), HelperKey, webctx)))
		})
	}
}

func langCookie(r *http.Request) string {
	cookie, _ := r.Cookie("langcode")
	if cookie != nil {
		val := cookie.Value
		if len(val) == 2 {
			return val
		}
	}
	return ""
}

func Ctx(r *http.Request) *WebCtx {
	x := r.Context().Value(HelperKey)
	if x != nil {
		y, ok := x.(*WebCtx)
		if ok {
			return y
		}
	}
	return &WebCtx{}
}

func bearerCode(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	prefix := "Bearer "

	if auth != "" && strings.HasPrefix(auth, prefix) {
		return auth[len(prefix):]
	}
	return ""
}
