package models

import "errors"

var (
	ErrKeyNotFound       = errors.New("api key not found")
	ErrKeyExpired        = errors.New("api key has expired")
	ErrKeyInactive       = errors.New("api key is inactive")
	ErrPermissionDenied  = errors.New("api key lacks required permission")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrKeyAlreadyExists  = errors.New("api key already exists")
)
