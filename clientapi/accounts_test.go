package clientapi

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/contribsys/sparq/db"
	"github.com/stretchr/testify/assert"
)

func jsonPayload(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	if w.Header().Get("Content-Type") != "application/json" {
		return map[string]interface{}{"content": w.Body.String()}
	}

	var payload map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &payload)
	// fmt.Printf("JSON response: %+v\n", payload)
	assert.NoError(t, err)
	return payload
}

func TestAccounts(t *testing.T) {
	stopper, err := db.TestDB("accounts")
	assert.NoError(t, err)
	defer stopper()
	ts := &testSvr{}
	root := rootRouter(ts)
	AddPublicEndpoints(ts, root.PathPrefix("/api/v1").Subrouter())

	t.Run("verify_credentials", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "http://localhost.dev:9494/api/v1/accounts/verify_credentials", nil)
		w := httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 204, w.Code)

		req = httptest.NewRequest("GET", "http://localhost.dev:9494/api/v1/accounts/verify_credentials", nil)
		req.Header.Add("Authorization", "Bearer 1234567")
		w = httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 401, w.Code)

		token, err := registerToken(t, ts)
		assert.NoError(t, err)

		req = httptest.NewRequest("GET", "http://localhost.dev:9494/api/v1/accounts/verify_credentials", nil)
		req.Header.Add("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)

		payload := jsonPayload(t, w)
		assert.NotNil(t, payload)
		assert.Contains(t, payload, "id")
	})

	t.Run("apps", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "http://localhost.dev:9494/api/v1/apps", nil)
		w := httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, w.Code, 204)
		assert.Contains(t, w.Body.String(), "")

		// GET not allowed
		req = httptest.NewRequest("GET", "http://localhost.dev:9494/api/v1/apps", nil)
		w = httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 400, w.Code)
		assert.Contains(t, w.Body.String(), "Bad method")

		// no body
		req = httptest.NewRequest("POST", "http://localhost.dev:9494/api/v1/apps", nil)
		w = httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 400, w.Code)
		assert.Contains(t, w.Body.String(), "")

		j := `{"client_name":"Pinafore",
		 "redirect_uris":"https://pinafore.social/settings/instances/add",
		 "scopes":"read write follow push",
		 "website":"https://pinafore.social"}`
		br := strings.NewReader(j)
		req = httptest.NewRequest("POST", "http://localhost.dev:9494/api/v1/apps", br)
		req.Header.Add("Content-Type", "application/json")
		w = httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "client_secret")
		assert.Contains(t, w.Body.String(), "Pinafore")

		var count int
		err := db.Database().QueryRow("select count(*) from oauth_clients where Name = 'Pinafore'").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}
