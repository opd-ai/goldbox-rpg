#!/bin/bash
# Asset Generation Setup Script for GoldBox RPG Engine
# 
# This script helps team members set up asset-generator for the project
#
# Usage:
#   ./scripts/assets/setup.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 GoldBox RPG Asset Generation Setup${NC}"
echo "===================================="
echo ""

# Check if asset-generator is already installed
if command -v asset-generator &> /dev/null; then
    echo -e "${GREEN}✅ asset-generator is already installed${NC}"
    asset-generator --version
    echo ""
else
    echo -e "${YELLOW}📥 Installing asset-generator...${NC}"
    
    # Detect OS and architecture
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac
    
    # Construct download URL
    BINARY_NAME="asset-generator-${OS}-${ARCH}"
    DOWNLOAD_URL="https://github.com/opd-ai/asset-generator/releases/latest/download/${BINARY_NAME}"
    
    echo "Downloading: $DOWNLOAD_URL"
    
    # Download and install
    if curl -sSL "$DOWNLOAD_URL" -o asset-generator; then
        chmod +x asset-generator
        sudo mv asset-generator /usr/local/bin/
        echo -e "${GREEN}✅ asset-generator installed successfully${NC}"
        asset-generator --version
    else
        echo -e "${RED}❌ Failed to download asset-generator${NC}"
        echo "Please download manually from: https://github.com/opd-ai/asset-generator/releases"
        exit 1
    fi
    echo ""
fi

# Initialize user configuration if it doesn't exist
echo -e "${BLUE}📝 Configuring asset-generator...${NC}"
if [ ! -f ~/.asset-generator/config.yaml ]; then
    echo "Initializing user configuration..."
    asset-generator config init
fi

# Prompt for API URL
echo ""
read -p "Enter SwarmUI API URL [http://localhost:7801]: " api_url
api_url=${api_url:-http://localhost:7801}
asset-generator config set api-url "$api_url"

# Test API connectivity
echo ""
echo -e "${BLUE}🌐 Testing API connectivity...${NC}"
if curl -s --max-time 5 "$api_url" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ API is accessible${NC}"
else
    echo -e "${YELLOW}⚠️  Warning: Cannot connect to $api_url${NC}"
    echo "   Make sure SwarmUI is running and accessible"
fi

# Optional API key
echo ""
read -p "Enter API key (leave blank if none): " api_key
if [ -n "$api_key" ]; then
    asset-generator config set api-key "$api_key"
    echo -e "${GREEN}✅ API key configured${NC}"
fi

# Create output directory
echo ""
echo -e "${BLUE}📁 Creating output directory...${NC}"
mkdir -p output
echo -e "${GREEN}✅ Output directory created: ./output/${NC}"

echo ""
echo -e "${GREEN}🎉 Setup complete!${NC}"
echo ""
echo -e "${BLUE}📋 Next steps:${NC}"
echo "   1. Review pipeline files in ./assets/"
echo "   2. Run 'make assets-preview' to see what will be generated"
echo "   3. Run 'make assets' to generate all assets"
echo "   4. Or use './scripts/assets/generate-all.sh' directly"
echo ""
echo -e "${BLUE}📖 Documentation:${NC}"
echo "   - Asset pipeline: ./assets/README.md"
echo "   - Project overview: ./README.md"
echo "   - Troubleshooting: Check SwarmUI logs if generation fails"