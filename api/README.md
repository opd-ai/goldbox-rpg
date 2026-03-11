# Gold Box RPG API Specification

This directory contains the OpenAPI 3.0 specification for the Gold Box RPG JSON-RPC API.

## Files

- `openapi.yaml` - Complete OpenAPI 3.0 specification covering all JSON-RPC methods

## Viewing the Documentation

The API documentation is available through Swagger UI when the server is running:

```bash
# Start the server
./server

# Access Swagger UI
http://localhost:8080/api/docs
```

## API Categories

The API is organized into the following categories:

### Core Game Methods
- **Character Actions**: move, attack, castSpell, useItem  
- **Combat Management**: startCombat, endTurn
- **Game State**: joinGame, leaveGame, getGameState

### Equipment and Inventory
- **Equipment**: equipItem, unequipItem, getEquipment
- **Item Management**: Item use and inventory operations

### Quest System
- **Quest Management**: startQuest, completeQuest, failQuest
- **Quest Queries**: getQuest, getActiveQuests, getQuestLog

### Spell System
- **Spell Queries**: getSpell, getSpellsByLevel, getSpellsBySchool
- **Spell Search**: getAllSpells, searchSpells

### Spatial Operations
- **Object Queries**: getObjectsInRange, getObjectsInRadius, getNearestObjects
- **Position-based Searches**: Efficient spatial indexing support

### Procedural Content Generation (PCG)
- **Content Generation**: generateContent, generateLevel, generateQuest
- **Terrain Generation**: regenerateTerrain with biome support
- **Item Generation**: generateItems with rarity and level scaling
- **PCG Management**: getPCGStats, validateContent

## OpenAPI Spec Details

The specification follows OpenAPI 3.0 standard and includes:

- Complete request/response schemas for all methods
- JSON-RPC 2.0 envelope structure
- Parameter validation rules
- Example requests and responses
- Error response formats

## Generating Client SDKs

You can use OpenAPI code generators to create client SDKs:

```bash
# TypeScript/JavaScript
npx @openapitools/openapi-generator-cli generate \
  -i api/openapi.yaml \
  -g typescript-fetch \
  -o clients/typescript

# Python
openapi-generator generate \
  -i api/openapi.yaml \
  -g python \
  -o clients/python

# Go
openapi-generator generate \
  -i api/openapi.yaml \
  -g go \
  -o clients/go
```

## Maintaining the Specification

When adding new RPC methods or modifying existing ones:

1. Update the OpenAPI specification in `openapi.yaml`
2. Update the corresponding method documentation in `pkg/README-RPC.md`
3. Test the changes through Swagger UI
4. Regenerate client SDKs if needed

## Additional Resources

- [Full RPC Documentation](../pkg/README-RPC.md)
- [OpenAPI Specification](https://spec.openapis.org/oas/v3.0.0)
- [Swagger UI Documentation](https://swagger.io/tools/swagger-ui/)
