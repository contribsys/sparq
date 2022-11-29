package model

import (
	"fmt"
	"time"

	"github.com/contribsys/sparq/db"
	"github.com/contribsys/sparq/util"
)

type Community struct {
	Hostname string
}

type RoleMask int32

const (
	RoleUser RoleMask = 1 << iota
	RoleModerator
	RoleAdmin

	RoleAll RoleMask = RoleUser | RoleModerator | RoleAdmin
)

type User struct {
	Id       int64
	Sfid     string
	FullName string
	Nick     string
	Email    string
	RoleMask RoleMask
	// Profile  *AccountProfile
	*UserSecurity
}

type OauthApp struct {
	ClientName   string `json:"name"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectUris string `json:"redirect_uri"`
	Website      string `json:"website"`
	Scopes       string
	CreatedAt    time.Time
}

func (u *User) URI() string {
	return fmt.Sprintf("https://%s/@%s", db.InstanceHostname, u.Nick)
}

var (
	Snowflakes = util.NewSnowflake()
)

type AccountProfile struct {
	AccountID uint
	Bio       string
	Links     []AccountProfileMetadata
}

type AccountProfileMetadata struct {
	AccountID  uint
	Name       string
	Value      string
	VerifiedAt time.Time
}

type AccountSecurity struct {
	AccountId    uint
	PasswordHash []byte
	PublicKey    []byte
	PrivateKey   []byte
}

type Post struct {
	URI          string
	InReplyTo    string
	AttributedTo string
	Author       *Actor
	Summary      string
	Content      string
	Visibility
	CreatedAt time.Time
}

type Actor struct {
	Id     string
	UserId int
	Inbox  string
}

type UserSecurity struct {
	Id           int
	UserId       int
	PasswordHash string
	PublicKey    string
	PrivateKey   string
}
