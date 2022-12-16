package oauth2

import (
	"bytes"
	"context"
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// NewAccessGenerate create to generate the access token instance
func NewAccessGenerate() *accessGenerate {
	return &accessGenerate{}
}

// AccessGenerate generate the access token
type accessGenerate struct {
}

// Token based on the UUID generated token
func (ag *accessGenerate) Token(ctx context.Context, clientId, userId string, createdAt time.Time, isGenRefresh bool) (string, string, error) {
	buf := bytes.NewBufferString(clientId)
	buf.WriteString(userId)
	buf.WriteString(strconv.FormatInt(createdAt.UnixNano(), 10))

	access := base64.URLEncoding.EncodeToString([]byte(uuid.NewMD5(uuid.Must(uuid.NewRandom()), buf.Bytes()).String()))
	access = strings.ToUpper(strings.TrimRight(access, "="))
	refresh := ""
	if isGenRefresh {
		refresh = base64.URLEncoding.EncodeToString([]byte(uuid.NewSHA1(uuid.Must(uuid.NewRandom()), buf.Bytes()).String()))
		refresh = strings.ToUpper(strings.TrimRight(refresh, "="))
	}

	return access, refresh, nil
}
