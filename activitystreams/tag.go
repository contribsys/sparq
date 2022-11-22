package activitystreams

type Tag struct {
	Type TagType `json:"type"`
	HRef string  `json:"href"`
	Name string  `json:"name"`
}

type TagType string

const (
	TagHashtag TagType = "Hashtag"
	TagMention TagType = "Mention"
)
