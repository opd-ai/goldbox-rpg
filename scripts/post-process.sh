#!/bin/bash

# GoldBox RPG Engine - Asset Post-Processing Script
# Optimizes generated assets for production use
# Usage: ./post-process.sh [options]

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
STRIP_METADATA=true  # Always strip metadata for production assets
OPTIMIZE_PNG=true
VERBOSE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --no-optimize)
      OPTIMIZE_PNG=false
      shift
      ;;
    --verbose|-v)
      VERBOSE=true
      shift
      ;;
    --assets-dir)
      ASSETS_DIR="$2"
      shift 2
      ;;
    --help|-h)
      echo "Usage: $0 [options]"
      echo ""
      echo "Options:"
      echo "  --no-optimize      Don't optimize PNG compression"
      echo "  --verbose, -v      Show detailed output"
      echo "  --assets-dir DIR   Process assets in specified directory"
      echo "  --help, -h         Show this help message"
      exit 0
      ;;
    *)
      echo -e "${RED}Unknown option: $1${NC}"
      exit 1
      ;;
  esac
done

# Print header
echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  GoldBox RPG Engine - Asset Post-Processing                 ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if assets directory exists
if [ ! -d "${ASSETS_DIR}" ]; then
    echo -e "${RED}ERROR: Assets directory not found: ${ASSETS_DIR}${NC}"
    exit 1
fi

# Show configuration
echo -e "${GREEN}Configuration:${NC}"
echo "  Assets Directory: ${ASSETS_DIR}"
echo "  Strip Metadata: ${STRIP_METADATA}"
echo "  Optimize PNG: ${OPTIMIZE_PNG}"
echo "  Verbose: ${VERBOSE}"
echo ""

# Count assets
TOTAL_FILES=$(find "${ASSETS_DIR}" -type f -name "*.png" | wc -l)
echo -e "${BLUE}Found ${TOTAL_FILES} PNG files to process${NC}"
echo ""

# Check for optimization tools
HAS_OPTIPNG=false
HAS_PNGCRUSH=false
HAS_EXIFTOOL=false

if command -v optipng &> /dev/null; then
    HAS_OPTIPNG=true
    echo -e "${GREEN}✓ optipng available${NC}"
elif command -v pngcrush &> /dev/null; then
    HAS_PNGCRUSH=true
    echo -e "${GREEN}✓ pngcrush available${NC}"
fi

if command -v exiftool &> /dev/null; then
    HAS_EXIFTOOL=true
    echo -e "${GREEN}✓ exiftool available${NC}"
fi

if [ "$OPTIMIZE_PNG" = true ] && [ "$HAS_OPTIPNG" = false ] && [ "$HAS_PNGCRUSH" = false ]; then
    echo -e "${YELLOW}WARNING: No PNG optimization tool found (optipng or pngcrush)${NC}"
    echo -e "${YELLOW}Install with: apt-get install optipng (or) apt-get install pngcrush${NC}"
fi

if [ "$STRIP_METADATA" = true ] && [ "$HAS_EXIFTOOL" = false ]; then
    echo -e "${YELLOW}WARNING: exiftool not found for metadata stripping${NC}"
    echo -e "${YELLOW}Install with: apt-get install libimage-exiftool-perl${NC}"
fi

echo ""

# Start processing
PROCESSED=0
ERRORS=0

echo -e "${GREEN}Starting post-processing...${NC}"
echo ""

# Process each PNG file
find "${ASSETS_DIR}" -type f -name "*.png" | while read -r file; do
    PROCESSED=$((PROCESSED + 1))
    
    if [ "$VERBOSE" = true ]; then
        echo -e "${BLUE}Processing (${PROCESSED}/${TOTAL_FILES}): $(basename "$file")${NC}"
    fi
    
    # Strip metadata
    if [ "$STRIP_METADATA" = true ] && [ "$HAS_EXIFTOOL" = true ]; then
        exiftool -all= -overwrite_original "$file" 2>/dev/null || {
            echo -e "${YELLOW}WARNING: Failed to strip metadata from $(basename "$file")${NC}"
            ERRORS=$((ERRORS + 1))
        }
    fi
    
    # Optimize PNG
    if [ "$OPTIMIZE_PNG" = true ]; then
        if [ "$HAS_OPTIPNG" = true ]; then
            optipng -quiet -o7 "$file" 2>/dev/null || {
                echo -e "${YELLOW}WARNING: Failed to optimize $(basename "$file")${NC}"
                ERRORS=$((ERRORS + 1))
            }
        elif [ "$HAS_PNGCRUSH" = true ]; then
            pngcrush -q -ow "$file" 2>/dev/null || {
                echo -e "${YELLOW}WARNING: Failed to optimize $(basename "$file")${NC}"
                ERRORS=$((ERRORS + 1))
            }
        fi
    fi
done

# Show results
echo ""
echo -e "${GREEN}Post-processing complete!${NC}"
echo ""
echo "Statistics:"
echo "  Files processed: ${TOTAL_FILES}"
if [ $ERRORS -gt 0 ]; then
    echo -e "  ${YELLOW}Warnings: ${ERRORS}${NC}"
fi

# Calculate size savings if possible
if [ -f "${ASSETS_DIR}/.size_before" ]; then
    SIZE_BEFORE=$(cat "${ASSETS_DIR}/.size_before")
    SIZE_AFTER=$(du -sb "${ASSETS_DIR}" | cut -f1)
    SIZE_SAVED=$((SIZE_BEFORE - SIZE_AFTER))
    SIZE_SAVED_MB=$((SIZE_SAVED / 1024 / 1024))
    PERCENT_SAVED=$((SIZE_SAVED * 100 / SIZE_BEFORE))
    
    echo "  Size before: $((SIZE_BEFORE / 1024 / 1024)) MB"
    echo "  Size after: $((SIZE_AFTER / 1024 / 1024)) MB"
    echo -e "  ${GREEN}Saved: ${SIZE_SAVED_MB} MB (${PERCENT_SAVED}%)${NC}"
else
    # Store current size for next run
    du -sb "${ASSETS_DIR}" | cut -f1 > "${ASSETS_DIR}/.size_before"
fi

echo ""
echo -e "${GREEN}Assets are ready for production use!${NC}"
