package model

import (
	"fmt"
	"strconv"
	"time"

	"github.com/contribsys/sparq/oauth2"
)

type OauthClient struct {
	Id           uint64
	Name         string `json:"name"`
	ClientId     string `json:"client_id"`
	Secret       string `json:"client_secret"`
	RedirectUris string `json:"redirect_uri"`
	Website      string `json:"website"`
	AccountId    uint64
	Scopes       string
	CreatedAt    time.Time
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
	return fmt.Sprint(x.AccountId)
}

type OauthToken struct {
	ClientId            string
	AccountId           uint64
	RedirectUri         string
	Scope               string
	Code                string
	CodeChallenge       string
	CodeChallengeMethod string
	CodeCreatedAt       time.Time
	CodeExpiresIn       time.Duration
	Access              string
	AccessCreatedAt     time.Time
	AccessExpiresIn     time.Duration
	Refresh             string
	RefreshCreatedAt    time.Time
	RefreshExpiresIn    time.Duration
	CreatedAt           time.Time
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
	return fmt.Sprint(ot.AccountId)
}

func (ot *OauthToken) SetUserID(ui string) {
	uid, err := strconv.ParseUint(ui, 10, 64)
	if err != nil {
		panic("Unable to parse UserID: " + ui)
	}
	ot.AccountId = uid
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
	return ot.CodeCreatedAt
}

func (ot *OauthToken) SetCodeCreateAt(s time.Time) {
	ot.CodeCreatedAt = s
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
	return ot.AccessCreatedAt
}

func (ot *OauthToken) SetAccessCreateAt(s time.Time) {
	ot.AccessCreatedAt = s
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
	return ot.RefreshCreatedAt
}

func (ot *OauthToken) SetRefreshCreateAt(s time.Time) {
	ot.RefreshCreatedAt = s
}

func (ot *OauthToken) GetRefreshExpiresIn() time.Duration {
	return ot.RefreshExpiresIn
}

func (ot *OauthToken) SetRefreshExpiresIn(s time.Duration) {
	ot.RefreshExpiresIn = s
}
