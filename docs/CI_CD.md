# CI/CD Pipeline Documentation

## Overview

The GoldBox RPG Engine now has a comprehensive CI/CD pipeline implemented with GitHub Actions. This pipeline automates testing, linting, security scanning, building, and deployment processes.

## Pipeline Architecture

### Pull Request Workflow (`ci.yml`)

Triggered on every pull request and push to the main branch. Includes 6 parallel jobs:

#### 1. Test Job
- **Purpose**: Run all tests with race detection and enforce coverage thresholds
- **Steps**:
  - Run `go test ./... -v -race -coverprofile=coverage.out`
  - Generate and display coverage report
  - Enforce minimum 78% coverage threshold (current baseline)
  - Upload coverage report as artifact
- **Duration**: ~2-5 minutes
- **Why Important**: Catches regressions and race conditions before merge

#### 2. Lint Job
- **Purpose**: Run static code analysis with golangci-lint
- **Configuration**: `.golangci.yml`
- **Linters Enabled**:
  - errcheck, gosimple, govet, ineffassign, staticcheck, unused
  - gofmt, goimports, goconst, misspell, unconvert
  - gosec (security), revive, and more
- **Duration**: ~1-2 minutes
- **Why Important**: Maintains code quality and catches common bugs

#### 3. Format Check Job
- **Purpose**: Ensure all code follows consistent formatting
- **Tool**: gofumpt (stricter than gofmt)
- **Command**: `gofumpt -l -s -extra ./pkg`
- **Fix Locally**: Run `make fmt` to fix formatting issues
- **Duration**: ~30 seconds
- **Why Important**: Enforces consistent code style across all contributors

#### 4. Security Scan Job
- **Purpose**: Detect security vulnerabilities in dependencies
- **Tools**:
  - `govulncheck`: Scans for known vulnerabilities in Go dependencies
  - `go list -m -u all`: Checks for available dependency updates
- **Duration**: ~1-2 minutes
- **Why Important**: Prevents deployment of code with known vulnerabilities

#### 5. Build Job
- **Purpose**: Verify the server binary compiles successfully
- **Command**: `make build`
- **Output**: `bin/server` binary
- **Duration**: ~1-2 minutes
- **Why Important**: Catches compilation errors before merge

#### 6. Docker Build Job
- **Purpose**: Validate Dockerfile and test containerized deployment
- **Steps**:
  - Build Docker image: `goldbox-rpg:test`
  - Run container on port 8080
  - Test health endpoint: `GET /health`
  - Test readiness endpoint: `GET /ready`
  - Stop container
- **Cache**: GitHub Actions cache for faster builds
- **Duration**: ~2-3 minutes
- **Why Important**: Ensures Docker deployments work correctly

### Build and Push Workflow (`build.yml`)

Triggered on pushes to main branch and version tags.

#### Features
- **Multi-arch Support**: Builds for amd64 and arm64 (configurable)
- **Registry**: GitHub Container Registry (ghcr.io)
- **Image Tagging Strategy**:
  - `latest` - Latest commit on main branch
  - `main-<sha>` - Specific commit on main branch
  - `v1.2.3` - Semantic version tags
  - `v1.2` - Major.minor version
- **Metadata**: Includes build date, VCS ref, and version labels
- **Attestation**: Generates build provenance for supply chain security
- **Duration**: ~3-5 minutes

## Dependabot Configuration

Automated dependency updates configured in `.github/dependabot.yml`:

### Update Schedule
- **Frequency**: Weekly on Monday at 09:00
- **Go Modules**: Minor and patch updates grouped together
- **NPM**: Frontend dependency updates
- **GitHub Actions**: Action version updates
- **Docker**: Base image updates

## Linter Configuration

The `.golangci.yml` file configures 20+ linters for comprehensive code analysis.

## Local Development

### Before Committing

```bash
# 1. Format code
make fmt

# 2. Run tests with race detector
go test ./... -race

# 3. Run linter
golangci-lint run

# 4. Check coverage
make test-coverage

# 5. Build to verify
make build
```

## Troubleshooting

### Common Issues

#### Test Failures
**Solution**: Run tests with `-race` flag locally and check for environment differences.

#### Coverage Drop
**Solution**: Run `make find-untested` to identify uncovered files.

#### Linter Errors
**Solution**: Run `golangci-lint run` locally and fix reported issues.

## Integration with Roadmap

This CI/CD implementation completes **Phase 1.2** from ROADMAP.md:

### Completed ✅
- [x] CI/CD pipeline for automated testing
- [x] Race detector on every PR
- [x] Coverage enforcement (78% baseline)
- [x] Security vulnerability scanning
- [x] Dockerfile validation
- [x] Build status badges in README
- [x] Dependabot configuration
- [x] Linter configuration

### Next Steps (Phase 1.3-1.4)
- [ ] Secrets management implementation
- [ ] Deployment manifests (K8s, Docker Compose)

---

**Last Updated**: 2025-10-29
**Owner**: DevOps Team
**Status**: ✅ Active and Production-Ready
