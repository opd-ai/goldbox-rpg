#!/usr/bin/env bash
# FILENAME: build.sh
# PURPOSE: Script to cross-compile Go service for Android ARM64 and place into jniLibs.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/android/app/src/main/jniLibs/arm64-v8a"

mkdir -p "${OUTPUT_DIR}"

echo "==> Building Go web service for Android (android/arm64)..."
CGO_ENABLED=0 GOOS=android GOARCH=arm64 \
  go build -trimpath -ldflags="-s -w" \
  -o "${OUTPUT_DIR}/libwebservice.so" \
  "${SCRIPT_DIR}"

echo "==> Binary size: $(du -h "${OUTPUT_DIR}/libwebservice.so" | cut -f1)"
echo "==> Output: ${OUTPUT_DIR}/libwebservice.so"

if command -v file &>/dev/null; then
  echo "==> Binary info: $(file "${OUTPUT_DIR}/libwebservice.so")"
fi

echo "==> Build complete."
echo ""
echo "Next steps:"
echo "  1. Open cmd/android-service/android/ in Android Studio"
echo "  2. Build and deploy to an ARM64 Android device"
