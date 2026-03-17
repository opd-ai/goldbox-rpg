#!/usr/bin/env bash
# FILENAME: build.sh
# PURPOSE: Script to cross-compile Go service for Android ARM64 and copy into Android assets.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/android/app/src/main/assets"

mkdir -p "${OUTPUT_DIR}"

echo "==> Building Go web service for Android (android/arm64)..."
CGO_ENABLED=0 GOOS=android GOARCH=arm64 \
  go build -trimpath -ldflags="-s -w" \
  -o "${OUTPUT_DIR}/webservice" \
  "${SCRIPT_DIR}"

echo "==> Binary size: $(du -h "${OUTPUT_DIR}/webservice" | cut -f1)"
echo "==> Output: ${OUTPUT_DIR}/webservice"

if command -v file &>/dev/null; then
  echo "==> Binary info: $(file "${OUTPUT_DIR}/webservice")"
fi

echo "==> Build complete."
echo ""
echo "Next steps:"
echo "  1. Open cmd/android-service/android/ in Android Studio"
echo "  2. Build and deploy to an ARM64 Android device"
