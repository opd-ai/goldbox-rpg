# GoldBox RPG Engine

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.23.0-blue)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![Last Updated](https://img.shields.io/badge/last%20updated-2025--08--20-blue)

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
  - ✅ Advanced spatial indexing (R-tree-like structure for efficient queries)
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

## 🚀 Getting Started

### Prerequisites
- Go 1.23.0 or higher
- Node.js 18+ and npm (for frontend development)
- Make (for build automation)
- **Docker** (recommended for easy setup)

### Installation

```bash
# Clone the repository
git clone https://github.com/opd-ai/goldbox-rpg.git

# Navigate to the project directory
cd goldbox-rpg

# Install dependencies
go mod download

# Install frontend dependencies
npm install

# Build the project
make build
```

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

# In another terminal, start the frontend development server
npm run watch

# Access the application at http://localhost:8080
```

### Running Tests

```bash
# Run Go backend tests
make test

# Run Go tests with coverage
make test-coverage

# Run TypeScript type checking
npm run typecheck
```
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
├── src/               # TypeScript frontend source
├── web/               # Web assets and static files
├── data/              # Game data (spells, items, PCG templates)
├── scripts/           # Build and utility scripts
└── test/              # Integration tests
```

For complete API documentation, see [`pkg/README-RPC.md`](pkg/README-RPC.md) which includes all available JSON-RPC methods, parameters, and examples.

### Frontend Architecture

```
src/
├── core/           # Base components and infrastructure
├── game/           # Game logic and state management
├── network/        # RPC client and WebSocket management
├── ui/             # User interface components
├── utils/          # Utility functions and helpers
└── types/          # TypeScript type definitions
```
## 🛠️ Technical Details

### Technology Stack
- **Backend**: Go 1.23.0+ with native HTTP server
- **Protocol**: JSON-RPC 2.0 over HTTP and WebSockets
- **Dependencies**: 
  - Gorilla WebSocket v1.5.3 for real-time communication
  - Sirupsen Logrus v1.9.3 for structured logging
  - Prometheus client v1.22.0 for metrics collection
  - YAML v3.0.1 for configuration management
- **Frontend**: TypeScript with ES2020 target and ESBuild bundling
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

### Frontend (src/)
- TypeScript-based client architecture
- Component-based UI system
- Real-time state synchronization
- Canvas-based game rendering
- Event-driven communication

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
- [ ] Advanced NPC AI behaviors
- [ ] Enhanced combat mechanics
- [ ] Additional spell effects
- [ ] World editor tools
- [ ] Network optimization
- [ ] Content creation utilities
- [ ] Player progression persistence
- [ ] Guild and faction systems

Last Updated: 2025-08-20