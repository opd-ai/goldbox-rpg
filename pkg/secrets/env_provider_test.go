package secrets

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnvSecretProvider(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{
			name:   "with prefix",
			prefix: "GOLDBOX_",
		},
		{
			name:   "without prefix",
			prefix: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewEnvSecretProvider(tt.prefix)
			assert.NotNil(t, provider)
			assert.Equal(t, tt.prefix, provider.prefix)
			assert.NotNil(t, provider.cache)
		})
	}
}

func TestEnvSecretProvider_GetSecret(t *testing.T) {
	ctx := context.Background()
	provider := NewEnvSecretProvider("TEST_")

	// Set a test environment variable
	testKey := "TEST_SECRET_KEY"
	testValue := "test-secret-value"
	t.Setenv(testKey, testValue)

	tests := []struct {
		name      string
		key       string
		want      string
		wantErr   error
		setupFunc func()
	}{
		{
			name:    "existing secret",
			key:     testKey,
			want:    testValue,
			wantErr: nil,
		},
		{
			name:    "non-existent secret",
			key:     "TEST_NONEXISTENT",
			want:    "",
			wantErr: ErrSecretNotFound,
		},
		{
			name: "cached secret",
			key:  testKey,
			want: testValue,
			setupFunc: func() {
				// First call to populate cache
				_, _ = provider.GetSecret(ctx, testKey)
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			got, err := provider.GetSecret(ctx, tt.key)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestEnvSecretProvider_GetSecret_ContextCancellation(t *testing.T) {
	provider := NewEnvSecretProvider("")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := provider.GetSecret(ctx, "ANY_KEY")
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestEnvSecretProvider_GetSecret_ContextTimeout(t *testing.T) {
	provider := NewEnvSecretProvider("")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout occurs

	_, err := provider.GetSecret(ctx, "ANY_KEY")
	assert.Error(t, err)
}

func TestEnvSecretProvider_SetSecret(t *testing.T) {
	ctx := context.Background()
	provider := NewEnvSecretProvider("")

	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{
			name:    "valid secret",
			key:     "TEST_SET_KEY",
			value:   "test-value",
			wantErr: false,
		},
		{
			name:    "empty key",
			key:     "",
			value:   "test-value",
			wantErr: true,
		},
		{
			name:    "empty value",
			key:     "TEST_EMPTY_VALUE",
			value:   "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.SetSecret(ctx, tt.key, tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the secret was set
				if tt.key != "" {
					got, err := provider.GetSecret(ctx, tt.key)
					assert.NoError(t, err)
					assert.Equal(t, tt.value, got)
				}
			}
		})
	}
}

func TestEnvSecretProvider_SetSecret_ContextCancellation(t *testing.T) {
	provider := NewEnvSecretProvider("")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := provider.SetSecret(ctx, "ANY_KEY", "value")
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestEnvSecretProvider_ListSecrets(t *testing.T) {
	ctx := context.Background()

	// Set up test environment variables
	t.Setenv("GOLDBOX_SECRET_1", "value1")
	t.Setenv("GOLDBOX_SECRET_2", "value2")
	t.Setenv("OTHER_SECRET", "value3")

	tests := []struct {
		name           string
		providerPrefix string
		listPrefix     string
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:           "list with provider prefix",
			providerPrefix: "GOLDBOX_",
			listPrefix:     "",
			wantContains:   []string{"GOLDBOX_SECRET_1", "GOLDBOX_SECRET_2"},
			wantNotContain: []string{"OTHER_SECRET"},
		},
		{
			name:           "list with specific prefix",
			providerPrefix: "",
			listPrefix:     "GOLDBOX_",
			wantContains:   []string{"GOLDBOX_SECRET_1", "GOLDBOX_SECRET_2"},
			wantNotContain: []string{"OTHER_SECRET"},
		},
		{
			name:           "list all secrets",
			providerPrefix: "",
			listPrefix:     "",
			wantContains:   []string{"GOLDBOX_SECRET_1", "GOLDBOX_SECRET_2", "OTHER_SECRET"},
			wantNotContain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewEnvSecretProvider(tt.providerPrefix)
			secrets, err := provider.ListSecrets(ctx, tt.listPrefix)

			assert.NoError(t, err)
			assert.NotNil(t, secrets)

			for _, want := range tt.wantContains {
				assert.Contains(t, secrets, want)
			}

			for _, notWant := range tt.wantNotContain {
				assert.NotContains(t, secrets, notWant)
			}
		})
	}
}

func TestEnvSecretProvider_ListSecrets_ContextCancellation(t *testing.T) {
	provider := NewEnvSecretProvider("")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := provider.ListSecrets(ctx, "")
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestEnvSecretProvider_HealthCheck(t *testing.T) {
	ctx := context.Background()
	provider := NewEnvSecretProvider("")

	err := provider.HealthCheck(ctx)
	assert.NoError(t, err)
}

func TestEnvSecretProvider_HealthCheck_ContextCancellation(t *testing.T) {
	provider := NewEnvSecretProvider("")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := provider.HealthCheck(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestEnvSecretProvider_ClearCache(t *testing.T) {
	ctx := context.Background()
	provider := NewEnvSecretProvider("")

	// Add a secret to cache
	testKey := "TEST_CACHE_KEY"
	testValue := "test-value"
	t.Setenv(testKey, testValue)

	// Get secret to populate cache
	_, err := provider.GetSecret(ctx, testKey)
	require.NoError(t, err)
	assert.Equal(t, 1, provider.GetCachedCount())

	// Clear cache
	provider.ClearCache()
	assert.Equal(t, 0, provider.GetCachedCount())

	// Verify secret can still be retrieved from environment
	got, err := provider.GetSecret(ctx, testKey)
	assert.NoError(t, err)
	assert.Equal(t, testValue, got)
	assert.Equal(t, 1, provider.GetCachedCount())
}

func TestEnvSecretProvider_GetCachedCount(t *testing.T) {
	ctx := context.Background()
	provider := NewEnvSecretProvider("")

	// Initially empty
	assert.Equal(t, 0, provider.GetCachedCount())

	// Add secrets
	t.Setenv("TEST_KEY_1", "value1")
	t.Setenv("TEST_KEY_2", "value2")

	_, _ = provider.GetSecret(ctx, "TEST_KEY_1")
	assert.Equal(t, 1, provider.GetCachedCount())

	_, _ = provider.GetSecret(ctx, "TEST_KEY_2")
	assert.Equal(t, 2, provider.GetCachedCount())

	// Missing key shouldn't increase cache count
	_, _ = provider.GetSecret(ctx, "NONEXISTENT")
	assert.Equal(t, 2, provider.GetCachedCount())
}

func TestEnvSecretProvider_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	provider := NewEnvSecretProvider("")

	t.Setenv("CONCURRENT_KEY", "concurrent-value")

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = provider.GetSecret(ctx, "CONCURRENT_KEY")
			_ = provider.SetSecret(ctx, "CONCURRENT_SET", "value")
			_, _ = provider.ListSecrets(ctx, "")
			provider.ClearCache()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify provider is still functional
	err := provider.HealthCheck(ctx)
	assert.NoError(t, err)
}

func TestSecretProviderInterface(t *testing.T) {
	// Verify EnvSecretProvider implements SecretProvider
	var _ SecretProvider = (*EnvSecretProvider)(nil)
}

func TestCommonErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrSecretNotFound", ErrSecretNotFound},
		{"ErrSecretInvalid", ErrSecretInvalid},
		{"ErrProviderUnavailable", ErrProviderUnavailable},
		{"ErrPermissionDenied", ErrPermissionDenied},
		{"ErrNotImplemented", ErrNotImplemented},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.err)
			assert.NotEmpty(t, tt.err.Error())

			// Verify error can be compared with errors.Is
			assert.True(t, errors.Is(tt.err, tt.err))
		})
	}
}
