# GoldBox RPG Engine - Production Container
# Multi-stage build for minimal production image
# syntax=docker/dockerfile:1

# Stage 1: Build stage
FROM golang:1.26-bookworm AS builder

WORKDIR /build

# Copy all source code and dependencies
COPY go.mod go.sum ./
COPY cmd/ ./cmd/
COPY pkg/ ./pkg/
COPY data/ ./data/

# Build with optimizations for size and security
# Note: Build-time secrets can be mounted using --secret flag:
#   docker build --secret id=vault_token,src=.vault-token .
#   docker build --secret id=aws_credentials,src=~/.aws/credentials .
RUN --mount=type=secret,id=build_secrets,required=false \
    if [ -f /run/secrets/build_secrets ]; then \
      echo "Loading build-time secrets..." && \
      export $(cat /run/secrets/build_secrets | xargs); \
    fi && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build  -ldflags="-w -s -extldflags '-static'" \
    -trimpath \
    -o server ./cmd/server

# Stage 2: Production runtime
FROM gcr.io/distroless/static-debian12:nonroot

# Copy binary from builder
COPY --from=builder /build/server /server

# Copy runtime data (spells, items, etc.)
COPY --from=builder /build/data/ /data/

# Create directory for runtime secrets
# Runtime secrets should be mounted at /run/secrets/ using Docker secrets or Kubernetes secrets
# Example with Docker Swarm:
#   docker service create --secret goldbox_vault_token goldbox-rpg:prod
# Example with Docker run:
#   docker run -v /path/to/secrets:/run/secrets:ro goldbox-rpg:prod
# Example with Kubernetes:
#   Mount secrets as volumes to /run/secrets/

# Use non-root user (distroless provides uid 65532)
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Health check using the binary itself (distroless has no curl)
# Note: For proper health checks, use Kubernetes livenessProbe or Docker HEALTHCHECK with external tools
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD ["/server", "health"] || exit 1

# Run the server
# Server will automatically load secrets from:
#   1. /run/secrets/ directory (Docker/K8s secrets)
#   2. Environment variables (fallback for development)
#   3. Vault/AWS Secrets Manager (if configured)
ENTRYPOINT ["/server"]
