package oauth2

import (
	"net/http"
	"time"
)

// Config configuration parameters
type ConfigConfig struct {
	TokenType                   string         // token type
	AllowGetAccessRequest       bool           // to allow GET requests for the token
	AllowedResponseTypes        []ResponseType // allow the authorization type
	AllowedGrantTypes           []GrantType    // allow the grant type
	AllowedCodeChallengeMethods []CodeChallengeMethod
	ForcePKCE                   bool
}

// NewConfig create to configuration instance
func NewConfigConfig() *ConfigConfig {
	return &ConfigConfig{
		TokenType:            "Bearer",
		AllowedResponseTypes: []ResponseType{CodeType, TokenType},
		AllowedGrantTypes: []GrantType{
			AuthorizationCode,
			PasswordCredentials,
			ClientCredentials,
			Refreshing,
		},
		AllowedCodeChallengeMethods: []CodeChallengeMethod{
			CodeChallengePlain,
			CodeChallengeS256,
		},
	}
}

// AuthorizeRequest authorization request
type AuthorizeRequest struct {
	ResponseType        ResponseType
	ClientID            string
	Scope               string
	RedirectURI         string
	State               string
	UserID              string
	CodeChallenge       string
	CodeChallengeMethod CodeChallengeMethod
	AccessTokenExp      time.Duration
	Request             *http.Request
}
