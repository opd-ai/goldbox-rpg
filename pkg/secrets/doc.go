// Package secrets provides a unified interface for managing sensitive configuration
// values across different environments and secret backends.
//
// Supported Providers:
//   - EnvSecretProvider: Environment variable-based secrets (recommended for all environments)
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
//	provider := secrets.NewEnvSecretProvider("GOLDBOX_")
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
//  4. Validate secret formats before use
//  5. Clear sensitive data from memory when done
//
// Production Deployment:
//
// For production deployments, use environment variables injected by your
// orchestration platform (Kubernetes Secrets, Docker Secrets, AWS Secrets Manager,
// or HashiCorp Vault via environment injection). The EnvSecretProvider reads
// these injected values transparently.
//
// Extending with Additional Providers:
//
// To add support for other secret backends (e.g., Vault, AWS Secrets Manager):
//  1. Implement the SecretProvider interface
//  2. Add the necessary client library as a dependency
//  3. Handle authentication and connection management
//  4. Add comprehensive tests for error cases
package secrets
