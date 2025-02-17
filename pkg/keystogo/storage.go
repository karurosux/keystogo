package keystogo

import (
	"time"

	"github.com/karurosux/keystogo/pkg/models"
)

type Filter struct {
	Active      *bool      // Filter by active status
	Name        string     // Filter by name (partial match)
	Tags        []string   // Filter by tags
	CreatedFrom *time.Time // Filter by creation date range
	CreatedTo   *time.Time
	Metadata    map[string]string // Filter by metadata
}

// Page represents pagination parameters
type Page struct {
	Offset int
	Limit  int
}

// Storage defines the interface for API key persistence
type Storage interface {
	// Core CRUD operations
	Get(hashedKey string) (*models.APIKey, error)
	Create(apiKey *models.APIKey) error
	Update(apiKey *models.APIKey) error
	Delete(hashedKey string) error

	// Query operations
	List(page Page, filter Filter) ([]models.APIKey, int64, error)

	// Batch operations
	BatchCreate(apiKeys []*models.APIKey) error
	BatchDelete(hashedKeys []string) error

	// Optional operations - implementations can return ErrNotSupported
	Ping() error  // Check storage connection
	Clear() error // Clear all keys (useful for testing)
}
