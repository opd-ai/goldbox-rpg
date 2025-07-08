#!/bin/bash

# Advanced script to analyze Go test coverage and find untested files
# Usage: ./analyze_test_coverage.sh [options] [directory]
#
# Options:
#   -v, --verbose    Show detailed information
#   -j, --json      Output results in JSON format
#   -e, --exclude   Exclude patterns (comma-separated)
#   -h, --help      Show this help message

set -euo pipefail

# Default values
ROOT_DIR="."
VERBOSE=false
JSON_OUTPUT=false
EXCLUDE_PATTERNS=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

show_help() {
    cat << EOF
Usage: $0 [options] [directory]

Analyze Go test coverage and find source files without corresponding test files.

Options:
  -v, --verbose     Show detailed information including file sizes and package info
  -j, --json       Output results in JSON format
  -e, --exclude    Exclude patterns (comma-separated, e.g., "vendor,scripts")
  -h, --help       Show this help message

Examples:
  $0                          # Analyze current directory
  $0 ./pkg                    # Analyze specific directory
  $0 -v                       # Verbose output
  $0 -e "vendor,scripts"      # Exclude vendor and scripts directories
  $0 -j                       # JSON output for scripting

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -j|--json)
            JSON_OUTPUT=true
            shift
            ;;
        -e|--exclude)
            EXCLUDE_PATTERNS="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        -*)
            echo "Unknown option $1"
            show_help
            exit 1
            ;;
        *)
            ROOT_DIR="$1"
            shift
            ;;
    esac
done

# Check if directory exists
if [[ ! -d "$ROOT_DIR" ]]; then
    echo "Error: Directory '$ROOT_DIR' does not exist" >&2
    exit 1
fi

# Build find command with exclusions
find_cmd="find \"$ROOT_DIR\" -name \"*.go\""

if [[ -n "$EXCLUDE_PATTERNS" ]]; then
    IFS=',' read -ra patterns <<< "$EXCLUDE_PATTERNS"
    for pattern in "${patterns[@]}"; do
        find_cmd+=" -not -path \"*/$pattern/*\""
    done
fi

# Function to get file info
get_file_info() {
    local file="$1"
    local size=$(stat -c%s "$file" 2>/dev/null || echo "0")
    local package=$(head -1 "$file" | grep -o 'package [a-zA-Z_][a-zA-Z0-9_]*' | cut -d' ' -f2 || echo "unknown")
    echo "$size:$package"
}

# Find all Go source files (excluding test files and main.go)
if [[ "$JSON_OUTPUT" == "false" && "$VERBOSE" == "false" ]]; then
    echo -e "${BLUE}Scanning for Go source files without test files in: $ROOT_DIR${NC}"
    echo "============================================================"
fi

# Get source files
source_files=$(eval "$find_cmd -not -name \"*_test.go\" -not -name \"main.go\"" | sort)

untested_files=()
tested_files=()
total_source_files=0

for source_file in $source_files; do
    total_source_files=$((total_source_files + 1))
    
    # Get the directory and base name
    dir=$(dirname "$source_file")
    base_name=$(basename "$source_file" .go)
    
    # Check if corresponding test file exists
    test_file="${dir}/${base_name}_test.go"
    
    if [[ ! -f "$test_file" ]]; then
        if [[ "$VERBOSE" == "true" ]]; then
            file_info=$(get_file_info "$source_file")
            size=$(echo "$file_info" | cut -d':' -f1)
            package=$(echo "$file_info" | cut -d':' -f2)
            untested_files+=("$source_file:$size:$package")
        else
            untested_files+=("$source_file")
        fi
    else
        if [[ "$VERBOSE" == "true" ]]; then
            file_info=$(get_file_info "$source_file")
            size=$(echo "$file_info" | cut -d':' -f1)
            package=$(echo "$file_info" | cut -d':' -f2)
            tested_files+=("$source_file:$size:$package")
        else
            tested_files+=("$source_file")
        fi
    fi
done

# Calculate statistics
tested_count=${#tested_files[@]}
untested_count=${#untested_files[@]}
coverage_percentage=$(( (tested_count * 100) / total_source_files ))

# Output results
if [[ "$JSON_OUTPUT" == "true" ]]; then
    # JSON output
    echo "{"
    echo "  \"summary\": {"
    echo "    \"total_files\": $total_source_files,"
    echo "    \"tested_files\": $tested_count,"
    echo "    \"untested_files\": $untested_count,"
    echo "    \"coverage_percentage\": $coverage_percentage"
    echo "  },"
    echo "  \"untested_files\": ["
    
    for i in "${!untested_files[@]}"; do
        file="${untested_files[$i]}"
        if [[ "$VERBOSE" == "true" ]]; then
            filepath=$(echo "$file" | cut -d':' -f1)
            size=$(echo "$file" | cut -d':' -f2)
            package=$(echo "$file" | cut -d':' -f3)
            rel_path=$(realpath --relative-to="$ROOT_DIR" "$filepath")
            echo -n "    {\"file\": \"$rel_path\", \"size\": $size, \"package\": \"$package\"}"
        else
            rel_path=$(realpath --relative-to="$ROOT_DIR" "$file")
            echo -n "    {\"file\": \"$rel_path\"}"
        fi
        
        if [[ $i -lt $((${#untested_files[@]} - 1)) ]]; then
            echo ","
        else
            echo ""
        fi
    done
    
    echo "  ]"
    echo "}"
else
    # Human-readable output
    if [[ $untested_count -eq 0 ]]; then
        echo -e "${GREEN}✅ All Go source files have corresponding test files!${NC}"
        echo -e "${GREEN}Test coverage: 100% ($tested_count/$total_source_files files)${NC}"
    else
        echo -e "${RED}❌ Found $untested_count Go source files without test files:${NC}"
        echo ""
        
        for file in "${untested_files[@]}"; do
            if [[ "$VERBOSE" == "true" ]]; then
                filepath=$(echo "$file" | cut -d':' -f1)
                size=$(echo "$file" | cut -d':' -f2)
                package=$(echo "$file" | cut -d':' -f3)
                rel_path=$(realpath --relative-to="$ROOT_DIR" "$filepath")
                size_kb=$((size / 1024))
                echo -e "  ${YELLOW}$rel_path${NC} (${size_kb}KB, package: $package)"
            else
                rel_path=$(realpath --relative-to="$ROOT_DIR" "$file")
                echo "  $rel_path"
            fi
        done
        
        echo ""
        echo -e "${YELLOW}Summary:${NC}"
        echo "  • Total source files: $total_source_files"
        echo "  • Files with tests: $tested_count"
        echo "  • Files without tests: $untested_count"
        echo "  • Test coverage: $coverage_percentage%"
        
        if [[ "$VERBOSE" == "true" && $tested_count -gt 0 ]]; then
            echo ""
            echo -e "${GREEN}Files with test coverage:${NC}"
            for file in "${tested_files[@]}"; do
                filepath=$(echo "$file" | cut -d':' -f1)
                size=$(echo "$file" | cut -d':' -f2)
                package=$(echo "$file" | cut -d':' -f3)
                rel_path=$(realpath --relative-to="$ROOT_DIR" "$filepath")
                size_kb=$((size / 1024))
                echo -e "  ${GREEN}$rel_path${NC} (${size_kb}KB, package: $package)"
            done
        fi
    fi
fi
