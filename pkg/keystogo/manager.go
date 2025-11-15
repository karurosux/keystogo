package keystogo

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
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

func (m *Manager) ValidateKey(ctx context.Context, key string, requiredPermissions []models.Permission) models.ValidationResult {
	if key == "" {
		return models.ValidationResult{Valid: false, Error: models.ErrEmptyKey()}
	}

	hashedKey := HashKey(key)

	apiKey, err := m.Storage.GetByHashedKey(ctx, hashedKey)
	if err != nil {
		return models.ValidationResult{Valid: false, Error: models.ErrKeyNotFound()}
	}

	if !apiKey.Active {
		return models.ValidationResult{Valid: false, Error: models.ErrKeyInactive()}
	}

	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return models.ValidationResult{Valid: false, Error: models.ErrKeyExpired()}
	}

	if requiredPermissions != nil && len(requiredPermissions) > 0 {
		if apiKey.Permissions == nil {
			return models.ValidationResult{Valid: false, Error: models.ErrPermissionDenied()}
		}
		if !hasRequiredPermissions(*apiKey.Permissions, requiredPermissions) {
			return models.ValidationResult{Valid: false, Error: models.ErrPermissionDenied()}
		}
	}

	now := time.Now()
	if err := m.Storage.Update(ctx, apiKey.ID, models.ApiKeyUpdate{
		LastUsedAt: &now,
	}); err != nil {
		return models.ValidationResult{Valid: false, Error: err}
	}

	return models.ValidationResult{
		Valid:  true,
		APIKey: apiKey,
	}
}

func (m *Manager) DisableKey(ctx context.Context, id string) error {
	if id == "" {
		return models.ErrEmptyID()
	}

	active := false
	if err := m.Storage.Update(ctx, id, models.ApiKeyUpdate{
		Active: &active,
	}); err != nil {
		return err
	}
	return nil
}

func (m *Manager) EnableKey(ctx context.Context, id string) error {
	if id == "" {
		return models.ErrEmptyID()
	}

	active := true
	if err := m.Storage.Update(ctx, id, models.ApiKeyUpdate{
		Active: &active,
	}); err != nil {
		return err
	}
	return nil
}

func (m *Manager) DeleteKey(ctx context.Context, id string) error {
	if id == "" {
		return models.ErrEmptyID()
	}
	return m.Storage.Delete(ctx, id)
}

func (m *Manager) RenewKey(ctx context.Context, key string) (models.APIKey, string, error) {
	if key == "" {
		return models.APIKey{}, "", models.ErrEmptyKey()
	}
	oldHash := HashKey(key)
	oldKey, err := m.Storage.GetByHashedKey(ctx, oldHash)
	if err != nil {
		return models.APIKey{}, "", err
	}

	newKey, keyStr, err := m.GenerateApiKey(
		ctx,
		oldKey.Name,
		oldKey.Permissions,
		oldKey.Metadata,
		oldKey.ExpiresAt,
	)
	if err != nil {
		return models.APIKey{}, "", err
	}

	if err := m.DisableKey(ctx, oldKey.ID); err != nil {
		return models.APIKey{}, "", err
	}

	return newKey, keyStr, nil
}

func (m *Manager) ListKeys(ctx context.Context, page models.Page, filter models.Filter) ([]models.APIKey, int64, error) {
	return m.Storage.List(ctx, page, filter)
}

func (m *Manager) GenerateApiKey(ctx context.Context, name string, permissions *[]models.Permission, metadata *map[string]any, expiresAt *time.Time) (models.APIKey, string, error) {
	if name == "" {
		return models.APIKey{}, "", models.ErrEmptyName()
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

	if err := m.Storage.Create(ctx, &apiKey); err != nil {
		return models.APIKey{}, "", err
	}

	return apiKey, keyStr, nil
}

func (m *Manager) Update(ctx context.Context, key string, update models.ApiKeyUpdate) error {
	if key == "" {
		return models.ErrEmptyID()
	}

	if update.Name != nil && *update.Name == "" {
		return models.ErrEmptyName()
	}

	return m.Storage.Update(ctx, key, update)
}

func (m *Manager) CleanupExpired(ctx context.Context) (int64, error) {
	return m.Storage.CleanupExpired(ctx)
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
