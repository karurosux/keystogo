package keystogo

import (
	"github.com/karurosux/keystogo/pkg/models"
)

// Storage defines the interface for API key persistence
type Storage interface {
	// Core CRUD operations
	Get(hashedKey string) (*models.APIKey, error)
	Create(apiKey *models.APIKey) error
	Update(hashedKye string, apiKey models.ApiKeyUpdate) error
	Delete(hashedKey string) error

	// Query operations
	List(page models.Page, filter models.Filter) ([]models.APIKey, int64, error)

	// Optional operations - implementations can return ErrNotSupported
	Ping() error  // Check storage connection
	Clear() error // Clear all keys (useful for testing)
}
