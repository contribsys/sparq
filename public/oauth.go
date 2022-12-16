package public

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/db"
	"github.com/contribsys/sparq/model"
	"github.com/contribsys/sparq/oauth2"
	"github.com/contribsys/sparq/util"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// This is the code necessary to integrate go-oauth/oauth2 into Sparq.

type SqliteOauthStore struct {
	DB *sqlx.DB
}

func (scs *SqliteOauthStore) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	var client model.OauthClient
	util.Infof("Finding OAuth client %s", id)
	row := scs.DB.QueryRowxContext(ctx, "select * from oauth_clients where ClientId = ?", id)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "wrapped")
	}
	err := row.StructScan(&client)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to struct")
	}
	return &client, nil
}

func (scs *SqliteOauthStore) Set(ctx context.Context, id string, cli oauth2.ClientInfo) error {
	return errors.New("Set not implemented")
}

func (scs *SqliteOauthStore) Delete(ctx context.Context, id string) error {
	return wrap(scs.DB.ExecContext(ctx, "delete from oauth_clients where ClientId = ?", id))
}

func (scs *SqliteOauthStore) Create(ctx context.Context, info oauth2.TokenInfo) error {
	fmt.Printf("Created OAuth token: %+v\n", info)
	_, err := scs.DB.ExecContext(ctx, `INSERT INTO oauth_tokens (
			ClientId, UserId, RedirectUri, Scope, CodeChallenge,
			Code, CodeCreatedAt, CodeExpiresIn,
			Access, AccessCreatedAt, AccessExpiresIn,
			Refresh, RefreshCreatedAt, RefreshExpiresIn)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		info.GetClientID(), info.GetUserID(), info.GetRedirectURI(), info.GetScope(), info.GetCodeChallenge(),
		info.GetCode(), info.GetCodeCreateAt(), info.GetCodeExpiresIn(),
		info.GetAccess(), info.GetAccessCreateAt(), info.GetAccessExpiresIn(),
		info.GetRefresh(), info.GetRefreshCreateAt(), info.GetRefreshExpiresIn())
	if err != nil {
		return errors.Wrap(err, "insert")
	}
	return nil
}
func wrap(_ any, err error) error {
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return errors.Wrap(err, "Wrap")
	}
	return nil
}
func (scs *SqliteOauthStore) RemoveByCode(ctx context.Context, code string) error {
	return wrap(scs.DB.ExecContext(ctx, "delete from oauth_tokens where code = ?", code))
}
func (scs *SqliteOauthStore) RemoveByAccess(ctx context.Context, access string) error {
	return wrap(scs.DB.ExecContext(ctx, "delete from oauth_tokens where access = ?", access))
}
func (scs *SqliteOauthStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	return wrap(scs.DB.ExecContext(ctx, "delete from oauth_tokens where refresh = ?", refresh))
}
func (scs *SqliteOauthStore) getBy(ctx context.Context, name, value string) (oauth2.TokenInfo, error) {
	fmt.Printf("get %s %s\n", name, value)
	var token model.OauthToken
	err := scs.DB.QueryRowxContext(ctx, fmt.Sprintf("select * from oauth_tokens where %s = ?", name), value).StructScan(&token)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "getBy")
		}
		return nil, nil
	}
	return &token, nil
}
func (scs *SqliteOauthStore) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {
	return scs.getBy(ctx, "code", code)
}
func (scs *SqliteOauthStore) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	return scs.getBy(ctx, "access", access)
}
func (scs *SqliteOauthStore) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	return scs.getBy(ctx, "refresh", refresh)
}

func BearerAuth(store oauth2.TokenStore) func(http.Handler) http.Handler {
	return func(pass http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			prefix := "Bearer "
			token := ""

			if auth != "" && strings.HasPrefix(auth, prefix) {
				token = auth[len(prefix):]
				ti, err := store.GetByAccess(r.Context(), token)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if ti == nil || ti.GetAccessCreateAt().Add(ti.GetAccessExpiresIn()).Before(time.Now()) {
					// access token has expired
					w.Header().Add("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte(`{ "error": "invalid_token", "error_description": "The access token expired" }`))
					return
				}
				session, _ := sessionStore.Get(r, "sparq-session")
				val, ok := session.Values["uid"]
				if !ok {
					session.Values["uid"] = ti.GetUserID()
				} else if val != ti.GetUserID() {
					fmt.Println("Bearer: ", val, ti.GetUserID())
				}
			}
			pass.ServeHTTP(w, r)
		})
	}
}

func IntegrateOauth(s sparq.Server, root *mux.Router) mux.MiddlewareFunc {
	manager := oauth2.NewDefaultManager()
	store := &SqliteOauthStore{db.Database()}
	manager.MapTokenStorage(store)
	manager.MapClientStorage(store)

	sc := &oauth2.ConfigConfig{
		TokenType:             "Bearer",
		AllowGetAccessRequest: false,
		AllowedResponseTypes:  []oauth2.ResponseType{oauth2.CodeType},
		AllowedGrantTypes: []oauth2.GrantType{
			oauth2.AuthorizationCode,
		},
		AllowedCodeChallengeMethods: []oauth2.CodeChallengeMethod{
			oauth2.CodeChallengePlain, oauth2.CodeChallengeS256},
	}
	srv := oauth2.NewServer(sc, manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(oauth2.ClientFormHandler)
	srv.SetInternalErrorHandler(func(err error) (re *oauth2.Response) {
		fmt.Println("internal oauth error:", err.Error())
		if er, ok := err.(stackTracer); ok {
			for _, f := range er.StackTrace() {
				fmt.Printf("%+s:%d\n", f, f)
			}
		}
		return
	})
	srv.SetResponseErrorHandler(func(re *oauth2.Response) {
		fmt.Println("oauth response error:", re.Error.Error())
		if err, ok := re.Error.(stackTracer); ok {
			for _, f := range err.StackTrace() {
				fmt.Printf("%+s:%d\n", f, f)
			}
		}
	})
	root.HandleFunc("/oauth/authorize", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		session, _ := sessionStore.Get(r, "sparq-session")
		v, ok := session.Values["returnForm"]
		if ok {
			r.Form = v.(url.Values)
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		clientId := r.Form.Get("client_id")
		util.Infof("Authorizing client %s", clientId)
		client, err := srv.Manager.GetClient(r.Context(), clientId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		oc := client.(*model.OauthClient)

		// Denied, bye!
		if r.Form.Get("Deny") == "1" {
			delete(session.Values, "returnForm")
			_ = session.Save(r, w)
			err := store.Delete(r.Context(), r.Form.Get("client_id"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			u, err := url.Parse(oc.RedirectUris)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			q := u.Query()
			q.Add("error", "access_denied")
			q.Add("error_description", "Access was denied")
			u.RawQuery = q.Encode()
			http.Redirect(w, r, u.String(), 302)
			return
		}

		if r.Method == "POST" && r.Form.Get("Approve") == "1" {
			delete(session.Values, "returnForm")
			_ = session.Save(r, w)
			err := srv.HandleAuthorizeRequest(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		ego_oauth_authorize(w, r, oc)
	}))

	root.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") == "application/json" {
			bytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			data := map[string]string{}
			err = json.Unmarshal(bytes, &data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			r.Form = make(url.Values)
			for k, v := range data {
				r.Form.Set(k, v)
			}
		}
		err := srv.HandleTokenRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})
	srv.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (string, error) {
		session, _ := sessionStore.Get(r, "sparq-session")
		uid := session.Values["uid"]
		return fmt.Sprint(uid), nil
	})
	return BearerAuth(store)
}