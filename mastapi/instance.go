package mastapi

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"text/template"

	"github.com/contribsys/sparq"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func jsonHashBody(r *http.Request) (map[string]string, error) {
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, errors.New("Unexpected Content-Type")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read body")
	}
	result := map[string]string{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse JSON")
	}
	return result, nil
}

// /api/v1/apps
func appsHandler(svr sparq.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Add("Cache-Control", "public, max-age=3600")
			w.WriteHeader(204)
			return
		}
		if r.Method != "POST" {
			http.Error(w, "Bad method", http.StatusBadRequest)
			return
		}

		// {"client_name":"Pinafore",
		// "redirect_uris":"https://pinafore.social/settings/instances/add",
		// "scopes":"read write follow push",
		// "website":"https://pinafore.social"}
		hash, err := jsonHashBody(r)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		for _, v := range hash {
			if len(v) > 500 {
				http.Error(w, "Input too long", http.StatusBadRequest)
				return
			}
		}

		clientId := uuid.NewString()
		secret := make([]byte, 16)
		_, err = io.ReadFull(rand.Reader, secret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		clientSecret := hex.EncodeToString(secret)
		// save new OAuth2 application record with client_id and client_secret
		_, err = svr.DB().ExecContext(r.Context(), `insert into oauth_apps (
			ClientName, ClientId, ClientSecret, RedirectUris, Scopes, Website) values (
				?, ?, ?, ?, ?, ?
			)`, hash["client_name"], clientId, clientSecret,
			hash["redirect_uris"], hash["scopes"], hash["website"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(map[string]interface{}{
			"client_id":     clientId,
			"client_secret": clientSecret,
			"redirect_uri":  hash["redirect_uris"],
			"name":          hash["client_name"],
			"website":       "http://localhost:9494",
			"vapid_key":     nil,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write(b)
	}
}

// /api/v1/instance
func instanceHandler(svr sparq.Server) http.HandlerFunc {
	instanceTemplate = template.Must(template.New("instance").Parse(instanceText))
	admin := map[string]interface{}{}
	err := svr.DB().QueryRowx("select * from users where id = 1").MapScan(admin)
	if err != nil {
		panic(err.Error())
	}
	inst := &Instance{
		Description:     "The littlest Sparq can ignite a bonfire",
		SoftwareName:    sparq.Name,
		SoftwareVersion: sparq.Version,
		Domain:          svr.Hostname(),
		Admin:           admin,
	}
	buf := new(bytes.Buffer)
	err = instanceTemplate.Execute(buf, inst)
	if err != nil {
		panic(err.Error())
	}
	code := 200
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(code)
		_, _ = w.Write(buf.Bytes())
	}
}

type Instance struct {
	Description     string
	SoftwareName    string
	SoftwareVersion string
	Domain          string
	Admin           map[string]interface{}
}

type User struct {
}

var (
	instanceTemplate *template.Template
	instanceText     string = `{
		"domain": "{{.Domain}}",
		"title": "{{.SoftwareName}}",
		"version": "{{.SoftwareVersion}}",
		"source_url": "https://github.com/contribsys/sparq",
		"description": "{{.Description}}",
		"usage": {
			"users": {
				"active_month": 0
			}
		},
		"thumbnail": {
			"url": "https://{{.Domain}}/static/logo.png"
		},
		"languages": ["en"],
		"configuration": {
		},
		"registrations": {
			"enabled": true,
			"approval_required": false,
			"message": null
		},
		"contact": {
			"email": "{{.Admin.Email}}",
			"account": {
				"id": "{{.Admin.Id}}",
				"username": "{{.Admin.Nick}}",
				"acct": "{{.Admin.Nick}}",
				"display_name": "{{.Admin.FullName}}",
				"created_at": "{{.Admin.CreatedAt}}",
				"note": "TODO",
				"url": "https://{{.Domain}}/@{{.Admin.Nick}}",
				"avatar": "https://{{.Domain}}/static/logo.png",
				"avatar_static": "https://{{.Domain}}/static/logo.png",
				"header": "https://{{.Domain}}/static/logo.png",
				"header_static": "https://{{.Domain}}/static/logo.png",
				"followers_count": 0,
				"following_count": 0,
				"statuses_count": 0
			}
		},
		"rules": [
			{ "id": "1",
				"text": "Sexually explicit or violent media must be marked as sensitive when posting"
			}, { "id": "2",
				"text": "No racism, sexism, homophobia, transphobia, xenophobia, or casteism"
			}, { "id": "3",
				"text": "No incitement of violence or promotion of violent ideologies"
			}, { "id": "4",
				"text": "No harassment, dogpiling or doxxing of other users"
			}, { "id": "5",
				"text": "No content illegal in Germany"
			}, {"id": "6",
				"text": "Do not share intentionally false or misleading information"
			}
		]
	}
`
)
