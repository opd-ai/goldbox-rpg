# Secrets Management Operations Guide

## Overview

The GoldBox RPG Engine uses a unified secrets management system that supports multiple backends for different deployment environments. This guide covers operational procedures for managing secrets securely across development, staging, and production environments.

## Architecture

### Secret Providers

The secrets system implements a provider interface that abstracts backend-specific details:

**Currently Available:**
- **EnvSecretProvider**: Environment variable-based secrets for development and testing

**Planned (Not Yet Implemented):**
- **VaultSecretProvider**: HashiCorp Vault integration for production
- **AWS Secrets Manager**: AWS-based secrets

All providers implement the `SecretProvider` interface in `pkg/secrets/provider.go`:

```go
type SecretProvider interface {
    GetSecret(ctx context.Context, key string) (string, error)
    SetSecret(ctx context.Context, key, value string) error
    ListSecrets(ctx context.Context, prefix string) ([]string, error)
    HealthCheck(ctx context.Context) error
}
```

## Secret Naming Conventions

All secrets must follow the `GOLDBOX_` prefix convention for consistency and security:

### Standard Secret Names

| Secret Name | Purpose | Required |
|-------------|---------|----------|
| `GOLDBOX_DB_PASSWORD` | Database connection password | Production |
| `GOLDBOX_API_KEY` | External API authentication | Production |
| `GOLDBOX_JWT_SECRET` | JWT token signing key | All environments |
| `GOLDBOX_ENCRYPTION_KEY` | Data encryption key | All environments |
| `GOLDBOX_WEBSOCKET_SECRET` | WebSocket authentication | All environments |
| `GOLDBOX_SESSION_KEY` | Session encryption key | All environments |

### Naming Rules

1. Always use the `GOLDBOX_` prefix
2. Use uppercase with underscores (SCREAMING_SNAKE_CASE)
3. Be descriptive but concise
4. Group related secrets with common prefixes (e.g., `GOLDBOX_DB_*`)

## Secret Rotation Procedures

Secret rotation is a critical security practice that limits the window of exposure if a secret is compromised. This section describes procedures for rotating different types of secrets.

### Rotation Strategy

**Recommended Rotation Schedule:**

| Secret Type | Rotation Frequency | Risk Level |
|-------------|-------------------|------------|
| Database credentials | 90 days | High |
| API keys | 90 days | High |
| JWT signing keys | 30 days | Critical |
| Encryption keys | 365 days* | Critical |
| Session keys | 30 days | Medium |

*Note: Encryption key rotation requires re-encrypting data. Plan accordingly.

### Pre-Rotation Checklist

Before rotating any secret:

1. ✅ **Identify Dependencies**: List all services using the secret
2. ✅ **Plan Deployment Window**: Choose low-traffic period if possible
3. ✅ **Backup Current Values**: Store securely for rollback
4. ✅ **Test Rotation Process**: Verify procedure in staging first
5. ✅ **Notify Stakeholders**: Alert team of upcoming rotation
6. ✅ **Monitor Readiness**: Check all systems are healthy

### Rotation Procedures by Secret Type

> **Note:** The Vault commands shown below are for planned future integration. Currently, only environment variable-based secrets (EnvSecretProvider) are supported.

#### 1. Database Password Rotation

**Procedure:**

```bash
# Step 1: Generate new password
NEW_PASSWORD=$(openssl rand -base64 32)

# Step 2: Update database user password
psql -h db-host -U admin -c "ALTER USER goldbox_user PASSWORD '$NEW_PASSWORD';"

# Step 3: Update secret in provider
# For environment provider (development):
export GOLDBOX_DB_PASSWORD="$NEW_PASSWORD"

# For Vault provider (production):
vault kv put secret/goldbox/db password="$NEW_PASSWORD"

# Step 4: Restart application (graceful)
kubectl rollout restart deployment/goldbox-rpg

# Step 5: Verify connectivity
curl http://localhost:8080/health | jq '.database'
```

**Rollback Procedure:**

```bash
# Revert to previous password
psql -h db-host -U admin -c "ALTER USER goldbox_user PASSWORD '$OLD_PASSWORD';"
export GOLDBOX_DB_PASSWORD="$OLD_PASSWORD"
kubectl rollout restart deployment/goldbox-rpg
```

#### 2. JWT Secret Rotation

**Procedure:**

JWT secret rotation requires a dual-key approach to avoid invalidating active sessions:

```bash
# Step 1: Generate new secret
NEW_JWT_SECRET=$(openssl rand -hex 64)

# Step 2: Add new secret with versioning
export GOLDBOX_JWT_SECRET_NEW="$NEW_JWT_SECRET"
export GOLDBOX_JWT_SECRET_OLD="$CURRENT_JWT_SECRET"

# Step 3: Deploy application with dual-key validation
# Application should verify tokens with both keys
kubectl apply -f deploy/k8s/deployment.yaml

# Step 4: Wait for token expiration period (typically 24 hours)
sleep 86400

# Step 5: Promote new secret to primary
export GOLDBOX_JWT_SECRET="$NEW_JWT_SECRET"
unset GOLDBOX_JWT_SECRET_OLD

# Step 6: Deploy again to remove old key
kubectl apply -f deploy/k8s/deployment.yaml
```

**Critical Notes:**
- Never rotate JWT secret without dual-key support
- Maintain old key until all issued tokens expire
- Monitor failed authentication attempts during rotation

#### 3. API Key Rotation

**Procedure:**

```bash
# Step 1: Request new API key from provider
NEW_API_KEY=$(curl -X POST https://api-provider.com/keys \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.key')

# Step 2: Update secret
export GOLDBOX_API_KEY="$NEW_API_KEY"

# Step 3: Restart service
kubectl rollout restart deployment/goldbox-rpg

# Step 4: Verify API connectivity
curl http://localhost:8080/health | jq '.external_api'

# Step 5: Revoke old API key (after 24 hours grace period)
curl -X DELETE https://api-provider.com/keys/$OLD_API_KEY \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

#### 4. Encryption Key Rotation

**WARNING**: Encryption key rotation requires re-encrypting all encrypted data.

**Procedure:**

```bash
# Step 1: Generate new encryption key
NEW_ENCRYPTION_KEY=$(openssl rand -hex 32)

# Step 2: Add new key alongside old key
export GOLDBOX_ENCRYPTION_KEY_NEW="$NEW_ENCRYPTION_KEY"
export GOLDBOX_ENCRYPTION_KEY_OLD="$CURRENT_ENCRYPTION_KEY"

# Step 3: Deploy application with dual-key decryption
kubectl apply -f deploy/k8s/deployment.yaml

# Step 4: Run data re-encryption job
kubectl apply -f deploy/k8s/jobs/reencrypt-data.yaml

# Step 5: Monitor re-encryption progress
kubectl logs -f job/reencrypt-data

# Step 6: Verify all data re-encrypted
./scripts/verify-encryption.sh

# Step 7: Promote new key to primary
export GOLDBOX_ENCRYPTION_KEY="$NEW_ENCRYPTION_KEY"
unset GOLDBOX_ENCRYPTION_KEY_OLD

# Step 8: Deploy final configuration
kubectl apply -f deploy/k8s/deployment.yaml
```

**Estimated Time**: 1-4 hours depending on data volume

### Automated Rotation (Future)

When Vault integration is complete, enable automated rotation:

```hcl
# Vault dynamic database credentials example
path "database/creds/goldbox-role" {
  capabilities = ["read"]
}

# Lease duration: 90 days
# Automatic renewal: enabled
# Max lease time: 365 days
```

## Secret Health Checks

### Health Check Implementation

All secret providers implement health checks via the `HealthCheck` method:

```go
// Example: Check environment provider health
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

provider := secrets.NewEnvSecretProvider("GOLDBOX_")
if err := provider.HealthCheck(ctx); err != nil {
    log.Errorf("Secret provider unhealthy: %v", err)
}
```

### Health Check Endpoints

The application exposes health endpoints that include secret provider status:

```bash
# Comprehensive health check (includes secret provider)
curl http://localhost:8080/health

# Example response:
{
  "status": "healthy",
  "timestamp": "2026-03-11T09:00:00Z",
  "checks": {
    "secrets": {
      "status": "healthy",
      "provider": "env",
      "message": "Environment provider operational"
    },
    "database": { "status": "healthy" },
    "websocket": { "status": "healthy" }
  }
}
```

### Monitoring Secret Provider Health

**Automated Health Checks:**

```bash
# Add to monitoring system (Prometheus example)
# Alert if secret provider fails health check for 5 minutes

ALERT SecretProviderUnhealthy
  IF goldbox_secrets_health_status != 1
  FOR 5m
  LABELS { severity = "critical" }
  ANNOTATIONS {
    summary = "Secret provider health check failing"
  }
```

**Manual Health Verification:**

```bash
# Check all secrets are accessible
./scripts/verify-secrets.sh

# Example script:
#!/bin/bash
REQUIRED_SECRETS=(
  "GOLDBOX_DB_PASSWORD"
  "GOLDBOX_JWT_SECRET"
  "GOLDBOX_API_KEY"
  "GOLDBOX_ENCRYPTION_KEY"
)

for secret in "${REQUIRED_SECRETS[@]}"; do
  if [ -z "${!secret}" ]; then
    echo "ERROR: Missing secret: $secret"
    exit 1
  fi
done
echo "All required secrets present"
```

## Environment-Specific Configuration

### Development Environment

**Setup:**

```bash
# .env file for local development
cat > .env <<EOF
GOLDBOX_DB_PASSWORD=dev_password_123
GOLDBOX_JWT_SECRET=$(openssl rand -hex 64)
GOLDBOX_API_KEY=dev_api_key
GOLDBOX_ENCRYPTION_KEY=$(openssl rand -hex 32)
EOF

# Load environment variables
source .env

# Start server
make run
```

**Security Notes:**
- Development secrets should never be used in production
- Commit `.env.example` but never `.env`
- Use weak secrets for dev (no need for complex rotation)

### Staging Environment

**Setup:**

```bash
# Use Kubernetes secrets for staging
kubectl create secret generic goldbox-secrets \
  --from-literal=db-password="staging_password" \
  --from-literal=jwt-secret="$(openssl rand -hex 64)" \
  --from-literal=api-key="staging_api_key" \
  --namespace=staging

# Deployment will mount as environment variables
kubectl apply -f deploy/k8s/staging/deployment.yaml
```

**Rotation Schedule:**
- Rotate quarterly (90 days)
- Test rotation procedures before production

### Production Environment

**Setup (Vault - Future):**

```bash
# Enable Vault secrets engine
vault secrets enable -path=goldbox kv-v2

# Create secrets
vault kv put goldbox/db \
  password="$(openssl rand -base64 32)"

vault kv put goldbox/jwt \
  secret="$(openssl rand -hex 64)"

# Create access policy
vault policy write goldbox-read - <<EOF
path "goldbox/*" {
  capabilities = ["read", "list"]
}
EOF

# Enable Kubernetes auth
vault auth enable kubernetes
vault write auth/kubernetes/role/goldbox \
  bound_service_account_names=goldbox-sa \
  bound_service_account_namespaces=production \
  policies=goldbox-read \
  ttl=24h
```

**Rotation Schedule:**
- Critical secrets: Monthly
- Standard secrets: Quarterly
- Audit all rotations

### Docker and Container Deployment

The Dockerfile supports multiple methods for injecting secrets into containers securely, following Docker and Kubernetes best practices.

#### Build-Time Secrets (Docker BuildKit)

For secrets needed during the build process (e.g., private repository access):

```bash
# Create a secrets file (DO NOT commit to version control)
cat > .build-secrets <<EOF
GITHUB_TOKEN=ghp_xxxxxxxxxxxxx
NPM_TOKEN=npm_xxxxxxxxxxxxx
EOF

# Build with BuildKit secret mount (secrets not stored in image layers)
DOCKER_BUILDKIT=1 docker build \
  --secret id=build_secrets,src=.build-secrets \
  -t goldbox-rpg:prod .

# Secrets are only available during build, not in final image
```

**Security Benefits:**
- Secrets never appear in image layers or `docker history`
- Automatic cleanup after build
- No risk of accidental secret exposure in pushed images

#### Runtime Secrets (Docker Run)

For development and testing with Docker run:

```bash
# Option 1: Environment variables (least secure, use only for development)
docker run -d \
  -e GOLDBOX_JWT_SECRET="dev_secret_123" \
  -e GOLDBOX_DB_PASSWORD="dev_password" \
  -p 8080:8080 \
  goldbox-rpg:prod

# Option 2: Environment file (better isolation)
cat > .env.prod <<EOF
GOLDBOX_JWT_SECRET=prod_secret_xyz
GOLDBOX_DB_PASSWORD=prod_password_abc
GOLDBOX_ENCRYPTION_KEY=enc_key_456
EOF

docker run -d \
  --env-file .env.prod \
  -p 8080:8080 \
  goldbox-rpg:prod

# Option 3: Volume-mounted secrets (recommended for Docker)
mkdir -p ./secrets
echo "prod_jwt_secret_xyz" > ./secrets/jwt_secret
echo "prod_db_password_abc" > ./secrets/db_password
chmod 600 ./secrets/*

docker run -d \
  -v $(pwd)/secrets:/run/secrets:ro \
  -p 8080:8080 \
  goldbox-rpg:prod
```

**Security Notes:**
- Option 1 exposes secrets in `docker inspect` output
- Option 2 requires protecting `.env.prod` file
- Option 3 is most secure: secrets only accessible inside container

#### Runtime Secrets (Docker Swarm)

For production deployments with Docker Swarm:

```bash
# Create secrets in Swarm
echo "prod_jwt_secret_xyz" | docker secret create goldbox_jwt_secret -
echo "prod_db_password_abc" | docker secret create goldbox_db_password -

# Deploy service with secrets
docker service create \
  --name goldbox-rpg \
  --secret goldbox_jwt_secret \
  --secret goldbox_db_password \
  --env GOLDBOX_JWT_SECRET_FILE=/run/secrets/goldbox_jwt_secret \
  --env GOLDBOX_DB_PASSWORD_FILE=/run/secrets/goldbox_db_password \
  --publish 8080:8080 \
  goldbox-rpg:prod

# Update secrets (rotation)
echo "new_jwt_secret_abc" | docker secret create goldbox_jwt_secret_v2 -
docker service update \
  --secret-rm goldbox_jwt_secret \
  --secret-add source=goldbox_jwt_secret_v2,target=goldbox_jwt_secret \
  goldbox-rpg
```

**Security Features:**
- Secrets encrypted at rest in Swarm
- Encrypted in transit to containers
- Only accessible to authorized services
- Automatic rotation support

#### Runtime Secrets (Kubernetes)

For production deployments with Kubernetes:

```bash
# Create secret from literal values
kubectl create secret generic goldbox-secrets \
  --from-literal=jwt-secret="$(openssl rand -hex 64)" \
  --from-literal=db-password="$(openssl rand -base64 32)" \
  --from-literal=encryption-key="$(openssl rand -hex 32)" \
  --namespace=production

# Or create from files
kubectl create secret generic goldbox-secrets \
  --from-file=jwt-secret=./secrets/jwt.txt \
  --from-file=db-password=./secrets/db.txt \
  --namespace=production

# Reference in deployment manifest
cat > deployment.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: goldbox-rpg
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: goldbox-rpg
  template:
    metadata:
      labels:
        app: goldbox-rpg
    spec:
      containers:
      - name: server
        image: goldbox-rpg:prod
        ports:
        - containerPort: 8080
        # Mount secrets as environment variables
        env:
        - name: GOLDBOX_JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: goldbox-secrets
              key: jwt-secret
        - name: GOLDBOX_DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: goldbox-secrets
              key: db-password
        # Or mount as files (more secure)
        volumeMounts:
        - name: secrets
          mountPath: /run/secrets
          readOnly: true
      volumes:
      - name: secrets
        secret:
          secretName: goldbox-secrets
          defaultMode: 0400
EOF

kubectl apply -f deployment.yaml
```

**Secret Rotation in Kubernetes:**

```bash
# Method 1: Update existing secret (requires pod restart)
kubectl create secret generic goldbox-secrets \
  --from-literal=jwt-secret="$(openssl rand -hex 64)" \
  --dry-run=client -o yaml | \
  kubectl apply -f -

# Restart pods to pick up new secret
kubectl rollout restart deployment/goldbox-rpg -n production

# Method 2: Create new secret version (zero-downtime)
kubectl create secret generic goldbox-secrets-v2 \
  --from-literal=jwt-secret="$(openssl rand -hex 64)" \
  --namespace=production

# Update deployment to use new secret
kubectl patch deployment goldbox-rpg -n production \
  -p '{"spec":{"template":{"spec":{"volumes":[{"name":"secrets","secret":{"secretName":"goldbox-secrets-v2"}}]}}}}'

# Delete old secret after successful rollout
kubectl delete secret goldbox-secrets -n production
```

#### Secret Loading Priority

The application loads secrets in the following priority order (first found wins):

1. **File-based secrets** (`/run/secrets/GOLDBOX_*`)
   - Docker Swarm: `/run/secrets/goldbox_jwt_secret`
   - Kubernetes: `/run/secrets/jwt-secret`

2. **Environment variable files** (`GOLDBOX_*_FILE`)
   - Points to file containing secret value
   - Example: `GOLDBOX_JWT_SECRET_FILE=/run/secrets/jwt.txt`

3. **Direct environment variables** (`GOLDBOX_*`)
   - Standard environment variable
   - Example: `GOLDBOX_JWT_SECRET=secret_value`

4. **Vault/AWS Secrets Manager** (if configured)
   - Requires `GOLDBOX_SECRETS_PROVIDER=vault` or `aws`
   - Higher security, centralized management

**Implementation Example:**

```go
// Application automatically checks multiple sources
func loadSecret(key string) (string, error) {
    // 1. Check file-based secret
    if val, err := os.ReadFile("/run/secrets/" + key); err == nil {
        return string(val), nil
    }
    
    // 2. Check environment variable file
    if file := os.Getenv(key + "_FILE"); file != "" {
        if val, err := os.ReadFile(file); err == nil {
            return string(val), nil
        }
    }
    
    // 3. Check direct environment variable
    if val := os.Getenv(key); val != "" {
        return val, nil
    }
    
    // 4. Check secrets provider (Vault/AWS)
    return secretProvider.GetSecret(context.Background(), key)
}
```

## Security Best Practices

### 1. Never Log Secret Values

```go
// ❌ WRONG - logs secret value
log.Infof("DB password: %s", password)

// ✅ CORRECT - logs only presence
log.Info("DB password retrieved successfully")
```

### 2. Use Context Timeouts

```go
// Always use context with timeout for secret operations
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

secret, err := provider.GetSecret(ctx, "GOLDBOX_DB_PASSWORD")
```

### 3. Clear Sensitive Data

```go
// Clear secrets from memory when done
password := []byte(secret)
defer func() {
    for i := range password {
        password[i] = 0
    }
}()
```

### 4. Validate Secret Formats

```go
// Validate secrets before use
if len(jwtSecret) < 32 {
    return fmt.Errorf("JWT secret too short: minimum 32 bytes required")
}
```

### 5. Implement Secret Versioning

```go
// Support multiple secret versions for rotation
type SecretWithVersion struct {
    Value   string
    Version int
    Created time.Time
}
```

## Troubleshooting

### Secret Not Found

**Symptom:** `ErrSecretNotFound: GOLDBOX_DB_PASSWORD`

**Solutions:**
1. Verify secret name matches convention
2. Check environment variables are loaded
3. Confirm secret provider is initialized
4. Review provider-specific configuration

```bash
# Debug environment variables
env | grep GOLDBOX_

# Test secret provider directly
go run cmd/validator-demo/main.go
```

### Secret Provider Unavailable

**Symptom:** `ErrProviderUnavailable: connection timeout`

**Solutions:**
1. Check network connectivity to Vault/AWS
2. Verify credentials are valid
3. Review firewall rules
4. Check provider health status

```bash
# Test Vault connectivity
vault status

# Check Kubernetes secrets mounting
kubectl describe pod goldbox-rpg-xxx | grep -A5 Mounts
```

### Health Check Failing

**Symptom:** Health endpoint reports `"secrets": { "status": "unhealthy" }`

**Solutions:**
1. Check provider implementation
2. Verify backend connectivity
3. Review error logs
4. Test manual secret retrieval

```bash
# Check application logs
kubectl logs goldbox-rpg-xxx | grep -i secret

# Manual health check
curl http://localhost:8080/health | jq '.checks.secrets'
```

## Audit and Compliance

### Access Logging

All secret access should be logged (without values):

```go
logrus.WithFields(logrus.Fields{
    "function": "GetSecret",
    "key":      key,
    "user":     ctx.Value("user_id"),
}).Info("Secret accessed")
```

### Rotation Audit Trail

Maintain a rotation log:

```bash
# /var/log/goldbox/secret-rotation.log
2026-03-11T09:00:00Z ROTATION_START secret=GOLDBOX_JWT_SECRET user=admin
2026-03-11T09:05:00Z ROTATION_COMPLETE secret=GOLDBOX_JWT_SECRET status=success
```

### Compliance Requirements

For SOC2, GDPR, and similar compliance:

- ✅ Secrets never in version control
- ✅ Rotation procedures documented and tested
- ✅ Access logging enabled
- ✅ Encryption at rest and in transit
- ✅ Regular rotation schedule enforced
- ✅ Audit trail maintained

## References

### Related Documentation

- [pkg/secrets/doc.go](../pkg/secrets/doc.go) - Package documentation
- [pkg/secrets/provider.go](../pkg/secrets/provider.go) - Provider interface
- [ERROR_HANDLING.md](./ERROR_HANDLING.md) - Error handling patterns

### External Resources

- [HashiCorp Vault Documentation](https://www.vaultproject.io/docs)
- [AWS Secrets Manager Guide](https://docs.aws.amazon.com/secretsmanager/)
- [OWASP Secrets Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [NIST SP 800-57 Key Management](https://csrc.nist.gov/publications/detail/sp/800-57-part-1/rev-5/final)

## Future Enhancements

- [ ] Implement VaultSecretProvider with full Vault API integration
- [ ] Add AWS Secrets Manager provider
- [ ] Automate rotation scheduling with cron jobs
- [ ] Implement secret versioning and rollback
- [ ] Add Prometheus metrics for rotation tracking
- [ ] Create secret rotation dashboard
- [ ] Implement emergency secret revocation procedures
- [ ] Add multi-region secret replication

---

**Last Updated:** 2026-03-11  
**Maintained by:** DevOps Team  
**Review Schedule:** Quarterly
