package models

import "time"

type ApiKeyUpdate struct {
	Name        *string
	Active      *bool
	ExpiresAt   *time.Time
	LastUsedAt  *time.Time
	Metadata    *map[string]any
	Permissions *[]Permission
}

type Filter struct {
	Active      *bool      // Filter by active status
	Name        *string    // Filter by name (partial match)
	Tags        *[]string  // Filter by tags
	CreatedFrom *time.Time // Filter by creation date range
	CreatedTo   *time.Time
	Metadata    *map[string]string // Filter by metadata
}

// Page represents pagination parameters
type Page struct {
	Offset int
	Limit  int
}
