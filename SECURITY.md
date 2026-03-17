# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x     | :white_check_mark: |

## Security Measures

### WebSocket Security

This project uses [Coder WebSocket v1.8.14](https://github.com/coder/websocket) (a maintained fork of nhooyr.io/websocket) for production server WebSocket handling. The test client retains gorilla/websocket v1.5.3 for protocol compatibility testing.

**Note:** gorilla/websocket v1.5.3 includes fixes for **CVE-2020-27813** (CVSS 7.5), an integer overflow vulnerability patched in v1.4.2+. This applies only to the test client code.

#### Production Configuration

**Critical:** For production deployments, configure `WEBSOCKET_ALLOWED_ORIGINS` to restrict WebSocket connections to trusted domains:

```bash
export WEBSOCKET_ALLOWED_ORIGINS="https://yourdomain.com,https://www.yourdomain.com"
```

Without this environment variable, WebSocket connections from any origin are accepted (development mode).

### Input Validation

All JSON-RPC endpoints use the validation framework in `pkg/validation/` to:

- Sanitize user inputs
- Validate request size limits (default 1MB)
- Check parameter types and ranges
- Prevent injection attacks

### Session Security

- Sessions expire after 30 minutes (configurable via `GOLDBOX_SESSION_TIMEOUT`)
- Session IDs are generated using cryptographic randomness
- Thread-safe session management with proper mutex locking

### Rate Limiting

API endpoints are protected with rate limiting via `golang.org/x/time` to prevent abuse.

## Reporting a Vulnerability

If you discover a security vulnerability, please report it by:

1. Opening a private security advisory on GitHub
2. Emailing the maintainers directly

Please do **not** open a public issue for security vulnerabilities.

## Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `WEBSOCKET_ALLOWED_ORIGINS` | Comma-separated list of allowed WebSocket origins | None (all origins allowed) |
| `GOLDBOX_SESSION_TIMEOUT` | Session expiration duration | 30m |
| `GOLDBOX_LOG_LEVEL` | Logging verbosity | info |
| `GOLDBOX_PORT` | Server listening port | 8080 |

## Dependency Security

This project regularly updates dependencies. Key security-relevant dependencies:

- `github.com/coder/websocket v1.8.14` - WebSocket handling (production server)
- `github.com/gorilla/websocket v1.5.3` - WebSocket handling (E2E test client only, patched for CVE-2020-27813)
- `github.com/prometheus/client_golang v1.23.2` - Metrics collection
- `github.com/sirupsen/logrus v1.9.4` - Structured logging
