---
name: Security Upgrade
about: Track Go toolchain security updates
title: "[SECURITY] Go version upgrade required"
labels: security, maintenance
assignees: ''
---

## Current Go Version
<!-- Output of `go version` -->

## Required Version
<!-- Minimum Go version needed for security fix -->

## Vulnerabilities Fixed
<!-- List of CVEs addressed by the upgrade -->

## Upgrade Steps
1. Update `go.mod` to new Go version
2. Update `.github/workflows/*.yml` Go version
3. Update `Dockerfile` Go version
4. Run `go mod tidy`
5. Run full test suite: `go test -race ./...`
6. Run vulnerability check: `govulncheck ./...`

## Checklist
- [ ] CI passes with new Go version
- [ ] No breaking changes in dependencies
- [ ] Documentation updated if needed
- [ ] CHANGELOG updated
