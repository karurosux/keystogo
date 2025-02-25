package models

import (
	"time"
)

// Permission represents a single capability that an API key can have
type Permission string

// APIKey represents the structure of an API key with its metadata
type APIKey struct {
	ID          string         `json:"id"`
	Key         string         `json:"key"`  // Hashed key value
	Name        string         `json:"name"` // Human readable name
	Permissions []Permission   `json:"permissions"`
	Metadata    map[string]any `json:"metadata"`
	CreatedAt   time.Time      `json:"createdAt"`
	ExpiresAt   *time.Time     `json:"expiresAt"` // Optional expiration
	LastUsedAt  *time.Time     `json:"lastUsedAt"`
	Active      bool           `json:"active"`
}
