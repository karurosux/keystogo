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
    "github.com/karurosux/keystogo/pkg/storage/memory"
)

func main() {
    // Initialize with in-memory storage
    storage := memory.NewMemoryStorage()
    manager := keystogo.NewManager(storage)

    // Validate an API key with required permissions
    result := manager.ValidateKey("your-api-key", []keystogo.Permission{"read:users"})

    if !result.Valid {
        log.Printf("Key validation failed: %v", result.Error)
        return
    }

    log.Printf("Key is valid with permissions: %v", result.Permissions)
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
    GetAPIKey(hashedKey string) (*APIKey, error)
    UpdateLastUsed(keyID string, usedAt time.Time) error
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

func (s *CustomStorage) GetAPIKey(hashedKey string) (*keystogo.APIKey, error) {
    // Your implementation
}

func (s *CustomStorage) UpdateLastUsed(keyID string, usedAt time.Time) error {
    // Your implementation
}

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
