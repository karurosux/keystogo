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
	Active *bool   // Filter by active status
	Name   *string // Filter by name (partial match)
}

// Page represents pagination parameters
type Page struct {
	Offset int
	Limit  int
}
