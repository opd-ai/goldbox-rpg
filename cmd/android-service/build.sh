#!/usr/bin/env bash
# FILENAME: build.sh
# PURPOSE: Build WASM client, bundle web + data assets, and cross-compile Go service for Android ARM64.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/android/app/src/main/jniLibs/arm64-v8a"
ASSETS_DIR="${SCRIPT_DIR}/android/app/src/main/assets"

mkdir -p "${OUTPUT_DIR}"

# ── Step 1: Build WASM client ──
echo "==> Building WASM UI (js/wasm)..."
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" \
  -o "${REPO_ROOT}/web/static/js/game.wasm" \
  "${REPO_ROOT}/cmd/wasm-ui"
echo "    game.wasm size: $(du -h "${REPO_ROOT}/web/static/js/game.wasm" | cut -f1)"

# ── Step 2: Copy wasm_exec.js from Go installation ──
echo "==> Copying wasm_exec.js..."
WASM_EXEC=""
if [ -f "$(go env GOROOT)/lib/wasm/wasm_exec.js" ]; then
  WASM_EXEC="$(go env GOROOT)/lib/wasm/wasm_exec.js"
elif [ -f "$(go env GOROOT)/misc/wasm/wasm_exec.js" ]; then
  WASM_EXEC="$(go env GOROOT)/misc/wasm/wasm_exec.js"
else
  echo "ERROR: wasm_exec.js not found in Go installation"
  exit 1
fi
cp "${WASM_EXEC}" "${REPO_ROOT}/web/static/js/wasm_exec.js"

# ── Step 3: Bundle web assets for Android ──
echo "==> Bundling web assets into Android assets..."
rm -rf "${ASSETS_DIR}/web" "${ASSETS_DIR}/data"
mkdir -p "${ASSETS_DIR}/web/static/js"

cp "${REPO_ROOT}/web/index.html" "${ASSETS_DIR}/web/"
cp "${REPO_ROOT}/web/static/js/game.wasm" "${ASSETS_DIR}/web/static/js/"
cp "${REPO_ROOT}/web/static/js/wasm_exec.js" "${ASSETS_DIR}/web/static/js/"

# Copy static content (adventures, sprites)
if [ -d "${REPO_ROOT}/web/static/adventures" ]; then
  cp -r "${REPO_ROOT}/web/static/adventures" "${ASSETS_DIR}/web/static/"
fi
if [ -d "${REPO_ROOT}/web/static/assets" ]; then
  cp -r "${REPO_ROOT}/web/static/assets" "${ASSETS_DIR}/web/static/"
fi

# ── Step 4: Bundle game data for Android ──
echo "==> Bundling game data into Android assets..."
cp -r "${REPO_ROOT}/data" "${ASSETS_DIR}/data"

# ── Step 5: Cross-compile Go service for Android ARM64 ──
echo "==> Building Go server for Android (android/arm64)..."
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
