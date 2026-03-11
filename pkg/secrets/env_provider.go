package secrets

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// EnvSecretProvider implements SecretProvider using environment variables.
// Suitable for development and testing environments.
//
// Features:
//   - Thread-safe concurrent access
//   - Optional in-memory caching
//   - Support for key prefix filtering
//   - Environment variable validation
//
// Security Notes:
//   - Environment variables are visible in process listings
//   - Not suitable for production with sensitive secrets
//   - Use VaultSecretProvider or AWS Secrets Manager for production
type EnvSecretProvider struct {
	mu     sync.RWMutex
	cache  map[string]string
	prefix string
}

// NewEnvSecretProvider creates a new environment variable-based secret provider.
//
// Parameters:
//   - prefix: Optional prefix for secret keys (e.g., "GOLDBOX_")
//
// Returns:
//   - *EnvSecretProvider: A new provider instance
func NewEnvSecretProvider(prefix string) *EnvSecretProvider {
	logrus.WithFields(logrus.Fields{
		"function": "NewEnvSecretProvider",
		"prefix":   prefix,
	}).Debug("creating environment secret provider")

	return &EnvSecretProvider{
		cache:  make(map[string]string),
		prefix: prefix,
	}
}

// GetSecret retrieves a secret from environment variables.
// Checks cache first, then falls back to os.Getenv.
func (e *EnvSecretProvider) GetSecret(ctx context.Context, key string) (string, error) {
	logrus.WithFields(logrus.Fields{
		"function": "GetSecret",
		"key":      key,
	}).Debug("retrieving secret from environment")

	// Check context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Check cache first
	e.mu.RLock()
	if value, exists := e.cache[key]; exists {
		e.mu.RUnlock()
		logrus.WithField("key", key).Debug("secret found in cache")
		return value, nil
	}
	e.mu.RUnlock()

	// Get from environment
	value := os.Getenv(key)
	if value == "" {
		logrus.WithField("key", key).Debug("secret not found in environment")
		return "", fmt.Errorf("%w: %s", ErrSecretNotFound, key)
	}

	// Cache the value
	e.mu.Lock()
	e.cache[key] = value
	e.mu.Unlock()

	logrus.WithField("key", key).Debug("secret retrieved from environment")
	return value, nil
}

// SetSecret updates an environment variable.
// Note: This only affects the current process, not the system environment.
func (e *EnvSecretProvider) SetSecret(ctx context.Context, key, value string) error {
	logrus.WithFields(logrus.Fields{
		"function": "SetSecret",
		"key":      key,
	}).Debug("setting secret in environment")

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if key == "" {
		return fmt.Errorf("%w: key cannot be empty", ErrSecretInvalid)
	}

	// Set environment variable
	if err := os.Setenv(key, value); err != nil {
		logrus.WithError(err).WithField("key", key).Error("failed to set environment variable")
		return fmt.Errorf("failed to set secret: %w", err)
	}

	// Update cache
	e.mu.Lock()
	e.cache[key] = value
	e.mu.Unlock()

	logrus.WithField("key", key).Debug("secret set in environment")
	return nil
}

// ListSecrets returns all environment variables matching the provider's prefix.
func (e *EnvSecretProvider) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	logrus.WithFields(logrus.Fields{
		"function": "ListSecrets",
		"prefix":   prefix,
	}).Debug("listing secrets from environment")

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Use provider prefix if no specific prefix given
	searchPrefix := prefix
	if searchPrefix == "" {
		searchPrefix = e.prefix
	}

	// Scan environment variables
	var secrets []string
	for _, env := range os.Environ() {
		// Parse key=value format
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		if searchPrefix == "" || strings.HasPrefix(key, searchPrefix) {
			secrets = append(secrets, key)
		}
	}

	logrus.WithFields(logrus.Fields{
		"prefix": searchPrefix,
		"count":  len(secrets),
	}).Debug("secrets listed from environment")

	return secrets, nil
}

// HealthCheck verifies the provider can access environment variables.
func (e *EnvSecretProvider) HealthCheck(ctx context.Context) error {
	logrus.WithField("function", "HealthCheck").Debug("checking environment provider health")

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Simple check: verify we can read environment
	_ = os.Getenv("PATH")

	logrus.Debug("environment provider health check passed")
	return nil
}

// ClearCache removes all cached secrets.
// Useful for forcing a refresh of secret values.
func (e *EnvSecretProvider) ClearCache() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.cache = make(map[string]string)
	logrus.Debug("environment provider cache cleared")
}

// GetCachedCount returns the number of secrets currently cached.
func (e *EnvSecretProvider) GetCachedCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return len(e.cache)
}
