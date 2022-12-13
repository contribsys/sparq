package public

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"net/http"
	"strings"
)

var (
	//go:embed static/*.css static/*.js static/*.png
	staticFiles embed.FS

	//go:embed static/locales/*.yml
	localeFiles embed.FS
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
	files, err := localeFiles.ReadDir("static/locales")
	if err != nil {
		panic(err)
	}
	for idx := range files {
		name := strings.Split(files[idx].Name(), ".")[0]
		locales[name] = nil
	}
	// util.Debugf("Initialized %d locales", len(files))
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

func t(r *http.Request, word string) string {
	return ctx(r).translate(word)
}

func ctx(r *http.Request) *webCtx {
	return r.Context().Value(ctxKey).(*webCtx)
}

func loggedIn(r *http.Request) bool {
	session, _ := sessionStore.Get(r, "sparq-session")
	return session.Values["username"] != nil
}

func currentNick(r *http.Request) string {
	session, _ := sessionStore.Get(r, "sparq-session")
	return session.Values["username"].(string)
}

func flashes(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "sparq-session")
	if flashy := session.Flashes(); len(flashy) > 0 {
		ego_flashes(w, r, flashy)
		_ = session.Save(r, w)
	}
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
