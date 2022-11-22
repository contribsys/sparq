package activitystreams

import (
	"time"
)

const (
	Namespace = "https://www.w3.org/ns/activitystreams"
	Public    = "https://www.w3.org/ns/activitystreams#Public"
)

// Allows us to drop more elements into @context on Create/Update
var Extensions = map[string]string{}

// Activity describes an event in the ActivityStream
type Activity struct {
	BaseObject
	Actor     string    `json:"actor"`
	Published time.Time `json:"published,omitempty"`
	To        []string  `json:"to,omitempty"`
	CC        []string  `json:"cc,omitempty"`
	Object    *Object   `json:"object"`
}

type FollowActivity struct {
	BaseObject
	Actor     string    `json:"actor"`
	Published time.Time `json:"published,omitempty"`
	To        []string  `json:"to,omitempty"`
	CC        []string  `json:"cc,omitempty"`
	Object    string    `json:"object"`
}

func NewCreateActivity(o *Object) *Activity {
	a := Activity{
		BaseObject: BaseObject{
			Context: []interface{}{
				Namespace,
				Extensions,
			},
			ID:   o.ID,
			Type: "Create",
		},
		Actor:     o.AttributedTo,
		Object:    o,
		Published: o.Published,
	}
	return &a
}

func NewUpdateActivity(o *Object) *Activity {
	a := Activity{
		BaseObject: BaseObject{
			Context: []interface{}{
				Namespace,
				Extensions,
			},
			ID:   o.ID,
			Type: "Update",
		},
		Actor:     o.AttributedTo,
		Object:    o,
		Published: o.Published,
	}
	return &a
}

func NewDeleteActivity(o *Object) *Activity {
	a := Activity{
		BaseObject: BaseObject{
			Context: []interface{}{
				Namespace,
			},
			ID:   o.ID,
			Type: "Delete",
		},
		Actor:  o.AttributedTo,
		Object: o,
	}
	return &a
}

func NewFollowActivity(actorIRI, followeeIRI string) *FollowActivity {
	a := FollowActivity{
		BaseObject: BaseObject{
			Context: []interface{}{
				Namespace,
			},
			Type: "Follow",
		},
		Actor:  actorIRI,
		Object: followeeIRI,
	}
	return &a
}

// Object is the primary base type for the Activity Streams vocabulary.
type Object struct {
	BaseObject
	Published    time.Time         `json:"published,omitempty"`
	Summary      *string           `json:"summary,omitempty"`
	InReplyTo    *string           `json:"inReplyTo,omitempty"`
	URL          string            `json:"url"`
	AttributedTo string            `json:"attributedTo,omitempty"`
	To           []string          `json:"to,omitempty"`
	CC           []string          `json:"cc,omitempty"`
	Name         string            `json:"name,omitempty"`
	Content      string            `json:"content,omitempty"`
	ContentMap   map[string]string `json:"contentMap,omitempty"`
	Tag          []Tag             `json:"tag,omitempty"`
	Attachment   []Attachment      `json:"attachment,omitempty"`

	// Person
	Inbox             string     `json:"inbox,omitempty"`
	Outbox            string     `json:"outbox,omitempty"`
	Following         string     `json:"following,omitempty"`
	Followers         string     `json:"followers,omitempty"`
	PreferredUsername string     `json:"preferredUsername,omitempty"`
	Icon              *Image     `json:"icon,omitempty"`
	PublicKey         *PublicKey `json:"publicKey,omitempty"`
	Endpoints         *Endpoints `json:"endpoints,omitempty"`
}

func NewNoteObject() *Object {
	o := Object{
		BaseObject: BaseObject{
			Type: "Note",
		},
		To: []string{
			Public,
		},
	}
	return &o
}

func NewArticleObject() *Object {
	o := Object{
		BaseObject: BaseObject{
			Type: "Article",
		},
		To: []string{
			Public,
		},
	}
	return &o
}

func NewPersonObject() *Object {
	o := Object{
		BaseObject: BaseObject{
			Type: "Person",
		},
	}
	return &o
}
