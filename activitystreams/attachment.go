package activitystreams

import (
	"mime"
	"strings"
)

type Attachment struct {
	Type      AttachmentType `json:"type"`
	URL       string         `json:"url"`
	MediaType string         `json:"mediaType"`
	Name      string         `json:"name"`
}

type AttachmentType string

const (
	TypeImage    AttachmentType = "Image"
	TypeDocument AttachmentType = "Document"
)

func NewImageAttachment(url string) Attachment {
	return newAttachment(url, TypeImage)
}

func NewDocumentAttachment(url string) Attachment {
	return newAttachment(url, TypeDocument)
}

func newAttachment(url string, attachType AttachmentType) Attachment {
	var fileType string
	extIdx := strings.LastIndexByte(url, '.')
	if extIdx > -1 {
		fileType = mime.TypeByExtension(url[extIdx:])
	}
	return Attachment{
		Type:      attachType,
		URL:       url,
		MediaType: fileType,
	}
}
