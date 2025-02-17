package models

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
	Metadata    map[string]any
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
