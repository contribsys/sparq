package model

import (
	"time"
)

type Community struct {
	Hostname string
}

type Post struct {
	URI          string
	InReplyTo    string
	AttributedTo string
	Author       *Actor
	Summary      string
	Content      string
	PostVisibility
	CreatedAt time.Time
}

type Actor struct {
	Id     string
	UserId int
	Inbox  string
}
