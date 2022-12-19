package clientapi

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"text/template"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/model"
	"github.com/contribsys/sparq/util"
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
		_, err = svr.DB().ExecContext(r.Context(), `insert into oauth_clients (
			Name, ClientId, Secret, RedirectUris, Scopes, Website) values (
				?, ?, ?, ?, ?, ?
			)`, hash["client_name"], clientId, clientSecret,
			hash["redirect_uris"], hash["scopes"], hash["website"])
		if err != nil {
			util.Error("oauth_clients reg", err)
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
	var admin model.Account
	err := svr.DB().Get(&admin, "select * from accounts where id = 1")
	if err != nil {
		panic(err.Error())
	}
	err = svr.DB().Get(&admin, "select * from account_profiles where accountid = 1")
	if err != nil {
		panic(err.Error())
	}
	fields := [][]interface{}{}
	rows, err := svr.DB().Queryx("select * from account_fields where accountid = 1")
	if err != nil {
		panic(err.Error())
	}
	for rows.Next() {
		cols, err := rows.SliceScan()
		if err != nil {
			panic(err.Error())
		}
		fields = append(fields, cols)
	}
	// fmt.Printf("Admin: %+v %+v\n", admin, fields)
	inst := &Instance{
		Description:     "The littlest Sparq can ignite a bonfire",
		SoftwareName:    sparq.Name,
		SoftwareVersion: sparq.Version,
		Domain:          svr.Hostname(),
		Admin:           admin,
		Fields:          fields,
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
	Thumbnail       string
	Domain          string
	Admin           model.Account
	Fields          [][]interface{}
}

type User struct {
}

var (
	instanceTemplate *template.Template
	instanceText     string = `
	{
		"uri": "{{.Domain}}",
		"title": "{{.Domain}}",
		"short_description": "{{.Description}}",
		"description": "{{.Description}}",
		"email": "admin@{{.Domain}}",
		"name": "{{.SoftwareName}}",
		"version": "{{.SoftwareVersion}}",
		"urls": { "streaming_api": "wss://{{.Domain}}" },
		"stats": { "user_count": 0, "status_count": 0, "domain_count": 0 },
		"thumbnail": "{{.Thumbnail}}",
		"languages": [ "en" ],
		"registrations": true,
		"approval_required": true,
		"invites_enabled": false,
		"configuration": {
			"accounts": { "max_featured_tags": 10 },
			"statuses": { "max_characters": 500, "max_media_attachments": 4, "characters_reserved_per_url": 23 },
			"media_attachments": {
				"supported_mime_types": [
					"image/jpeg",
					"image/png",
					"image/gif",
					"image/heic",
					"image/heif",
					"image/webp",
					"image/avif",
					"video/webm",
					"video/mp4",
					"video/quicktime",
					"video/ogg",
					"audio/wave",
					"audio/wav",
					"audio/x-wav",
					"audio/x-pn-wave",
					"audio/vnd.wave",
					"audio/ogg",
					"audio/vorbis",
					"audio/mpeg",
					"audio/mp3",
					"audio/webm",
					"audio/flac",
					"audio/aac",
					"audio/m4a",
					"audio/x-m4a",
					"audio/mp4",
					"audio/3gpp",
					"video/x-ms-asf"
				],
				"image_size_limit": 10485760,
				"image_matrix_limit": 16777216,
				"video_size_limit": 41943040,
				"video_frame_rate_limit": 60,
				"video_matrix_limit": 2304000
			},
			"polls": {
				"max_options": 4,
				"max_characters_per_option": 50,
				"min_expiration": 300,
				"max_expiration": 2629746
			}
		},
		"contact_account": {
			"id": "1",
			"username": "admin",
			"acct": "admin",
			"display_name": "Administrator",
			"locked": false,
			"bot": false,
			"discoverable": true,
			"group": false,
			"created_at": "{{.Admin.CreatedAt}}",
			"note": "{{.Admin.AccountProfile.Note}}",
			"url": "https://{{.Domain}}/@admin",
			"avatar": "https://{{.Domain}}{{.Admin.AccountProfile.Avatar}}",
			"avatar_static": "https://{{.Domain}}{{.Admin.AccountProfile.Avatar}}",
			"header": "https://{{.Domain}}{{.Admin.AccountProfile.Header}}",
			"header_static": "https://{{.Domain}}{{.Admin.AccountProfile.Header}}",
			"followers_count": 0,
			"following_count": 0,
			"statuses_count": 0,
			"last_status_at": "2022-12-17",
			"noindex": false,
			"emojis": [],
			"fields": [
				{{ range $idx, $field := .Fields -}}
					{{if $idx}},{{end}}
					{ "name": "{{index $field 1}}", "value": "{{index $field 2}}", "verified_at": {{- with index $field 3}} "{{.}}" {{- else -}}null{{ end -}} }
				{{- end }}
			]
		},
		"rules": [
			{ "id": "2",
				"text": "No bots\r\n\r\n... unless they're cute, funny, or useful. Bots should avoid posting to the public timeline.\r\n" },
			{ "id": "3",
				"text": "No pornography or nudity, or something that would be considered \"NSFW\" in a workplace, even behind a NSFW warning. In particular, no sexual depictions of children." },
			{ "id": "4",
				"text": "No gore or graphic violence. Again, if you wouldn't post it at work, you definitely shouldn't be posting it here." },
			{ "id": "5",
				"text": "Use content warnings. Content that is borderline NSFW or could be construed as NSFW at a glance should be put behind a CW (Content Warning)" },
			{ "id": "6",
				"text": "No racism. This should go without saying, but unfortunately it doesn't." },
			{ "id": "7",
				"text": "No sexism\r\n\r\n... or pretty much any other -ism" },
			{ "id": "8",
				"text": "No discrimination. Any discrimination based on gender, sexual minority, sexual orientation, disability, physical appearance, body size, race, ethnicity, religion (or lack thereof), or national origin, will result in your account and content being removed." },
			{ "id": "9",
				"text": "No xenophobia. This includes violent nationalism." },
			{ "id": "11",
				"text": "No holocaust denialism. No Nazi symbolism, promotion of National Socialism, or anything that is illegal in the European Union" },
			{ "id": "12",
				"text": "No stalking or harassment of any kind" },
			{ "id": "13",
				"text": "No fake accounts, even celebrity joke accounts.\r\n\r\nNote this rule does not necessarily apply to remote accounts." },
			{ "id": "14",
				"text": "No \"follow bots.\" We block these and our instance is on a \"no bots\" list" },
			{ "id": "15",
				"text": "No advertising or excessive promotion. By far the biggest problem we've had running this server are spam accounts that sign up only to promote some company (often nothing even to do with Ruby), and so we take a fairly strong stance against letting such accounts even register on the server now."
			}
		]
	}
`
)
