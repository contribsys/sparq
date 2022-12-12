package oauth2

import (
	"context"
	"errors"
	"sync"
)

// NewClientStore create client store
func NewClientStore() ClientStore {
	return &clientStore{
		data: make(map[string]ClientInfo),
	}
}

type ClientStore interface {
	GetByID(ctx context.Context, id string) (ClientInfo, error)
	Set(ctx context.Context, id string, cli ClientInfo) error
	Delete(ctx context.Context, id string) error
}

type clientStore struct {
	sync.RWMutex
	data map[string]ClientInfo
}

func (cs *clientStore) GetByID(ctx context.Context, id string) (ClientInfo, error) {
	cs.RLock()
	defer cs.RUnlock()

	if c, ok := cs.data[id]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}

func (cs *clientStore) Set(ctx context.Context, id string, cli ClientInfo) (err error) {
	cs.Lock()
	defer cs.Unlock()

	cs.data[id] = cli
	return
}

func (cs *clientStore) Delete(ctx context.Context, id string) error {
	delete(cs.data, id)
	return nil
}
