package finger

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/contribsys/sparq/model"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("User not found")
)

func Lookup(ctx context.Context, db *gorm.DB, username, host string) (*Resource, error) {

	var acct model.Account
	err := db.Where("username = ?", username).First(&acct).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return &Resource{
		Subject: "acct:" + username + "@" + host,
		Aliases: []string{
			// acct.CanonicalURL(),
			// acct.FederatedAccount(),
		},
		Links: []Link{
			{
				// HRef: c.CanonicalURL(),
				Type: "text/html",
				Rel:  "https://webfinger.net/rel/profile-page",
			},
			{
				// HRef: c.FederatedAccount(),
				Type: "application/activity+json",
				Rel:  "self",
			},
		},
	}, nil

}

func HttpHandler(db *gorm.DB, hostname string) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		username := req.URL.Query().Get("resource")
		fmt.Println(username)
		fmt.Println(hostname)
		if username == "" || !strings.HasSuffix(username, "@"+hostname) {
			http.Error(resp, "Invalid input", http.StatusBadRequest)
			return
		}
	}
}
