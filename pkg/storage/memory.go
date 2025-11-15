package storage

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/karurosux/keystogo/pkg/keystogo"
	"github.com/karurosux/keystogo/pkg/models"
)

func NewMemoryStorage() keystogo.Storage {
	return &MemoryStorage{
		keys:      make(map[string]*models.APIKey),
		hashIndex: make(map[string]string),
	}
}

type MemoryStorage struct {
	mu        sync.RWMutex
	keys      map[string]*models.APIKey
	hashIndex map[string]string
}

func (m *MemoryStorage) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keys = make(map[string]*models.APIKey)
	m.hashIndex = make(map[string]string)
	return nil
}

func (m *MemoryStorage) Create(ctx context.Context, apiKey *models.APIKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if apiKey.ID == "" {
		apiKey.ID = m.getRandomKey()
	}
	m.keys[apiKey.ID] = apiKey
	m.hashIndex[apiKey.Key] = apiKey.ID
	return nil
}

func (m *MemoryStorage) getRandomKey() string {
	return uuid.NewString()
}

func (m *MemoryStorage) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if apiKey, ok := m.keys[id]; ok {
		delete(m.hashIndex, apiKey.Key)
	}
	delete(m.keys, id)
	return nil
}

func (m *MemoryStorage) GetByID(ctx context.Context, id string) (*models.APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if apiKey, ok := m.keys[id]; ok {
		return apiKey, nil
	}

	return nil, models.ErrKeyNotFound()
}

func (m *MemoryStorage) GetByHashedKey(ctx context.Context, hashedKey string) (*models.APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id, ok := m.hashIndex[hashedKey]
	if !ok {
		return nil, models.ErrKeyNotFound()
	}

	apiKey, ok := m.keys[id]
	if !ok {
		return nil, models.ErrKeyNotFound()
	}

	return apiKey, nil
}

func (m *MemoryStorage) List(ctx context.Context, page models.Page, filter models.Filter) ([]models.APIKey, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]models.APIKey, 0, len(m.keys))
	now := time.Now()

	for _, apiKey := range m.keys {
		if apiKey.ExpiresAt != nil && now.After(*apiKey.ExpiresAt) {
			continue
		}

		matches := true

		if filter.Name != nil && *filter.Name != "" {
			matches = matches && (apiKey.Name != "" && containsIgnoreCase(apiKey.Name, *filter.Name))
		}

		if filter.Active != nil {
			matches = matches && (apiKey.Active == *filter.Active)
		}

		if matches {
			result = append(result, *apiKey)
		}
	}

	total := int64(len(result))

	if page.Limit > 0 {
		start := page.Offset
		end := start + page.Limit

		if start < len(result) {
			if end > len(result) {
				end = len(result)
			}
			result = result[start:end]
		} else {
			result = []models.APIKey{}
		}
	}

	return result, total, nil
}

func containsIgnoreCase(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}

func (m *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (m *MemoryStorage) Update(ctx context.Context, id string, apiUpdate models.ApiKeyUpdate) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	apiKey, ok := m.keys[id]
	if !ok {
		return models.ErrKeyNotFound()
	}

	if apiUpdate.Active != nil {
		apiKey.Active = *apiUpdate.Active
	}
	if apiUpdate.Name != nil {
		apiKey.Name = *apiUpdate.Name
	}
	if apiUpdate.ExpiresAt != nil {
		apiKey.ExpiresAt = apiUpdate.ExpiresAt
	}
	if apiUpdate.Metadata != nil {
		apiKey.Metadata = apiUpdate.Metadata
	}
	if apiUpdate.Permissions != nil {
		apiKey.Permissions = apiUpdate.Permissions
	}
	if apiUpdate.LastUsedAt != nil {
		apiKey.LastUsedAt = apiUpdate.LastUsedAt
	}

	m.keys[id] = apiKey

	return nil
}

func (m *MemoryStorage) CleanupExpired(ctx context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	count := int64(0)

	for id, apiKey := range m.keys {
		if apiKey.ExpiresAt != nil && now.After(*apiKey.ExpiresAt) {
			delete(m.hashIndex, apiKey.Key)
			delete(m.keys, id)
			count++
		}
	}

	return count, nil
}
