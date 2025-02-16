package keystogo

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

var (
	ErrKeyNotFound       = errors.New("api key not found")
	ErrKeyExpired        = errors.New("api key has expired")
	ErrKeyInactive       = errors.New("api key is inactive")
	ErrPermissionDenied  = errors.New("api key lacks required permission")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

// Manager handles API key validation and permission checking
type Manager struct {
	storage Storage
}

// NewManager creates a new API key manager with the given storage implementation
func NewManager(storage Storage) *Manager {
	return &Manager{
		storage: storage,
	}
}

// ValidateKey checks if an API key is valid and has the required permissions
func (m *Manager) ValidateKey(key string, requiredPermissions []Permission) ValidationResult {
	hashedKey := hashKey(key)

	apiKey, err := m.storage.GetAPIKey(hashedKey)
	if err != nil {
		return ValidationResult{Valid: false, Error: ErrKeyNotFound}
	}

	// Check if key is active
	if !apiKey.Active {
		return ValidationResult{Valid: false, Error: ErrKeyInactive}
	}

	// Check expiration
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return ValidationResult{Valid: false, Error: ErrKeyExpired}
	}

	// Check permissions
	if !hasRequiredPermissions(apiKey.Permissions, requiredPermissions) {
		return ValidationResult{Valid: false, Error: ErrPermissionDenied}
	}

	// Update last used timestamp
	_ = m.storage.UpdateLastUsed(apiKey.ID, time.Now())

	return ValidationResult{
		Valid:       true,
		APIKey:      apiKey,
		Permissions: apiKey.Permissions,
	}
}

// Helper function to hash API keys
func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// Helper function to check if an API key has all required permissions
func hasRequiredPermissions(keyPerms []Permission, requiredPerms []Permission) bool {
	if len(requiredPerms) == 0 {
		return true
	}

	permMap := make(map[Permission]bool)
	for _, p := range keyPerms {
		permMap[p] = true
	}

	for _, required := range requiredPerms {
		if !permMap[required] {
			return false
		}
	}

	return true
}
