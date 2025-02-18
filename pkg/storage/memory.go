package storage

import (
	"errors"
	"strings"
	"sync"

	"github.com/karurosux/keystogo/pkg/keystogo"
	"github.com/karurosux/keystogo/pkg/models"
)

func NewMemoryStorage() keystogo.Storage {
	return &MemoryStorage{
		keys: make(map[string]models.APIKey),
	}
}

type MemoryStorage struct {
	mu   sync.RWMutex
	keys map[string]models.APIKey
}

// BatchCreate implements keystogo.Storage.
func (m *MemoryStorage) BatchCreate(apiKeys []*models.APIKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, apiKey := range apiKeys {
		m.keys[apiKey.Key] = *apiKey
	}

	return nil
}

// BatchDelete implements keystogo.Storage.
func (m *MemoryStorage) BatchDelete(hashedKeys []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, hashedKey := range hashedKeys {
		delete(m.keys, hashedKey)
	}

	return nil
}

// Clear implements keystogo.Storage.
func (m *MemoryStorage) Clear() error {
	m.keys = make(map[string]models.APIKey)
	return nil
}

// Create implements keystogo.Storage.
func (m *MemoryStorage) Create(apiKey *models.APIKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.keys[apiKey.Key] = *apiKey
	return nil
}

// Delete implements keystogo.Storage.
func (m *MemoryStorage) Delete(hashedKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.keys, hashedKey)
	return nil
}

// Get implements keystogo.Storage.
func (m *MemoryStorage) Get(hashedKey string) (*models.APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if apiKey, ok := m.keys[hashedKey]; ok {
		return &apiKey, nil
	}

	return nil, errors.New("api key not found")
}

// List implements keystogo.Storage.
func (m *MemoryStorage) List(page keystogo.Page, filter keystogo.Filter) ([]models.APIKey, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []models.APIKey

	for _, apiKey := range m.keys {
		matches := true

		if filter.Name != "" {
			matches = matches && (apiKey.Name != "" && containsIgnoreCase(apiKey.Name, filter.Name))
		}

		if matches {
			result = append(result, apiKey)
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
func (m *MemoryStorage) Update(apiKey *models.APIKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.keys[apiKey.Key] = *apiKey
	return nil
}
