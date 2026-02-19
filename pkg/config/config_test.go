package config

import (
	"os"
	"testing"
	"time"

	"goldbox-rpg/pkg/retry"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		validate    func(t *testing.T, config *Config)
	}{
		{
			name:        "default configuration",
			envVars:     map[string]string{},
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, 8080, config.ServerPort)
				assert.Equal(t, "./web", config.WebDir)
				assert.Equal(t, 30*time.Minute, config.SessionTimeout)
				assert.Equal(t, "info", config.LogLevel)
				assert.Equal(t, []string{}, config.AllowedOrigins)
				assert.Equal(t, int64(1*1024*1024), config.MaxRequestSize)
				assert.Equal(t, true, config.EnableDevMode)
				assert.Equal(t, 30*time.Second, config.RequestTimeout)
			},
		},
		{
			name: "custom configuration from environment",
			envVars: map[string]string{
				"SERVER_PORT":      "9090",
				"WEB_DIR":          "/custom/web",
				"SESSION_TIMEOUT":  "45m",
				"LOG_LEVEL":        "debug",
				"ALLOWED_ORIGINS":  "http://localhost:3000,https://example.com",
				"MAX_REQUEST_SIZE": "2097152", // 2MB
				"ENABLE_DEV_MODE":  "true",
				"REQUEST_TIMEOUT":  "45s",
			},
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, 9090, config.ServerPort)
				assert.Equal(t, "/custom/web", config.WebDir)
				assert.Equal(t, 45*time.Minute, config.SessionTimeout)
				assert.Equal(t, "debug", config.LogLevel)
				assert.Equal(t, []string{"http://localhost:3000", "https://example.com"}, config.AllowedOrigins)
				assert.Equal(t, int64(2*1024*1024), config.MaxRequestSize)
				assert.Equal(t, true, config.EnableDevMode)
				assert.Equal(t, 45*time.Second, config.RequestTimeout)
			},
		},
		{
			name: "invalid port",
			envVars: map[string]string{
				"SERVER_PORT": "99999",
			},
			expectError: true,
		},
		{
			name: "invalid log level",
			envVars: map[string]string{
				"LOG_LEVEL": "invalid",
			},
			expectError: true,
		},
		{
			name: "session timeout too short",
			envVars: map[string]string{
				"SESSION_TIMEOUT": "30s",
			},
			expectError: true,
		},
		{
			name: "request timeout too short",
			envVars: map[string]string{
				"REQUEST_TIMEOUT": "500ms",
			},
			expectError: true,
		},
		{
			name: "max request size too small",
			envVars: map[string]string{
				"MAX_REQUEST_SIZE": "512",
			},
			expectError: true,
		},
		{
			name: "production mode without allowed origins",
			envVars: map[string]string{
				"ENABLE_DEV_MODE": "false",
			},
			expectError: true,
		},
		{
			name: "production mode with allowed origins",
			envVars: map[string]string{
				"ENABLE_DEV_MODE": "false",
				"ALLOWED_ORIGINS": "https://production.example.com",
			},
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, false, config.EnableDevMode)
				assert.Equal(t, []string{"https://production.example.com"}, config.AllowedOrigins)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			clearTestEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			config, err := Load()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				if tt.validate != nil {
					tt.validate(t, config)
				}
			}
		})
	}
}

func TestConfig_OriginAllowed(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		origin         string
		expectedResult bool
	}{
		{
			name: "dev mode allows all origins",
			config: &Config{
				EnableDevMode:  true,
				AllowedOrigins: []string{"https://example.com"},
			},
			origin:         "https://unknown.com",
			expectedResult: true,
		},
		{
			name: "production mode allows listed origin",
			config: &Config{
				EnableDevMode:  false,
				AllowedOrigins: []string{"https://example.com", "https://app.example.com"},
			},
			origin:         "https://example.com",
			expectedResult: true,
		},
		{
			name: "production mode blocks unlisted origin",
			config: &Config{
				EnableDevMode:  false,
				AllowedOrigins: []string{"https://example.com"},
			},
			origin:         "https://malicious.com",
			expectedResult: false,
		},
		{
			name: "production mode blocks empty origin",
			config: &Config{
				EnableDevMode:  false,
				AllowedOrigins: []string{"https://example.com"},
			},
			origin:         "",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.OriginAllowed(tt.origin)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestGetEnvHelpers(t *testing.T) {
	// Clean environment
	clearTestEnv()

	t.Run("getEnvAsString", func(t *testing.T) {
		// Test default value
		assert.Equal(t, "default", getEnvAsString("TEST_STRING", "default"))

		// Test environment value
		os.Setenv("TEST_STRING", "custom")
		defer os.Unsetenv("TEST_STRING")
		assert.Equal(t, "custom", getEnvAsString("TEST_STRING", "default"))
	})

	t.Run("getEnvAsInt", func(t *testing.T) {
		// Test default value
		assert.Equal(t, 42, getEnvAsInt("TEST_INT", 42))

		// Test valid environment value
		os.Setenv("TEST_INT", "100")
		defer os.Unsetenv("TEST_INT")
		assert.Equal(t, 100, getEnvAsInt("TEST_INT", 42))

		// Test invalid environment value falls back to default
		os.Setenv("TEST_INT_INVALID", "not-a-number")
		defer os.Unsetenv("TEST_INT_INVALID")
		assert.Equal(t, 42, getEnvAsInt("TEST_INT_INVALID", 42))
	})

	t.Run("getEnvAsInt64", func(t *testing.T) {
		// Test default value
		assert.Equal(t, int64(42), getEnvAsInt64("TEST_INT64", 42))

		// Test valid environment value
		os.Setenv("TEST_INT64", "9223372036854775807")
		defer os.Unsetenv("TEST_INT64")
		assert.Equal(t, int64(9223372036854775807), getEnvAsInt64("TEST_INT64", 42))
	})

	t.Run("getEnvAsBool", func(t *testing.T) {
		// Test default value
		assert.Equal(t, true, getEnvAsBool("TEST_BOOL", true))

		// Test valid environment values
		testCases := []struct {
			value    string
			expected bool
		}{
			{"true", true},
			{"false", false},
			{"1", true},
			{"0", false},
			{"TRUE", true},
			{"FALSE", false},
		}

		for _, tc := range testCases {
			os.Setenv("TEST_BOOL", tc.value)
			assert.Equal(t, tc.expected, getEnvAsBool("TEST_BOOL", false), "value: %s", tc.value)
		}
		os.Unsetenv("TEST_BOOL")
	})

	t.Run("getEnvAsDuration", func(t *testing.T) {
		// Test default value
		assert.Equal(t, 5*time.Minute, getEnvAsDuration("TEST_DURATION", 5*time.Minute))

		// Test valid environment value
		os.Setenv("TEST_DURATION", "2h30m")
		defer os.Unsetenv("TEST_DURATION")
		assert.Equal(t, 2*time.Hour+30*time.Minute, getEnvAsDuration("TEST_DURATION", 5*time.Minute))
	})

	t.Run("getEnvAsStringSlice", func(t *testing.T) {
		// Test default value
		defaultSlice := []string{"a", "b"}
		assert.Equal(t, defaultSlice, getEnvAsStringSlice("TEST_SLICE", defaultSlice))

		// Test valid environment value
		os.Setenv("TEST_SLICE", "one,two,three")
		defer os.Unsetenv("TEST_SLICE")
		assert.Equal(t, []string{"one", "two", "three"}, getEnvAsStringSlice("TEST_SLICE", defaultSlice))

		// Test environment value with whitespace
		os.Setenv("TEST_SLICE_WHITESPACE", " one , two , three ")
		defer os.Unsetenv("TEST_SLICE_WHITESPACE")
		assert.Equal(t, []string{"one", "two", "three"}, getEnvAsStringSlice("TEST_SLICE_WHITESPACE", defaultSlice))

		// Test environment value with empty parts
		os.Setenv("TEST_SLICE_EMPTY", "one,,three,")
		defer os.Unsetenv("TEST_SLICE_EMPTY")
		assert.Equal(t, []string{"one", "three"}, getEnvAsStringSlice("TEST_SLICE_EMPTY", defaultSlice))
	})
}

// clearTestEnv removes all environment variables that might affect tests
func clearTestEnv() {
	testVars := []string{
		"SERVER_PORT", "WEB_DIR", "SESSION_TIMEOUT", "LOG_LEVEL",
		"ALLOWED_ORIGINS", "MAX_REQUEST_SIZE", "ENABLE_DEV_MODE", "REQUEST_TIMEOUT",
		"TEST_STRING", "TEST_INT", "TEST_INT_INVALID", "TEST_INT64", "TEST_BOOL",
		"TEST_DURATION", "TEST_SLICE", "TEST_SLICE_WHITESPACE", "TEST_SLICE_EMPTY",
	}

	for _, v := range testVars {
		os.Unsetenv(v)
	}
}

func TestConfig_GetRetryConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		expectedConfig retry.RetryConfig
	}{
		{
			name: "default retry configuration",
			config: &Config{
				RetryMaxAttempts:       3,
				RetryInitialDelay:      100 * time.Millisecond,
				RetryMaxDelay:          30 * time.Second,
				RetryBackoffMultiplier: 2.0,
				RetryJitterPercent:     10,
			},
			expectedConfig: retry.RetryConfig{
				MaxAttempts:       3,
				InitialDelay:      100 * time.Millisecond,
				MaxDelay:          30 * time.Second,
				BackoffMultiplier: 2.0,
				JitterMaxPercent:  10,
				RetryableErrors:   []error{},
			},
		},
		{
			name: "custom retry configuration",
			config: &Config{
				RetryMaxAttempts:       5,
				RetryInitialDelay:      200 * time.Millisecond,
				RetryMaxDelay:          60 * time.Second,
				RetryBackoffMultiplier: 3.0,
				RetryJitterPercent:     20,
			},
			expectedConfig: retry.RetryConfig{
				MaxAttempts:       5,
				InitialDelay:      200 * time.Millisecond,
				MaxDelay:          60 * time.Second,
				BackoffMultiplier: 3.0,
				JitterMaxPercent:  20,
				RetryableErrors:   []error{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetRetryConfig()

			// Verify the return type is retry.RetryConfig
			var _ retry.RetryConfig = result

			// Verify field values
			assert.Equal(t, tt.expectedConfig.MaxAttempts, result.MaxAttempts)
			assert.Equal(t, tt.expectedConfig.InitialDelay, result.InitialDelay)
			assert.Equal(t, tt.expectedConfig.MaxDelay, result.MaxDelay)
			assert.Equal(t, tt.expectedConfig.BackoffMultiplier, result.BackoffMultiplier)
			assert.Equal(t, tt.expectedConfig.JitterMaxPercent, result.JitterMaxPercent)
			assert.Equal(t, tt.expectedConfig.RetryableErrors, result.RetryableErrors)
		})
	}
}

func TestConfig_GetRetryConfig_UsableWithRetrier(t *testing.T) {
	// Test that the returned config can be used directly with retry.NewRetrier
	config := &Config{
		RetryMaxAttempts:       3,
		RetryInitialDelay:      100 * time.Millisecond,
		RetryMaxDelay:          30 * time.Second,
		RetryBackoffMultiplier: 2.0,
		RetryJitterPercent:     10,
	}

	retryConfig := config.GetRetryConfig()

	// Create a retrier with the config - this verifies type compatibility
	retrier := retry.NewRetrier(retryConfig)
	require.NotNil(t, retrier, "retrier should be created successfully with config")
}

func TestLoad_ServerLifecycleTimeouts(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		validate    func(t *testing.T, config *Config)
	}{
		{
			name:        "default lifecycle timeouts",
			envVars:     map[string]string{},
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, 60*time.Second, config.BootstrapTimeout, "BootstrapTimeout default")
				assert.Equal(t, 30*time.Second, config.ShutdownTimeout, "ShutdownTimeout default")
				assert.Equal(t, 1*time.Second, config.ShutdownGracePeriod, "ShutdownGracePeriod default")
			},
		},
		{
			name: "custom lifecycle timeouts from environment",
			envVars: map[string]string{
				"BOOTSTRAP_TIMEOUT":     "120s",
				"SHUTDOWN_TIMEOUT":      "45s",
				"SHUTDOWN_GRACE_PERIOD": "2s",
			},
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, 120*time.Second, config.BootstrapTimeout, "BootstrapTimeout custom")
				assert.Equal(t, 45*time.Second, config.ShutdownTimeout, "ShutdownTimeout custom")
				assert.Equal(t, 2*time.Second, config.ShutdownGracePeriod, "ShutdownGracePeriod custom")
			},
		},
		{
			name: "lifecycle timeouts with duration format variants",
			envVars: map[string]string{
				"BOOTSTRAP_TIMEOUT":     "2m",
				"SHUTDOWN_TIMEOUT":      "1m30s",
				"SHUTDOWN_GRACE_PERIOD": "500ms",
			},
			expectError: false,
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, 2*time.Minute, config.BootstrapTimeout)
				assert.Equal(t, 90*time.Second, config.ShutdownTimeout)
				assert.Equal(t, 500*time.Millisecond, config.ShutdownGracePeriod)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			clearTestEnv()
			clearLifecycleTimeoutEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			config, err := Load()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				if tt.validate != nil {
					tt.validate(t, config)
				}
			}
		})
	}
}

// clearLifecycleTimeoutEnv removes lifecycle timeout environment variables
func clearLifecycleTimeoutEnv() {
	timeoutVars := []string{
		"BOOTSTRAP_TIMEOUT", "SHUTDOWN_TIMEOUT", "SHUTDOWN_GRACE_PERIOD",
	}
	for _, v := range timeoutVars {
		os.Unsetenv(v)
	}
}

// TestConfig_OriginAllowed_ThreadSafety tests that OriginAllowed is safe for concurrent access
func TestConfig_OriginAllowed_ThreadSafety(t *testing.T) {
	config := &Config{
		EnableDevMode:  false,
		AllowedOrigins: []string{"https://example.com", "https://app.example.com"},
	}

	// Run concurrent reads to verify thread safety
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = config.OriginAllowed("https://example.com")
				_ = config.OriginAllowed("https://unknown.com")
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
