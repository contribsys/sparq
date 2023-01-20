package model

import (
	"fmt"
	"time"

	"github.com/contribsys/sparq/db"
)

type Toot struct {
	Sid                string `json:"id"`
	URI                string `json:"uri"`
	AccountId          uint64
	ActorId            uint64
	BoostOfId          *string
	InReplyTo          *string `json:"in_reply_to_id,omitempty"`
	InReplyToAccountId *uint64 `json:"in_reply_to_account_id,omitempty"`
	Summary            string  `json:"spoiler_text"`
	Content            string  `json:"content"`
	Lang               string  `json:"language"`
	Visibility         PostVisibility
	CreatedAt          time.Time `json:"created_at"`
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

func (t *Toot) Viz() string {
	return FromVis(t.Visibility)
}

/*
	{
	  "id": "22345792",
	  "type": "image",
	  "url": "https://files.mastodon.social/media_attachments/files/022/345/792/original/57859aede991da25.jpeg",
	  "preview_url": "https://files.mastodon.social/media_attachments/files/022/345/792/small/57859aede991da25.jpeg",
	  "remote_url": null,
	  "text_url": "https://mastodon.social/media/2N4uvkuUtPVrkZGysms",
	  "meta": {
	    "original": {
	      "width": 640,
	      "height": 480,
	      "size": "640x480",
	      "aspect": 1.3333333333333333
	    },
	    "small": {
	      "width": 461,
	      "height": 346,
	      "size": "461x346",
	      "aspect": 1.3323699421965318
	    },
	    "focus": {
	      "x": -0.27,
	      "y": 0.51
	    }
	  },
	  "description": "test media description",
	  "blurhash": "UFBWY:8_0Jxv4mx]t8t64.%M-:IUWGWAt6M}"
	}
*/
type TootMedia struct {
	Id            uint64
	Sid           string
	AccountId     string
	Salt          string
	MimeType      string
	Path          string
	ThumbMimeType string
	ThumbPath     string
	Meta          string
	Description   string
	Blurhash      string
	CreatedAt     time.Time
}

func (tm *TootMedia) DiskPath(variant string) string {
	c := tm.CreatedAt
	return fmt.Sprintf("/media/%d/%d/%d/%s-%s.jpg", c.Year(), c.Month(), c.Day(), variant, tm.Salt)
}

func (tm *TootMedia) PublicUri(variant string) string {
	c := tm.CreatedAt
	return fmt.Sprintf("https://%s/media/%d/%d/%d/%s-%s.jpg", db.InstanceHostname, c.Year(), c.Month(), c.Day(), variant, tm.Salt)
}

type TootTag struct {
	Sid       string
	Name      string
	CreatedAt time.Time
}
