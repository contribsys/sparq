package oauth2

import (
	"net/url"
	"strings"
)

type (
	// ValidateURIHandler validates that redirectURI is contained in baseURI
	ValidateURIHandler func(baseURI, redirectURI string) error
)

// DefaultValidateURI validates that redirectURI is contained in baseURI
func DefaultValidateURI(baseURI string, redirectURI string) error {
	base, err := url.Parse(baseURI)
	if err != nil {
		return err
	}

	redirect, err := url.Parse(redirectURI)
	if err != nil {
		return err
	}
	if strings.Contains(redirectURI, ":oob") {
		return nil
	}
	if !strings.HasSuffix(redirect.Host, base.Host) {
		return ErrInvalidRedirectURI
	}
	return nil
}
