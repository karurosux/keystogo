package storage

import (
	"sync"
	"time"

	"github.com/karurosux/keystogo/pkg/keystogo"
)

type MemoryStorage struct {
	mu   sync.RWMutex
	keys map[string]keystogo.APIKey
}

func NewMemoryStorage() keystogo.Storage {
	return &MemoryStorage{
		keys: make(map[string]keystogo.APIKey),
	}
}

func (s *MemoryStorage) GetAPIKey(hashedKey string) (*keystogo.APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if key, exists := s.keys[hashedKey]; exists {
		return &key, nil
	}
	return nil, keystogo.ErrKeyNotFound
}

func (s *MemoryStorage) UpdateLastUsed(keyID string, usedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range s.keys {
		if v.ID == keyID {
			v.LastUsedAt = &usedAt
			s.keys[k] = v
			return nil
		}
	}
	return keystogo.ErrKeyNotFound
}
