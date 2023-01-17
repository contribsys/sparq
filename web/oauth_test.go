package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestPublicOauth(t *testing.T) {
	ts, stopper := NewTestServer(t, "oauth")
	defer stopper()

	r := RootRouter(ts)
	// AddPublicEndpoints(s, r)
	IntegrateOauth(ts, r)

	route(r, "/nosuch", func(w *httptest.ResponseRecorder, req *http.Request) {
		assert.Equal(t, 404, w.Code)
		assert.Contains(t, w.Body.String(), "not found")
	})
	route(r, "/oauth/authorize", func(w *httptest.ResponseRecorder, req *http.Request) {
		assert.Equal(t, 302, w.Code)
		assert.Contains(t, w.Body.String(), "/login")
	})
	route(r, "/oauth/token", func(w *httptest.ResponseRecorder, req *http.Request) {
		assert.Equal(t, 401, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})
	route(r, "/oauth/authorize?client_id=93e60c83-3c57-42ac-abaf-be6bc7ad2e68&redirect_uri=http%3A%2F%2Flocalhost%3A4002%2Fsettings%2Finstances%2Fadd&response_type=code&scope=read%20write%20follow%20push", func(w *httptest.ResponseRecorder, req *http.Request) {
		assert.Equal(t, 302, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "/login")
	})

	t.Run("with http redirect", func(t *testing.T) {
		_, err := ts.DB().Exec(`insert into oauth_clients
		(ClientId, Name, Secret, RedirectUris, Website, Scopes) values 
		("93e60c83-3c57-42ac-abaf-be6bc7ad2e68", "Pinafore", "123456789abcdef", "http://localhost:4002/settings/instances/add", "http://localhost:4002", "read write follow push")`)
		assert.NoError(t, err)

		sos := &SqliteOauthStore{DB: ts.DB()}
		ci, err := sos.GetByID(context.Background(), "93e60c83-3c57-42ac-abaf-be6bc7ad2e68")
		assert.NoError(t, err)
		assert.NotNil(t, ci)

		req := httptest.NewRequest("GET", "http://localhost.dev:9494/oauth/authorize?client_id=93e60c83-3c57-42ac-abaf-be6bc7ad2e68&redirect_uri=http%3A%2F%2Flocalhost%3A4002%2Fsettings%2Finstances%2Fadd&response_type=code&scope=read%20write%20follow%20push", nil)
		w := httptest.NewRecorder()
		session, err := SessionStore.Get(req, "sparq-session")
		assert.NoError(t, err)
		session.Values["uid"] = "1"
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Authorize Application?")

		req = httptest.NewRequest("POST", "http://localhost.dev:9494/oauth/authorize?client_id=93e60c83-3c57-42ac-abaf-be6bc7ad2e68&redirect_uri=http%3A%2F%2Flocalhost%3A4002%2Fsettings%2Finstances%2Fadd&response_type=code&scope=read%20write%20follow%20push&Approve=1", nil)
		w = httptest.NewRecorder()
		session, err = SessionStore.Get(req, "sparq-session")
		assert.NoError(t, err)
		session.Values["uid"] = "1"
		session.Values["username"] = "admin"
		r.ServeHTTP(w, req)
		assert.Equal(t, 302, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "http://localhost:4002/settings/instances/add?code=")

		req = httptest.NewRequest("POST", "http://localhost.dev:9494/oauth/authorize?client_id=93e60c83-3c57-42ac-abaf-be6bc7ad2e68&redirect_uri=http%3A%2F%2Flocalhost%3A4002%2Fsettings%2Finstances%2Fadd&response_type=code&scope=read%20write%20follow%20push&Deny=1", nil)
		w = httptest.NewRecorder()
		session, err = SessionStore.Get(req, "sparq-session")
		assert.NoError(t, err)
		session.Values["uid"] = "1"
		session.Values["username"] = "admin"
		r.ServeHTTP(w, req)
		assert.Equal(t, 302, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "http://localhost:4002/settings/instances/add?error")
		var count int
		err = ts.DB().QueryRow("select count(*) from oauth_clients").Scan(&count)
		assert.NoError(t, err)
		assert.EqualValues(t, 0, count)
	})

	t.Run("with OOB redirect", func(t *testing.T) {
		_, err := ts.DB().Exec(`insert into oauth_clients
		(ClientId, Name, Secret, RedirectUris, Website, Scopes) values 
		("93e60c83-3c57-42ac-abaf-be6bc7ad2e69", "Tut", "987654321abcdef", "urn:ietf:wg:oauth:2.0:oob", "http://localhost:4002", "read write follow")`)
		assert.NoError(t, err)

		sos := &SqliteOauthStore{DB: ts.DB()}
		ci, err := sos.GetByID(context.Background(), "93e60c83-3c57-42ac-abaf-be6bc7ad2e69")
		assert.NoError(t, err)
		assert.NotNil(t, ci)

		req := httptest.NewRequest("GET", "http://localhost.dev:9494/oauth/authorize?client_id=93e60c83-3c57-42ac-abaf-be6bc7ad2e69&redirect_uri=urn%3Aietf%3Awg%3Aoauth%3A2.0%3Aoob&response_type=code&scope=read%20write%20follow", nil)
		w := httptest.NewRecorder()
		session, err := SessionStore.Get(req, "sparq-session")
		assert.NoError(t, err)
		session.Values["uid"] = "1"
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Authorize Application?")

		req = httptest.NewRequest("POST", "http://localhost.dev:9494/oauth/authorize?client_id=93e60c83-3c57-42ac-abaf-be6bc7ad2e69&redirect_uri=urn%3Aietf%3Awg%3Aoauth%3A2.0%3Aoob&response_type=code&scope=read%20write%20follow&Approve=1", nil)
		w = httptest.NewRecorder()
		session, err = SessionStore.Get(req, "sparq-session")
		assert.NoError(t, err)
		session.Values["uid"] = "1"
		session.Values["username"] = "admin"
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, w.Header().Get("Location"), "")
		assert.Contains(t, w.Body.String(), "Your authorization code is")
	})
}

func route(r *mux.Router, query string, fn func(w *httptest.ResponseRecorder, req *http.Request)) {
	req := httptest.NewRequest("GET", "http://localhost.dev:9494"+query, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	fn(w, req)
}
