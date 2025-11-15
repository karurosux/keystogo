package models

import "errors"

type ErrFactory func() error

var (
	ErrKeyNotFound ErrFactory = func() error {
		return errors.New("api key not found")
	}
	ErrKeyExpired ErrFactory = func() error {
		return errors.New("api key has expired")
	}
	ErrKeyInactive ErrFactory = func() error {
		return errors.New("api key is inactive")
	}
	ErrPermissionDenied ErrFactory = func() error {
		return errors.New("api key lacks required permission")
	}
	ErrKeyAlreadyExists ErrFactory = func() error {
		return errors.New("api key already exists")
	}
	ErrEmptyKey ErrFactory = func() error {
		return errors.New("key cannot be empty")
	}
	ErrEmptyID ErrFactory = func() error {
		return errors.New("id cannot be empty")
	}
	ErrEmptyName ErrFactory = func() error {
		return errors.New("name cannot be empty")
	}
	ErrInvalidID ErrFactory = func() error {
		return errors.New("invalid id format")
	}
)
