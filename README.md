# GoldBox RPG Engine

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.20-blue)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![Last Updated](https://img.shields.io/badge/last%20updated-2025--01-blue)

A modern, Go-based RPG engine inspired by the classic SSI Gold Box series of role-playing games. This engine provides a comprehensive framework for creating and managing turn-based RPG games with robust combat systems, character management, and world interactions.

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
  - Advanced spatial indexing
  - Object and NPC management

### Event System
- **Event-Driven Architecture**
  - Combat events
  - Quest updates
  - Item interactions
  - Spell casting
  - Level progression

## 🚀 Getting Started

### Prerequisites
- Go 1.20 or higher
- Make (for build automation)

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

### Running Tests

```bash
make test
```

## 📖 Project Structure

```
goldbox-rpg/
├── cmd/
│   └── server/      # Server entry point
├── pkg/
│   ├── game/       # Core game mechanics and systems
│   └── server/     # Server implementation
├── internal/       # Internal packages
└── test/          # Test suites
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

Last Updated: 2025-01-17