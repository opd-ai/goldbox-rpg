// Package secrets provides a unified interface for managing sensitive configuration
// values across different environments and secret backends.
//
// The package supports multiple secret providers:
//   - EnvSecretProvider: Environment variable-based secrets (recommended, fully implemented)
//   - VaultSecretProvider: HashiCorp Vault integration (stub, requires additional setup)
//
// Current Implementation Status:
//   - EnvSecretProvider: Fully functional for all environments
//   - VaultSecretProvider: Returns ErrNotImplemented; requires vault/api dependency
//
// Secret Naming Conventions:
//
// All secrets should follow the GOLDBOX_ prefix convention:
//   - GOLDBOX_DB_PASSWORD
//   - GOLDBOX_API_KEY
//   - GOLDBOX_JWT_SECRET
//   - GOLDBOX_ENCRYPTION_KEY
//
// Usage Example:
//
//	provider := secrets.NewEnvSecretProvider()
//	dbPassword, err := provider.GetSecret(ctx, "GOLDBOX_DB_PASSWORD")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Security Best Practices:
//
//  1. Never log secret values
//  2. Use context.Context for cancellation and timeouts
//  3. Implement secret rotation where supported
//  4. Use different providers for different environments
//  5. Validate secret formats before use
//  6. Clear sensitive data from memory when done
//
// Production Deployment:
//
// For production deployments requiring Vault, extend VaultSecretProvider with:
//   - github.com/hashicorp/vault/api client library
//   - TLS-enabled Vault connection
//   - AppRole or Kubernetes auth
//   - Dynamic secret generation where possible
//   - Audit logging enabled
package secrets
