package config

import (
	"context"
	"os"
	"testing"
	"time"

	"goldbox-rpg/pkg/secrets"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadWithSecrets(t *testing.T) {
	tests := []struct {
		name           string
		setupSecrets   func(*secrets.EnvSecretProvider)
		setupEnv       func()
		cleanupEnv     func()
		expectedPort   int
		expectedLogLevel string
		expectError    bool
	}{
		{
			name: "load from secrets provider",
			setupSecrets: func(provider *secrets.EnvSecretProvider) {
				ctx := context.Background()
				_ = provider.SetSecret(ctx, "SERVER_PORT", "9090")
				_ = provider.SetSecret(ctx, "LOG_LEVEL", "debug")
			},
			setupEnv:     func() {},
			cleanupEnv:   func() {},
			expectedPort: 9090,
			expectedLogLevel: "debug",
			expectError:  false,
		},
		{
			name:         "fallback to environment variables",
			setupSecrets: func(provider *secrets.EnvSecretProvider) {},
			setupEnv: func() {
				os.Setenv("SERVER_PORT", "7070")
				os.Setenv("LOG_LEVEL", "warn")
			},
			cleanupEnv: func() {
				os.Unsetenv("SERVER_PORT")
				os.Unsetenv("LOG_LEVEL")
			},
			expectedPort:     7070,
			expectedLogLevel: "warn",
			expectError:      false,
		},
		{
			name: "secrets override environment variables",
			setupSecrets: func(provider *secrets.EnvSecretProvider) {
				ctx := context.Background()
				_ = provider.SetSecret(ctx, "SERVER_PORT", "6060")
			},
			setupEnv: func() {
				os.Setenv("SERVER_PORT", "5050")
			},
			cleanupEnv: func() {
				os.Unsetenv("SERVER_PORT")
			},
			expectedPort:     6060,
			expectedLogLevel: "info", // default
			expectError:      false,
		},
		{
			name:         "use defaults when no secrets or env vars",
			setupSecrets: func(provider *secrets.EnvSecretProvider) {},
			setupEnv:     func() {},
			cleanupEnv:   func() {},
			expectedPort: 8080, // default
			expectedLogLevel: "info", // default
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			provider := secrets.NewEnvSecretProvider("GOLDBOX_")
			tt.setupSecrets(provider)
			tt.setupEnv()
			defer tt.cleanupEnv()

			// Execute
			config, err := LoadWithSecrets(provider)

			// Verify
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, config)
			assert.Equal(t, tt.expectedPort, config.ServerPort)
			assert.Equal(t, tt.expectedLogLevel, config.LogLevel)
		})
	}
}

func TestLoadWithSecretsNilProvider(t *testing.T) {
	config, err := LoadWithSecrets(nil)

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "secret provider cannot be nil")
}

func TestLoadWithSecretsComplexTypes(t *testing.T) {
	provider := secrets.NewEnvSecretProvider("GOLDBOX_")
	ctx := context.Background()

	// Set complex types in secrets
	_ = provider.SetSecret(ctx, "SESSION_TIMEOUT", "45m")
	_ = provider.SetSecret(ctx, "ENABLE_DEV_MODE", "false")
	_ = provider.SetSecret(ctx, "ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080")
	_ = provider.SetSecret(ctx, "MAX_REQUEST_SIZE", "2097152") // 2MB
	_ = provider.SetSecret(ctx, "RATE_LIMIT_REQUESTS_PER_SECOND", "10.5")

	config, err := LoadWithSecrets(provider)

	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 45*time.Minute, config.SessionTimeout)
	assert.False(t, config.EnableDevMode)
	assert.Equal(t, []string{"http://localhost:3000", "http://localhost:8080"}, config.AllowedOrigins)
	assert.Equal(t, int64(2097152), config.MaxRequestSize)
	assert.Equal(t, 10.5, config.RateLimitRequestsPerSecond)
}

func TestLoadWithSecretsValidation(t *testing.T) {
	tests := []struct {
		name          string
		setupSecrets  func(*secrets.EnvSecretProvider)
		expectError   bool
		errorContains string
	}{
		{
			name: "invalid port number",
			setupSecrets: func(provider *secrets.EnvSecretProvider) {
				ctx := context.Background()
				_ = provider.SetSecret(ctx, "SERVER_PORT", "99999")
			},
			expectError:   true,
			errorContains: "server port must be between",
		},
		{
			name: "invalid log level",
			setupSecrets: func(provider *secrets.EnvSecretProvider) {
				ctx := context.Background()
				_ = provider.SetSecret(ctx, "LOG_LEVEL", "invalid")
			},
			expectError:   true,
			errorContains: "log level must be one of",
		},
		{
			name: "session timeout too short",
			setupSecrets: func(provider *secrets.EnvSecretProvider) {
				ctx := context.Background()
				_ = provider.SetSecret(ctx, "SESSION_TIMEOUT", "30s")
			},
			expectError:   true,
			errorContains: "session timeout must be at least 1 minute",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment before test to prevent pollution from previous tests
			os.Unsetenv("SERVER_PORT")
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("SESSION_TIMEOUT")

			// Create fresh provider for each test
			provider := secrets.NewEnvSecretProvider("GOLDBOX_")

			// Setup secrets
			tt.setupSecrets(provider)

			// Execute
			config, err := LoadWithSecrets(provider)

			// Verify
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
			}

			// Clean up after test
			os.Unsetenv("SERVER_PORT")
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("SESSION_TIMEOUT")
		})
	}
}

func TestGetSecretHelperFunctions(t *testing.T) {
	provider := secrets.NewEnvSecretProvider("TEST_")
	ctx := context.Background()

	t.Run("getSecretAsString", func(t *testing.T) {
		_ = provider.SetSecret(ctx, "STRING_KEY", "test_value")
		value := getSecretAsString(ctx, provider, "STRING_KEY", "default")
		assert.Equal(t, "test_value", value)
	})

	t.Run("getSecretAsString fallback", func(t *testing.T) {
		value := getSecretAsString(ctx, provider, "NONEXISTENT_KEY", "default")
		assert.Equal(t, "default", value)
	})

	t.Run("getSecretAsInt", func(t *testing.T) {
		_ = provider.SetSecret(ctx, "INT_KEY", "42")
		value := getSecretAsInt(ctx, provider, "INT_KEY", 0)
		assert.Equal(t, 42, value)
	})

	t.Run("getSecretAsInt fallback", func(t *testing.T) {
		value := getSecretAsInt(ctx, provider, "NONEXISTENT_INT", 99)
		assert.Equal(t, 99, value)
	})

	t.Run("getSecretAsInt64", func(t *testing.T) {
		_ = provider.SetSecret(ctx, "INT64_KEY", "1234567890")
		value := getSecretAsInt64(ctx, provider, "INT64_KEY", 0)
		assert.Equal(t, int64(1234567890), value)
	})

	t.Run("getSecretAsBool", func(t *testing.T) {
		_ = provider.SetSecret(ctx, "BOOL_KEY", "true")
		value := getSecretAsBool(ctx, provider, "BOOL_KEY", false)
		assert.True(t, value)
	})

	t.Run("getSecretAsBool fallback", func(t *testing.T) {
		value := getSecretAsBool(ctx, provider, "NONEXISTENT_BOOL", true)
		assert.True(t, value)
	})

	t.Run("getSecretAsDuration", func(t *testing.T) {
		_ = provider.SetSecret(ctx, "DURATION_KEY", "5m")
		value := getSecretAsDuration(ctx, provider, "DURATION_KEY", time.Second)
		assert.Equal(t, 5*time.Minute, value)
	})

	t.Run("getSecretAsStringSlice", func(t *testing.T) {
		_ = provider.SetSecret(ctx, "SLICE_KEY", "a,b,c")
		value := getSecretAsStringSlice(ctx, provider, "SLICE_KEY", []string{})
		assert.Equal(t, []string{"a", "b", "c"}, value)
	})

	t.Run("getSecretAsStringSlice with spaces", func(t *testing.T) {
		_ = provider.SetSecret(ctx, "SLICE_KEY2", "a, b , c")
		value := getSecretAsStringSlice(ctx, provider, "SLICE_KEY2", []string{})
		assert.Equal(t, []string{"a", "b", "c"}, value)
	})

	t.Run("getSecretAsFloat64", func(t *testing.T) {
		_ = provider.SetSecret(ctx, "FLOAT_KEY", "3.14")
		value := getSecretAsFloat64(ctx, provider, "FLOAT_KEY", 0.0)
		assert.Equal(t, 3.14, value)
	})
}

func TestLoadWithSecretsBackwardCompatibility(t *testing.T) {
	// Test that LoadWithSecrets can fall back to all environment variables
	// ensuring backward compatibility with existing deployments

	// Set environment variables
	os.Setenv("SERVER_PORT", "8888")
	os.Setenv("LOG_LEVEL", "error")
	os.Setenv("SESSION_TIMEOUT", "60m")
	os.Setenv("ENABLE_DEV_MODE", "false")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("SESSION_TIMEOUT")
		os.Unsetenv("ENABLE_DEV_MODE")
	}()

	// Create provider that won't have these secrets
	provider := secrets.NewEnvSecretProvider("GOLDBOX_")

	// Load configuration
	config, err := LoadWithSecrets(provider)

	// Verify it falls back to environment variables
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 8888, config.ServerPort)
	assert.Equal(t, "error", config.LogLevel)
	assert.Equal(t, 60*time.Minute, config.SessionTimeout)
	assert.False(t, config.EnableDevMode)
}
