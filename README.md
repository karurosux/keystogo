# keystogo

A Go library for managing API keys. Generate, validate, and manage API keys with permission controls and customizable storage backends.

[![Go Reference](https://pkg.go.dev/badge/github.com/karurosux/keystogo.svg)](https://pkg.go.dev/github.com/karurosux/keystogo)
[![Go Report Card](https://goreportcard.com/badge/github.com/karurosux/keystogo)](https://goreportcard.com/report/github.com/karurosux/keystogo)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- Generate cryptographically secure API keys (32-byte random values)
- Validate keys with permission-based access control
- Key lifecycle management (enable, disable, renew, delete)
- Key expiration support
- Pluggable storage backends
- SHA-256 hashing for secure key storage

## Installation

```bash
go get github.com/karurosux/keystogo
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/karurosux/keystogo/pkg/keystogo"
    "github.com/karurosux/keystogo/pkg/models"
    "github.com/karurosux/keystogo/pkg/storage"
)

func main() {
    ctx := context.Background()

    storage := storage.NewMemoryStorage()
    manager := keystogo.NewManager(storage)

    expiresAt := time.Now().Add(365 * 24 * time.Hour)
    permissions := &[]models.Permission{"read:users", "write:users"}

    apiKey, plainKey, err := manager.GenerateApiKey(ctx, "my-service", permissions, nil, &expiresAt)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Generated key ID: %s", apiKey.ID)
    log.Printf("Plain key (save this, it won't be shown again): %s", plainKey)

    result := manager.ValidateKey(ctx, plainKey, []models.Permission{"read:users"})
    if !result.Valid {
        log.Printf("Validation failed: %v", result.Error)
        return
    }

    log.Println("Key is valid")
}
```

## How It Works

Keys are generated as cryptographically random 32-byte values and hashed with SHA-256 before storage. The plain text key is only returned once during generation and cannot be recovered later. This library is designed for system-generated keys, not user-provided keys.

## Manager Methods

The Manager provides methods for key management:

```go
type Manager interface {
    ValidateKey(ctx context.Context, key string, requiredPermissions []models.Permission) models.ValidationResult
    DisableKey(ctx context.Context, id string) error
    EnableKey(ctx context.Context, id string) error
    DeleteKey(ctx context.Context, id string) error
    RenewKey(ctx context.Context, key string) (models.APIKey, string, error)
    ListKeys(ctx context.Context, page models.Page, filter models.Filter) ([]models.APIKey, int64, error)
    GenerateApiKey(ctx context.Context, name string, permissions *[]models.Permission, metadata *map[string]any, expiresAt *time.Time) (models.APIKey, string, error)
    Update(ctx context.Context, id string, update models.ApiKeyUpdate) error
}
```

## Key Expiration

Keys can be created with an optional expiration time. Expired keys are automatically filtered from list operations and rejected during validation.

```go
expiresAt := time.Now().Add(30 * 24 * time.Hour)
apiKey, plainKey, err := manager.GenerateApiKey(ctx, "temp-key", nil, nil, &expiresAt)

result := manager.ValidateKey(ctx, plainKey, nil)
if !result.Valid {
    log.Printf("Key expired: %v", result.Error)
}
```

### Automatic Expiration Handling

All storage backends automatically filter expired keys from list operations. Additionally, you can manually clean up expired keys:

```go
count, err := manager.CleanupExpired(ctx)
if err != nil {
    log.Fatal(err)
}
log.Printf("Cleaned up %d expired keys", count)
```

### Storage-Specific Behavior

- **Redis**: Uses native TTL expiration. Keys are automatically removed when they expire.
- **PostgreSQL**: Expired keys remain in the database until `CleanupExpired()` is called. List operations filter them using SQL `WHERE` clauses.
- **Memory**: Expired keys remain in memory until `CleanupExpired()` is called. List operations filter them at runtime.

## Storage Backends

Included storage implementations:

- In-memory (for testing and development)
- PostgreSQL (with configurable table names and column mapping)
- Redis (with configurable key prefixes and storage modes)

### PostgreSQL Storage

Use the same database connection for multiple key tables (licenses, API keys, service tokens, etc.):

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
    "github.com/karurosux/keystogo/pkg/storage"
)

db, _ := sql.Open("postgres", "postgres://user:pass@localhost/db?sslmode=disable")

licenseConfig := storage.PostgresConfig{
    TableName: "licenses",
    SchemaName: "public",
    ColumnMapping: storage.DefaultPostgresConfig().ColumnMapping,
}
licenseStorage := storage.NewPostgresStorage(db, &licenseConfig)
licenseManager := keystogo.NewManager(licenseStorage)

apiKeyConfig := storage.PostgresConfig{
    TableName: "api_keys",
    SchemaName: "public",
    ColumnMapping: storage.DefaultPostgresConfig().ColumnMapping,
}
apiKeyStorage := storage.NewPostgresStorage(db, &apiKeyConfig)
apiKeyManager := keystogo.NewManager(apiKeyStorage)
```

### Redis Storage

Configure key prefixes for different key types:

```go
import (
    "github.com/redis/go-redis/v9"
    "github.com/karurosux/keystogo/pkg/storage"
)

client := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

licenseConfig := storage.RedisConfig{
    KeyPrefix: "license:",
    HashIndexKey: "license:hash_index",
    UseHashStorage: true,
}
licenseStorage := storage.NewRedisStorage(client, &licenseConfig)
licenseManager := keystogo.NewManager(licenseStorage)

apiKeyConfig := storage.RedisConfig{
    KeyPrefix: "apikey:",
    HashIndexKey: "apikey:hash_index",
    UseHashStorage: true,
}
apiKeyStorage := storage.NewRedisStorage(client, &apiKeyConfig)
apiKeyManager := keystogo.NewManager(apiKeyStorage)
```

### Custom Storage

You can implement your own storage backend by implementing the `Storage` interface:

```go
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
```

## Usage Examples

### Basic Validation

```go
ctx := context.Background()
result := manager.ValidateKey(ctx, apiKey, []models.Permission{"read:users"})
if !result.Valid {
    return fmt.Errorf("invalid API key: %v", result.Error)
}
```

### With Custom Storage

```go
type CustomStorage struct {
    // Your storage implementation
}

func NewCustomStorage() keystogo.Storage {
    return &CustomStorage{}
}

func (s *CustomStorage) GetByHashedKey(ctx context.Context, hashedKey string) (*models.APIKey, error) {
    // Your implementation
}

func (s *CustomStorage) GetByID(ctx context.Context, id string) (*models.APIKey, error) {
    // Your implementation
}

// ... Implement other Storage interface methods

storage := NewCustomStorage()
manager := keystogo.NewManager(storage)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- [Documentation](https://pkg.go.dev/github.com/karurosux/keystogo)
- [Issue Tracker](https://github.com/karurosux/keystogo/issues)
- [Discussions](https://github.com/karurosux/keystogo/discussions)

## Roadmap

- [x] PostgreSQL storage implementation
- [x] Redis storage implementation
- [ ] Key rotation capabilities
- [ ] Batch key operations
- [ ] Key generation utilities
