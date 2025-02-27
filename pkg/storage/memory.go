package storage

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/karurosux/keystogo/pkg/keystogo"
	"github.com/karurosux/keystogo/pkg/models"
)

func NewMemoryStorage() keystogo.Storage {
	return &MemoryStorage{
		keys: make(map[string]*models.APIKey),
	}
}

type MemoryStorage struct {
	mu   sync.RWMutex
	keys map[string]*models.APIKey
}

// Clear implements keystogo.Storage.
func (m *MemoryStorage) Clear() error {
	m.keys = make(map[string]*models.APIKey)
	return nil
}

// Create implements keystogo.Storage.
func (m *MemoryStorage) Create(apiKey *models.APIKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if apiKey.ID == "" {
		apiKey.ID = m.getRandomKey()
	}
	fmt.Println("Printing apii key => ", apiKey.ID)
	m.keys[apiKey.ID] = apiKey
	return nil
}

func (m *MemoryStorage) getRandomKey() string {
	return uuid.NewString()
}

// Delete implements keystogo.Storage.
func (m *MemoryStorage) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.keys, id)
	return nil
}

// GetByID implements keystogo.Storage.
func (m *MemoryStorage) GetByID(id string) (*models.APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if apiKey, ok := m.keys[id]; ok {
		return apiKey, nil
	}

	return nil, errors.New("api key not found")
}

// GetByHashedKey implements keystogo.Storage.
func (m *MemoryStorage) GetByHashedKey(hashedKey string) (*models.APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, apiKey := range m.keys {
		if apiKey.Key == hashedKey {
			return apiKey, nil
		}
	}

	return nil, errors.New("api key not found")
}

// List implements keystogo.Storage.
func (m *MemoryStorage) List(page models.Page, filter models.Filter) ([]models.APIKey, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []models.APIKey

	for _, apiKey := range m.keys {
		matches := true

		if filter.Name != nil && *filter.Name != "" {
			matches = matches && (apiKey.Name != "" && containsIgnoreCase(apiKey.Name, *filter.Name))
		}

		if matches {
			result = append(result, *apiKey)
		}
	}

	total := int64(len(result))

	if page.Limit > 0 {
		start := page.Limit * page.Offset
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

// Ping implements keystogo.Storage.
func (m *MemoryStorage) Ping() error {
	// Just do nothing in this case.
	return nil
}

// Update implements keystogo.Storage.
func (m *MemoryStorage) Update(id string, apiUpdate models.ApiKeyUpdate) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	apiKey, ok := m.keys[id]
	if !ok {
		return models.ErrKeyNotFound
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
