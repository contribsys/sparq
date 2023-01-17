package clientapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/contribsys/sparq/web"
	"github.com/stretchr/testify/assert"
)

func TestMediaUpload(t *testing.T) {
	ts, stopper := web.NewTestServer(t, "media")
	defer stopper()
	token, err := registerToken(t, ts)
	assert.NoError(t, err)
	root := rootRouter(ts)
	AddPublicEndpoints(ts, root.PathPrefix("/api/v1").Subrouter())
	root.PathPrefix("/media/").Handler(http.StripPrefix("/media", http.FileServer(http.FS(os.DirFS(ts.MediaRoot())))))

	t.Run("PostMinimal", func(t *testing.T) {
		buf, wr, err := web.MultipartTestForm("file", "fixtures/cat.png", map[string]string{
			"description": "nice kitty",
			"focus":       "0.5,0.5",
		})
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "http://localhost.dev:9494/api/v1/media", bytes.NewReader(buf.Bytes()))
		req.Header.Set("Content-Type", wr.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accepts", "application/json")
		w := httptest.NewRecorder()

		root.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var testy map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &testy)
		assert.NoError(t, err)
		assert.NotNil(t, testy)

		assert.NotNil(t, testy["url"])
		u := testy["url"].(string)
		assert.FileExists(t, ts.Root()+u)
		assert.FileExists(t, ts.Root()+testy["preview_url"].(string))

		req = httptest.NewRequest("GET", "http://localhost.dev:9494"+u, nil)
		w = httptest.NewRecorder()
		root.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/jpeg", w.Header().Get("Content-Type"))
		assert.Equal(t, "114123", w.Header().Get("Content-Length"))
	})
}
