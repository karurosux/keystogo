package keystogo

import (
	"context"

	"github.com/karurosux/keystogo/pkg/models"
)

type Storage interface {
	GetByID(ctx context.Context, id string) (*models.APIKey, error)
	GetByHashedKey(ctx context.Context, hashedKey string) (*models.APIKey, error)
	Create(ctx context.Context, apiKey *models.APIKey) error
	Update(ctx context.Context, id string, apiKey models.ApiKeyUpdate) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page models.Page, filter models.Filter) ([]models.APIKey, int64, error)
	Ping(ctx context.Context) error
	Clear(ctx context.Context) error
	CleanupExpired(ctx context.Context) (int64, error)
}
