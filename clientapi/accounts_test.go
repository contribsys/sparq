package clientapi

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/contribsys/sparq/web"
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
	ts, stopper := web.NewTestServer(t, "accounts")
	defer stopper()

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
}
