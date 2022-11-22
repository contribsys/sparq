package activitystreams

import "fmt"

type (
	BaseObject struct {
		Context []interface{} `json:"@context,omitempty"`
		Type    string        `json:"type"`
		ID      string        `json:"id"`
	}

	PublicKey struct {
		ID           string `json:"id"`
		Owner        string `json:"owner"`
		PublicKeyPEM string `json:"publicKeyPem"`
		privateKey   []byte
	}

	Endpoints struct {
		SharedInbox string `json:"sharedInbox,omitempty"`
	}

	Image struct {
		Type      string `json:"type"`
		MediaType string `json:"mediaType"`
		URL       string `json:"url"`
	}
)

type OrderedCollection struct {
	BaseObject
	TotalItems int    `json:"totalItems"`
	First      string `json:"first"`
	Last       string `json:"last,omitempty"`
}

func NewOrderedCollection(accountRoot, collType string, items int) *OrderedCollection {
	oc := OrderedCollection{
		BaseObject: BaseObject{
			Context: []interface{}{
				Namespace,
			},
			ID:   accountRoot + "/" + collType,
			Type: "OrderedCollection",
		},
		First:      accountRoot + "/" + collType + "?page=1",
		TotalItems: items,
	}
	return &oc
}

type OrderedCollectionPage struct {
	BaseObject
	TotalItems   int           `json:"totalItems"`
	PartOf       string        `json:"partOf"`
	Next         string        `json:"next,omitempty"`
	Prev         string        `json:"prev,omitempty"`
	OrderedItems []interface{} `json:"orderedItems,omitempty"`
}

func NewOrderedCollectionPage(accountRoot, collType string, items, page int) *OrderedCollectionPage {
	ocp := OrderedCollectionPage{
		BaseObject: BaseObject{
			Context: []interface{}{
				Namespace,
			},
			ID:   fmt.Sprintf("%s/%s?page=%d", accountRoot, collType, page),
			Type: "OrderedCollectionPage",
		},
		TotalItems: items,
		PartOf:     accountRoot + "/" + collType,
		Next:       fmt.Sprintf("%s/%s?page=%d", accountRoot, collType, page+1),
	}
	return &ocp
}
