package model

import "time"

type Actor struct {
	Id             string
	Type           string
	Email          string
	PrivateKey     []byte
	PrivateKeySalt []byte
	PublicKey      string
	CreatedAt      time.Time
	Properties     string
}

type ActorFollowing struct {
	Id                 string
	ActorId            string
	TargetActorId      string
	TargetActorAccount string
	State              string
	CreatedAt          time.Time
}

type Object struct {
	Id               string
	MastodonId       string
	Type             string
	CreatedAt        time.Time
	OriginalActorId  string
	OriginalObjectId string
	ReplyToObjectId  string
	Properties       string
	Local            int
}

type InboxObject struct {
	Id        string
	ActorId   string
	ObjectId  string
	CreatedAt time.Time
}

type OutboxObjects struct {
	Id          string
	ActorId     string
	ObjectId    string
	Target      string
	CreatedAt   time.Time
	PublishedAt time.Time
}

type ActorNotification struct {
	Id          int
	Type        string
	ActorId     string
	FromActorId string
	ObjectId    string
	CreatedAt   time.Time
}

type ActorFavorite struct {
	Id        string
	ActorId   string
	ObjectId  string
	CreatedAt time.Time
}

type ActorReblog struct {
	Id        string
	ActorId   string
	ObjectId  string
	CreatedAt time.Time
}

type Subscription struct {
	Id                 string
	ActorId            string
	ClientId           string
	Endpoint           string
	KeyP256dh          string
	KeyAuth            string
	AlertMention       int
	AlertStatus        int
	AlertReblog        int
	AlertFollow        int
	AlertFollowRequest int
	AlertFavorite      int
	AlertPoll          int
	AlertUpdate        int
	AlertAdminSignUp   int
	AlertAdminReport   int
	Policy             string
	CreatedAt          time.Time
}

type ActorReply struct {
	Id                string
	ActorId           string
	ObjectId          string
	InReplyToObjectId string
	CreatedAt         time.Time
}
