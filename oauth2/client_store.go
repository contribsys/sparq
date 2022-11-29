package oauth2

import (
	"context"
	"errors"
	"sync"
)

// NewClientStore create client store
func NewClientStore() *clientStore {
	return &clientStore{
		data: make(map[string]ClientInfo),
	}
}

// ClientStore client information store
type clientStore struct {
	sync.RWMutex
	data map[string]ClientInfo
}

// GetByID according to the ID for the client information
func (cs *clientStore) GetByID(ctx context.Context, id string) (ClientInfo, error) {
	cs.RLock()
	defer cs.RUnlock()

	if c, ok := cs.data[id]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}

// Set set client information
func (cs *clientStore) Set(id string, cli ClientInfo) (err error) {
	cs.Lock()
	defer cs.Unlock()

	cs.data[id] = cli
	return
}
