package keystogo_test

import (
	"testing"
	"time"

	"github.com/karurosux/keystogo/pkg/keystogo"
	"github.com/karurosux/keystogo/pkg/models"
	"github.com/karurosux/keystogo/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestManager_EnableKey(t *testing.T) {
	const fakeId = "123"
	const fakeKey = "fake-key"

	a := assert.New(t)
	strg := storage.NewMemoryStorage()
	mngr := keystogo.NewManager(strg)

	err := mngr.EnableKey("")
	a.Error(err, "should fail on empty key")

	err = mngr.EnableKey("non-existent-key")
	a.Error(err, "should fail on non-existent key")

	strg.Create(&models.APIKey{
		ID:     fakeId,
		Name:   "test-key",
		Key:    keystogo.HashKey(fakeKey),
		Active: false,
	})

	err = mngr.EnableKey(fakeId)
	a.NoError(err, "should succeed for existing key by id.")

	apiKey, err := strg.GetByHashedKey(keystogo.HashKey(fakeKey))
	a.NoError(err, "should find the key")
	a.True(apiKey.Active, "key should be enabled")

	err = mngr.EnableKey(fakeId)
	a.NoError(err, "should succeed when enabling already enabled key")
}

func TestManager_DisableKey(t *testing.T) {
	const fakeKey = "fake-key"

	a := assert.New(t)
	strg := storage.NewMemoryStorage()
	mngr := keystogo.NewManager(strg)

	err := mngr.DisableKey("")
	a.Error(err, "should fail on empty key")

	err = mngr.DisableKey("non-existent-key")
	a.Error(err, "should fail on non-existent key")

	strg.Create(&models.APIKey{
		ID:     "123",
		Name:   "test-key",
		Key:    keystogo.HashKey(fakeKey),
		Active: true,
	})

	err = mngr.DisableKey("123")
	a.NoError(err, "should succeed for existing key")

	apiKey, err := strg.GetByHashedKey(keystogo.HashKey(fakeKey))
	a.NoError(err, "should find the key")
	a.False(apiKey.Active, "key should be disabled")

	err = mngr.DisableKey("123")
	a.NoError(err, "should succeed when disabling already disabled key")
}

func TestManager_DeleteKey(t *testing.T) {
	const fakeKey = "fake-key"
	const fakeId = "123"

	a := assert.New(t)
	strg := storage.NewMemoryStorage()
	mngr := keystogo.NewManager(strg)

	err := mngr.DeleteKey("")
	a.Error(err, "should fail on empty key")

	strg.Create(&models.APIKey{
		ID:     fakeId,
		Name:   "test-key",
		Key:    keystogo.HashKey(fakeKey),
		Active: true,
	})

	err = mngr.DeleteKey(fakeId)
	a.NoError(err, "should succeed for existing key")

	_, err = strg.GetByHashedKey(keystogo.HashKey(fakeKey))
	a.Error(err, "should failt for removed api key")
}

func TestManager_RenewKey(t *testing.T) {
	const originalId = "123"
	const originalKey = "fake-key"

	a := assert.New(t)
	hashedKey := keystogo.HashKey(originalKey)

	strg := storage.NewMemoryStorage()
	mngr := keystogo.NewManager(strg)

	apikey, key, err := mngr.RenewKey("")

	a.Error(err, "should fail for empty key")
	a.Empty(apikey, "should return empty apikey when key is empty")
	a.Empty(key, "should return empty key when key is empty")

	apikey, key, err = mngr.RenewKey("wrong-key")

	a.Error(err, "should fail for non-existing key")
	a.Empty(apikey, "should return empty apikey when key is non-existing")
	a.Empty(key, "should return empty key when key is non-existing")

	strg.Create(&models.APIKey{
		Name:   "test-key",
		Key:    hashedKey,
		Active: true,
	})

	apikey, key, err = mngr.RenewKey(originalKey)

	a.NoError(err, "should succeed for existing key")
	a.NotEqual(hashedKey, apikey.Key, "should return apikey with new hashed key")
	a.NotEqual(originalKey, key, "should return same new unhashed key")

	oldApiKey, _ := strg.GetByHashedKey(hashedKey)
	a.False(oldApiKey.Active, "should disable old key")
}

func TestManager_ListKeys(t *testing.T) {
	const fakeKey = "fake-key"

	a := assert.New(t)
	strg := storage.NewMemoryStorage()
	mngr := keystogo.NewManager(strg)

	keys, total, _ := mngr.ListKeys(models.Page{}, models.Filter{})
	a.Len(keys, 0, "should return empty list")
	a.Equal(total, int64(0), "should return total of 0")

	strg.Create(&models.APIKey{
		Name:   "test-key",
		Key:    keystogo.HashKey(fakeKey),
		Active: true,
	})

	keys, total, err := mngr.ListKeys(models.Page{Limit: 10, Offset: 0}, models.Filter{})
	a.NoError(err, "should succeed for existing key")
	a.Len(keys, 1, "should return 1 key")
	a.Equal(total, int64(1), "should return total of 1")
}

func TestManager_GenerateApiKey(t *testing.T) {
	a := assert.New(t)
	strg := storage.NewMemoryStorage()
	mngr := keystogo.NewManager(strg)

	name := "test-key"
	permissions := &[]models.Permission{"modes:read", "modes:write"}
	metadata := &map[string]any{"test": "value"}
	expiresAt := time.Now().Add(time.Hour)

	apiKey, key, err := mngr.GenerateApiKey(name, permissions, metadata, &expiresAt)

	a.NoError(err, "should succeed")
	a.NotEmpty(apiKey.ID, "should generate a new ID")
	a.NotEmpty(apiKey.Key, "should generate a new key")
	a.Equal(name, apiKey.Name, "should set name")
	a.Equal(permissions, apiKey.Permissions, "should set permissions")
	a.Equal(metadata, apiKey.Metadata, "should set metadata")
	a.Equal(expiresAt, *apiKey.ExpiresAt, "should set expiresAt")
	a.True(apiKey.Active, "should set active to true")
	// key is meant to be public.
	a.NotEqual(key, apiKey.Key, "should return different key from api key and key.")
}

func TestManager_ValidateKey(t *testing.T) {
	const fakeKey = "fake-key"

	a := assert.New(t)
	strg := storage.NewMemoryStorage()
	mngr := keystogo.NewManager(strg)

	res := mngr.ValidateKey("", []models.Permission{})
	a.False(res.Valid, "should fail on empty key")
	a.Equal(models.ErrKeyNotFound, res.Error, "should return ErrKeyNotFound")

	strg.Create(&models.APIKey{
		ID:     "id",
		Name:   "test-key",
		Key:    keystogo.HashKey(fakeKey),
		Active: true,
	})

	res = mngr.ValidateKey(fakeKey, nil)
	a.True(res.Valid, "should succeed for existing key")
	a.Equal(keystogo.HashKey(fakeKey), res.APIKey.Key, "should return same key")

	res = mngr.ValidateKey(fakeKey, []models.Permission{"read:users"})
	a.False(res.Valid, "should fail if key does not have required permissions")

	strg.Create(&models.APIKey{
		ID:          "id",
		Name:        "test-key",
		Key:         keystogo.HashKey(fakeKey),
		Active:      true,
		Permissions: &[]models.Permission{"read:users"},
	})

	res = mngr.ValidateKey(fakeKey, []models.Permission{"read:users"})
	a.True(res.Valid, "should succeed if key has required permissions")
}
