package model

import (
	"time"
)

type Post struct {
	ID          int64
	URI         string
	InReplyTo   string
	AuthorID    int64
	PollID      int64
	WarningText string
	Content     string
	Lang        string
	Visibility  PostVisibility
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
