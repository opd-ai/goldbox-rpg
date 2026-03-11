# GoldBox RPG Helm Chart

A Helm chart for deploying the GoldBox RPG Engine - a modern Go-based RPG engine inspired by the classic SSI Gold Box series.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- PersistentVolume provisioner support in the underlying infrastructure
- (Optional) cert-manager for TLS certificate management
- (Optional) Prometheus Operator for ServiceMonitor support

## Installing the Chart

To install the chart with the release name `my-goldbox-rpg`:

```bash
# From the deploy/helm directory
helm install my-goldbox-rpg ./goldbox-rpg

# Or with custom values
helm install my-goldbox-rpg ./goldbox-rpg -f my-values.yaml

# Install in a specific namespace
helm install my-goldbox-rpg ./goldbox-rpg --namespace goldbox-rpg --create-namespace
```

## Uninstalling the Chart

To uninstall/delete the `my-goldbox-rpg` deployment:

```bash
helm uninstall my-goldbox-rpg

# If installed in a specific namespace
helm uninstall my-goldbox-rpg --namespace goldbox-rpg
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Upgrading the Chart

To upgrade the `my-goldbox-rpg` deployment:

```bash
# Upgrade with new values
helm upgrade my-goldbox-rpg ./goldbox-rpg -f my-values.yaml

# Upgrade with specific version
helm upgrade my-goldbox-rpg ./goldbox-rpg --version 1.0.0
```

## Configuration

The following table lists the configurable parameters of the GoldBox RPG chart and their default values.

### Global Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of GoldBox RPG replicas to deploy | `3` |
| `nameOverride` | String to partially override goldbox-rpg.fullname | `""` |
| `fullnameOverride` | String to fully override goldbox-rpg.fullname | `""` |

### Image Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | GoldBox RPG image repository | `goldbox-rpg` |
| `image.pullPolicy` | GoldBox RPG image pull policy | `Always` |
| `image.tag` | GoldBox RPG image tag (overrides Chart.AppVersion) | `"latest"` |
| `imagePullSecrets` | Docker registry secret names as an array | `[]` |

### Service Account Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `serviceAccount.create` | Specifies whether a service account should be created | `true` |
| `serviceAccount.automount` | Automatically mount a ServiceAccount's API credentials | `true` |
| `serviceAccount.annotations` | Annotations to add to the service account | `{}` |
| `serviceAccount.name` | The name of the service account to use | `""` |

### Security Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `podSecurityContext.runAsNonRoot` | Run containers as non-root user | `true` |
| `podSecurityContext.runAsUser` | User ID to run containers as | `65532` |
| `podSecurityContext.fsGroup` | Group ID for filesystem access | `65532` |
| `securityContext.allowPrivilegeEscalation` | Allow privilege escalation | `false` |
| `securityContext.readOnlyRootFilesystem` | Use read-only root filesystem | `true` |

### Service Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `service.type` | Kubernetes Service type | `ClusterIP` |
| `service.port` | Service HTTP port | `80` |
| `service.targetPort` | Target container port | `8080` |
| `service.sessionAffinity` | Session affinity type | `ClientIP` |
| `service.sessionAffinityTimeoutSeconds` | Session affinity timeout | `10800` |
| `loadBalancer.enabled` | Enable LoadBalancer service | `false` |
| `loadBalancer.annotations` | LoadBalancer service annotations | `{}` |

### Ingress Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress controller resource | `true` |
| `ingress.className` | Ingress class name | `"nginx"` |
| `ingress.annotations` | Ingress annotations | See values.yaml |
| `ingress.hosts` | Ingress hosts configuration | See values.yaml |
| `ingress.tls` | Ingress TLS configuration | See values.yaml |

### Resource Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.limits.cpu` | CPU resource limits | `500m` |
| `resources.limits.memory` | Memory resource limits | `512Mi` |
| `resources.requests.cpu` | CPU resource requests | `100m` |
| `resources.requests.memory` | Memory resource requests | `128Mi` |

### Autoscaling Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `autoscaling.enabled` | Enable Horizontal Pod Autoscaler | `true` |
| `autoscaling.minReplicas` | Minimum number of replicas | `3` |
| `autoscaling.maxReplicas` | Maximum number of replicas | `10` |
| `autoscaling.targetCPUUtilizationPercentage` | Target CPU utilization percentage | `70` |
| `autoscaling.targetMemoryUtilizationPercentage` | Target memory utilization percentage | `80` |

### Persistence Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `persistence.enabled` | Enable persistence using PVC | `true` |
| `persistence.storageClass` | PVC Storage Class | `""` |
| `persistence.accessMode` | PVC Access Mode | `ReadWriteOnce` |
| `persistence.size` | PVC Storage Request | `10Gi` |
| `persistence.mountPath` | Data mount path | `/data` |

### Configuration Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.port` | Server port | `"8080"` |
| `config.logLevel` | Logging level | `"info"` |
| `config.sessionTimeout` | Session timeout duration | `"30m"` |
| `config.enableSessionPersistence` | Enable session persistence | `"true"` |
| `config.rateLimitEnabled` | Enable rate limiting | `"true"` |
| `config.rateLimitRps` | Rate limit requests per second | `"10.0"` |

See [values.yaml](./values.yaml) for the complete list of configuration parameters.

### Secrets Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `secrets.websocketAllowedOrigins` | Allowed WebSocket origins (comma-separated) | `"https://yourdomain.com,https://www.yourdomain.com"` |

## Example Configurations

### Development Environment

```yaml
# dev-values.yaml
replicaCount: 1

config:
  logLevel: "debug"
  devMode: "true"

autoscaling:
  enabled: false

persistence:
  size: 1Gi

ingress:
  enabled: false
```

Install with:
```bash
helm install goldbox-rpg-dev ./goldbox-rpg -f dev-values.yaml
```

### Staging Environment

```yaml
# staging-values.yaml
replicaCount: 2

config:
  logLevel: "info"
  devMode: "false"

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 5

persistence:
  size: 5Gi
  storageClass: "fast-ssd"

ingress:
  enabled: true
  hosts:
    - host: goldbox-rpg-staging.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: goldbox-rpg-staging-tls
      hosts:
        - goldbox-rpg-staging.yourdomain.com
```

Install with:
```bash
helm install goldbox-rpg-staging ./goldbox-rpg \
  -f staging-values.yaml \
  --namespace goldbox-rpg-staging \
  --create-namespace
```

### Production Environment

```yaml
# prod-values.yaml
replicaCount: 3

image:
  tag: "1.0.0"  # Use specific version, not latest

config:
  logLevel: "warn"
  devMode: "false"
  alertingEnabled: "true"

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10

persistence:
  size: 20Gi
  storageClass: "fast-ssd"

resources:
  limits:
    cpu: 1000m
    memory: 1024Mi
  requests:
    cpu: 250m
    memory: 256Mi

ingress:
  enabled: true
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/rate-limit: "100"
  hosts:
    - host: goldbox-rpg.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: goldbox-rpg-prod-tls
      hosts:
        - goldbox-rpg.yourdomain.com

secrets:
  websocketAllowedOrigins: "https://goldbox-rpg.yourdomain.com"

serviceMonitor:
  enabled: true
  interval: 30s

networkPolicy:
  enabled: true
```

Install with:
```bash
helm install goldbox-rpg ./goldbox-rpg \
  -f prod-values.yaml \
  --namespace goldbox-rpg \
  --create-namespace
```

## Health Checks

The chart configures three types of health checks:

1. **Liveness Probe** (`/live`) - Restarts the pod if it becomes unresponsive
2. **Readiness Probe** (`/ready`) - Removes pod from service if not ready
3. **Startup Probe** (`/health`) - Allows slow-starting pods to initialize

## Monitoring

### Prometheus Metrics

The application exposes Prometheus metrics at `/metrics`. To enable automatic scraping:

1. Enable ServiceMonitor (requires Prometheus Operator):
   ```yaml
   serviceMonitor:
     enabled: true
   ```

2. Or use pod annotations (enabled by default):
   ```yaml
   podAnnotations:
     prometheus.io/scrape: "true"
     prometheus.io/port: "8080"
     prometheus.io/path: "/metrics"
   ```

## Persistence

The chart mounts a Persistent Volume at `/data` for:
- Game state persistence
- Character data
- Session snapshots

If you want to disable persistence:
```yaml
persistence:
  enabled: false
```

**Warning**: Disabling persistence will cause data loss on pod restarts.

## Network Policy

The chart includes a NetworkPolicy to restrict network access. To enable:

```yaml
networkPolicy:
  enabled: true
```

This restricts:
- **Ingress**: Only from ingress-nginx namespace on port 8080
- **Egress**: DNS (port 53) and HTTPS (port 443)

## Upgrading

### From 0.x to 1.x

No breaking changes. Follow standard upgrade procedure.

## Troubleshooting

### Pods are not starting

Check pod status:
```bash
kubectl get pods -n goldbox-rpg
kubectl describe pod <pod-name> -n goldbox-rpg
kubectl logs <pod-name> -n goldbox-rpg
```

### Ingress not accessible

Verify ingress configuration:
```bash
kubectl get ingress -n goldbox-rpg
kubectl describe ingress goldbox-rpg -n goldbox-rpg
```

Check ingress controller logs:
```bash
kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx
```

### WebSocket connections failing

Ensure WebSocket origins are configured correctly:
```yaml
secrets:
  websocketAllowedOrigins: "https://yourdomain.com"
```

Check ingress annotations for WebSocket support:
```yaml
ingress:
  annotations:
    nginx.ingress.kubernetes.io/websocket-services: "goldbox-rpg"
```

### Persistence issues

Check PVC status:
```bash
kubectl get pvc -n goldbox-rpg
kubectl describe pvc goldbox-rpg-data -n goldbox-rpg
```

Verify storage class exists:
```bash
kubectl get storageclass
```

## Support

For issues and feature requests, please visit:
- GitHub: https://github.com/opd-ai/goldbox-rpg
- Issues: https://github.com/opd-ai/goldbox-rpg/issues

## License

This Helm chart is licensed under the MIT License. See the [LICENSE](../../../LICENSE) file for details.
