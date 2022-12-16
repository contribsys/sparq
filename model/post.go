package model

import (
	"fmt"

	"github.com/contribsys/sparq/db"
)

type Visibility int

var (
	Public        Visibility = 0
	MentionedOnly Visibility = 1
)

func NewPost(author *Account, inReplyTo *Post, summary string, content string, visibility Visibility) *Post {
	sid := Snowflakes.NextID()
	p := &Post{
		AttributedTo: author.URI(),
		URI:          fmt.Sprintf("https://%s/@%s/statuses/%d", db.InstanceHostname, author.Nick, sid),
		Summary:      summary,
		Content:      content,
		Visibility:   Public,
	}
	if inReplyTo != nil {
		p.InReplyTo = inReplyTo.URI
	}
	return p
}
