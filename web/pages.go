package web

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/contribsys/sparq/db"
	"github.com/contribsys/sparq/model"
	"github.com/contribsys/sparq/util"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
)

var (
	pageTemplates = map[string]*template.Template{}

	//go:embed locales/*.yml
	localeFiles embed.FS

	//go:embed public/*.gotmpl
	templateFiles embed.FS
)

type localeMap map[string]map[string]string
type assetLookup func(string) ([]byte, error)

var (
	AssetLookups = []assetLookup{
		localeFiles.ReadFile,
	}
	locales = localeMap{}
)

func init() {
	RegisterPages("public/authorize")

	files, err := localeFiles.ReadDir("locales")
	if err != nil {
		panic(err)
	}
	for idx := range files {
		name := strings.Split(files[idx].Name(), ".")[0]
		locales[name] = nil
	}
	// util.Debugf("Initialized %d locales", len(files))
}

func RegisterPages(pages ...string) {
	// Include the navigation partial in the template files.
	for _, page := range pages {
		files := []string{
			"public/base.gotmpl",
			"public/nav.gotmpl",
			"public/flashes.gotmpl",
			"public/newpost.gotmpl",
			"public/timeline.gotmpl",
			page + ".gotmpl",
		}

		ts := template.New(page)
		ts.Funcs(template.FuncMap{
			"now":      util.Nows,
			"hostname": func() string { return db.InstanceHostname },
			"relative": func(when time.Time) string { return time.Since(when).String() },
		})
		ts, err := ts.ParseFS(templateFiles, files...)
		if err != nil {
			panic(err)
		}
		pageTemplates[page] = ts
	}
}

func Render(w http.ResponseWriter, r *http.Request, page string, custom any) {
	ts := pageTemplates[page]
	if ts == nil {
		panic(errors.New("No registered page: " + page))
	}
	err := ts.ExecuteTemplate(w, "base", &PageData{r: r, w: w, Locale: "en", Custom: custom})
	if err != nil {
		HttpError(w, err, 503)
	}
}

type PageData struct {
	act    *model.Account
	r      *http.Request
	w      http.ResponseWriter
	Locale string
	Custom any
}

func (pd *PageData) T(text string) string {
	trns, ok := locales[pd.Locale][text]
	if ok {
		return trns
	}
	return text
}

func (pd *PageData) Session() *sessions.Session {
	session, err := SessionStore.Get(pd.r, "sparq-session")
	if err != nil {
		panic(err)
	}
	return session
}

func (pd *PageData) CurrentAccount() *model.Account {
	if pd.act != nil {
		return pd.act
	}
	uid, ok := pd.Session().Values["uid"]
	if ok {
		var acct model.Account
		err := db.Database().Get(&acct, `
		select a.*, ap.* from accounts a
		join account_profiles ap on ap.accountid = a.id
		where a.id = ?`, uid)
		if err != nil {
			util.Error("Unable to fetch account", err)
			return nil
		}
		pd.act = &acct
	}
	return pd.act
}

func translations(locale string) map[string]string {
	strs, ok := locales[locale]
	if strs != nil {
		return strs
	}

	if !ok {
		return nil
	}

	if ok {
		// util.Debugf("Booting the %s locale", locale)
		strs := map[string]string{}
		for _, finder := range AssetLookups {
			content, err := finder(fmt.Sprintf("static/locales/%s.yml", locale))
			if err != nil {
				continue
			}

			scn := bufio.NewScanner(bytes.NewReader(content))
			for scn.Scan() {
				kv := strings.Split(scn.Text(), ":")
				if len(kv) == 2 {
					strs[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
				}
			}
		}
		locales[locale] = strs
		return strs
	}

	panic("Shouldn't get here")
}

type webCtx struct {
	locale  string
	prefix  string
	strings map[string]string
}

func defaultCtx() *webCtx {
	return &webCtx{
		locale: "en",
		prefix: "",
	}
}

type ctxType int

var (
	ctxKey ctxType = 1
)

func setCtx(pass http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := r.WithContext(context.WithValue(r.Context(), ctxKey, newCtx(w, r)))
		pass.ServeHTTP(w, req)
	})
}

func t(r *http.Request, word string) string {
	return ctx(r).translate(word)
}

func ctx(r *http.Request) *webCtx {
	return r.Context().Value(ctxKey).(*webCtx)
}

func loggedIn(r *http.Request) bool {
	session, _ := SessionStore.Get(r, "sparq-session")
	return session.Values["username"] != nil
}

func currentNick(r *http.Request) string {
	session, _ := SessionStore.Get(r, "sparq-session")
	return session.Values["username"].(string)
}

func (d *webCtx) translate(str string) string {
	val, ok := d.strings[str]
	if ok {
		return val
	}
	return str
}

func newCtx(w http.ResponseWriter, req *http.Request) *webCtx {
	// set locale via cookie
	localeCookie, _ := req.Cookie("locale")

	var locale string
	if localeCookie != nil {
		locale = localeCookie.Value
	}

	if locale == "" {
		// fall back to browser language
		locale = localeFromHeader(req.Header.Get("Accept-Language"))
	}

	w.Header().Set("Content-Language", locale)
	ctx := defaultCtx()
	ctx.prefix = req.Header.Get("X-Script-Name")
	ctx.strings = translations(locale)

	return ctx
}

func acceptableLanguages(header string) []string {
	langs := []string{}
	pairs := strings.Split(header, ",")
	// we ignore the q weighting and just assume the
	// values are sorted by acceptability
	for idx := range pairs {
		trimmed := strings.Trim(pairs[idx], " ")
		split := strings.Split(trimmed, ";")
		langs = append(langs, strings.ToLower(split[0]))
	}
	return langs
}

func localeFromHeader(value string) string {
	if value == "" {
		return "en"
	}

	langs := acceptableLanguages(value)
	// util.Debugf("A-L: %s %v", value, langs)
	for idx := range langs {
		strs := translations(langs[idx])
		if strs != nil {
			return langs[idx]
		}
	}

	// fallback by checking the language component of any dialect pairs, e.g. "sv-se"
	for idx := range langs {
		pair := strings.Split(langs[idx], "-")
		if len(pair) == 2 {
			baselang := pair[0]
			strs := translations(baselang)
			if strs != nil {
				return baselang
			}
		}
	}

	return "en"
}
