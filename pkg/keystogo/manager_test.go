package keystogo_test

import (
	"testing"

	"github.com/karurosux/keystogo/pkg/keystogo"
	"github.com/karurosux/keystogo/pkg/models"
	"github.com/karurosux/keystogo/pkg/storage"
)

func getManager() (*keystogo.Manager, keystogo.Storage) {
	storage := storage.NewMemoryStorage()
	manager := keystogo.NewManager(storage)
	return manager, storage
}

func TestManager_EnableKey(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		storage keystogo.Storage
		// Named input parameters for target function.
		key     string
		wantErr bool
	}{
		{name: "should fail for empty key", storage: storage.NewMemoryStorage(), key: "", wantErr: true},
		{name: "should fail for non-existing key", storage: storage.NewMemoryStorage(), key: "non-existing-key", wantErr: true},
		{name: "should enable existing key", storage: (func() keystogo.Storage {
			storage := storage.NewMemoryStorage()
			storage.Create(&models.APIKey{
				Name:   "test-key",
				Key:    keystogo.HashKey("fake-key"),
				Active: false,
			})
			return storage
		})(), key: "fake-key", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := keystogo.NewManager(tt.storage)
			gotErr := m.EnableKey(tt.key)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("EnableKey() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("EnableKey() succeeded unexpectedly")
			}
		})
	}
}

func TestManager_DisableKey(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		storage keystogo.Storage
		// Named input parameters for target function.
		key     string
		wantErr bool
	}{
		{name: "should fail for empty key", storage: storage.NewMemoryStorage(), key: "", wantErr: true},
		{name: "should fail for non-existing key", storage: storage.NewMemoryStorage(), key: "non-existing-key", wantErr: true},
		{name: "should disable existing key", storage: (func() keystogo.Storage {
			storage := storage.NewMemoryStorage()
			storage.Create(&models.APIKey{
				Name:   "test-key",
				Key:    keystogo.HashKey("fake-key"),
				Active: true,
			})
			return storage
		})(), key: "fake-key", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := keystogo.NewManager(tt.storage)
			gotErr := m.DisableKey(tt.key)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DisableKey() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DisableKey() succeeded unexpectedly")
			}
		})
	}
}

func TestManager_DeleteKey(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		storage keystogo.Storage
		// Named input parameters for target function.
		key     string
		wantErr bool
	}{
		{name: "should fail for empty key", storage: storage.NewMemoryStorage(), key: "", wantErr: true},
		{name: "should delete existing key", storage: (func() keystogo.Storage {
			storage := storage.NewMemoryStorage()
			storage.Create(&models.APIKey{
				Name:   "test-key",
				Key:    keystogo.HashKey("fake-key"),
				Active: true,
			})
			return storage
		})(), key: "fake-key", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := keystogo.NewManager(tt.storage)
			gotErr := m.DeleteKey(tt.key)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteKey() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteKey() succeeded unexpectedly")
			}
		})
	}
}

func TestManager_RenewKey(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		storage keystogo.Storage
		// Named input parameters for target function.
		key     string
		want    models.APIKey
		want2   string
		wantErr bool
	}{
		{name: "should fail for empty key", storage: storage.NewMemoryStorage(), key: "", wantErr: true},
		{name: "should fail for non-existing key", storage: storage.NewMemoryStorage(), key: "non-existing-key", wantErr: true},
		// {
		// 	name: "should renew existing key",
		// 	storage: (func() keystogo.Storage {
		// 		storage := storage.NewMemoryStorage()
		// 		storage.Create(&models.APIKey{
		// 			Name:   "test-key",
		// 			Key:    keystogo.HashKey("fake-key"),
		// 			Active: true,
		// 		})
		// 		return storage
		// 	})(),
		// 	key: "fake-key",
		// 	want: models.APIKey{
		// 		Name:   "test-key",
		// 		Key:    keystogo.HashKey("fake-key"),
		// 		Active: true,
		// 	}, want2: "fake-key", wantErr: false,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := keystogo.NewManager(tt.storage)
			got, got2, gotErr := m.RenewKey(tt.key)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("RenewKey() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("RenewKey() succeeded unexpectedly")
			}
			if got.Key != tt.want.Key {
				t.Errorf("RenewKey() = %v, want %v", got, tt.want)
			}
			if got2 != tt.want2 {
				t.Errorf("RenewKey() = %v, want %v", got2, tt.want2)
			}
		})
	}
}
