package model

import (
	"time"
)

type Toot struct {
	Sid                string `json:"id"`
	URI                string `json:"uri"`
	AccountId          uint64
	ActorId            uint64
	BoostOfId          *string
	InReplyTo          *string        `json:"in_reply_to_id,omitempty"`
	InReplyToAccountId *uint64        `json:"in_reply_to_account_id,omitempty"`
	Summary            string         `json:"spoiler_text"`
	Content            string         `json:"content"`
	Lang               string         `json:"language"`
	Visibility         PostVisibility `json:"visibility"`
	CreatedAt          time.Time      `json:"created_at"`
	AuthorID           string
	CollectionId       *uint64
	PollID             *uint64
	AppID              *uint64
	LastEditAt         *time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

type PostVisibility int

var (
	VisPublic    PostVisibility = 0
	VisUnlisted  PostVisibility = 1
	VisPrivate   PostVisibility = 2
	VisDirect    PostVisibility = 3
	VisLimited   PostVisibility = 4
	Visibilities                = map[PostVisibility]string{
		VisPublic:   "public",
		VisUnlisted: "unlisted",
		VisPrivate:  "private",
		VisDirect:   "direct",
		VisLimited:  "limited",
	}
)

func ToVis(word string) PostVisibility {
	if word == "public" {
		return VisPublic
	} else if word == "unlisted" {
		return VisUnlisted
	} else if word == "private" {
		return VisPrivate
	} else if word == "direct" {
		return VisDirect
	} else if word == "limited" {
		return VisLimited
	}
	return VisPublic
}
func FromVis(value PostVisibility) string {
	return Visibilities[value]
}
