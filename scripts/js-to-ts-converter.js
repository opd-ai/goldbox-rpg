#!/bin/bash

# TypeScript to JavaScript migration helper script
# Converts existing JavaScript files to TypeScript with proper typing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SRC_DIR="$PROJECT_ROOT/src"
JS_DIR="$PROJECT_ROOT/web/static/js"

echo "🔄 JavaScript to TypeScript Migration Helper"
echo "============================================"

# Function to convert a JavaScript file to TypeScript
convert_js_to_ts() {
    local js_file="$1"
    local target_dir="$2"
    local filename=$(basename "$js_file" .js)
    
    echo "📝 Converting $js_file..."
    
    # Create target directory if it doesn't exist
    mkdir -p "$target_dir"
    
    # Copy file and rename to .ts
    local ts_file="$target_dir/${filename}.ts"
    cp "$js_file" "$ts_file"
    
    # Add basic TypeScript annotations (basic conversion)
    # This is a simple conversion - manual review will be needed
    
    # Add type imports at the top
    sed -i '1i\/**' "$ts_file"
    sed -i '2i\ * Migrated from JavaScript to TypeScript' "$ts_file"
    sed -i '3i\ * TODO: Add proper type annotations and review implementation' "$ts_file"
    sed -i '4i\ */' "$ts_file"
    sed -i '5i\\' "$ts_file"
    
    echo "✅ Converted $js_file -> $ts_file"
}

# Function to backup existing files
backup_files() {
    local backup_dir="$PROJECT_ROOT/backup_$(date +%Y%m%d_%H%M%S)"
    echo "📦 Creating backup at $backup_dir"
    
    mkdir -p "$backup_dir"
    cp -r "$JS_DIR" "$backup_dir/"
    
    echo "✅ Backup created successfully"
}

# Main migration process
main() {
    echo "🎯 Target directories:"
    echo "   Source: $JS_DIR"
    echo "   Target: $SRC_DIR"
    echo ""
    
    # Create backup
    backup_files
    
    # Create migration directories
    mkdir -p "$SRC_DIR/game"
    mkdir -p "$SRC_DIR/rendering"
    mkdir -p "$SRC_DIR/network"
    mkdir -p "$SRC_DIR/ui"
    mkdir -p "$SRC_DIR/services"
    
    echo "📂 Migration plan:"
    echo "   game.js -> src/game/GameState.ts"
    echo "   combat.js -> src/game/CombatManager.ts"
    echo "   render.js -> src/rendering/GameRenderer.ts"
    echo "   rpc.js -> src/network/RPCClient.ts"
    echo "   ui.js -> src/ui/UIManager.ts"
    echo "   spatial.js -> src/utils/SpatialQueryManager.ts (already done)"
    echo ""
    
    # Convert specific files (examples - adjust paths as needed)
    if [ -f "$JS_DIR/game.js" ]; then
        convert_js_to_ts "$JS_DIR/game.js" "$SRC_DIR/game"
        mv "$SRC_DIR/game/game.ts" "$SRC_DIR/game/GameState.ts"
    fi
    
    if [ -f "$JS_DIR/combat.js" ]; then
        convert_js_to_ts "$JS_DIR/combat.js" "$SRC_DIR/game"
        mv "$SRC_DIR/game/combat.ts" "$SRC_DIR/game/CombatManager.ts"
    fi
    
    if [ -f "$JS_DIR/render.js" ]; then
        convert_js_to_ts "$JS_DIR/render.js" "$SRC_DIR/rendering"
        mv "$SRC_DIR/rendering/render.ts" "$SRC_DIR/rendering/GameRenderer.ts"
    fi
    
    if [ -f "$JS_DIR/rpc.js" ]; then
        convert_js_to_ts "$JS_DIR/rpc.js" "$SRC_DIR/network"
        mv "$SRC_DIR/network/rpc.ts" "$SRC_DIR/network/RPCClient.ts"
    fi
    
    if [ -f "$JS_DIR/ui.js" ]; then
        convert_js_to_ts "$JS_DIR/ui.js" "$SRC_DIR/ui"
        mv "$SRC_DIR/ui/ui.ts" "$SRC_DIR/ui/UIManager.ts"
    fi
    
    echo ""
    echo "🎉 Migration completed!"
    echo ""
    echo "📋 Next steps:"
    echo "   1. Review generated TypeScript files"
    echo "   2. Add proper type annotations"
    echo "   3. Fix any TypeScript errors"
    echo "   4. Run 'npm run typecheck' to validate"
    echo "   5. Run 'npm run build' to test compilation"
    echo ""
    echo "💡 Tip: Use 'npm run watch' for development with auto-compilation"
}

# Check if this is being run from the correct directory
if [ ! -f "$PROJECT_ROOT/package.json" ]; then
    echo "❌ Error: Run this script from the project root directory"
    exit 1
fi

# Run main function
main
