#!/bin/bash
# download-assets.sh - Downloads pre-generated asset pack from GitHub releases
#
# Usage: ./scripts/download-assets.sh [--verify] [--force]
#   --verify  Run asset verification after download
#   --force   Overwrite existing assets
#
# This script fetches the asset pack from the latest GitHub release.
# If no release is found or the pack doesn't exist, it falls back to
# placeholder asset generation.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ASSETS_DIR="$PROJECT_ROOT/web/static/assets/sprites"
REPO="opd-ai/goldbox-rpg"
VERIFY=false
FORCE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --verify)
            VERIFY=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [--verify] [--force]"
            echo "  --verify  Run asset verification after download"
            echo "  --force   Overwrite existing assets"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Check for existing assets
existing_count=$(find "$ASSETS_DIR" -name "*.png" 2>/dev/null | wc -l)
if [[ $existing_count -gt 200 && $FORCE != true ]]; then
    echo "Found $existing_count existing assets. Use --force to overwrite."
    echo "Current assets will be preserved."
    if [[ $VERIFY == true ]]; then
        "$SCRIPT_DIR/verify-assets.sh"
    fi
    exit 0
fi

echo "Checking for pre-generated asset pack..."

# Fetch latest release tag
LATEST_TAG=$(curl -sL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4 || true)

if [[ -z "$LATEST_TAG" ]]; then
    echo "No release found. Using placeholder assets instead."
    "$SCRIPT_DIR/generate-placeholders.sh"
    if [[ $VERIFY == true ]]; then
        "$SCRIPT_DIR/verify-assets.sh"
    fi
    exit 0
fi

echo "Found release: $LATEST_TAG"

# Construct asset URL
ASSET_URL="https://github.com/$REPO/releases/download/$LATEST_TAG/assets.tar.gz"

# Create temp file
TMP_FILE=$(mktemp)
trap "rm -f $TMP_FILE" EXIT

# Attempt download
echo "Downloading assets from $ASSET_URL..."
if curl -sfL "$ASSET_URL" -o "$TMP_FILE" 2>/dev/null; then
    echo "Download complete. Extracting assets..."
    
    # Create directories if needed
    mkdir -p "$ASSETS_DIR"
    
    # Extract assets
    tar -xzf "$TMP_FILE" -C "$ASSETS_DIR"
    
    echo "Assets downloaded from release $LATEST_TAG"
    
    # Count extracted assets
    new_count=$(find "$ASSETS_DIR" -name "*.png" | wc -l)
    echo "Total assets: $new_count"
else
    echo "Asset pack not found in release. Using placeholder assets instead."
    "$SCRIPT_DIR/generate-placeholders.sh"
fi

# Verify if requested
if [[ $VERIFY == true ]]; then
    echo ""
    echo "Running asset verification..."
    "$SCRIPT_DIR/verify-assets.sh"
fi

echo ""
echo "Asset download complete!"
