# GoldBox RPG Engine - Asset Integration Guide

**Version:** 1.0  
**Last Updated:** 2025-10-21

## Table of Contents

1. [Introduction](#introduction)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [Quick Start](#quick-start)
5. [Pipeline Structure](#pipeline-structure)
6. [Generation Workflow](#generation-workflow)
7. [Build System Integration](#build-system-integration)
8. [Customization](#customization)
9. [Troubleshooting](#troubleshooting)
10. [Best Practices](#best-practices)

## Introduction

This guide explains how to use the automated asset generation pipeline for the GoldBox RPG Engine. The pipeline is designed to generate all visual assets needed by the game, including:

- Character portraits and battle sprites
- Monster sprites
- Item icons and equipment
- Terrain tiles
- Combat effects
- UI elements

The pipeline is based on the comprehensive analysis documented in [ASSET_ANALYSIS.md](./ASSET_ANALYSIS.md) and uses a YAML configuration file ([game-assets.yaml](./game-assets.yaml)) to define all assets.

## Prerequisites

### Required Tools

1. **Asset Generator Tool**
   - The pipeline requires an asset generation tool (e.g., Stable Diffusion, DALL-E, Midjourney CLI, or similar)
   - The tool should support command-line batch generation
   - Should accept prompts, seeds, and output specifications

2. **Image Optimization Tools** (Optional but recommended)
   ```bash
   # Install on Ubuntu/Debian
   sudo apt-get install optipng exiftool
   
   # Install on macOS
   brew install optipng exiftool
   ```

3. **Standard Unix Tools**
   - bash (4.0+)
   - find, grep, sed
   - Basic utilities (cat, wc, du, etc.)

### System Requirements

- **Disk Space:** At least 500 MB free for generated assets
- **Memory:** 4GB RAM minimum for asset generation
- **Network:** Internet connection for downloading asset generation tools

## Installation

### Step 1: Verify Repository Setup

Ensure you have cloned the GoldBox RPG repository:

```bash
git clone https://github.com/opd-ai/goldbox-rpg.git
cd goldbox-rpg
```

### Step 2: Install Asset Generation Tool

**Option A: Using a local Stable Diffusion setup**

```bash
# Install Stable Diffusion locally (example)
# Follow instructions at: https://github.com/AUTOMATIC1111/stable-diffusion-webui
# Or use the API endpoint
```

**Option B: Using an online service**

Configure your API credentials for services like:
- OpenAI DALL-E API
- Midjourney API (via third-party)
- Replicate API

```bash
# Set API credentials
export ASSET_GENERATOR_API_KEY="your-api-key"
```

**Option C: Using a custom tool**

If you have a custom asset-generator tool, ensure it supports the following command structure:

```bash
asset-generator pipeline \
  --file game-assets.yaml \
  --output-dir ./output \
  --base-seed 42 \
  --auto-crop \
  --downscale-width 1024
```

### Step 3: Verify Installation

```bash
# Check if asset-generator is available
which asset-generator

# Test with dry-run
./scripts/generate-all.sh --dry-run
```

## Quick Start

### Generate All Assets

The fastest way to generate all assets:

```bash
# Generate complete asset library (4-6 hours)
./scripts/generate-all.sh

# With verbose output
./scripts/generate-all.sh --verbose

# Preview what would be generated (dry-run)
./scripts/generate-all.sh --dry-run
```

### Generate Priority Assets Only

For quick testing, generate only critical assets:

```bash
# Generate essential assets first (~30 minutes)
./scripts/generate-priority1.sh
```

### Verify Generated Assets

After generation, verify all assets were created correctly:

```bash
./scripts/verify-assets.sh
```

### Post-Process Assets

Optimize assets for production:

```bash
./scripts/post-process.sh
```

### Build and Run

Integrate assets into the game:

```bash
# Build the project
make build

# Run the game server
make run

# Or use Docker
make docker
```

## Pipeline Structure

### YAML Configuration

The `game-assets.yaml` file defines the complete asset pipeline using a hierarchical structure:

```yaml
metadata:
  style: "fantasy RPG pixel art..."
  quality: "high detail, clean lines..."
  
config:
  output_base: "web/static/assets/sprites"
  image_format: "png"
  
assets:
  - name: Character Portraits
    output_dir: characters/portraits
    seed_offset: 0
    metadata: { ... }
    subgroups:
      - name: Fighter Portraits
        assets:
          - id: fighter_human_male
            prompt: "male human fighter..."
            filename: "portrait_fighter_human_male.png"
```

### Key Concepts

1. **Metadata Cascading**: Child groups inherit parent metadata
2. **Seed Offsets**: Reproducible generation with logical seed values
3. **Hierarchical Structure**: Organizes assets by category and subcategory
4. **Output Directory**: Mirrors or enhances game's asset structure

### Asset Categories

| Category | Count | Seed Offset | Directory |
|----------|-------|-------------|-----------|
| Character Portraits | 48 | 0 | characters/portraits |
| Monsters | 35 | 1000 | monsters |
| Items | 60 | 2000 | items |
| Terrain | 80 | 3000 | terrain |
| Effects | 150 | 4000 | effects |
| UI Elements | 148 | 5000 | ui |

## Generation Workflow

### Full Pipeline

```bash
# 1. Clean previous generation (optional)
make assets-clean

# 2. Preview generation
./scripts/generate-all.sh --dry-run

# 3. Generate assets
./scripts/generate-all.sh --seed 42

# 4. Post-process for optimization
./scripts/post-process.sh

# 5. Verify all assets
./scripts/verify-assets.sh

# 6. Integrate with build
make assets

# 7. Build and test
make build && make run
```

### Incremental Workflow

For iterative development:

```bash
# Generate Priority 1 (critical) assets
./scripts/generate-priority1.sh

# Test with minimal assets
make build && make run

# Generate remaining categories as needed
# (Use category-specific generation commands)

# Final verification
./scripts/verify-assets.sh
```

### Regeneration Workflow

To regenerate specific assets:

```bash
# Regenerate with different seed
./scripts/generate-all.sh --seed 123

# Regenerate specific category
# (Edit game-assets.yaml to enable only desired category)

# Or regenerate individual assets
# (Use asset-generator directly with specific IDs)
```

## Build System Integration

### Makefile Targets

The following targets have been added to the Makefile:

```makefile
# Generate all assets
make assets

# Preview asset generation (dry-run)
make assets-preview

# Clean generated assets
make assets-clean

# Post-process assets
make assets-optimize

# Verify assets
make assets-verify
```

### Integration Points

1. **Development Build**: Assets can be generated on-demand
2. **CI/CD Pipeline**: Assets can be cached or generated during deployment
3. **Docker Build**: Assets are included in the Docker image

### Example CI/CD Integration

```yaml
# .github/workflows/build.yml
- name: Generate Assets
  run: |
    ./scripts/generate-all.sh --seed 42
    ./scripts/post-process.sh
    
- name: Verify Assets
  run: ./scripts/verify-assets.sh
  
- name: Build
  run: make build
```

## Customization

### Modifying Prompts

Edit `game-assets.yaml` to customize asset generation:

```yaml
assets:
  - name: Character Portraits
    metadata:
      style: "YOUR CUSTOM STYLE HERE"
    subgroups:
      - name: Fighters
        assets:
          - id: fighter_human_male
            prompt: "YOUR CUSTOM PROMPT HERE"
```

### Adding New Assets

1. **Add to YAML Configuration**:
   ```yaml
   - id: new_asset_id
     name: new-asset-name
     prompt: "detailed prompt for new asset"
     filename: "new_asset.png"
     seed_offset: N
   ```

2. **Regenerate**:
   ```bash
   ./scripts/generate-all.sh
   ```

3. **Verify**:
   ```bash
   ./scripts/verify-assets.sh
   ```

### Changing Art Style

To change the overall art style:

1. **Update Global Metadata** in `game-assets.yaml`:
   ```yaml
   metadata:
     style: "YOUR NEW STYLE (e.g., 'anime style', '3D render')"
     quality: "YOUR QUALITY PREFERENCES"
     negative: "THINGS TO AVOID"
   ```

2. **Regenerate All Assets**:
   ```bash
   ./scripts/generate-all.sh --seed NEW_SEED
   ```

### Output Directory Structure

Customize the output structure by modifying `output_dir` in YAML:

```yaml
assets:
  - name: Characters
    output_dir: "custom/path/characters"  # Custom path
```

## Troubleshooting

### Common Issues

#### Issue: "asset-generator command not found"

**Solution:**
```bash
# Install or configure your asset generation tool
# Ensure it's in your PATH
export PATH="$PATH:/path/to/asset-generator"

# Or create a symlink
ln -s /path/to/actual-tool /usr/local/bin/asset-generator
```

#### Issue: "Generation fails with API errors"

**Solution:**
```bash
# Check API credentials
echo $ASSET_GENERATOR_API_KEY

# Verify API quota/rate limits
# Try with fewer concurrent requests
# Use --delay option if available
```

#### Issue: "Generated assets are too large"

**Solution:**
```bash
# Use post-processing to optimize
./scripts/post-process.sh

# Or adjust generation parameters
./scripts/generate-all.sh --downscale-width 512

# Use compression tools
optipng -o7 assets/**/*.png
```

#### Issue: "Assets don't match game style"

**Solution:**
1. Review prompts in `game-assets.yaml`
2. Adjust `style`, `quality`, and `negative` metadata
3. Try different seed values
4. Regenerate specific categories

#### Issue: "Missing assets after generation"

**Solution:**
```bash
# Run verification
./scripts/verify-assets.sh

# Check generation logs for errors
./scripts/generate-all.sh --verbose 2>&1 | tee generation.log

# Regenerate missing assets
# (identify missing IDs and regenerate individually)
```

### Debugging Tips

1. **Use Dry-Run Mode**:
   ```bash
   ./scripts/generate-all.sh --dry-run
   ```

2. **Enable Verbose Logging**:
   ```bash
   ./scripts/generate-all.sh --verbose
   ```

3. **Check Individual Assets**:
   ```bash
   # Verify specific file exists
   ls -lh web/static/assets/sprites/characters/portraits/fighters/portrait_fighter_human_male.png
   ```

4. **Validate YAML Syntax**:
   ```bash
   # Use a YAML validator
   python3 -c "import yaml; yaml.safe_load(open('game-assets.yaml'))"
   ```

## Best Practices

### Generation Best Practices

1. **Use Version Control**: Commit `game-assets.yaml` to track asset definitions
2. **Document Seed Values**: Record seed values used for each generation
3. **Incremental Generation**: Start with Priority 1, then expand
4. **Regular Verification**: Run `verify-assets.sh` after each generation
5. **Post-Process Always**: Optimize assets before committing

### Performance Optimization

1. **Batch Generation**: Generate all assets in one run when possible
2. **Caching**: Cache generated assets in CI/CD
3. **Parallel Generation**: Use parallel processing if supported
4. **Selective Regeneration**: Only regenerate changed assets

### Asset Management

1. **Naming Conventions**: Follow the established naming scheme
2. **Organization**: Keep assets organized by category
3. **Documentation**: Document custom assets and modifications
4. **Backup**: Keep backups of successfully generated asset sets

### Integration Best Practices

1. **Test Incrementally**: Test each asset category as it's generated
2. **Version Assets**: Tag asset generations with version numbers
3. **Separate Concerns**: Keep asset generation separate from game logic
4. **Automate Verification**: Include asset verification in CI/CD

## Asset Loading in Game

### Frontend Integration

The game loads assets via `src/rendering/GameRenderer.ts`:

```typescript
const spriteUrls = {
  terrain: './static/assets/sprites/terrain.png',
  characters: './static/assets/sprites/characters.png',
  effects: './static/assets/sprites/effects.png',
  ui: './static/assets/sprites/ui.png',
};
```

### Adding New Asset References

To add new asset categories to the game:

1. **Update GameRenderer.ts**:
   ```typescript
   const spriteUrls = {
     // ... existing
     monsters: './static/assets/sprites/monsters.png',
   };
   ```

2. **Rebuild Frontend**:
   ```bash
   npm run build
   ```

3. **Test Loading**:
   ```bash
   make run
   # Check browser console for asset loading
   ```

## Conclusion

This asset generation pipeline provides a complete solution for creating all visual assets needed by the GoldBox RPG Engine. The pipeline is:

- **Automated**: Run scripts to generate hundreds of assets
- **Reproducible**: Use seed values for consistent results
- **Customizable**: Edit YAML to adjust prompts and structure
- **Integrated**: Works with existing build system
- **Documented**: Comprehensive guides and troubleshooting

For additional help or to report issues, see:
- [ASSET_ANALYSIS.md](./ASSET_ANALYSIS.md) - Detailed codebase analysis
- [game-assets.yaml](./game-assets.yaml) - Asset pipeline configuration
- Project Issues: https://github.com/opd-ai/goldbox-rpg/issues

---

**Happy Asset Generating!** 🎮🎨
