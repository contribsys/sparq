package faktoryui

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/contribsys/faktory/util"
	"github.com/contribsys/sparq/faktory"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestLiveServer(t *testing.T) {
	bootRuntime(t, "webui", func(ui *WebUI, s *faktory.Server, t *testing.T, dispatch http.HandlerFunc) {
		t.Run("StaticAssets", func(t *testing.T) {
			req, err := ui.NewRequest("GET", "http://localhost:7420/static/application.js", nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			dispatch(w, req)
			assert.Equal(t, 200, w.Code)
			assert.True(t, strings.Contains(w.Body.String(), "Fuzzy"), w.Body.String())
		})

		t.Run("Debug", func(t *testing.T) {
			req, err := ui.NewRequest("GET", "http://localhost:7420/debug", nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			dispatch(w, req)
			assert.Equal(t, 200, w.Code)
			assert.True(t, strings.Contains(w.Body.String(), "Disk Usage"), w.Body.String())
		})

		t.Run("Health", func(t *testing.T) {
			req, err := ui.NewRequest("GET", "http://localhost:7420/health", nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			healthHandler(ui)(w, req)
			assert.Equal(t, 200, w.Code)
			assert.True(t, strings.Contains(w.Body.String(), "faktory_version"), w.Body.String())
		})

		t.Run("ComputeLocale", func(t *testing.T) {
			lang := localeFromHeader("")
			assert.Equal(t, "en", lang)
			lang = localeFromHeader(" 'fr-FR,fr;q=0.8,en-US;q=0.6,en;q=0.4,ru;q=0.2'")
			assert.Equal(t, "fr", lang)
			lang = localeFromHeader("zh-CN,zh;q=0.8,en-US;q=0.6,en;q=0.4,ru;q=0.2")
			assert.Equal(t, "zh-cn", lang)
			lang = localeFromHeader("en-US,sv-SE;q=0.8,sv;q=0.6,en;q=0.4")
			assert.Equal(t, "sv", lang)
			lang = localeFromHeader("nb-NO,nb;q=0.2")
			assert.Equal(t, "nb", lang)
			lang = localeFromHeader("en-us")
			assert.Equal(t, "en", lang)
			lang = localeFromHeader("sv-se")
			assert.Equal(t, "sv", lang)
			lang = localeFromHeader("pt-BR,pt;q=0.8,en-US;q=0.6,en;q=0.4")
			assert.Equal(t, "pt-br", lang)
			lang = localeFromHeader("pt-PT,pt;q=0.8,en-US;q=0.6,en;q=0.4")
			assert.Equal(t, "pt", lang)
			lang = localeFromHeader("pt-br")
			assert.Equal(t, "pt-br", lang)
			lang = localeFromHeader("pt-pt")
			assert.Equal(t, "pt", lang)
			lang = localeFromHeader("pt")
			assert.Equal(t, "pt", lang)
			lang = localeFromHeader("en-us; *")
			assert.Equal(t, "en", lang)
			lang = localeFromHeader("en-US,en;q=0.8")
			assert.Equal(t, "en", lang)
			lang = localeFromHeader("en-GB,en-US;q=0.8,en;q=0.6")
			assert.Equal(t, "en", lang)
			lang = localeFromHeader("ru,en")
			assert.Equal(t, "ru", lang)
			lang = localeFromHeader("*")
			assert.Equal(t, "en", lang)
		})
	})
}

func bootRuntime(t *testing.T, name string, fn func(*WebUI, *faktory.Server, *testing.T, http.HandlerFunc)) {
	dir := fmt.Sprintf("/tmp/faktory-test-%s", name)
	defer os.RemoveAll(dir)

	sock := fmt.Sprintf("%s/redis.sock", dir)

	b := "localhost:9494"
	s, err := faktory.NewServer(faktory.Options{
		RedisSock:        sock,
		StorageDirectory: dir,
	})
	if err != nil {
		fmt.Println("Panic: " + err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = s.Run(ctx)
	if err != nil {
		fmt.Println("Panic: " + err.Error())
		return
	}

	root := mux.NewRouter()
	web := NewWeb(s, b)
	web.enabledCSRF = false
	web.Embed(root, "")
	root.NotFoundHandler = NotFound()

	fn(web, s, t, func(w http.ResponseWriter, r *http.Request) {
		root.ServeHTTP(w, r)
	})

	s.Store().Flush(context.Background())
	s.Close()
}

func NotFound() http.Handler {
	pass := http.NotFoundHandler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass.ServeHTTP(w, r)
		fmt.Printf("\033[31m404\033[0m %s %s\n", r.Method, r.RequestURI)
		// util.Infof("[404] %s %s %+v", r.Method, r.RequestURI, r.Header)
	})
}

func fakeJob() (string, []byte) {
	jid := util.RandomJid()
	nows := util.Nows()
	return jid, []byte(fmt.Sprintf(`{
		"jid":"%s",
		"created_at":"%s",
		"queue":"default",
		"args":[1,2,3],
		"jobtype":"SomeWorker",
		"at":"%s",
		"enqueued_at":"%s",
		"failure":{
		"retry_count":0,
		"failed_at":"%s",
		"message":"Invalid argument",
			"errtype":"RuntimeError"
		},
		"custom":{
			"foo":"bar",
			"tenant":1
		}
	}`, jid, nows, nows, nows, nows))
}
