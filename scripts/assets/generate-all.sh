#!/bin/bash
# Asset Generation Master Script for GoldBox RPG Engine
# 
# This script generates all 2D game assets from the existing game metadata
# and procedural content system using asset-generator.
#
# Usage:
#   ./scripts/assets/generate-all.sh
#   ./scripts/assets/generate-all.sh --preview    # Dry run
#   ./scripts/assets/generate-all.sh --clean      # Clean and regenerate
#
# Environment Variables:
#   ASSET_GEN_API_URL     - SwarmUI API endpoint (default: http://localhost:7801)
#   ASSET_GEN_BASE_SEED   - Base seed for reproducibility (default: 42)
#   ASSET_GEN_STEPS       - Generation steps (default: 30)
#   ASSET_GEN_OUTPUT_DIR  - Output directory (default: ./output)

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ASSET_GEN="${ASSET_GEN:-asset-generator}"
API_URL="${ASSET_GEN_API_URL:-http://localhost:7801}"
BASE_SEED="${ASSET_GEN_BASE_SEED:-42}"
STEPS="${ASSET_GEN_STEPS:-30}"
OUTPUT_DIR="${ASSET_GEN_OUTPUT_DIR:-./output}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Parse command line arguments
PREVIEW_MODE=false
CLEAN_MODE=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --preview|--dry-run)
            PREVIEW_MODE=true
            shift
            ;;
        --clean)
            CLEAN_MODE=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --preview, --dry-run    Preview what would be generated"
            echo "  --clean                 Clean output directory first"
            echo "  --verbose, -v           Verbose output"
            echo "  --help, -h              Show this help"
            echo ""
            echo "Environment Variables:"
            echo "  ASSET_GEN_API_URL      SwarmUI API endpoint"
            echo "  ASSET_GEN_BASE_SEED    Base seed for reproducibility"
            echo "  ASSET_GEN_STEPS        Generation steps"
            echo "  ASSET_GEN_OUTPUT_DIR   Output directory"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Change to project root
cd "$PROJECT_ROOT"

echo -e "${BLUE}🎨 GoldBox RPG Asset Generation Pipeline${NC}"
echo "========================================"
echo "Project Root: $PROJECT_ROOT"
echo "Output Dir:   $OUTPUT_DIR"
echo "API URL:      $API_URL"
echo "Base Seed:    $BASE_SEED"
echo "Steps:        $STEPS"
if [[ "$PREVIEW_MODE" == "true" ]]; then
    echo -e "${YELLOW}Mode:         PREVIEW (dry run)${NC}"
fi
echo ""

# Check if asset-generator is installed
if ! command -v "$ASSET_GEN" &> /dev/null; then
    echo -e "${RED}❌ asset-generator not found${NC}"
    echo "📥 Install from: https://github.com/opd-ai/asset-generator/releases"
    echo ""
    echo "Quick install:"
    echo "curl -sSL https://github.com/opd-ai/asset-generator/releases/latest/download/asset-generator-linux-amd64 -o asset-generator"
    echo "chmod +x asset-generator"
    echo "sudo mv asset-generator /usr/local/bin/"
    exit 1
fi

# Verify asset-generator version
echo -e "${BLUE}🔧 Checking asset-generator...${NC}"
"$ASSET_GEN" --version
echo ""

# Check API connectivity
echo -e "${BLUE}🌐 Checking API connectivity...${NC}"
if ! curl -s --max-time 5 "$API_URL" > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Warning: Cannot connect to $API_URL${NC}"
    echo "   Make sure SwarmUI is running and accessible"
    echo "   You can override the URL with: ASSET_GEN_API_URL=http://your-server:port"
    echo ""
fi

# Clean output directory if requested
if [[ "$CLEAN_MODE" == "true" ]]; then
    echo -e "${YELLOW}🧹 Cleaning output directory...${NC}"
    rm -rf "$OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR"
    echo "✅ Output directory cleaned"
    echo ""
fi

# Function to generate assets with error handling
generate_pipeline() {
    local pipeline_file=$1
    local pipeline_name=$2
    local category=$3
    
    echo -e "${BLUE}📝 Processing ${pipeline_name}...${NC}"
    
    # Build command arguments
    local cmd_args=(
        "pipeline"
        "--file" "$pipeline_file"
        "--output-dir" "$OUTPUT_DIR"
        "--base-seed" "$BASE_SEED"
        "--steps" "$STEPS"
        "--api-url" "$API_URL"
        "--auto-crop"
        "--downscale-width" "512"
    )
    
    # Add preview mode if enabled
    if [[ "$PREVIEW_MODE" == "true" ]]; then
        cmd_args+=("--dry-run")
    fi
    
    # Add verbose flag if enabled
    if [[ "$VERBOSE" == "true" ]]; then
        cmd_args+=("--verbose")
    fi
    
    # Execute the command
    if "$ASSET_GEN" "${cmd_args[@]}"; then
        if [[ "$PREVIEW_MODE" == "true" ]]; then
            echo -e "${GREEN}✅ ${pipeline_name} preview complete${NC}"
        else
            echo -e "${GREEN}✅ ${pipeline_name} generated successfully${NC}"
            
            # Count generated files
            if [[ -d "$OUTPUT_DIR/$category" ]]; then
                local file_count=$(find "$OUTPUT_DIR/$category" -name "*.png" -type f | wc -l)
                echo "   Generated $file_count PNG files"
            fi
        fi
        echo ""
        return 0
    else
        echo -e "${RED}❌ ${pipeline_name} failed${NC}"
        echo ""
        return 1
    fi
}

# Track statistics
SUCCESSFUL_PIPELINES=0
FAILED_PIPELINES=0
START_TIME=$(date +%s)

# Generate all asset categories
echo -e "${BLUE}🚀 Starting asset generation...${NC}"
echo ""

# Character Assets
if generate_pipeline "assets/characters.yaml" "Character Portraits & NPCs" "characters"; then
    ((SUCCESSFUL_PIPELINES++))
else
    ((FAILED_PIPELINES++))
fi

# Spell Assets
if generate_pipeline "assets/spells.yaml" "Spell Effects & Icons" "spells"; then
    ((SUCCESSFUL_PIPELINES++))
else
    ((FAILED_PIPELINES++))
fi

# Item Assets
if generate_pipeline "assets/items.yaml" "Items & Equipment" "items"; then
    ((SUCCESSFUL_PIPELINES++))
else
    ((FAILED_PIPELINES++))
fi

# Environment Assets
if generate_pipeline "assets/environments.yaml" "Terrain & Environments" "environments"; then
    ((SUCCESSFUL_PIPELINES++))
else
    ((FAILED_PIPELINES++))
fi

# UI Assets
if generate_pipeline "assets/ui.yaml" "User Interface Elements" "ui"; then
    ((SUCCESSFUL_PIPELINES++))
else
    ((FAILED_PIPELINES++))
fi

# Calculate duration
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

# Final report
echo "========================================"
echo -e "${BLUE}📊 Asset Generation Summary${NC}"
echo "========================================"
echo "Successful pipelines: $SUCCESSFUL_PIPELINES"
echo "Failed pipelines:     $FAILED_PIPELINES"
echo "Total duration:       ${DURATION}s"

if [[ "$PREVIEW_MODE" == "false" ]]; then
    echo "Output directory:     $OUTPUT_DIR"
    
    # Count total generated files
    if [[ -d "$OUTPUT_DIR" ]]; then
        TOTAL_FILES=$(find "$OUTPUT_DIR" -name "*.png" -type f | wc -l)
        echo "Total PNG files:      $TOTAL_FILES"
        
        # Calculate total size
        TOTAL_SIZE=$(du -sh "$OUTPUT_DIR" 2>/dev/null | cut -f1)
        echo "Total size:           $TOTAL_SIZE"
    fi
fi

if [[ $FAILED_PIPELINES -eq 0 ]]; then
    echo -e "${GREEN}🎉 All assets generated successfully!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Review generated assets in $OUTPUT_DIR"
    echo "2. Integrate assets into your game engine"
    echo "3. Update asset paths in game configuration"
    exit 0
else
    echo -e "${RED}⚠️  Some pipelines failed. Check the output above for details.${NC}"
    exit 1
fi