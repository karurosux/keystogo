# keystogo

A flexible API key management library for Go applications. Keystogo provides a complete solution for handling API key validation, permissions, and lifecycle management with customizable storage backends.

[![Go Reference](https://pkg.go.dev/badge/github.com/karurosux/keystogo.svg)](https://pkg.go.dev/github.com/karurosux/keystogo)
[![Go Report Card](https://goreportcard.com/badge/github.com/karurosux/keystogo)](https://goreportcard.com/report/github.com/karurosux/keystogo)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- 🔐 API key validation and management
- 🔑 Flexible permission-based access control
- 💾 Pluggable storage backends
- ⏰ Key expiration management
- 🔄 Easy integration with existing Go applications

## Installation

```bash
go get github.com/karurosux/keystogo
```

## Quick Start

```go
package main

import (
    "log"
    "github.com/karurosux/keystogo/pkg/keystogo"
    "github.com/karurosux/keystogo/pkg/models"
    "github.com/karurosux/keystogo/pkg/storage/memory"
)

func main() {
    // Initialize with in-memory storage
    storage := memory.NewMemoryStorage()
    manager := keystogo.NewManager(storage)

    // Validate an API key with required permissions
    result := manager.ValidateKey("your-api-key", []models.Permission{"read:users"})

    if !result.Valid {
        log.Printf("Key validation failed: %v", result.Error)
        return
    }

    log.Printf("Key is valid with permissions: %v", result.Permissions)
}
```

## Using manager

The manager struct allow us to manage keys using the implemented backend, the manager has the next utilities:

```go

type KeyManager struct {
    // ValidateKey checks if an API key is valid and has the required permissions.
    // It verifies the key's existence, expiration status, and permission set.
    //
    // Parameters:
    //   - key: The API key string to validate
    //   - requiredPermissions: Slice of permissions that the key must have, this could be empty
    //    object if you only want to check license have not expired.
    //
    // Returns:
    //   models.ValidationResult containing:
    //   - Valid: true if key is valid and has all required permissions
    //   - Error: error message if validation fails
    //   - Permissions: permissions associated with the key
    ValidateKey(key string, requiredPermissions []models.Permission) models.ValidationResult

    // DisableKey disables an API key, preventing it from being used for authentication
    // while preserving the key and its associated data in storage.
    //
    // Parameters:
    //   - key: The API key string to disable
    //
    // Returns:
    //   - error: nil if successful, error if key not found or operation fails
    DisableKey(key string) error

    // EnableKey enables a previously disabled API key, allowing it to be used for authentication again.
    // If the key is already enabled, this operation has no effect.
    //
    // Parameters:
    //   - key: The API key string to enable
    //
    // Returns:
    //   - error: nil if successful, error if key not found or operation fails
    EnableKey(key string) error

    // DeleteKey permanently removes an API key from storage. Once deleted, the key can no longer
    // be used for authentication and all associated data (permissions, metadata, etc.) is removed.
    // This operation cannot be undone.
    //
    // Parameters:
    //   - key: The API key string to delete
    //
    // Returns:
    //   - error: nil if deletion was successful, error if key not found or deletion fails
    DeleteKey(key string) error

    // RenewKey creates a new API key while invalidating the old one. This operation
    // preserves all the original key's properties (permissions, metadata, expiration)
    // but generates a new key string and hash. The old key is automatically disabled.
    //
    // Parameters:
    //   - key: The current API key string to renew
    //
    // Returns:
    //   - models.APIKey: The newly created API key object
    //   - string: The plain text key string that should be provided to the client
    //   - error: nil if successful, error if key not found or operation fails
    RenewKey(key string) (models.APIKey, string, error)

    // ListKeys returns a paginated list of API keys with optional filtering.
    //
    // Parameters:
    //   - page: Pagination settings including offset and limit
    //   - filter: Filter criteria to narrow down the results (e.g., by name, status, or date range)
    //
    // Returns:
    //   - []models.APIKey: Slice of API keys matching the filter criteria for the requested page
    //   - int64: Total count of API keys matching the filter criteria (across all pages)
    //   - error: nil if successful, error if operation fails
    ListKeys(page Page, filter Filter) ([]models.APIKey, int64, error)

    // GenerateApiKey creates a new API key with the specified parameters.
    //
    // Parameters:
    //   - name: A human-readable identifier for the API key
    //   - permissions: A slice of permissions to be associated with the key
    //   - metadata: Optional key-value pairs for storing additional information
    //   - expiresAt: Optional expiration time for the key. If nil, the key never expires
    //
    // Returns:
    //   - models.APIKey: The created API key object containing all key details
    //   - string: The plain text API key that should be securely transmitted to the client
    //   - error: nil if successful, error if the operation fails
    GenerateApiKey(name string, permissions []models.Permission, metadata map[string]any, expiresAt *time.Time) (models.APIKey, string, error)
}
```

## Storage Backends

Keystogo comes with built-in support for:

- In-memory storage (for testing and development)
- PostgreSQL (coming soon)
- Redis (coming soon)

You can implement your own storage backend by implementing the `Storage` interface:

```go
type Storage interface {
    Get(hashedKey string) (*models.APIKey, error)
    Create(apiKey *models.APIKey) error
    Update(apiKey *models.APIKey) error
    Delete(hashedKey string) error

    List(page Page, filter Filter) ([]models.APIKey, int64, error)

    BatchCreate(apiKeys []*models.APIKey) error
    BatchDelete(hashedKeys []string) error

    Ping() error
    Clear() error
}
```

## Usage Examples

### Basic Validation

```go
result := manager.ValidateKey(apiKey, []keystogo.Permission{"read:users"})
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

func (s *CustomStorage) Get(hashedKey string) (*keystogo.APIKey, error) {
    // Your implementation
}

... Other implementations

// Use your custom storage
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

- 📚 [Documentation](https://pkg.go.dev/github.com/karurosux/keystogo)
- 🐛 [Issue Tracker](https://github.com/karurosux/keystogo/issues)
- 💬 [Discussions](https://github.com/karurosux/keystogo/discussions)

## Roadmap

- [ ] PostgreSQL storage implementation
- [ ] Redis storage implementation
- [ ] Key rotation capabilities
- [ ] Batch key operations
- [ ] Key generation utilities
