package keystogo_test

import (
	"context"
	"testing"
	"time"

	"github.com/karurosux/keystogo/pkg/keystogo"
	"github.com/karurosux/keystogo/pkg/models"
	"github.com/karurosux/keystogo/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestExpiration_ValidateRejectsExpired(t *testing.T) {
	ctx := context.Background()
	a := assert.New(t)

	strg := storage.NewMemoryStorage()
	mngr := keystogo.NewManager(strg)

	expiredTime := time.Now().Add(-1 * time.Hour)
	permissions := &[]models.Permission{"read:users"}

	_, plainKey, err := mngr.GenerateApiKey(ctx, "expired-key", permissions, nil, &expiredTime)
	a.NoError(err)
	a.NotEmpty(plainKey)

	result := mngr.ValidateKey(ctx, plainKey, []models.Permission{"read:users"})
	a.False(result.Valid, "expired key should not be valid")
	a.Equal(models.ErrKeyExpired(), result.Error)
}

func TestExpiration_ListExcludesExpired(t *testing.T) {
	ctx := context.Background()
	a := assert.New(t)

	strg := storage.NewMemoryStorage()
	mngr := keystogo.NewManager(strg)

	futureTime := time.Now().Add(1 * time.Hour)
	pastTime := time.Now().Add(-1 * time.Hour)

	mngr.GenerateApiKey(ctx, "valid-key", nil, nil, &futureTime)
	mngr.GenerateApiKey(ctx, "expired-key", nil, nil, &pastTime)
	mngr.GenerateApiKey(ctx, "no-expiry", nil, nil, nil)

	keys, total, err := mngr.ListKeys(ctx, models.Page{}, models.Filter{})
	a.NoError(err)
	a.Equal(int64(2), total, "should only count non-expired keys")
	a.Len(keys, 2, "should only return non-expired keys")

	for _, key := range keys {
		a.NotEqual("expired-key", key.Name)
	}
}

func TestExpiration_CleanupExpired(t *testing.T) {
	ctx := context.Background()
	a := assert.New(t)

	strg := storage.NewMemoryStorage()
	mngr := keystogo.NewManager(strg)

	futureTime := time.Now().Add(1 * time.Hour)
	pastTime := time.Now().Add(-1 * time.Hour)

	mngr.GenerateApiKey(ctx, "valid-key-1", nil, nil, &futureTime)
	mngr.GenerateApiKey(ctx, "expired-key-1", nil, nil, &pastTime)
	mngr.GenerateApiKey(ctx, "expired-key-2", nil, nil, &pastTime)
	mngr.GenerateApiKey(ctx, "no-expiry", nil, nil, nil)

	count, err := mngr.CleanupExpired(ctx)
	a.NoError(err)
	a.Equal(int64(2), count, "should cleanup 2 expired keys")

	keys, total, err := mngr.ListKeys(ctx, models.Page{}, models.Filter{})
	a.NoError(err)
	a.Equal(int64(2), total)
	a.Len(keys, 2)
}

func TestExpiration_NoExpiryKeys(t *testing.T) {
	ctx := context.Background()
	a := assert.New(t)

	strg := storage.NewMemoryStorage()
	mngr := keystogo.NewManager(strg)

	_, plainKey, err := mngr.GenerateApiKey(ctx, "no-expiry-key", nil, nil, nil)
	a.NoError(err)

	time.Sleep(10 * time.Millisecond)

	result := mngr.ValidateKey(ctx, plainKey, nil)
	a.True(result.Valid, "key with no expiry should always be valid")
}
