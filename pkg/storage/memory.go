package storage

import (
	"errors"
	"fmt"
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

	fmt.Println("get", hashedKey)
	fmt.Print(m.keys)
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
	var total int64

	for _, apiKey := range m.keys {
		// Just implementing name for the moment
		if filter.Name != "" && filter.Name == apiKey.Name {
			result = append(result, apiKey)
		}
	}

	total = int64(len(result))

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
