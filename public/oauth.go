package public

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/contribsys/sparq"
	"github.com/contribsys/sparq/db"
	"github.com/contribsys/sparq/model"
	"github.com/contribsys/sparq/oauth2"
	"github.com/contribsys/sparq/util"
	"github.com/contribsys/sparq/webutil"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

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
	// fmt.Printf("Created OAuth token: %+v\n", info)
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

func IntegrateOauth(s sparq.Server, root *mux.Router) *SqliteOauthStore {
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
		util.DumpError(err)
		return
	})
	srv.SetResponseErrorHandler(func(re *oauth2.Response) {
		util.DumpError(re.Error)
	})
	root.HandleFunc("/oauth/authorize", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		session, _ := webutil.SessionStore.Get(r, "sparq-session")
		v, ok := session.Values["returnForm"]
		if ok {
			r.Form = v.(url.Values)
		}

		err := r.ParseForm()
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}

		clientId := r.Form.Get("client_id")
		util.Infof("Authorizing client %s", clientId)
		client, err := srv.Manager.GetClient(r.Context(), clientId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				httpError(w, errors.New("Client authorization expired, please start over."), http.StatusBadRequest)
			} else {
				httpError(w, err, http.StatusBadRequest)
			}
			return
		}
		oc := client.(*model.OauthClient)

		// Denied, bye!
		if r.Form.Get("Deny") == "1" {
			delete(session.Values, "returnForm")
			_ = session.Save(r, w)
			err := store.Delete(r.Context(), r.Form.Get("client_id"))
			if err != nil {
				httpError(w, err, http.StatusBadRequest)
				return
			}
			if oc.RedirectUris == "urn:ietf:wg:oauth:2.0:oob" {
				util.Debugf("Rejected OOB authorization")
				http.Redirect(w, r, "/home", http.StatusFound)
				return
			}
			u, err := url.Parse(oc.RedirectUris)
			if err != nil {
				util.Debugf("Rejected OOB authorization")
				httpError(w, err, http.StatusBadRequest)
				return
			}
			q := u.Query()
			q.Add("error", "access_denied")
			q.Add("error_description", "Access was denied")
			u.RawQuery = q.Encode()
			http.Redirect(w, r, u.String(), http.StatusFound)
			return
		}

		if r.Method == "POST" && r.Form.Get("Approve") == "1" {
			delete(session.Values, "returnForm")
			_ = session.Save(r, w)
			code, err := srv.HandleAuthorizeRequest(w, r)
			if err != nil {
				httpError(w, err, http.StatusBadRequest)
				return
			}
			if code != "" {
				session.AddFlash(fmt.Sprintf("Your authorization code is %s", code))
			}
		}
		render(w, r, "authorize", oc)
	}))

	root.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") == "application/json" {
			bytes, err := io.ReadAll(r.Body)
			if err != nil {
				httpError(w, err, http.StatusBadRequest)
				return
			}
			data := map[string]string{}
			err = json.Unmarshal(bytes, &data)
			if err != nil {
				httpError(w, err, http.StatusBadRequest)
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
		session, _ := webutil.SessionStore.Get(r, "sparq-session")
		uid := session.Values["uid"]
		return fmt.Sprint(uid), nil
	})
	return store
}

func Auth(store oauth2.TokenStore) func(http.Handler) http.Handler {
	return func(pass http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			webctx := webutil.Ctx(r)
			code := webctx.BearerCode
			if code != "" {
				ti, err := store.GetByAccess(r.Context(), code)
				if err != nil {
					httpError(w, err, http.StatusInternalServerError)
					return
				}
				if ti == nil || ti.GetAccessCreateAt().Add(ti.GetAccessExpiresIn()).Before(time.Now()) {
					// access token has expired
					w.Header().Add("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte(`{ "error": "invalid_token", "error_description": "The access token expired" }`))
					return
				}
				webctx.CurrentUserID = ti.GetUserID()
			}

			pass.ServeHTTP(w, r)
		})
	}
}

func httpError(w http.ResponseWriter, err error, code int) {
	webutil.HttpError(w, err, code)
}
