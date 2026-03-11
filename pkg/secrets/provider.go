package secrets

import (
	"context"
	"errors"
)

// Common errors returned by secret providers
var (
	// ErrSecretNotFound is returned when a secret key doesn't exist
	ErrSecretNotFound = errors.New("secret not found")

	// ErrSecretInvalid is returned when a secret value is invalid or malformed
	ErrSecretInvalid = errors.New("secret invalid")

	// ErrProviderUnavailable is returned when the secret provider backend is unavailable
	ErrProviderUnavailable = errors.New("secret provider unavailable")

	// ErrPermissionDenied is returned when the caller lacks permission to access a secret
	ErrPermissionDenied = errors.New("permission denied")

	// ErrNotImplemented is returned for operations not yet supported
	ErrNotImplemented = errors.New("operation not implemented")
)

// SecretProvider defines the interface for accessing secrets from various backends.
// Implementations must be thread-safe for concurrent access.
//
// Context handling:
//   - All methods accept context.Context for cancellation and timeouts
//   - Providers should respect context cancellation
//   - Long-running operations should check ctx.Done() periodically
type SecretProvider interface {
	// GetSecret retrieves a secret value by key.
	// Returns ErrSecretNotFound if the key doesn't exist.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - key: Secret key name (e.g., "GOLDBOX_DB_PASSWORD")
	//
	// Returns:
	//   - string: The secret value
	//   - error: ErrSecretNotFound, ErrProviderUnavailable, or other errors
	GetSecret(ctx context.Context, key string) (string, error)

	// SetSecret stores or updates a secret value.
	// Not all providers support this operation (returns ErrNotImplemented).
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - key: Secret key name
	//   - value: Secret value to store
	//
	// Returns:
	//   - error: ErrNotImplemented, ErrPermissionDenied, or other errors
	SetSecret(ctx context.Context, key, value string) error

	// ListSecrets returns all secret keys matching a prefix.
	// Useful for discovery and validation.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - prefix: Key prefix to filter (e.g., "GOLDBOX_")
	//
	// Returns:
	//   - []string: List of matching secret keys
	//   - error: ErrProviderUnavailable or other errors
	ListSecrets(ctx context.Context, prefix string) ([]string, error)

	// HealthCheck verifies the provider is accessible and operational.
	// Should perform a lightweight check without excessive overhead.
	//
	// Returns:
	//   - error: nil if healthy, error describing the problem otherwise
	HealthCheck(ctx context.Context) error
}

// SecretCache provides an optional caching layer for secret providers
// to reduce backend calls and improve performance.
type SecretCache interface {
	// Get retrieves a cached secret value
	Get(key string) (string, bool)

	// Set stores a secret value in the cache
	Set(key, value string)

	// Invalidate removes a secret from the cache
	Invalidate(key string)

	// Clear removes all secrets from the cache
	Clear()
}
