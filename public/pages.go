package public

import (
	"html/template"
	"net/http"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/db"
	"github.com/contribsys/sparq/model"
	"github.com/contribsys/sparq/util"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
)

var (
	pageTemplates = map[string]*template.Template{}
)

func init() {
	// these are the pages which can be rendered
	prepare("index", "profile", "login", "authorize")
}

func prepare(pages ...string) {
	// Include the navigation partial in the template files.
	for _, page := range pages {
		files := []string{
			"base.gotmpl",
			"nav.gotmpl",
			"flashes.gotmpl",
			"footer.gotmpl",
			page + ".gotmpl",
		}

		ts := template.New(page)
		ts.Funcs(template.FuncMap{
			"now":      util.Nows,
			"hostname": func() string { return db.InstanceHostname },
		})
		ts, err := ts.ParseFS(templateFiles, files...)
		if err != nil {
			panic(err)
		}
		pageTemplates[page] = ts
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
	session, err := sparq.SessionStore.Get(pd.r, "sparq-session")
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
		err := db.Database().Get(&acct, "select * from accounts where id = ?", uid)
		if err != nil {
			util.Error("Unable to fetch account", err)
			return nil
		}
		pd.act = &acct
	}
	return pd.act
}

func render(w http.ResponseWriter, r *http.Request, page string, custom any) {
	ts := pageTemplates[page]
	if ts == nil {
		panic(errors.New("No registered page: " + page))
	}
	err := ts.ExecuteTemplate(w, "base", &PageData{r: r, w: w, Locale: "en", Custom: custom})
	if err != nil {
		httpError(w, err, 503)
	}
}
