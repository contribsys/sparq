package model

import (
	"fmt"

	"github.com/contribsys/sparq/db"
)

type PostVisibility int

var (
	ToPublic        PostVisibility = 0
	ToMentionedOnly PostVisibility = 1
)

func NewPost(author *Account, inReplyTo *Post, summary string, content string, visibility PostVisibility) *Post {
	sid := Snowflakes.NextID()
	p := &Post{
		AttributedTo:   author.URI(),
		URI:            fmt.Sprintf("https://%s/@%s/statuses/%d", db.InstanceHostname, author.Nick, sid),
		Summary:        summary,
		Content:        content,
		PostVisibility: ToPublic,
	}
	if inReplyTo != nil {
		p.InReplyTo = inReplyTo.URI
	}
	return p
}
