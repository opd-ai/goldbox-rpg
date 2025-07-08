#!/bin/bash

# Script to find Go source files without corresponding test files
# Usage: ./find_untested_files.sh [directory]

set -euo pipefail

# Default to current directory if no argument provided
ROOT_DIR="${1:-.}"

# Check if directory exists
if [[ ! -d "$ROOT_DIR" ]]; then
    echo "Error: Directory '$ROOT_DIR' does not exist" >&2
    exit 1
fi

echo "Scanning for Go source files without test files in: $ROOT_DIR"
echo "============================================================"

# Find all Go source files (excluding test files and main.go)
source_files=$(find "$ROOT_DIR" -name "*.go" -not -name "*_test.go" -not -name "main.go" | sort)

untested_files=()

for source_file in $source_files; do
    # Get the directory and base name
    dir=$(dirname "$source_file")
    base_name=$(basename "$source_file" .go)
    
    # Check if corresponding test file exists
    test_file="${dir}/${base_name}_test.go"
    
    if [[ ! -f "$test_file" ]]; then
        untested_files+=("$source_file")
    fi
done

# Display results
if [[ ${#untested_files[@]} -eq 0 ]]; then
    echo "✅ All Go source files have corresponding test files!"
else
    echo "❌ Found ${#untested_files[@]} Go source files without test files:"
    echo
    
    for file in "${untested_files[@]}"; do
        # Show relative path from ROOT_DIR
        rel_path=$(realpath --relative-to="$ROOT_DIR" "$file")
        echo "  $rel_path"
    done
    
    echo
    echo "Summary: ${#untested_files[@]} files need test coverage"
fi
