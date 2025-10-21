#!/bin/bash

# GoldBox RPG Engine - Asset Verification Script
# Validates that all required assets have been generated
# Usage: ./verify-assets.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Configuration
ASSETS_DIR="${PROJECT_ROOT}/web/static/assets/sprites"
PIPELINE_FILE="${PROJECT_ROOT}/game-assets.yaml"

# Print header
echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  GoldBox RPG Engine - Asset Verification                    ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if assets directory exists
if [ ! -d "${ASSETS_DIR}" ]; then
    echo -e "${RED}ERROR: Assets directory not found: ${ASSETS_DIR}${NC}"
    exit 1
fi

# Check required sprite sheets
echo -e "${BLUE}Checking core sprite sheets...${NC}"

CORE_SHEETS=("terrain.png" "characters.png" "effects.png" "ui.png")
MISSING_SHEETS=0

for sheet in "${CORE_SHEETS[@]}"; do
    if [ -f "${ASSETS_DIR}/${sheet}" ]; then
        SIZE=$(stat -f%z "${ASSETS_DIR}/${sheet}" 2>/dev/null || stat -c%s "${ASSETS_DIR}/${sheet}")
        SIZE_KB=$((SIZE / 1024))
        echo -e "${GREEN}✓ ${sheet} (${SIZE_KB} KB)${NC}"
    else
        echo -e "${RED}✗ ${sheet} - MISSING${NC}"
        MISSING_SHEETS=$((MISSING_SHEETS + 1))
    fi
done

echo ""

# Check generated asset directories
echo -e "${BLUE}Checking generated asset categories...${NC}"

ASSET_DIRS=(
    "characters/portraits"
    "monsters"
    "items"
    "terrain"
    "effects"
    "ui"
)

TOTAL_ASSETS=0
for dir in "${ASSET_DIRS[@]}"; do
    FULL_PATH="${ASSETS_DIR}/${dir}"
    if [ -d "${FULL_PATH}" ]; then
        COUNT=$(find "${FULL_PATH}" -type f -name "*.png" | wc -l | tr -d ' ')
        TOTAL_ASSETS=$((TOTAL_ASSETS + COUNT))
        echo -e "${GREEN}✓ ${dir} (${COUNT} files)${NC}"
    else
        echo -e "${YELLOW}⚠ ${dir} - Directory not found${NC}"
    fi
done

echo ""
echo -e "${BLUE}Asset Statistics:${NC}"
echo "  Total generated assets: ${TOTAL_ASSETS}"
echo "  Expected minimum: 100"

if [ $TOTAL_ASSETS -ge 500 ]; then
    echo -e "  ${GREEN}Status: COMPLETE (All assets generated)${NC}"
elif [ $TOTAL_ASSETS -ge 100 ]; then
    echo -e "  ${YELLOW}Status: PARTIAL (Priority assets generated)${NC}"
else
    echo -e "  ${YELLOW}Status: MINIMAL (Few assets generated)${NC}"
fi

# Check file sizes
echo ""
echo -e "${BLUE}Size Analysis:${NC}"

TOTAL_SIZE=$(du -sh "${ASSETS_DIR}" | cut -f1)
echo "  Total size: ${TOTAL_SIZE}"

# Check for issues
echo ""
echo -e "${BLUE}Checking for issues...${NC}"

ISSUES=0

# Check for empty files
EMPTY_FILES=$(find "${ASSETS_DIR}" -type f -name "*.png" -size 0)
if [ -n "$EMPTY_FILES" ]; then
    EMPTY_COUNT=$(echo "$EMPTY_FILES" | wc -l | tr -d ' ')
    echo -e "${RED}✗ Found ${EMPTY_COUNT} empty PNG files${NC}"
    ISSUES=$((ISSUES + 1))
else
    echo -e "${GREEN}✓ No empty files${NC}"
fi

# Check for very small files (potentially broken)
SMALL_FILES=$(find "${ASSETS_DIR}" -type f -name "*.png" -size -100c)
if [ -n "$SMALL_FILES" ]; then
    SMALL_COUNT=$(echo "$SMALL_FILES" | wc -l | tr -d ' ')
    echo -e "${YELLOW}⚠ Found ${SMALL_COUNT} very small PNG files (< 100 bytes)${NC}"
    ISSUES=$((ISSUES + 1))
else
    echo -e "${GREEN}✓ No suspiciously small files${NC}"
fi

# Check for consistent naming
echo -e "${GREEN}✓ File naming convention check (placeholder)${NC}"

# Summary
echo ""
echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
if [ $MISSING_SHEETS -eq 0 ] && [ $ISSUES -eq 0 ] && [ $TOTAL_ASSETS -ge 100 ]; then
    echo -e "${GREEN}║  ✓ Asset verification PASSED                                ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${GREEN}All assets are ready for use!${NC}"
    exit 0
else
    echo -e "${YELLOW}║  ⚠ Asset verification completed with warnings               ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    if [ $MISSING_SHEETS -gt 0 ]; then
        echo -e "${YELLOW}Missing ${MISSING_SHEETS} core sprite sheets${NC}"
    fi
    if [ $ISSUES -gt 0 ]; then
        echo -e "${YELLOW}Found ${ISSUES} potential issues${NC}"
    fi
    if [ $TOTAL_ASSETS -lt 100 ]; then
        echo -e "${YELLOW}Expected at least 100 assets, found ${TOTAL_ASSETS}${NC}"
    fi
    echo ""
    echo "Run './scripts/generate-all.sh' to generate missing assets"
    exit 1
fi
