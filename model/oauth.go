package model

import (
	"fmt"
	"strconv"
	"time"

	"github.com/contribsys/sparq/oauth2"
)

type OauthClient struct {
	Name         string    `db:"Name" json:"name"`
	ClientId     string    `db:"ClientId" json:"client_id"`
	Secret       string    `db:"Secret" json:"client_secret"`
	RedirectUris string    `db:"RedirectUris" json:"redirect_uri"`
	Website      string    `db:"Website" json:"website"`
	UserId       *int      `db:"UserId"`
	Scopes       string    `db:"Scopes"`
	CreatedAt    time.Time `db:"CreatedAt"`
}

func (x *OauthClient) GetID() string {
	return x.ClientId
}
func (x *OauthClient) GetSecret() string {
	return x.Secret
}
func (x *OauthClient) GetDomain() string {
	return x.Website
}
func (x *OauthClient) GetUserID() string {
	return fmt.Sprint(x.UserId)
}

type OauthToken struct {
	ClientId            string
	UserId              int64
	RedirectUri         string
	Scope               string
	Code                string
	CodeChallenge       string
	CodeChallengeMethod string
	CodeCreateAt        time.Time
	CodeExpiresIn       time.Duration
	Access              string
	AccessCreateAt      time.Time
	AccessExpiresIn     time.Duration
	Refresh             string
	RefreshCreateAt     time.Time
	RefreshExpiresIn    time.Duration
}

func (ot *OauthToken) New() oauth2.TokenInfo {
	return &OauthToken{}
}

func (ot *OauthToken) GetClientID() string {
	return ot.ClientId
}

func (ot *OauthToken) SetClientID(ci string) {
	ot.ClientId = ci
}

func (ot *OauthToken) GetUserID() string {
	return fmt.Sprint(ot.UserId)
}

func (ot *OauthToken) SetUserID(ui string) {
	uid, err := strconv.ParseInt(ui, 10, 64)
	if err != nil {
		panic("Unable to parse UserID: " + ui)
	}
	ot.UserId = uid
}

func (ot *OauthToken) GetRedirectURI() string {
	return ot.RedirectUri
}

func (ot *OauthToken) SetRedirectURI(ru string) {
	ot.RedirectUri = ru
}

func (ot *OauthToken) GetScope() string {
	return ot.Scope
}

func (ot *OauthToken) SetScope(s string) {
	ot.Scope = s
}

func (ot *OauthToken) GetCode() string {
	return ot.Code
}

func (ot *OauthToken) SetCode(s string) {
	ot.Code = s
}

func (ot *OauthToken) GetCodeCreateAt() time.Time {
	return ot.CodeCreateAt
}

func (ot *OauthToken) SetCodeCreateAt(s time.Time) {
	ot.CodeCreateAt = s
}

func (ot *OauthToken) GetCodeExpiresIn() time.Duration {
	return ot.CodeExpiresIn
}

func (ot *OauthToken) SetCodeExpiresIn(s time.Duration) {
	ot.CodeExpiresIn = s
}

func (ot *OauthToken) GetCodeChallenge() string {
	return ot.CodeChallenge
}

func (ot *OauthToken) SetCodeChallenge(s string) {
	ot.CodeChallenge = s
}

func (ot *OauthToken) GetCodeChallengeMethod() oauth2.CodeChallengeMethod {
	return oauth2.CodeChallengeS256
}

func (ot *OauthToken) SetCodeChallengeMethod(ccm oauth2.CodeChallengeMethod) {
	if ccm != oauth2.CodeChallengeS256 {
		panic("What? " + ccm)
	}
}

func (ot *OauthToken) GetAccess() string {
	return ot.Access
}

func (ot *OauthToken) SetAccess(s string) {
	ot.Access = s
}

func (ot *OauthToken) GetAccessCreateAt() time.Time {
	return ot.AccessCreateAt
}

func (ot *OauthToken) SetAccessCreateAt(s time.Time) {
	ot.AccessCreateAt = s
}

func (ot *OauthToken) GetAccessExpiresIn() time.Duration {
	return ot.AccessExpiresIn
}

func (ot *OauthToken) SetAccessExpiresIn(s time.Duration) {
	ot.AccessExpiresIn = s
}

func (ot *OauthToken) GetRefresh() string {
	return ot.Refresh
}

func (ot *OauthToken) SetRefresh(s string) {
	ot.Refresh = s
}

func (ot *OauthToken) GetRefreshCreateAt() time.Time {
	return ot.RefreshCreateAt
}

func (ot *OauthToken) SetRefreshCreateAt(s time.Time) {
	ot.RefreshCreateAt = s
}

func (ot *OauthToken) GetRefreshExpiresIn() time.Duration {
	return ot.RefreshExpiresIn
}

func (ot *OauthToken) SetRefreshExpiresIn(s time.Duration) {
	ot.RefreshExpiresIn = s
}

type OauthGrant struct {
	ClientId    string
	Token       string
	ExpiresIn   int64
	Scopes      string
	UserId      int64
	RedirectUri string
	RevokedAt   time.Time
	CreatedAt   time.Time
}
