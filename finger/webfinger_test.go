package finger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	fingerJson = map[string]string{
		"@getajobmike@ruby.social": `
			{"subject":"acct:getajobmike@ruby.social","aliases":["https://ruby.social/@getajobmike","https://ruby.social/users/getajobmike"],"links":[{"rel":"http://webfinger.net/rel/profile-page","type":"text/html","href":"https://ruby.social/@getajobmike"},{"rel":"self","type":"application/activity+json","href":"https://ruby.social/users/getajobmike"},{"rel":"http://ostatus.org/schema/1.0/subscribe","template":"https://ruby.social/authorize_interaction?uri={uri}"}]}
			`,
		"@karolat@stereophonic.space": `
			{"aliases":["https://stereophonic.space/users/karolat"],"links":[{"href":"https://stereophonic.space/users/karolat","rel":"http://webfinger.net/rel/profile-page","type":"text/html"},{"href":"https://stereophonic.space/users/karolat","rel":"self","type":"application/activity+json"},{"href":"https://stereophonic.space/users/karolat","rel":"self","type":"application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""},{"rel":"http://ostatus.org/schema/1.0/subscribe","template":"https://stereophonic.space/ostatus_subscribe?acct={uri}"}],"subject":"acct:karolat@stereophonic.space"}
			`,
	}
)

func xTestRemoteLookup(t *testing.T) {
	users := []string{
		"@getajobmike@ruby.social",
		"@karolat@stereophonic.space",
	}
	for idx := range users {
		handle := users[idx]
		data, err := RemoteLookup(handle, func(string) ([]byte, error) {
			return []byte(fingerJson[handle]), nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "", data)
	}
}

// Pelorma 2.4?
/*
{"aliases"=>["https://stereophonic.space/users/karolat"],
 "links"=>
  [{"href"=>"https://stereophonic.space/users/karolat",
    "rel"=>"http://webfinger.net/rel/profile-page",
    "type"=>"text/html"},
   {"href"=>"https://stereophonic.space/users/karolat",
    "rel"=>"self",
    "type"=>"application/activity+json"},
   {"href"=>"https://stereophonic.space/users/karolat",
    "rel"=>"self",
    "type"=>
     "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\""},
   {"rel"=>"http://ostatus.org/schema/1.0/subscribe",
    "template"=>"https://stereophonic.space/ostatus_subscribe?acct={uri}"}],
 "subject"=>"acct:karolat@stereophonic.space"}
*/

// Mastodon 3.5.3
/*
{"subject"=>"acct:getajobmike@ruby.social",
 "aliases"=>
  ["https://ruby.social/@getajobmike",
   "https://ruby.social/users/getajobmike"],
 "links"=>
  [{"rel"=>"http://webfinger.net/rel/profile-page",
    "type"=>"text/html",
    "href"=>"https://ruby.social/@getajobmike"},
   {"rel"=>"self",
    "type"=>"application/activity+json",
    "href"=>"https://ruby.social/users/getajobmike"},
   {"rel"=>"http://ostatus.org/schema/1.0/subscribe",
    "template"=>"https://ruby.social/authorize_interaction?uri={uri}"}]}
*/
