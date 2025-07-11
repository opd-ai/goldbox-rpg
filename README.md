# GoldBox RPG Engine

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.22.0-blue)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![Last Updated](https://img.shields.io/badge/last%20updated-2025--07-blue)

A modern, Go-based RPG engine inspired by the classic SSI Gold Box series of role-playing games. This engine provides a comprehensive framework for creating and managing turn-based RPG games with robust combat systems, character management, and world interactions through a JSON-RPC API with WebSocket support for real-time communication.

## 🎮 Features

### Core Game Systems
- **Character Management**
  - Flexible character attributes (Strength, Dexterity, Constitution, Intelligence, Wisdom, Charisma)
  - Class-based system (Fighter, Mage, Cleric, Thief, Ranger, Paladin)
  - Equipment and inventory management
  - Experience and level progression

### Combat & Effects
- **Comprehensive Effect System**
  - Status effects (Damage over Time, Healing over Time)
  - Combat conditions (Stun, Root, Burning, Bleeding, Poison)
  - Stat modifications (Boosts and Penalties)
  - Effect stacking and priority management
  - Immunity and resistance handling

### World Management
- **Dynamic World System**
  - Tile-based environments
  - Multiple damage types (Physical, Fire, Poison, Frost, Lightning)
  - ✅ Advanced spatial indexing (R-tree-like structure for efficient queries)
  - Object and NPC management

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

## 🚀 Getting Started

### Prerequisites
- Go 1.22.0 or higher
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

# Run frontend tests
npm test

# Run TypeScript type checking
npm run typecheck
```

## 📖 Project Structure

```
goldbox-rpg/
├── cmd/
│   └── server/      # Server entry point
├── pkg/
│   ├── game/       # Core game mechanics and systems
│   └── server/     # Server implementation
├── src/            # TypeScript frontend source
├── web/            # Web assets and static files
├── data/           # Game data (spells, items)
└── scripts/        # Build and utility scripts
```

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

- [ ] Advanced NPC AI behaviors
- [ ] Enhanced combat mechanics
- [ ] Additional spell effects
- [ ] World editor tools
- [ ] Network optimization
- [ ] Content creation utilities

Last Updated: 2025-07-09