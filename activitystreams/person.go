package activitystreams

type Person struct {
	BaseObject
	Inbox             string    `json:"inbox"`
	Outbox            string    `json:"outbox"`
	PreferredUsername string    `json:"preferredUsername"`
	URL               string    `json:"url"`
	Name              string    `json:"name"`
	Icon              Image     `json:"icon"`
	Following         string    `json:"following"`
	Followers         string    `json:"followers"`
	Summary           string    `json:"summary"`
	PublicKey         PublicKey `json:"publicKey"`
	Endpoints         Endpoints `json:"endpoints"`
}

func NewPerson(accountRoot string) *Person {
	p := Person{
		BaseObject: BaseObject{
			Type: "Person",
			Context: []interface{}{
				Namespace,
			},
			ID: accountRoot,
		},
		Following: accountRoot + "/following",
		Followers: accountRoot + "/followers",
		Inbox:     accountRoot + "/inbox",
		Outbox:    accountRoot + "/outbox",
	}

	return &p
}

func (p *Person) AddPubKey(k string) {
	p.Context = append(p.Context, "https://w3id.org/security/v1")
	p.PublicKey = PublicKey{
		ID:           p.ID + "#main-key",
		Owner:        p.ID,
		PublicKeyPEM: string(k),
	}
}

func (p *Person) SetPrivKey(k []byte) {
	p.PublicKey.privateKey = k
}

func (p *Person) GetPrivKey() []byte {
	return p.PublicKey.privateKey
}
