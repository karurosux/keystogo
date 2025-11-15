package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/karurosux/keystogo/pkg/keystogo"
	"github.com/karurosux/keystogo/pkg/models"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	KeyPrefix      string
	HashIndexKey   string
	DefaultTTL     time.Duration
	UseHashStorage bool
}

func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		KeyPrefix:      "apikey:",
		HashIndexKey:   "apikey:hash_index",
		DefaultTTL:     0,
		UseHashStorage: true,
	}
}

type RedisStorage struct {
	client *redis.Client
	config RedisConfig
}

func NewRedisStorage(client *redis.Client, config *RedisConfig) keystogo.Storage {
	if config == nil {
		defaultConfig := DefaultRedisConfig()
		config = &defaultConfig
	}

	return &RedisStorage{
		client: client,
		config: *config,
	}
}

func (r *RedisStorage) keyForID(id string) string {
	return r.config.KeyPrefix + "id:" + id
}

func (r *RedisStorage) GetByID(ctx context.Context, id string) (*models.APIKey, error) {
	key := r.keyForID(id)

	if r.config.UseHashStorage {
		return r.getFromHash(ctx, key)
	}
	return r.getFromString(ctx, key)
}

func (r *RedisStorage) GetByHashedKey(ctx context.Context, hashedKey string) (*models.APIKey, error) {
	id, err := r.client.HGet(ctx, r.config.HashIndexKey, hashedKey).Result()
	if err == redis.Nil {
		return nil, models.ErrKeyNotFound()
	}
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *RedisStorage) Create(ctx context.Context, apiKey *models.APIKey) error {
	key := r.keyForID(apiKey.ID)

	var err error
	if r.config.UseHashStorage {
		err = r.saveToHash(ctx, key, apiKey)
	} else {
		err = r.saveToString(ctx, key, apiKey)
	}
	if err != nil {
		return err
	}

	err = r.client.HSet(ctx, r.config.HashIndexKey, apiKey.Key, apiKey.ID).Err()
	if err != nil {
		return err
	}

	ttl := r.config.DefaultTTL
	if apiKey.ExpiresAt != nil {
		ttl = time.Until(*apiKey.ExpiresAt)
		if ttl < 0 {
			ttl = 0
		}
	}

	if ttl > 0 {
		r.client.Expire(ctx, key, ttl)
	}

	return nil
}

func (r *RedisStorage) Update(ctx context.Context, id string, update models.ApiKeyUpdate) error {
	existing, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if update.Active != nil {
		existing.Active = *update.Active
	}
	if update.Name != nil {
		existing.Name = *update.Name
	}
	if update.ExpiresAt != nil {
		existing.ExpiresAt = update.ExpiresAt
	}
	if update.Metadata != nil {
		existing.Metadata = update.Metadata
	}
	if update.Permissions != nil {
		existing.Permissions = update.Permissions
	}
	if update.LastUsedAt != nil {
		existing.LastUsedAt = update.LastUsedAt
	}

	key := r.keyForID(id)
	if r.config.UseHashStorage {
		err = r.saveToHash(ctx, key, existing)
	} else {
		err = r.saveToString(ctx, key, existing)
	}
	if err != nil {
		return err
	}

	ttl := r.config.DefaultTTL
	if existing.ExpiresAt != nil {
		ttl = time.Until(*existing.ExpiresAt)
		if ttl < 0 {
			ttl = 0
		}
	}

	if ttl > 0 {
		r.client.Expire(ctx, key, ttl)
	} else if ttl == 0 && existing.ExpiresAt == nil {
		r.client.Persist(ctx, key)
	}

	return nil
}

func (r *RedisStorage) Delete(ctx context.Context, id string) error {
	apiKey, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	key := r.keyForID(id)

	err = r.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	return r.client.HDel(ctx, r.config.HashIndexKey, apiKey.Key).Err()
}

func (r *RedisStorage) List(ctx context.Context, page models.Page, filter models.Filter) ([]models.APIKey, int64, error) {
	pattern := r.config.KeyPrefix + "id:*"
	var cursor uint64
	var keys []string

	for {
		var batch []string
		var err error
		batch, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, 0, err
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}

	result := make([]models.APIKey, 0)
	now := time.Now()

	for _, key := range keys {
		var apiKey *models.APIKey
		var err error

		if r.config.UseHashStorage {
			apiKey, err = r.getFromHash(ctx, key)
		} else {
			apiKey, err = r.getFromString(ctx, key)
		}

		if err != nil {
			continue
		}

		if apiKey.ExpiresAt != nil && now.After(*apiKey.ExpiresAt) {
			continue
		}

		matches := true
		if filter.Name != nil && *filter.Name != "" {
			matches = matches && strings.Contains(strings.ToLower(apiKey.Name), strings.ToLower(*filter.Name))
		}
		if filter.Active != nil {
			matches = matches && (apiKey.Active == *filter.Active)
		}

		if matches {
			result = append(result, *apiKey)
		}
	}

	total := int64(len(result))

	if page.Limit > 0 {
		start := page.Offset
		end := start + page.Limit

		if start < len(result) {
			if end > len(result) {
				end = len(result)
			}
			result = result[start:end]
		} else {
			result = []models.APIKey{}
		}
	}

	return result, total, nil
}

func (r *RedisStorage) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisStorage) Clear(ctx context.Context) error {
	pattern := r.config.KeyPrefix + "*"
	var cursor uint64

	for {
		var keys []string
		var err error
		keys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		if cursor == 0 {
			break
		}
	}

	return r.client.Del(ctx, r.config.HashIndexKey).Err()
}

func (r *RedisStorage) CleanupExpired(ctx context.Context) (int64, error) {
	pattern := r.config.KeyPrefix + "id:*"
	var cursor uint64
	var keys []string

	for {
		var batch []string
		var err error
		batch, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return 0, err
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}

	count := int64(0)
	now := time.Now()

	for _, key := range keys {
		var apiKey *models.APIKey
		var err error

		if r.config.UseHashStorage {
			apiKey, err = r.getFromHash(ctx, key)
		} else {
			apiKey, err = r.getFromString(ctx, key)
		}

		if err != nil {
			continue
		}

		if apiKey.ExpiresAt != nil && now.After(*apiKey.ExpiresAt) {
			if err := r.Delete(ctx, apiKey.ID); err == nil {
				count++
			}
		}
	}

	return count, nil
}

func (r *RedisStorage) saveToHash(ctx context.Context, key string, apiKey *models.APIKey) error {
	permissionsJSON, err := json.Marshal(apiKey.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	metadataJSON, err := json.Marshal(apiKey.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	fields := map[string]interface{}{
		"id":          apiKey.ID,
		"key":         apiKey.Key,
		"name":        apiKey.Name,
		"permissions": string(permissionsJSON),
		"metadata":    string(metadataJSON),
		"created_at":  apiKey.CreatedAt.Format(time.RFC3339),
		"active":      apiKey.Active,
	}

	if apiKey.ExpiresAt != nil {
		fields["expires_at"] = apiKey.ExpiresAt.Format(time.RFC3339)
	}

	if apiKey.LastUsedAt != nil {
		fields["last_used_at"] = apiKey.LastUsedAt.Format(time.RFC3339)
	}

	return r.client.HSet(ctx, key, fields).Err()
}

func (r *RedisStorage) getFromHash(ctx context.Context, key string) (*models.APIKey, error) {
	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, models.ErrKeyNotFound()
	}

	apiKey := &models.APIKey{}
	apiKey.ID = data["id"]
	apiKey.Key = data["key"]
	apiKey.Name = data["name"]

	if data["active"] == "true" || data["active"] == "1" {
		apiKey.Active = true
	}

	if data["permissions"] != "" {
		if err := json.Unmarshal([]byte(data["permissions"]), &apiKey.Permissions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}
	}

	if data["metadata"] != "" {
		if err := json.Unmarshal([]byte(data["metadata"]), &apiKey.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	if data["created_at"] != "" {
		createdAt, err := time.Parse(time.RFC3339, data["created_at"])
		if err == nil {
			apiKey.CreatedAt = createdAt
		}
	}

	if data["expires_at"] != "" {
		expiresAt, err := time.Parse(time.RFC3339, data["expires_at"])
		if err == nil {
			apiKey.ExpiresAt = &expiresAt
		}
	}

	if data["last_used_at"] != "" {
		lastUsedAt, err := time.Parse(time.RFC3339, data["last_used_at"])
		if err == nil {
			apiKey.LastUsedAt = &lastUsedAt
		}
	}

	return apiKey, nil
}

func (r *RedisStorage) saveToString(ctx context.Context, key string, apiKey *models.APIKey) error {
	data, err := json.Marshal(apiKey)
	if err != nil {
		return fmt.Errorf("failed to marshal api key: %w", err)
	}

	return r.client.Set(ctx, key, data, r.config.DefaultTTL).Err()
}

func (r *RedisStorage) getFromString(ctx context.Context, key string) (*models.APIKey, error) {
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, models.ErrKeyNotFound()
	}
	if err != nil {
		return nil, err
	}

	var apiKey models.APIKey
	if err := json.Unmarshal([]byte(data), &apiKey); err != nil {
		return nil, fmt.Errorf("failed to unmarshal api key: %w", err)
	}

	return &apiKey, nil
}
