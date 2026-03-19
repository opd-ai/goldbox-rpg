# GoldBox RPG Engine

[![CI](https://github.com/opd-ai/goldbox-rpg/actions/workflows/ci.yml/badge.svg)](https://github.com/opd-ai/goldbox-rpg/actions/workflows/ci.yml)
[![Build](https://github.com/opd-ai/goldbox-rpg/actions/workflows/build.yml/badge.svg)](https://github.com/opd-ai/goldbox-rpg/actions/workflows/build.yml)
![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.25.6-blue)
![Coverage](https://img.shields.io/badge/coverage-65--96%25-brightgreen)
![Assets](https://img.shields.io/badge/assets-521%20ready%20%2F%20521%20defined-brightgreen)
![Last Updated](https://img.shields.io/badge/last%20updated-2026--03--13-blue)

A modern, Go-based RPG engine inspired by the classic SSI Gold Box series of role-playing games. This engine provides a comprehensive framework for creating and managing turn-based RPG games with robust combat systems, character management, and world interactions through a JSON-RPC API with WebSocket support for real-time communication.

## 🎮 Features

### Core Game Systems
- **Character Management**
  - Six core attributes: Strength, Dexterity, Constitution, Intelligence, Wisdom, Charisma
  - Class-based system (Fighter, Mage, Cleric, Thief, Ranger, Paladin)
  - Multiple character creation methods: roll, standard array, point-buy, custom
  - Equipment and inventory management with class proficiency restrictions
  - Experience and level progression with automatic stat calculations

### Combat & Effects
- **Comprehensive Effect System**
  - Status effects (Damage over Time, Healing over Time)
  - Combat conditions (Stun, Root, Burning, Bleeding, Poison)
  - Stat modifications (Boosts and Penalties)
  - Effect stacking and priority management
  - Immunity and resistance handling

### World Management
- **Dynamic World System**
  - Tile-based environments with multiple terrain types
  - Multiple damage types (Physical, Fire, Poison, Frost, Lightning)
  - ✅ Advanced spatial indexing (Quadtree structure for efficient queries)
  - Object and NPC management with procedural generation
  - Combat positioning and line-of-sight calculations

### Event System
- **Event-Driven Architecture**
  - Combat events
  - Quest updates
  - Item interactions
  - Spell casting
  - Level progression

### Real-time Communication
- **WebSocket Integration**
  - Live game state updates
  - Real-time event broadcasting
  - Session-based multiplayer support
  - Concurrent player management

### Monitoring & Observability
- **Health Check Endpoints**
  - `/health` - Comprehensive health status with detailed checks
  - `/ready` - Kubernetes-style readiness probe
  - `/live` - Basic liveness probe for load balancers
- **Metrics Integration**
  - Prometheus metrics endpoint at `/metrics`
  - Request/response monitoring
  - Session and performance tracking
  - Memory and goroutine monitoring

### Procedural Content Generation
- **Dynamic Content Creation**
  - Terrain generation with biome-aware algorithms
  - Item generation using template-based systems
  - Quest generation with objectives and rewards
  - NPC generation with personalities and motivations
  - Deterministic seeding for reproducible content
  - Validation system for generated content integrity

### System Resilience
- **Circuit Breaker Patterns**
  - Protection against cascade failures
  - Automatic recovery mechanisms
  - Configurable failure thresholds
- **Retry Mechanisms**
  - Exponential backoff strategies
  - Transient failure handling
  - Customizable retry policies
- **Input Validation**
  - Comprehensive JSON-RPC parameter validation
  - Security against injection attacks
  - Request size limiting for DoS prevention

### Asset Generation Pipeline
- **Automated Asset Creation**
  - Complete pipeline for generating 521 game assets
  - Character portraits, monster sprites, item icons
  - Terrain tiles, combat effects, UI elements
  - YAML-based configuration with hierarchical structure
  - Reproducible generation with seed-based randomization
- **Generation Scripts**
  - Full asset generation (`make assets`)
  - Priority asset generation for quick testing
  - Post-processing optimization tools
  - Asset verification and validation
- **Comprehensive Documentation**
  - Detailed codebase analysis ([ASSET_ANALYSIS.md](./ASSET_ANALYSIS.md))
  - Complete integration guide ([ASSET_INTEGRATION.md](./ASSET_INTEGRATION.md))
  - See [Asset Generation](#-asset-generation) section below

## 📸 Screenshots

Captured automatically by headless browser playtests and published with each [nightly build](https://github.com/opd-ai/goldbox-rpg/releases/tag/nightly).

| Splash Screen | Main Menu |
|:---:|:---:|
| ![Splash Screen](https://github.com/opd-ai/goldbox-rpg/releases/download/nightly/screenshot-splash-screen.png) | ![Main Menu](https://github.com/opd-ai/goldbox-rpg/releases/download/nightly/screenshot-main-menu.png) |

| Character Creation | Gameplay |
|:---:|:---:|
| ![Character Creation](https://github.com/opd-ai/goldbox-rpg/releases/download/nightly/screenshot-character-creation.png) | ![Gameplay](https://github.com/opd-ai/goldbox-rpg/releases/download/nightly/screenshot-gameplay.png) |

## 🚀 Getting Started

### Prerequisites
- Go 1.25.6 or higher (toolchain 1.25.8)
- Make (for build automation)
- **Docker** (recommended for easy setup)
- **Asset Generation Tool** (optional - Stable Diffusion, DALL-E) - See [Asset Generation](#-asset-generation) section and [ASSET_INTEGRATION.md](./ASSET_INTEGRATION.md) for setup details. Production-ready sprite assets (521 PNG files) are already included in the repository.

### Installation

```bash
# Clone the repository
git clone https://github.com/opd-ai/goldbox-rpg.git

# Navigate to the project directory
cd goldbox-rpg

# Install dependencies
go mod download

# Build the project
make build
```

**✅ Asset Status:** The repository includes 521 production-ready sprite assets in `web/static/assets/sprites/`. The game is **fully functional** with these assets. Custom asset generation (for alternative art styles) requires an external AI image generation tool—see [Asset Generation](#-asset-generation) and [ASSET_INTEGRATION.md](./ASSET_INTEGRATION.md) for details.

### Running with Docker (Recommended)

The easiest way to run the GoldBox RPG Engine is using Docker:

```bash
# Build and run (that's it!)
docker build -t goldbox-rpg .
docker run -p 8080:8080 goldbox-rpg

# Open http://localhost:8080 in your browser and play!
```

The Docker container includes automatic health checks. You can verify the server status:

```bash
# Check health status
curl http://localhost:8080/health

# Check readiness (for load balancers)
curl http://localhost:8080/ready

# View metrics (Prometheus format)
curl http://localhost:8080/metrics
```

### Running Locally

For local development without Docker:

```bash
# Start the Go backend
make run

# In another terminal, build the WASM frontend
make wasm

# Access the application at http://localhost:8080
```

### Running Tests

```bash
# Run Go backend tests
make test

# Run Go tests with coverage
make test-coverage

# Run race detector
go test -race ./...
```

### OpenAPI Spec Generation

The project includes automatic OpenAPI specification generation from Go source code:

```bash
# Generate OpenAPI spec from RPC method constants
make openapi-gen

# Validate the generated spec (requires npx)
make openapi-validate
```

The generator parses `pkg/server/constants.go` to extract all RPC methods and updates `api/openapi.yaml` with:
- Complete list of available RPC methods
- Method categorization by feature group
- Automatic sync with code changes

**Note:** The generator preserves manual edits to request/response schemas while keeping method lists current.

### Asset Generation

The GoldBox RPG Engine includes a comprehensive asset generation pipeline for creating all visual assets:

**Asset Availability:**
- **Quick Start (Recommended)**: `make assets-download` - Download 500+ pre-generated assets from GitHub releases (~50MB, 30 seconds)
- **Priority Assets**: `make assets-priority` - Generate 50 high-priority assets (~30 minutes, requires AI tool)
- **Full Generation**: `make assets` - Generate all 521 assets from scratch (~4-6 hours, requires AI tool setup per [ASSET_INTEGRATION.md](./ASSET_INTEGRATION.md))

```bash
# Download pre-generated assets (recommended for quick start)
make assets-download

# Preview what assets would be generated (dry-run)
make assets-preview

# Generate priority assets for quick testing (~30 minutes)
make assets-priority

# Generate all game assets (~4-6 hours)
make assets

# Optimize generated assets for production
make assets-optimize

# Verify all required assets are present
make assets-verify

# Clean generated assets
make assets-clean
```

**Asset Pipeline Features:**
- 521 total assets across 6 categories (characters, monsters, items, terrain, effects, UI)
- YAML-based configuration ([game-assets.yaml](./game-assets.yaml))
- Reproducible generation with seed values
- Hierarchical organization with metadata cascading
- Detailed prompts for consistent art style
- Pre-generated asset packs available via GitHub releases

**Documentation:**
- [ASSET_ANALYSIS.md](./ASSET_ANALYSIS.md) - Complete codebase analysis for asset requirements
- [ASSET_INTEGRATION.md](./ASSET_INTEGRATION.md) - Comprehensive integration and usage guide

**Important Notes:**
- **Ready to Use**: 521 production-ready sprite assets are already included in the repository—no download or generation needed for development or deployment.
- **Custom Assets**: To create alternative art styles, an external AI image generation tool (Stable Diffusion, DALL-E) is required. See [ASSET_INTEGRATION.md](./ASSET_INTEGRATION.md) for setup.
- **Time Commitment:** Full custom asset generation takes 4-6 hours and processes 521 assets across 6 categories.

### Production Deployment

For production deployments, configure the following environment variables for security:

```bash
# Required for production WebSocket origin validation
export WEBSOCKET_ALLOWED_ORIGINS="https://yourdomain.com,https://www.yourdomain.com"
# Alternative: ALLOWED_ORIGINS for configuration-based origin validation

# Example production configuration
export GOLDBOX_PORT=8080
export GOLDBOX_LOG_LEVEL=warn
```

**Important:** The WebSocket origin validation is automatically enabled in production mode. Make sure to set `WEBSOCKET_ALLOWED_ORIGINS` to include all legitimate client domains to prevent unauthorized cross-origin connections.

## 📖 Project Structure

```
goldbox-rpg/
├── cmd/
│   ├── server/         # Main server entry point
│   ├── dungeon-demo/   # Dungeon generation demo
│   ├── events-demo/    # Event system demo
│   ├── metrics-demo/   # Metrics monitoring demo
│   └── validator-demo/ # Input validation demo
├── pkg/
│   ├── game/          # Core game mechanics and systems
│   ├── server/        # Server implementation
│   ├── pcg/           # Procedural Content Generation
│   ├── resilience/    # Circuit breaker patterns
│   ├── validation/    # Input validation framework
│   ├── retry/         # Retry mechanisms
│   ├── integration/   # Integration utilities
│   ├── config/        # Configuration management
│   └── README-RPC.md  # Complete JSON-RPC API documentation
├── web/               # Web UI (splash screen + Ebitengine/WASM)
├── data/              # Game data (spells, items, PCG templates)
├── scripts/           # Build and utility scripts
└── test/              # Integration tests
```

For complete API documentation, see [`pkg/README-RPC.md`](pkg/README-RPC.md) which includes all available JSON-RPC methods, parameters, and examples.

### Frontend Architecture

The frontend is an **Ebitengine/WASM** client compiled from Go. The HTML page at
`web/index.html` shows a splash screen, then loads the WASM binary and hands
control to Ebitengine.

```
pkg/wasmui/
├── game.go              # Ebitengine Game impl (Update/Draw/Layout)
├── rpc_client_wasm.go   # WebSocket JSON-RPC 2.0 client
├── types.go             # Shared game-state types
├── stub_native.go       # Stubs for non-WASM builds
├── types_test.go        # Table-driven tests
├── editor.go            # Map editor WASM client
├── map_editor.go        # Map editor UI components
└── quest_editor.go      # Quest editor UI components

cmd/wasm-ui/
└── main.go              # WASM entry point
```

### Browser-Based Content Editors

The GoldBox RPG Engine includes visual editors accessible via web browser:

**Map Editor** - Visual map creation and editing tool
- **URL**: `http://localhost:8080/editor.html`
- **Features**: Tile-based map editing, terrain placement, object positioning, real-time preview
- **Usage**: Start the server with `make run`, navigate to the editor URL, create or load maps
- **Export**: Maps can be saved to YAML format compatible with the adventure system

**Quest Builder** - Visual quest chain creation tool
- **URL**: `http://localhost:8080/quest-builder.html`
- **Features**: Quest objective creation, reward configuration, prerequisite chains, NPC dialogue
- **Usage**: Access via browser after starting the server
- **Export**: Quests saved as YAML files in the adventure format

**Complete Editor Guide**: See [docs/EDITOR_GUIDE.md](./docs/EDITOR_GUIDE.md) for comprehensive documentation including:
- Complete JSON-RPC API reference for all editor endpoints
- Real-time collaboration features via WebSocket
- Quest objective types and reward configuration
- Troubleshooting and best practices
- Example workflows with curl commands

**CLI Tools**: For scripting and automation, command-line tools are available in `cmd/`:
- `map-editor` - CLI map creation and editing
- `quest-builder` - CLI quest definition
- `content-creator` - Spell and item template generator
## 🛠️ Technical Details

### Technology Stack
- **Backend**: Go 1.25.6+ (toolchain 1.25.8) with native HTTP server
- **Protocol**: JSON-RPC 2.0 over HTTP and WebSockets
- **Dependencies**: 
  - Coder WebSocket v1.8.14 (nhooyr.io/websocket fork) for real-time communication
  - gorilla/websocket v1.5.3 retained for E2E test client only
  - Sirupsen Logrus v1.9.4 for structured logging
  - Prometheus client v1.23.2 for metrics collection
  - YAML v3.0.1 for configuration management
- **Frontend**: Ebitengine/WASM (Go compiled to WebAssembly)
- **Deployment**: Docker support with health checks

### Game Package (pkg/game)
- Character and NPC management
- Combat and effect systems
- World state management
- Equipment and inventory systems
- Quest and progression tracking
- Event handling

### Server Package (pkg/server)
- Game state management
- Session handling
- Combat coordination
- Time management
- Event scheduling
- JSON-RPC API endpoints
- WebSocket real-time communication

### Procedural Content Generation (pkg/pcg)
- Terrain generation with biome awareness
- Item generation using template systems
- Quest generation with dynamic objectives
- NPC generation with personalities
- Deterministic seeding for reproducibility
- Content validation before integration

### System Resilience (pkg/resilience, pkg/retry, pkg/validation)
- Circuit breaker patterns for fault tolerance
- Retry mechanisms with exponential backoff
- Comprehensive input validation framework
- Security against injection and DoS attacks
- Integration utilities for robust API endpoints

### Frontend (pkg/wasmui/)
- Ebitengine/WASM-based client architecture (Go compiled to WebAssembly)
- Canvas-based game rendering with Ebitengine
- WebSocket JSON-RPC 2.0 client for real-time communication
- Stateful game UI with Update/Draw/Layout lifecycle
- Event-driven communication with backend

## 🤝 Contributing

We welcome contributions! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Development Guidelines
- Follow Go best practices and coding standards
- Include tests for new features
- Update documentation as needed
- Use meaningful commit messages

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by the SSI Gold Box series of games
- Built with Go's robust standard library
- Special thanks to all contributors

## 📞 Contact & Support

For questions and support:
- Open an issue in the GitHub repository
- Contact repository owner: [@opd-ai](https://github.com/opd-ai)

## 🔄 Project Status

This project is under active development. Check the [Issues](../../issues) tab for current tasks and planned features.

## 🚧 Roadmap

- [x] Core RPG mechanics and character system
- [x] Combat and effect systems
- [x] WebSocket real-time communication
- [x] Procedural Content Generation system
- [x] Circuit breaker patterns and resilience
- [x] Comprehensive input validation
- [x] Health monitoring and metrics
- [x] **Asset generation pipeline with 521 defined assets** (pipeline complete, 521 assets ready - full AI art requires external tool setup per [ASSET_INTEGRATION.md](./ASSET_INTEGRATION.md))
- [x] Advanced NPC AI behaviors (A* pathfinding, tactical combat AI, behavior trees)
- [x] Enhanced combat mechanics (opportunity attacks, cover/flanking, morale system)
- [x] Complete spell system (levels 0-9, 60 spells across 10 YAML files)
- [x] World editor tools (CLI + browser-based visual editors at `/editor.html` and `/quest-builder.html`)
- [x] Network optimization (rate limiting, connection pooling, delta compression)
- [x] Content creation utilities (CLI tools + browser-based Map Editor and Quest Builder)
- [x] Player progression persistence
- [x] Guild and faction systems with full mechanics (ranks, permissions, treasury, perks)
- [x] **Embedded Adventures** (10 complete adventure packs with 100 maps, 37 quests, 30+ hours of content)

Last Updated: 2026-03-13