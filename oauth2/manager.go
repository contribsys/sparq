package oauth2

import (
	"context"
	"time"
)

// NewDefaultManager create to default authorization management instance
func NewDefaultManager() *manager {
	m := NewManager()
	// default implementation
	m.MapAuthorizeGenerate(NewAuthorizeGenerate())
	m.MapAccessGenerate(NewAccessGenerate())

	return m
}

// NewManager create to authorization management instance
func NewManager() *manager {
	return &manager{
		gtcfg:       make(map[GrantType]*Config),
		validateURI: DefaultValidateURI,
	}
}

// Manager provide authorization management
type manager struct {
	codeExp           time.Duration
	gtcfg             map[GrantType]*Config
	rcfg              *RefreshingConfig
	validateURI       ValidateURIHandler
	authorizeGenerate AuthorizeGenerate
	accessGenerate    AccessGenerate
	tokenStore        TokenStore
	clientStore       ClientStore
}

// get grant type config
func (m *manager) grantConfig(gt GrantType) *Config {
	if c, ok := m.gtcfg[gt]; ok && c != nil {
		return c
	}
	switch gt {
	case AuthorizationCode:
		return DefaultAuthorizeCodeTokenCfg
	case Implicit:
		return DefaultImplicitTokenCfg
	case PasswordCredentials:
		return DefaultPasswordTokenCfg
	case ClientCredentials:
		return DefaultClientTokenCfg
	}
	return &Config{}
}

// SetAuthorizeCodeExp set the authorization code expiration time
func (m *manager) SetAuthorizeCodeExp(exp time.Duration) {
	m.codeExp = exp
}

// SetAuthorizeCodeTokenCfg set the authorization code grant token config
func (m *manager) SetAuthorizeCodeTokenCfg(cfg *Config) {
	m.gtcfg[AuthorizationCode] = cfg
}

// SetImplicitTokenCfg set the implicit grant token config
func (m *manager) SetImplicitTokenCfg(cfg *Config) {
	m.gtcfg[Implicit] = cfg
}

// SetPasswordTokenCfg set the password grant token config
func (m *manager) SetPasswordTokenCfg(cfg *Config) {
	m.gtcfg[PasswordCredentials] = cfg
}

// SetClientTokenCfg set the client grant token config
func (m *manager) SetClientTokenCfg(cfg *Config) {
	m.gtcfg[ClientCredentials] = cfg
}

// SetRefreshTokenCfg set the refreshing token config
func (m *manager) SetRefreshTokenCfg(cfg *RefreshingConfig) {
	m.rcfg = cfg
}

// SetValidateURIHandler set the validates that RedirectURI is contained in baseURI
func (m *manager) SetValidateURIHandler(handler ValidateURIHandler) {
	m.validateURI = handler
}

// MapAuthorizeGenerate mapping the authorize code generate interface
func (m *manager) MapAuthorizeGenerate(gen AuthorizeGenerate) {
	m.authorizeGenerate = gen
}

// MapAccessGenerate mapping the access token generate interface
func (m *manager) MapAccessGenerate(gen AccessGenerate) {
	m.accessGenerate = gen
}

// MapClientStorage mapping the client store interface
func (m *manager) MapClientStorage(stor ClientStore) {
	m.clientStore = stor
}

// MustClientStorage mandatory mapping the client store interface
func (m *manager) MustClientStorage(stor ClientStore, err error) {
	if err != nil {
		panic(err.Error())
	}
	m.clientStore = stor
}

// MapTokenStorage mapping the token store interface
func (m *manager) MapTokenStorage(stor TokenStore) {
	m.tokenStore = stor
}

// MustTokenStorage mandatory mapping the token store interface
func (m *manager) MustTokenStorage(stor TokenStore, err error) {
	if err != nil {
		panic(err)
	}
	m.tokenStore = stor
}

// GetClient get the client information
func (m *manager) GetClient(ctx context.Context, clientID string) (cli ClientInfo, err error) {
	cli, err = m.clientStore.GetByID(ctx, clientID)
	if err != nil {
		return
	} else if cli == nil {
		err = ErrInvalidClient
	}
	return
}

// GenerateAuthToken generate the authorization token(code)
func (m *manager) GenerateAuthToken(ctx context.Context, rt ResponseType, tgr *TokenGenerateRequest) (TokenInfo, error) {
	cli, err := m.GetClient(ctx, tgr.ClientID)
	if err != nil {
		return nil, err
	} else if tgr.RedirectURI != "" {
		if err := m.validateURI(cli.GetDomain(), tgr.RedirectURI); err != nil {
			return nil, err
		}
	}

	ti := NewToken()
	ti.SetClientID(tgr.ClientID)
	ti.SetUserID(tgr.UserID)
	ti.SetRedirectURI(tgr.RedirectURI)
	ti.SetScope(tgr.Scope)

	createAt := time.Now()
	td := &GenerateBasic{
		Client:    cli,
		UserID:    tgr.UserID,
		CreateAt:  createAt,
		TokenInfo: ti,
		Request:   tgr.Request,
	}
	switch rt {
	case CodeType:
		codeExp := m.codeExp
		if codeExp == 0 {
			codeExp = DefaultCodeExp
		}
		ti.SetCodeCreateAt(createAt)
		ti.SetCodeExpiresIn(codeExp)
		if exp := tgr.AccessTokenExp; exp > 0 {
			ti.SetAccessExpiresIn(exp)
		}
		if tgr.CodeChallenge != "" {
			ti.SetCodeChallenge(tgr.CodeChallenge)
			ti.SetCodeChallengeMethod(tgr.CodeChallengeMethod)
		}

		tv, err := m.authorizeGenerate.Token(ctx, td)
		if err != nil {
			return nil, err
		}
		ti.SetCode(tv)
	case TokenType:
		// set access token expires
		icfg := m.grantConfig(Implicit)
		aexp := icfg.AccessTokenExp
		if exp := tgr.AccessTokenExp; exp > 0 {
			aexp = exp
		}
		ti.SetAccessCreateAt(createAt)
		ti.SetAccessExpiresIn(aexp)

		if icfg.IsGenerateRefresh {
			ti.SetRefreshCreateAt(createAt)
			ti.SetRefreshExpiresIn(icfg.RefreshTokenExp)
		}

		tv, rv, err := m.accessGenerate.Token(ctx, td.Client.GetID(), td.UserID, td.CreateAt, icfg.IsGenerateRefresh)
		if err != nil {
			return nil, err
		}
		ti.SetAccess(tv)

		if rv != "" {
			ti.SetRefresh(rv)
		}
	}

	err = m.tokenStore.Create(ctx, ti)
	if err != nil {
		return nil, err
	}
	return ti, nil
}

// get authorization code data
func (m *manager) getAuthorizationCode(ctx context.Context, code string) (TokenInfo, error) {
	ti, err := m.tokenStore.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	} else if ti == nil || ti.GetCode() != code || ti.GetCodeCreateAt().Add(ti.GetCodeExpiresIn()).Before(time.Now()) {
		return nil, ErrInvalidAuthorizeCode
	}
	return ti, nil
}

// delete authorization code data
func (m *manager) delAuthorizationCode(ctx context.Context, code string) error {
	return m.tokenStore.RemoveByCode(ctx, code)
}

// get and delete authorization code data
func (m *manager) getAndDelAuthorizationCode(ctx context.Context, tgr *TokenGenerateRequest) (TokenInfo, error) {
	code := tgr.Code
	ti, err := m.getAuthorizationCode(ctx, code)
	if err != nil {
		return nil, err
	} else if ti.GetClientID() != tgr.ClientID {
		return nil, ErrInvalidAuthorizeCode
	} else if codeURI := ti.GetRedirectURI(); codeURI != "" && codeURI != tgr.RedirectURI {
		return nil, ErrInvalidAuthorizeCode
	}

	err = m.delAuthorizationCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return ti, nil
}

func (m *manager) validateCodeChallenge(ti TokenInfo, ver string) error {
	cc := ti.GetCodeChallenge()
	// early return
	if cc == "" && ver == "" {
		return nil
	}
	if cc == "" {
		return ErrMissingCodeVerifier
	}
	if ver == "" {
		return ErrMissingCodeVerifier
	}
	ccm := ti.GetCodeChallengeMethod()
	if ccm.String() == "" {
		ccm = CodeChallengePlain
	}
	if !ccm.Validate(cc, ver) {
		return ErrInvalidCodeChallenge
	}
	return nil
}

// GenerateAccessToken generate the access token
func (m *manager) GenerateAccessToken(ctx context.Context, gt GrantType, tgr *TokenGenerateRequest) (TokenInfo, error) {
	cli, err := m.GetClient(ctx, tgr.ClientID)
	if err != nil {
		return nil, err
	}
	if cliPass, ok := cli.(ClientPasswordVerifier); ok {
		if !cliPass.VerifyPassword(tgr.ClientSecret) {
			return nil, ErrInvalidClient
		}
	} else if len(cli.GetSecret()) > 0 && tgr.ClientSecret != cli.GetSecret() {
		return nil, ErrInvalidClient
	}
	if tgr.RedirectURI != "" {
		if err := m.validateURI(cli.GetDomain(), tgr.RedirectURI); err != nil {
			return nil, err
		}
	}

	if gt == AuthorizationCode {
		ti, err := m.getAndDelAuthorizationCode(ctx, tgr)
		if err != nil {
			return nil, err
		}
		if err := m.validateCodeChallenge(ti, tgr.CodeVerifier); err != nil {
			return nil, err
		}
		tgr.UserID = ti.GetUserID()
		tgr.Scope = ti.GetScope()
		if exp := ti.GetAccessExpiresIn(); exp > 0 {
			tgr.AccessTokenExp = exp
		}
	}

	ti := NewToken()
	ti.SetClientID(tgr.ClientID)
	ti.SetUserID(tgr.UserID)
	ti.SetRedirectURI(tgr.RedirectURI)
	ti.SetScope(tgr.Scope)

	createAt := time.Now()
	ti.SetAccessCreateAt(createAt)

	// set access token expires
	gcfg := m.grantConfig(gt)
	aexp := gcfg.AccessTokenExp
	if exp := tgr.AccessTokenExp; exp > 0 {
		aexp = exp
	}
	ti.SetAccessExpiresIn(aexp)
	if gcfg.IsGenerateRefresh {
		ti.SetRefreshCreateAt(createAt)
		ti.SetRefreshExpiresIn(gcfg.RefreshTokenExp)
	}

	av, rv, err := m.accessGenerate.Token(ctx, cli.GetID(), tgr.UserID, createAt, gcfg.IsGenerateRefresh)
	if err != nil {
		return nil, err
	}
	ti.SetAccess(av)

	if rv != "" {
		ti.SetRefresh(rv)
	}

	err = m.tokenStore.Create(ctx, ti)
	if err != nil {
		return nil, err
	}

	return ti, nil
}

// RefreshAccessToken refreshing an access token
func (m *manager) RefreshAccessToken(ctx context.Context, tgr *TokenGenerateRequest) (TokenInfo, error) {
	ti, err := m.LoadRefreshToken(ctx, tgr.Refresh)
	if err != nil {
		return nil, err
	}

	cli, err := m.GetClient(ctx, ti.GetClientID())
	if err != nil {
		return nil, err
	}

	oldAccess, oldRefresh := ti.GetAccess(), ti.GetRefresh()

	rcfg := DefaultRefreshTokenCfg
	if v := m.rcfg; v != nil {
		rcfg = v
	}

	createAt := time.Now()
	ti.SetAccessCreateAt(createAt)
	if v := rcfg.AccessTokenExp; v > 0 {
		ti.SetAccessExpiresIn(v)
	}

	if v := rcfg.RefreshTokenExp; v > 0 {
		ti.SetRefreshExpiresIn(v)
	}

	if rcfg.IsResetRefreshTime {
		ti.SetRefreshCreateAt(createAt)
	}

	if scope := tgr.Scope; scope != "" {
		ti.SetScope(scope)
	}

	tv, rv, err := m.accessGenerate.Token(ctx, cli.GetID(), ti.GetUserID(), time.Now(), rcfg.IsGenerateRefresh)
	if err != nil {
		return nil, err
	}

	ti.SetAccess(tv)
	if rv != "" {
		ti.SetRefresh(rv)
	}

	if err := m.tokenStore.Create(ctx, ti); err != nil {
		return nil, err
	}

	if rcfg.IsRemoveAccess {
		// remove the old access token
		if err := m.tokenStore.RemoveByAccess(ctx, oldAccess); err != nil {
			return nil, err
		}
	}

	if rcfg.IsRemoveRefreshing && rv != "" {
		// remove the old refresh token
		if err := m.tokenStore.RemoveByRefresh(ctx, oldRefresh); err != nil {
			return nil, err
		}
	}

	if rv == "" {
		ti.SetRefresh("")
		ti.SetRefreshCreateAt(time.Now())
		ti.SetRefreshExpiresIn(0)
	}

	return ti, nil
}

// RemoveAccessToken use the access token to delete the token information
func (m *manager) RemoveAccessToken(ctx context.Context, access string) error {
	if access == "" {
		return ErrInvalidAccessToken
	}
	return m.tokenStore.RemoveByAccess(ctx, access)
}

// RemoveRefreshToken use the refresh token to delete the token information
func (m *manager) RemoveRefreshToken(ctx context.Context, refresh string) error {
	if refresh == "" {
		return ErrInvalidAccessToken
	}
	return m.tokenStore.RemoveByRefresh(ctx, refresh)
}

// LoadAccessToken according to the access token for corresponding token information
func (m *manager) LoadAccessToken(ctx context.Context, access string) (TokenInfo, error) {
	if access == "" {
		return nil, ErrInvalidAccessToken
	}

	ct := time.Now()
	ti, err := m.tokenStore.GetByAccess(ctx, access)
	if err != nil {
		return nil, err
	} else if ti == nil || ti.GetAccess() != access {
		return nil, ErrInvalidAccessToken
	} else if ti.GetRefresh() != "" && ti.GetRefreshExpiresIn() != 0 &&
		ti.GetRefreshCreateAt().Add(ti.GetRefreshExpiresIn()).Before(ct) {
		return nil, ErrExpiredRefreshToken
	} else if ti.GetAccessExpiresIn() != 0 &&
		ti.GetAccessCreateAt().Add(ti.GetAccessExpiresIn()).Before(ct) {
		return nil, ErrExpiredAccessToken
	}
	return ti, nil
}

// LoadRefreshToken according to the refresh token for corresponding token information
func (m *manager) LoadRefreshToken(ctx context.Context, refresh string) (TokenInfo, error) {
	if refresh == "" {
		return nil, ErrInvalidRefreshToken
	}

	ti, err := m.tokenStore.GetByRefresh(ctx, refresh)
	if err != nil {
		return nil, err
	} else if ti == nil || ti.GetRefresh() != refresh {
		return nil, ErrInvalidRefreshToken
	} else if ti.GetRefreshExpiresIn() != 0 && // refresh token set to not expire
		ti.GetRefreshCreateAt().Add(ti.GetRefreshExpiresIn()).Before(time.Now()) {
		return nil, ErrExpiredRefreshToken
	}
	return ti, nil
}
