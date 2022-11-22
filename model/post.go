package model

import (
	"fmt"
)

type Visibility int

var (
	Public        Visibility = 0
	MentionedOnly Visibility = 1
)

func NewPost(author *User, inReplyTo *Post, summary string, content string, visibility Visibility) *Post {
	sid := Snowflakes.NextID()
	p := &Post{
		AttributedTo: author.URI(),
		URI:          fmt.Sprintf("https://localhost.dev/@%s/statuses/%d", author.Nick, sid),
		Summary:      summary,
		Content:      content,
		Visibility:   Public,
	}
	if inReplyTo != nil {
		p.InReplyTo = inReplyTo.URI
	}
	return p
}
