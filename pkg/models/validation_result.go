package models

// ValidationResult represents the result of validating an API key
type ValidationResult struct {
	Valid       bool
	APIKey      *APIKey
	Error       error
	Permissions []Permission
}
