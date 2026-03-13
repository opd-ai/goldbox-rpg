#!/bin/bash
# check-go-vuln.sh - Check for Go vulnerabilities and recommend upgrades
#
# Usage: ./scripts/check-go-vuln.sh
#
# Exit codes:
#   0 - No vulnerabilities found
#   1 - Vulnerabilities detected, upgrade recommended

set -e

echo "=== Go Vulnerability Check ==="
echo "Current Go version: $(go version)"
echo ""

# Check if govulncheck is installed
if ! command -v govulncheck &> /dev/null; then
    echo "Installing govulncheck..."
    go install golang.org/x/vuln/cmd/govulncheck@latest
fi

echo "Running govulncheck..."
echo ""

# Run govulncheck and capture exit code
if govulncheck ./...; then
    echo ""
    echo "✅ No vulnerabilities found in dependencies"
    exit 0
else
    echo ""
    echo "⚠️  Vulnerabilities detected!"
    echo ""
    echo "Recommended actions:"
    echo "1. Check for Go toolchain updates: https://go.dev/dl/"
    echo "2. Update dependencies: go get -u ./... && go mod tidy"
    echo "3. Review vulnerability details above"
    echo "4. Create a security upgrade issue using the template"
    echo ""
    echo "For Go stdlib vulnerabilities, upgrade to the latest patch version"
    echo "when available (e.g., Go 1.24.12+ or Go 1.25.8+)"
    exit 1
fi
