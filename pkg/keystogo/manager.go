package keystogo

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/karurosux/keystogo/pkg/models"
)

// Manager handles API key validation and permission checking
type Manager struct {
	Storage Storage
}

// NewManager creates a new API key manager with the given storage implementation
func NewManager(storage Storage) *Manager {
	return &Manager{
		Storage: storage,
	}
}

// ValidateKey checks if an API key is valid and has the required permissions
func (m *Manager) ValidateKey(key string, requiredPermissions []models.Permission) models.ValidationResult {
	hashedKey := HashKey(key)

	apiKey, err := m.Storage.GetByHashedKey(hashedKey)
	if err != nil {
		return models.ValidationResult{Valid: false, Error: models.ErrKeyNotFound}
	}

	// Check if key is active
	if !apiKey.Active {
		return models.ValidationResult{Valid: false, Error: models.ErrKeyInactive}
	}

	// Check expiration
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return models.ValidationResult{Valid: false, Error: models.ErrKeyExpired}
	}

	// Check permissions
	if apiKey.Permissions != nil && !hasRequiredPermissions(*apiKey.Permissions, requiredPermissions) {
		return models.ValidationResult{Valid: false, Error: models.ErrPermissionDenied}
	} else if apiKey.Permissions == nil && requiredPermissions != nil {
		return models.ValidationResult{Valid: false, Error: models.ErrPermissionDenied}
	}

	// Update last used timestamp
	now := time.Now()
	if err := m.Storage.Update(apiKey.ID, models.ApiKeyUpdate{
		LastUsedAt: &now,
	}); err != nil {
		return models.ValidationResult{Valid: false, Error: err}
	}

	return models.ValidationResult{
		Valid:  true,
		APIKey: apiKey,
	}
}

func (m *Manager) DisableKey(id string) error {
	if id == "" {
		return errors.New("key is required")
	}

	active := false
	if err := m.Storage.Update(id, models.ApiKeyUpdate{
		Active: &active,
	}); err != nil {
		return err
	}
	return nil
}

func (m *Manager) EnableKey(id string) error {
	if id == "" {
		return errors.New("key is required")
	}

	active := true
	if err := m.Storage.Update(id, models.ApiKeyUpdate{
		Active: &active,
	}); err != nil {
		return err
	}
	return nil
}

func (m *Manager) DeleteKey(id string) error {
	if id == "" {
		return errors.New("key is required")
	}
	return m.Storage.Delete(id)
}

// RenewKey creates a new API key while invalidating the old one
func (m *Manager) RenewKey(key string) (models.APIKey, string, error) {
	if key == "" {
		return models.APIKey{}, "", errors.New("key is required")
	}
	oldHash := HashKey(key)
	oldKey, err := m.Storage.GetByHashedKey(oldHash)
	if err != nil {
		return models.APIKey{}, "", err
	}

	// Create new key with same permissions and metadata
	newKey, keyStr, err := m.GenerateApiKey(
		oldKey.Name,
		oldKey.Permissions,
		oldKey.Metadata,
		oldKey.ExpiresAt,
	)
	if err != nil {
		return models.APIKey{}, "", err
	}

	// Disable old key
	if err := m.DisableKey(oldKey.ID); err != nil {
		return models.APIKey{}, "", err
	}

	return newKey, keyStr, nil
}

// ListKeys returns a paginated list of API keys
func (m *Manager) ListKeys(page models.Page, filter models.Filter) ([]models.APIKey, int64, error) {
	return m.Storage.List(page, filter)
}

func (m *Manager) GenerateApiKey(name string, permissions *[]models.Permission, metadata *map[string]any, expiresAt *time.Time) (models.APIKey, string, error) {
	if name == "" {
		return models.APIKey{}, "", errors.New("name is required")
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return models.APIKey{}, "", fmt.Errorf("failed to generate random key: %w", err)
	}

	keyStr := hex.EncodeToString(key)

	apiKey := models.APIKey{
		ID:          uuid.NewString(),
		Name:        name,
		Key:         HashKey(keyStr),
		Permissions: permissions,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		Active:      true,
	}

	if err := m.Storage.Create(&apiKey); err != nil {
		return models.APIKey{}, "", err
	}

	return apiKey, keyStr, nil
}

func (m *Manager) Update(key string, update models.ApiKeyUpdate) error {
	return m.Storage.Update(key, update)
}

// Helper function to check if an API key has all required permissions
func hasRequiredPermissions(keyPerms []models.Permission, requiredPerms []models.Permission) bool {
	if len(requiredPerms) == 0 {
		return true
	}

	permMap := make(map[models.Permission]bool)
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

// Helper function to hash API keys
func HashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
