# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- **WebSocket Library**: Migrated server from `gorilla/websocket` to `nhooyr.io/websocket` (2026-03-13)
  - Server now uses actively-maintained nhooyr.io/websocket for all WebSocket connections
  - Added `WebSocketConn` adapter interface for library abstraction
  - Test clients retain gorilla/websocket (protocol is standard, interoperability confirmed)
  - All WebSocket features preserved: delta compression, origin validation, editor protocol
  - See `docs/WEBSOCKET_MIGRATION.md` for full details
- **Dependencies**: All dependencies verified as current for Go 1.23.8 (2026-03-11)
  - `github.com/prometheus/client_golang` v1.23.2 (latest for Go 1.23)
  - `github.com/stretchr/testify` v1.11.1 (latest)
  - All other dependencies at latest Go 1.23-compatible versions

### Added
- `pkg/server/websocket_upgrade.go` - nhooyr-based WebSocket upgrade and origin validation
- `CheckOriginAllowed()` method for testing origin validation without full connection

### Removed
- `pkg/server/websocket_gorilla.go` - gorilla adapter no longer needed for server
   
### Security
- **Known Vulnerabilities**: 18 Go standard library vulnerabilities identified by `govulncheck`
  - All vulnerabilities require Go 1.24.12+ or Go 1.25.8 to resolve
  - Vulnerabilities affect: `crypto/tls`, `crypto/x509`, `net/http`, `net/url`, `html/template`, `os`
  - **Recommended Action**: Upgrade to Go 1.24.12+ or Go 1.25.8 when available
  - No immediate risk for typical use cases, but upgrade recommended for production
  - See: https://pkg.go.dev/vuln/ for detailed vulnerability information

### Notes
- Dependencies cannot be updated beyond current versions without upgrading Go toolchain
- Latest versions of `github.com/hajimehoshi/ebiten/v2`, `github.com/ebitengine/gomobile`, and several other dependencies require Go 1.24.0+
- `golang.org/x/time` v0.15.0+ requires Go 1.25.0+
- Project remains on Go 1.23.8 until toolchain upgrade is planned

### Tested
- ✅ All tests passing with current dependency versions
- ✅ All E2E WebSocket tests pass with nhooyr server + gorilla client
- ✅ No performance regressions detected in benchmarks
- ✅ `go mod tidy` confirms dependency graph is clean
- ✅ No deprecated packages found (`github.com/golang/protobuf` not in use)

## Previous Changes

See git history for changes prior to CHANGELOG introduction.
