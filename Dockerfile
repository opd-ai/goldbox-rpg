# GoldBox RPG Engine - Turnkey Container
FROM golang:1.22-bookworm

# Install curl for health checks
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# Set up workspace
WORKDIR /app

# Copy everything needed
COPY . .

# Build Go backend
RUN CGO_ENABLED=0 go build -o server ./cmd/server

# Create user and set permissions
RUN useradd -m gameuser && chown -R gameuser:gameuser /app
USER gameuser

# Expose port and health check
EXPOSE 8080
HEALTHCHECK CMD curl -f http://localhost:8080/health || exit 1

# Run the server
CMD ["./server"]
