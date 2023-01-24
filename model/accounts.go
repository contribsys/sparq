package model

import (
	"fmt"
	"time"

	"github.com/contribsys/sparq/db"
	"github.com/contribsys/sparq/util"
)

type RoleMask int32

const (
	RoleUser RoleMask = 1 << iota
	RoleModerator
	RoleAdmin

	RoleAll RoleMask = RoleUser | RoleModerator | RoleAdmin
)

type AccountVisibility int32

const (
	Public    AccountVisibility = 0
	Protected AccountVisibility = 1
	Private   AccountVisibility = 2
)

type Account struct {
	Id         int64
	Sid        string
	FullName   string
	Nick       string
	Email      string
	Visibility AccountVisibility
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
	RoleMask   RoleMask
	*AccountProfile
	*AccountSecurity
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
	Note      string
	Avatar    string
	Header    string
	Fields    []AccountField
}

type AccountField struct {
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
