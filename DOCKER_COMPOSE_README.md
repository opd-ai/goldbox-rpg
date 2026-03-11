# Docker Compose Local Development Environment

This directory contains Docker Compose configuration for running the GoldBox RPG Engine locally with full observability stack.

## Quick Start

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f goldbox-server

# Stop all services
docker-compose down

# Stop and remove volumes (clean state)
docker-compose down -v
```

## Services

### Game Server (goldbox-server)
- **URL**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **Metrics**: http://localhost:8080/metrics
- **Data**: Persisted in `./data` directory

### Prometheus (metrics collection)
- **URL**: http://localhost:9090
- **Configuration**: `monitoring/prometheus.yml`
- **Data**: Persisted in Docker volume `prometheus-data`

### Grafana (metrics visualization)
- **URL**: http://localhost:3000
- **Username**: admin
- **Password**: goldbox
- **Dashboards**: Auto-provisioned from `monitoring/grafana/dashboards/`

## Configuration

### Game Server Environment Variables

Edit the `environment` section in `docker-compose.yml`:

```yaml
environment:
  - GOLDBOX_PORT=8080
  - GOLDBOX_LOG_LEVEL=debug          # debug, info, warn, error
  - GOLDBOX_SESSION_TIMEOUT=30m       # Session expiration
  - ENABLE_SESSION_PERSISTENCE=true   # Persist sessions across restarts
  - SESSION_PERSISTENCE_INTERVAL=60s  # Auto-save interval
  - DATA_DIR=/data                    # Data directory path
```

### Prometheus Configuration

Edit `monitoring/prometheus.yml` to customize:
- Scrape intervals
- Target endpoints
- Alerting rules (optional)

### Grafana Dashboards

Add custom dashboards to `monitoring/grafana/dashboards/` as JSON files. They will be auto-loaded on startup.

## Health Checks

All services include health checks:

```bash
# Check game server
curl http://localhost:8080/health

# Check Prometheus
curl http://localhost:9090/-/healthy

# Check Grafana
curl http://localhost:3000/api/health
```

## Data Persistence

The following data is persisted:

1. **Game Data** (`./data` directory):
   - Game state (`gamestate.yaml`)
   - Character files (`characters/*.yaml`)
   - Session snapshots (`sessions/*.yaml`)

2. **Prometheus Data** (Docker volume):
   - Time-series metrics
   - Automatically managed by Docker

3. **Grafana Data** (Docker volume):
   - Dashboard customizations
   - User settings
   - Automatically managed by Docker

## Monitoring

### Grafana Dashboard

Access Grafana at http://localhost:3000 (admin/goldbox) to view:

- **RPC Request Rate**: Requests per second by method
- **RPC P95 Latency**: 95th percentile response times
- **Active Sessions**: Current player sessions
- **Active Goroutines**: Concurrent Go routines
- **Memory Usage**: Heap allocation
- **Error Rate**: Errors by method and type

### Prometheus Queries

Example queries you can run in Prometheus (http://localhost:9090):

```promql
# Request rate by method
rate(goldbox_rpc_requests_total[5m])

# P95 latency
histogram_quantile(0.95, rate(goldbox_rpc_duration_seconds_bucket[5m]))

# Active sessions
goldbox_active_sessions

# Error rate
rate(goldbox_rpc_errors_total[5m])

# Memory usage
go_memstats_alloc_bytes
```

## Development Workflow

### Making Code Changes

```bash
# Rebuild and restart server after code changes
docker-compose up -d --build goldbox-server

# View logs
docker-compose logs -f goldbox-server
```

### Testing Metrics

```bash
# Trigger some game actions
curl -X POST http://localhost:8080/rpc -d '{"jsonrpc":"2.0","method":"session.create","params":{"player_id":"test"},"id":1}'

# View metrics
curl http://localhost:8080/metrics

# Check Grafana dashboard
open http://localhost:3000
```

### Debugging

```bash
# Access server logs
docker-compose logs goldbox-server

# Follow logs in real-time
docker-compose logs -f goldbox-server

# Execute commands inside container
docker-compose exec goldbox-server sh

# View configuration
docker-compose config
```

## Troubleshooting

### Server Won't Start

```bash
# Check logs
docker-compose logs goldbox-server

# Verify port availability
lsof -i :8080

# Rebuild from scratch
docker-compose down -v
docker-compose build --no-cache
docker-compose up -d
```

### Prometheus Can't Scrape Metrics

```bash
# Verify server is exposing metrics
curl http://localhost:8080/metrics

# Check Prometheus targets
open http://localhost:9090/targets

# Verify network connectivity
docker-compose exec prometheus wget -O- http://goldbox-server:8080/metrics
```

### Grafana Dashboard Not Loading

```bash
# Check datasource configuration
docker-compose exec grafana cat /etc/grafana/provisioning/datasources/prometheus.yml

# Verify Grafana can reach Prometheus
docker-compose exec grafana wget -O- http://prometheus:9090/api/v1/status/config

# Restart Grafana
docker-compose restart grafana
```

## Clean Up

```bash
# Stop all services
docker-compose down

# Remove all data (WARNING: This deletes game state!)
docker-compose down -v
rm -rf ./data/*

# Remove all images
docker-compose down --rmi all
```

## Production Deployment

This Docker Compose setup is for **local development only**. For production:

1. Use Kubernetes manifests in `deploy/k8s/`
2. Use Helm chart in `deploy/helm/`
3. Configure proper secrets management
4. Set up external persistence (not local volumes)
5. Configure proper resource limits
6. Enable TLS/HTTPS
7. Use production-grade monitoring (external Prometheus/Grafana)

See `docs/DEPLOYMENT.md` for production deployment guide.
