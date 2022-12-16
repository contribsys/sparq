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

type Account struct {
	Id        int64      `db:"Id"`
	Sfid      string     `db:"Sfid"`
	FullName  string     `db:"FullName"`
	Nick      string     `db:"Nick"`
	Email     string     `db:"Email"`
	CreatedAt *time.Time `db:"CreatedAt"`
	UpdatedAt *time.Time `db:"UpdatedAt"`
	RoleMask  RoleMask   `db:"RoleMask"`
	// Profile  *AccountProfile
	*UserSecurity
}

func (a *Account) URI() string {
	return fmt.Sprintf("https://%s/@%s", db.InstanceHostname, a.Nick)
}

func (a *Account) Created() string {
	return util.Thens(*a.CreatedAt)
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
