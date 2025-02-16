package keystogo

import (
	"time"
)

// Permission represents a single capability that an API key can have
type Permission string

// APIKey represents the structure of an API key with its metadata
type APIKey struct {
	ID          string
	Key         string // Hashed key value
	Name        string // Human readable name
	Permissions []Permission
	CreatedAt   time.Time
	ExpiresAt   *time.Time // Optional expiration
	LastUsedAt  *time.Time
	RateLimit   *RateLimit // Optional rate limiting
	Active      bool
}

// RateLimit defines rate limiting parameters for an API key
type RateLimit struct {
	RequestsPerMinute int
	RequestsPerHour   int
	RequestsPerDay    int
}

// Storage defines the interface that must be implemented by persistence layers
type Storage interface {
	// GetAPIKey retrieves an API key by its hashed value
	GetAPIKey(hashedKey string) (*APIKey, error)

	// UpdateLastUsed updates the LastUsedAt timestamp
	UpdateLastUsed(keyID string, usedAt time.Time) error
}

// ValidationResult represents the result of validating an API key
type ValidationResult struct {
	Valid       bool
	APIKey      *APIKey
	Error       error
	Permissions []Permission
}
