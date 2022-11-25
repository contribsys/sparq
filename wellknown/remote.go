package wellknown

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Link struct {
	HRef       string             `json:"href"`
	Type       string             `json:"type,omitempty"`
	Rel        string             `json:"rel"`
	Properties map[string]*string `json:"properties,omitempty"`
	Titles     map[string]string  `json:"titles,omitempty"`
}

type Resource struct {
	Subject    string            `json:"subject,omitempty"`
	Aliases    []string          `json:"aliases,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
	Links      []Link            `json:"links"`
}

type resolverFn func(string) ([]byte, error)

func defaultResolver(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

// Look up a user handle of the form "@username@remotehost.ext"
// and return the WebFinger payload
func RemoteLookup(handle string, fn resolverFn) (data string, err error) {
	handle = strings.TrimLeft(handle, "@")
	parts := strings.Split(handle, "@")

	if fn == nil {
		fn = defaultResolver
	}
	body, err := fn(fmt.Sprintf("https://%s/.well-known/webfinger?resource=acct:%s", parts[1], handle))
	if err != nil {
		return
	}

	var result Resource
	err = json.Unmarshal(body, &result)
	if err != nil {
		return
	}

	var href string
	// iterate over webfinger links and find the one with
	// a self "rel"
	for _, link := range result.Links {
		if link.Rel == "self" {
			href = link.HRef
		}
	}

	// if we didn't find it with the above then
	// try using aliases
	if href == "" {
		// take the last alias because mastodon has the
		// https://instance.tld/@user first which
		// doesn't work as an href
		href = result.Aliases[len(result.Aliases)-1]
	}

	return href, nil
}
