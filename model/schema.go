package model

type Community struct {
	Hostname string
}

type Actor struct {
	Id     string
	UserId int
	Inbox  string
}
