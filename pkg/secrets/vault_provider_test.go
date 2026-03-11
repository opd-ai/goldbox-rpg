package secrets

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVaultSecretProvider(t *testing.T) {
	tests := []struct {
		name      string
		address   string
		token     string
		namespace string
		wantErr   bool
	}{
		{
			name:      "valid configuration",
			address:   "https://vault.example.com:8200",
			token:     "test-token",
			namespace: "test-namespace",
			wantErr:   false,
		},
		{
			name:      "without namespace",
			address:   "https://vault.example.com:8200",
			token:     "test-token",
			namespace: "",
			wantErr:   false,
		},
		{
			name:      "missing address",
			address:   "",
			token:     "test-token",
			namespace: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewVaultSecretProvider(tt.address, tt.token, tt.namespace)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrSecretInvalid)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, provider)
				assert.Equal(t, tt.address, provider.address)
				assert.Equal(t, tt.token, provider.token)
				assert.Equal(t, tt.namespace, provider.namespace)
			}
		})
	}
}

func TestVaultSecretProvider_GetSecret_NotImplemented(t *testing.T) {
	provider, err := NewVaultSecretProvider("https://vault.example.com:8200", "token", "")
	require.NoError(t, err)

	ctx := context.Background()
	_, err = provider.GetSecret(ctx, "test-key")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotImplemented)
}

func TestVaultSecretProvider_SetSecret_NotImplemented(t *testing.T) {
	provider, err := NewVaultSecretProvider("https://vault.example.com:8200", "token", "")
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.SetSecret(ctx, "test-key", "test-value")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotImplemented)
}

func TestVaultSecretProvider_ListSecrets_NotImplemented(t *testing.T) {
	provider, err := NewVaultSecretProvider("https://vault.example.com:8200", "token", "")
	require.NoError(t, err)

	ctx := context.Background()
	_, err = provider.ListSecrets(ctx, "prefix")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotImplemented)
}

func TestVaultSecretProvider_HealthCheck_NotImplemented(t *testing.T) {
	provider, err := NewVaultSecretProvider("https://vault.example.com:8200", "token", "")
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.HealthCheck(ctx)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotImplemented)
}
