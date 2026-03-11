# Kubernetes Deployment for GoldBox RPG Engine

This directory contains Kubernetes manifests for deploying the GoldBox RPG Engine to a Kubernetes cluster.

## Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured to access your cluster
- Container registry with the goldbox-rpg image
- (Optional) Helm 3.x for Helm chart deployment
- (Optional) cert-manager for automatic TLS certificate management
- (Optional) Prometheus Operator for metrics collection

## Quick Start

### Deploy with kubectl

```bash
# Create namespace and deploy all resources
kubectl apply -f namespace.yaml
kubectl apply -f serviceaccount.yaml
kubectl apply -f configmap.yaml

# Update secret.yaml with your actual values first!
kubectl apply -f secret.yaml

kubectl apply -f persistentvolumeclaim.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f hpa.yaml
kubectl apply -f pdb.yaml
kubectl apply -f ingress.yaml
kubectl apply -f networkpolicy.yaml

# Optional: ServiceMonitor (requires Prometheus Operator)
kubectl apply -f servicemonitor.yaml
```

### Deploy with Kustomize

```bash
# Deploy all resources at once
kubectl apply -k .

# Or customize for your environment
kubectl kustomize . | kubectl apply -f -
```

## Configuration

### Required Secrets

Before deploying, update `secret.yaml` with your values:

```yaml
stringData:
  WEBSOCKET_ALLOWED_ORIGINS: "https://yourdomain.com,https://www.yourdomain.com"
```

For production deployments, consider using external secret management:
- Kubernetes External Secrets Operator
- HashiCorp Vault
- AWS Secrets Manager
- Azure Key Vault
- Google Secret Manager

### ConfigMap Customization

Edit `configmap.yaml` to adjust:
- Log level (`GOLDBOX_LOG_LEVEL`)
- Session timeout (`GOLDBOX_SESSION_TIMEOUT`)
- Rate limiting (`GOLDBOX_RATE_LIMIT_RPS`, `GOLDBOX_RATE_LIMIT_BURST`)
- Resource limits in `deployment.yaml`

### Ingress Configuration

Update `ingress.yaml` with your domain:

```yaml
spec:
  tls:
  - hosts:
    - goldbox-rpg.yourdomain.com  # Replace with your domain
  rules:
  - host: goldbox-rpg.yourdomain.com  # Replace with your domain
```

## Resource Requirements

### Default Resource Requests/Limits

- **Requests**: 100m CPU, 128Mi memory
- **Limits**: 500m CPU, 512Mi memory

Adjust based on your workload in `deployment.yaml`:

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

## High Availability

### Replicas

Default: 3 replicas for high availability

```yaml
spec:
  replicas: 3  # Adjust as needed
```

### Horizontal Pod Autoscaling

HPA automatically scales between 3-10 pods based on:
- CPU utilization (target: 70%)
- Memory utilization (target: 80%)

Edit `hpa.yaml` to customize scaling behavior.

### Pod Disruption Budget

PDB ensures at least 2 pods are available during voluntary disruptions (updates, node drains).

```yaml
spec:
  minAvailable: 2
```

## Monitoring

### Health Checks

The deployment includes:
- **Liveness probe**: `/live` endpoint
- **Readiness probe**: `/ready` endpoint
- **Startup probe**: `/health` endpoint

### Prometheus Metrics

Metrics are exposed at `/metrics` and automatically scraped by Prometheus (if ServiceMonitor is deployed).

Key metrics:
- HTTP request duration
- Active sessions
- WebSocket connections
- Game state operations
- System resources

## Networking

### Services

- **ClusterIP**: Internal cluster access
- **LoadBalancer**: External access with session affinity

### Ingress

- HTTPS with automatic TLS (cert-manager)
- WebSocket support
- Rate limiting
- Security headers

### Network Policy

Restricts traffic to:
- Ingress controller → Pods
- Within namespace
- DNS resolution
- External HTTPS (if needed)

## Storage

### Persistent Volume

A 10Gi PersistentVolumeClaim is created for game data persistence.

```yaml
resources:
  requests:
    storage: 10Gi
```

Session data, character files, and game state are stored in `/data` within the container.

## Security

### Pod Security

- Runs as non-root user (uid 65532)
- Read-only root filesystem
- Drops all capabilities
- Seccomp profile enabled
- No privilege escalation

### Network Security

- NetworkPolicy restricts ingress/egress
- TLS encryption for external traffic
- WebSocket origin validation

### Secrets Management

- Sensitive data in Kubernetes Secrets
- Consider using external secret providers
- Secrets mounted as environment variables

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n goldbox-rpg
kubectl describe pod <pod-name> -n goldbox-rpg
kubectl logs <pod-name> -n goldbox-rpg
```

### Check Service

```bash
kubectl get svc -n goldbox-rpg
kubectl describe svc goldbox-rpg -n goldbox-rpg
```

### Check Ingress

```bash
kubectl get ingress -n goldbox-rpg
kubectl describe ingress goldbox-rpg -n goldbox-rpg
```

### Test Health Endpoints

```bash
# From within the cluster
kubectl run -it --rm debug --image=busybox --restart=Never -n goldbox-rpg -- wget -qO- http://goldbox-rpg/health

# From outside (if LoadBalancer is configured)
curl http://<EXTERNAL-IP>/health
curl http://<EXTERNAL-IP>/ready
curl http://<EXTERNAL-IP>/live
```

### Check HPA Status

```bash
kubectl get hpa -n goldbox-rpg
kubectl describe hpa goldbox-rpg -n goldbox-rpg
```

### Common Issues

1. **Pods not starting**: Check image pull permissions and container registry access
2. **Health checks failing**: Verify health endpoint configuration and startup time
3. **Ingress not working**: Ensure ingress controller is installed and domain DNS is configured
4. **Metrics not appearing**: Verify Prometheus ServiceMonitor and scraping configuration

## Cleanup

```bash
# Delete all resources
kubectl delete -k .

# Or delete individually
kubectl delete namespace goldbox-rpg
```

## Production Checklist

- [ ] Update `secret.yaml` with production values
- [ ] Configure domain in `ingress.yaml`
- [ ] Set up TLS certificates (cert-manager or manual)
- [ ] Configure container registry authentication
- [ ] Adjust resource limits based on load testing
- [ ] Set up monitoring and alerting
- [ ] Configure log aggregation
- [ ] Test disaster recovery procedures
- [ ] Document runbooks for common operations
- [ ] Set up backup strategy for persistent data

## Environment-Specific Deployments

### Development

```bash
# Use smaller resource limits
kubectl apply -k . --dry-run=client -o yaml | \
  sed 's/replicas: 3/replicas: 1/' | \
  sed 's/cpu: 500m/cpu: 200m/' | \
  kubectl apply -f -
```

### Staging

```bash
# Use staging namespace
kubectl apply -k . -n goldbox-rpg-staging
```

### Production

```bash
# Use production optimizations
kubectl apply -k .
# Monitor rollout
kubectl rollout status deployment/goldbox-rpg -n goldbox-rpg
```

## Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Kustomize Documentation](https://kustomize.io/)
- [Ingress NGINX Documentation](https://kubernetes.github.io/ingress-nginx/)
- [Prometheus Operator Documentation](https://prometheus-operator.dev/)
- [cert-manager Documentation](https://cert-manager.io/docs/)
