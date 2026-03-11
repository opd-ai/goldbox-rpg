# GoldBox RPG Engine - Production Container
# Multi-stage build for minimal production image

# Stage 1: Build stage
FROM golang:1.23-bookworm AS builder

WORKDIR /build

# Copy all source code and dependencies
COPY go.mod go.sum ./
COPY vendor/ ./vendor/
COPY cmd/ ./cmd/
COPY pkg/ ./pkg/
COPY data/ ./data/

# Build with optimizations for size and security
# Use -mod=vendor to use vendored dependencies
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -mod=vendor -ldflags="-w -s -extldflags '-static'" \
    -trimpath \
    -o server ./cmd/server

# Stage 2: Production runtime
FROM gcr.io/distroless/static-debian12:nonroot

# Copy binary from builder
COPY --from=builder /build/server /server

# Copy runtime data (spells, items, etc.)
COPY --from=builder /build/data/ /data/

# Use non-root user (distroless provides uid 65532)
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Health check using the binary itself (distroless has no curl)
# Note: For proper health checks, use Kubernetes livenessProbe or Docker HEALTHCHECK with external tools
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD ["/server", "health"] || exit 1

# Run the server
ENTRYPOINT ["/server"]
