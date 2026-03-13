#!/bin/bash
# Adventure Placeholder Asset Generation Script for GoldBox RPG Engine
#
# Generates placeholder PNG images for adventure-specific assets
# Reads adventure YAML files and creates placeholders for NPCs, items, and maps
#
# Usage:
#   ./scripts/generate-adventure-placeholders.sh                    # Generate for all adventures
#   ./scripts/generate-adventure-placeholders.sh sunken-sanctum     # Generate for specific adventure
#   ./scripts/generate-adventure-placeholders.sh --dry-run          # Preview only

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ADVENTURES_DATA="$PROJECT_ROOT/data/adventures"
ADVENTURES_ASSETS="$PROJECT_ROOT/web/static/adventures"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

DRY_RUN=false
TARGET_ADVENTURE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [--dry-run] [adventure-slug]"
            echo "  --dry-run         Preview assets to generate without creating files"
            echo "  adventure-slug    Generate for specific adventure (default: all)"
            exit 0
            ;;
        *)
            TARGET_ADVENTURE="$1"
            shift
            ;;
    esac
done

echo -e "${BLUE}🎮 GoldBox RPG Adventure Placeholder Generator${NC}"
echo "================================================"
echo "Data Directory:   $ADVENTURES_DATA"
echo "Output Directory: $ADVENTURES_ASSETS"
echo ""

if [[ ! -d "$ADVENTURES_DATA" ]]; then
    echo -e "${RED}❌ Adventures data directory not found: $ADVENTURES_DATA${NC}"
    exit 1
fi

# Create Go placeholder generator
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

cat > "$TEMP_DIR/generator.go" << 'GOEOF'
package main

import (
	"crypto/md5"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Adventure YAML structure (minimal for asset extraction)
type Adventure struct {
	Slug  string          `yaml:"adventure_slug"`
	Title string          `yaml:"adventure_title"`
	NPCs  []AdventureNPC  `yaml:"adventure_npcs"`
	Items []AdventureItem `yaml:"adventure_items"`
	Maps  []AdventureMap  `yaml:"adventure_maps"`
}

type AdventureNPC struct {
	ID      string `yaml:"npc_id"`
	Name    string `yaml:"npc_name"`
	Role    string `yaml:"npc_role"`
	Hostile bool   `yaml:"npc_hostile"`
}

type AdventureItem struct {
	ID   string `yaml:"item_id"`
	Name string `yaml:"item_name"`
	Type string `yaml:"item_type"`
}

type AdventureMap struct {
	ID   string `yaml:"map_id"`
	Name string `yaml:"map_name"`
}

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <yaml-file> <output-dir> <dry-run>\n", os.Args[0])
		os.Exit(1)
	}

	yamlFile := os.Args[1]
	outputDir := os.Args[2]
	dryRun := os.Args[3] == "true"

	data, err := os.ReadFile(yamlFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", yamlFile, err)
		os.Exit(1)
	}

	var adv Adventure
	if err := yaml.Unmarshal(data, &adv); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", yamlFile, err)
		os.Exit(1)
	}

	if adv.Slug == "" {
		fmt.Fprintf(os.Stderr, "Adventure has no slug\n")
		os.Exit(1)
	}

	advDir := filepath.Join(outputDir, adv.Slug)
	
	if !dryRun {
		if err := os.MkdirAll(advDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", advDir, err)
			os.Exit(1)
		}
	}

	generated := 0

	// Generate NPC portraits
	for _, npc := range adv.NPCs {
		filename := fmt.Sprintf("npc-%s.png", npc.ID)
		path := filepath.Join(advDir, filename)
		
		var baseColor color.RGBA
		switch {
		case npc.Role == "boss":
			baseColor = color.RGBA{150, 0, 0, 255} // Dark red for bosses
		case npc.Hostile:
			baseColor = color.RGBA{200, 50, 50, 255} // Red for hostile
		case npc.Role == "merchant":
			baseColor = color.RGBA{50, 150, 50, 255} // Green for merchants
		default:
			baseColor = color.RGBA{50, 50, 200, 255} // Blue for friendly
		}

		if dryRun {
			fmt.Printf("Would create: %s (NPC: %s)\n", path, npc.Name)
		} else {
			if err := generatePlaceholder(path, 64, 64, baseColor, npc.Name); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating %s: %v\n", path, err)
				continue
			}
			fmt.Printf("Created: %s\n", filename)
		}
		generated++
	}

	// Generate item icons
	for _, item := range adv.Items {
		filename := fmt.Sprintf("item-%s.png", item.ID)
		path := filepath.Join(advDir, filename)
		
		var baseColor color.RGBA
		switch item.Type {
		case "weapon":
			baseColor = color.RGBA{150, 150, 150, 255} // Gray for weapons
		case "armor":
			baseColor = color.RGBA{100, 100, 150, 255} // Blue-gray for armor
		case "key":
			baseColor = color.RGBA{200, 180, 50, 255} // Gold for keys
		case "consumable":
			baseColor = color.RGBA{50, 200, 50, 255} // Green for consumables
		case "quest_item":
			baseColor = color.RGBA{200, 150, 50, 255} // Orange for quest items
		default:
			baseColor = color.RGBA{150, 100, 150, 255} // Purple for accessories
		}

		if dryRun {
			fmt.Printf("Would create: %s (Item: %s)\n", path, item.Name)
		} else {
			if err := generatePlaceholder(path, 32, 32, baseColor, item.Name); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating %s: %v\n", path, err)
				continue
			}
			fmt.Printf("Created: %s\n", filename)
		}
		generated++
	}

	// Generate map thumbnails
	for _, m := range adv.Maps {
		filename := fmt.Sprintf("map-%s.png", m.ID)
		path := filepath.Join(advDir, filename)
		baseColor := color.RGBA{80, 60, 40, 255} // Brown for maps

		if dryRun {
			fmt.Printf("Would create: %s (Map: %s)\n", path, m.Name)
		} else {
			if err := generatePlaceholder(path, 128, 128, baseColor, m.Name); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating %s: %v\n", path, err)
				continue
			}
			fmt.Printf("Created: %s\n", filename)
		}
		generated++
	}

	// Generate adventure banner
	bannerPath := filepath.Join(advDir, "banner.png")
	if dryRun {
		fmt.Printf("Would create: %s (Adventure Banner)\n", bannerPath)
	} else {
		bannerColor := color.RGBA{30, 60, 100, 255} // Dark blue for banner
		if err := generatePlaceholder(bannerPath, 320, 180, bannerColor, adv.Title); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating banner: %v\n", err)
		} else {
			fmt.Printf("Created: banner.png\n")
		}
	}
	generated++

	fmt.Printf("\nTotal assets: %d\n", generated)
}

func generatePlaceholder(path string, width, height int, baseColor color.RGBA, label string) error {
	// Create image with slight color variation based on label hash
	hash := md5.Sum([]byte(label))
	r := int(baseColor.R) + int(hash[0]%30) - 15
	g := int(baseColor.G) + int(hash[1]%30) - 15
	b := int(baseColor.B) + int(hash[2]%30) - 15
	
	r = clamp(r, 0, 255)
	g = clamp(g, 0, 255)
	b = clamp(b, 0, 255)
	
	finalColor := color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{finalColor}, image.Point{}, draw.Src)
	
	// Add a border
	borderColor := color.RGBA{
		uint8(clamp(int(finalColor.R)-40, 0, 255)),
		uint8(clamp(int(finalColor.G)-40, 0, 255)),
		uint8(clamp(int(finalColor.B)-40, 0, 255)),
		255,
	}
	
	for x := 0; x < width; x++ {
		img.Set(x, 0, borderColor)
		img.Set(x, height-1, borderColor)
	}
	for y := 0; y < height; y++ {
		img.Set(0, y, borderColor)
		img.Set(width-1, y, borderColor)
	}
	
	// Add simple pattern (diagonal lines for texture)
	patternColor := color.RGBA{
		uint8(clamp(int(finalColor.R)+20, 0, 255)),
		uint8(clamp(int(finalColor.G)+20, 0, 255)),
		uint8(clamp(int(finalColor.B)+20, 0, 255)),
		255,
	}
	for y := 2; y < height-2; y += 4 {
		for x := 2; x < width-2; x += 4 {
			img.Set(x, y, patternColor)
		}
	}

	// Add label initial in center for larger images
	if width >= 64 && height >= 64 {
		initial := strings.ToUpper(string(label[0]))
		cx, cy := width/2, height/2
		textColor := color.RGBA{255, 255, 255, 200}
		
		// Draw a simple "letter" using pixels (5x7 grid)
		drawInitial(img, cx-2, cy-3, initial, textColor)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

func drawInitial(img *image.RGBA, x, y int, letter string, c color.Color) {
	// Simple 5x7 pixel font for common letters
	patterns := map[string][]string{
		"A": {"01110", "10001", "11111", "10001", "10001", "10001", "10001"},
		"B": {"11110", "10001", "11110", "10001", "10001", "10001", "11110"},
		"C": {"01110", "10001", "10000", "10000", "10000", "10001", "01110"},
		"D": {"11100", "10010", "10001", "10001", "10001", "10010", "11100"},
		"E": {"11111", "10000", "11110", "10000", "10000", "10000", "11111"},
		"F": {"11111", "10000", "11110", "10000", "10000", "10000", "10000"},
		"G": {"01110", "10001", "10000", "10111", "10001", "10001", "01110"},
		"H": {"10001", "10001", "11111", "10001", "10001", "10001", "10001"},
		"I": {"11111", "00100", "00100", "00100", "00100", "00100", "11111"},
		"K": {"10001", "10010", "10100", "11000", "10100", "10010", "10001"},
		"L": {"10000", "10000", "10000", "10000", "10000", "10000", "11111"},
		"M": {"10001", "11011", "10101", "10001", "10001", "10001", "10001"},
		"O": {"01110", "10001", "10001", "10001", "10001", "10001", "01110"},
		"P": {"11110", "10001", "11110", "10000", "10000", "10000", "10000"},
		"R": {"11110", "10001", "11110", "10100", "10010", "10001", "10001"},
		"S": {"01111", "10000", "01110", "00001", "00001", "10001", "01110"},
		"T": {"11111", "00100", "00100", "00100", "00100", "00100", "00100"},
		"U": {"10001", "10001", "10001", "10001", "10001", "10001", "01110"},
		"V": {"10001", "10001", "10001", "10001", "01010", "01010", "00100"},
		"W": {"10001", "10001", "10001", "10101", "10101", "11011", "10001"},
	}
	
	pattern, ok := patterns[letter]
	if !ok {
		// Default to a simple box for unknown letters
		pattern = []string{"11111", "10001", "10001", "10001", "10001", "10001", "11111"}
	}
	
	for py, row := range pattern {
		for px, ch := range row {
			if ch == '1' {
				img.Set(x+px, y+py, c)
			}
		}
	}
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
GOEOF

# Build the generator
echo -e "${BLUE}Building placeholder generator...${NC}"
cd "$TEMP_DIR"
go mod init placeholder-gen > /dev/null 2>&1
go get gopkg.in/yaml.v3 > /dev/null 2>&1
go build -o generator generator.go

TOTAL_GENERATED=0
ADVENTURES_PROCESSED=0

# Process adventures
for adv_dir in "$ADVENTURES_DATA"/*/; do
    if [[ ! -d "$adv_dir" ]]; then
        continue
    fi
    
    slug=$(basename "$adv_dir")
    
    # If specific adventure requested, skip others
    if [[ -n "$TARGET_ADVENTURE" && "$slug" != "$TARGET_ADVENTURE" ]]; then
        continue
    fi
    
    yaml_file="$adv_dir/adventure.yaml"
    if [[ ! -f "$yaml_file" ]]; then
        echo -e "${YELLOW}⚠️  Skipping $slug: no adventure.yaml${NC}"
        continue
    fi
    
    echo ""
    echo -e "${GREEN}Processing adventure: $slug${NC}"
    echo "-----------------------------------"
    
    dry_run_flag="false"
    if [[ "$DRY_RUN" == "true" ]]; then
        dry_run_flag="true"
    fi
    
    output=$("$TEMP_DIR/generator" "$yaml_file" "$ADVENTURES_ASSETS" "$dry_run_flag")
    echo "$output"
    
    # Extract count from output
    count=$(echo "$output" | grep "Total assets:" | grep -oP '\d+')
    if [[ -n "$count" ]]; then
        TOTAL_GENERATED=$((TOTAL_GENERATED + count))
    fi
    
    ADVENTURES_PROCESSED=$((ADVENTURES_PROCESSED + 1))
done

echo ""
echo "================================================"
echo -e "${GREEN}✅ Processing complete${NC}"
echo "Adventures processed: $ADVENTURES_PROCESSED"
echo "Total assets: $TOTAL_GENERATED"

if [[ "$DRY_RUN" == "true" ]]; then
    echo -e "${YELLOW}(Dry run - no files created)${NC}"
fi
