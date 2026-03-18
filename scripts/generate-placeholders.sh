#!/bin/bash
# ============================================================================
# DEPRECATED: Placeholder Asset Generation Script
# ============================================================================
# NOTE: This script is DEPRECATED. Production-ready AI-generated assets (521 PNGs)
# are now included in the repository under web/static/assets/sprites/. This script
# is retained for custom art style exploration only.
#
# For normal development and deployment, no asset generation is required.
# ============================================================================
#
# Generates simple colored PNG placeholder images from game-assets.yaml
# This allows development and testing without the external AI asset generator
#
# Usage:
#   ./scripts/generate-placeholders.sh
#   ./scripts/generate-placeholders.sh --dry-run    # Preview only
#   ./scripts/generate-placeholders.sh --clean      # Remove and regenerate

set -e

echo -e "\033[1;33mWARNING: This script is DEPRECATED. Production assets are already committed.\033[0m"
echo "See web/static/assets/sprites/ for 521 ready-to-use PNG assets."
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
YAML_FILE="$PROJECT_ROOT/game-assets.yaml"
OUTPUT_BASE="$PROJECT_ROOT/web/static/assets/sprites"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

DRY_RUN=false
CLEAN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --clean)
            CLEAN=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [--dry-run] [--clean]"
            exit 0
            ;;
        *)
            shift
            ;;
    esac
done

echo -e "${BLUE}🎨 GoldBox RPG Placeholder Asset Generator${NC}"
echo "==========================================="
echo "Project Root: $PROJECT_ROOT"
echo "YAML File:    $YAML_FILE"
echo "Output Base:  $OUTPUT_BASE"
echo ""

if [[ ! -f "$YAML_FILE" ]]; then
    echo -e "${RED}❌ game-assets.yaml not found at $YAML_FILE${NC}"
    exit 1
fi

# Check for required tools
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is required but not installed${NC}"
    exit 1
fi

# Create placeholder generator Go program
TEMP_DIR=$(mktemp -d)
cat > "$TEMP_DIR/placeholder.go" << 'GOEOF'
package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: placeholder <yaml_file> <output_base> <dry_run>")
		os.Exit(1)
	}

	yamlFile := os.Args[1]
	outputBase := os.Args[2]
	dryRun := os.Args[3] == "true"

	file, err := os.Open(yamlFile)
	if err != nil {
		fmt.Printf("Error opening YAML: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Parse YAML for asset definitions
	filenameRe := regexp.MustCompile(`^\s*filename:\s*["']?([^"'\s]+)["']?\s*$`)
	outputDirRe := regexp.MustCompile(`^\s*output_dir:\s*(.+)$`)
	idRe := regexp.MustCompile(`^\s*- id:\s*(.+)$`)

	var currentDir string
	var assetID string
	var created, skipped int

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Track current output directory
		if match := outputDirRe.FindStringSubmatch(line); len(match) > 1 {
			dir := strings.TrimSpace(match[1])
			// Only update if it's a subpath, not a full path
			if !strings.HasPrefix(dir, "/") && !strings.Contains(dir, ":") {
				currentDir = dir
			}
		}

		// Track asset ID for deterministic colors
		if match := idRe.FindStringSubmatch(line); len(match) > 1 {
			assetID = strings.TrimSpace(match[1])
		}

		// Generate placeholder for each filename
		if match := filenameRe.FindStringSubmatch(line); len(match) > 1 {
			filename := strings.TrimSpace(match[1])
			fullPath := filepath.Join(outputBase, currentDir, filename)

			if dryRun {
				fmt.Printf("Would create: %s\n", fullPath)
				created++
				continue
			}

			// Skip if file already exists
			if _, err := os.Stat(fullPath); err == nil {
				skipped++
				continue
			}

			// Create directory
			dir := filepath.Dir(fullPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("Error creating directory %s: %v\n", dir, err)
				continue
			}

			// Generate color based on asset ID
			clr := hashColor(assetID)

			// Create image
			img := image.NewRGBA(image.Rect(0, 0, 128, 128))
			draw.Draw(img, img.Bounds(), &image.Uniform{clr}, image.Point{}, draw.Src)

			// Add simple border
			borderClr := darken(clr)
			for x := 0; x < 128; x++ {
				for _, y := range []int{0, 1, 126, 127} {
					img.Set(x, y, borderClr)
				}
			}
			for y := 0; y < 128; y++ {
				for _, x := range []int{0, 1, 126, 127} {
					img.Set(x, y, borderClr)
				}
			}

			// Add simple X pattern to indicate placeholder
			for i := 0; i < 128; i++ {
				if i > 10 && i < 118 {
					img.Set(i, i, borderClr)
					img.Set(127-i, i, borderClr)
				}
			}

			// Save image
			f, err := os.Create(fullPath)
			if err != nil {
				fmt.Printf("Error creating %s: %v\n", fullPath, err)
				continue
			}

			if err := png.Encode(f, img); err != nil {
				f.Close()
				fmt.Printf("Error encoding %s: %v\n", fullPath, err)
				continue
			}
			f.Close()
			created++
		}
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Created: %d\n", created)
	fmt.Printf("  Skipped: %d (already exist)\n", skipped)
}

// hashColor generates a deterministic color from a string
func hashColor(s string) color.RGBA {
	h := md5.Sum([]byte(s))
	return color.RGBA{
		R: uint8(100 + int(h[0])%100), // Range 100-200 for muted colors
		G: uint8(100 + int(h[1])%100),
		B: uint8(100 + int(h[2])%100),
		A: 255,
	}
}

// darken creates a darker version of a color
func darken(c color.RGBA) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * 0.5),
		G: uint8(float64(c.G) * 0.5),
		B: uint8(float64(c.B) * 0.5),
		A: c.A,
	}
}
GOEOF

echo -e "${BLUE}🔧 Compiling placeholder generator...${NC}"
cd "$TEMP_DIR"
go build -o placeholder placeholder.go

echo -e "${BLUE}🖼️  Generating placeholder assets...${NC}"
echo ""

if [[ "$CLEAN" == "true" ]]; then
    echo -e "${YELLOW}🧹 Cleaning existing placeholders...${NC}"
    find "$OUTPUT_BASE" -name "*.png" -path "*/characters/*" -delete 2>/dev/null || true
    find "$OUTPUT_BASE" -name "*.png" -path "*/items/*" -delete 2>/dev/null || true
    find "$OUTPUT_BASE" -name "*.png" -path "*/spells/*" -delete 2>/dev/null || true
    find "$OUTPUT_BASE" -name "*.png" -path "*/terrain/*" -delete 2>/dev/null || true
    find "$OUTPUT_BASE" -name "*.png" -path "*/effects/*" -delete 2>/dev/null || true
    find "$OUTPUT_BASE" -name "*.png" -path "*/ui/*" -delete 2>/dev/null || true
fi

if [[ "$DRY_RUN" == "true" ]]; then
    ./placeholder "$YAML_FILE" "$OUTPUT_BASE" "true"
else
    ./placeholder "$YAML_FILE" "$OUTPUT_BASE" "false"
fi

# Cleanup
rm -rf "$TEMP_DIR"

echo ""
echo -e "${GREEN}✅ Placeholder generation complete!${NC}"

# Count total assets
TOTAL=$(find "$OUTPUT_BASE" -name "*.png" -type f 2>/dev/null | wc -l)
echo "Total PNG assets: $TOTAL"
