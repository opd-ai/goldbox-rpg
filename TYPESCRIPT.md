# TypeScript Migration Plan for GoldBox RPG Engine

**Migration Target**: JavaScript frontend codebase only (Go backend excluded)  
**Timeline**: 4-week phased approach with immediate benefits  
**Goal**: Improve type safety, developer experience, and code organization

## Current State Assessment

- **Core JS Files**: 7 main modules (rpc.js, game.js, render.js, ui.js, combat.js, spatial.js, error-handler.js, logger.js)
- **Test Files**: 9 test modules (test-*.js pattern)
- **Existing TypeScript Setup**: tsconfig.json configured, types.d.ts foundation available
- **Dependencies**: Minimal external dependencies, clean EventEmitter architecture

## Migration Strategy: Gradual with Reorganization

### Phase 1: Foundation & Utilities (Week 1)
**Target**: Infrastructure and shared utilities

```bash
# 1.1 Setup build pipeline
npm install --save-dev typescript @types/node
npm install --save-dev esbuild concurrently

# 1.2 Migrate core utilities (small, isolated modules)
web/static/js/logger.js → src/utils/Logger.ts
web/static/js/error-handler.js → src/utils/ErrorHandler.ts
web/static/js/spatial.js → src/utils/SpatialIndex.ts
```

**Reorganization Benefits**: Move utilities to dedicated `src/utils/` directory for better organization

### Phase 2: Type Definitions & Interfaces (Week 2)
**Target**: Expand type system and convert base classes

```bash
# 2.1 Enhanced type definitions
types.d.ts → src/types/GameTypes.ts (expand existing definitions)
           → src/types/RPCTypes.ts (JSON-RPC 2.0 interfaces)
           → src/types/UITypes.ts (DOM and event interfaces)

# 2.2 Base classes conversion
src/core/EventEmitter.ts (extract from existing modules)
src/core/BaseComponent.ts (shared component lifecycle)
```

**Key Types to Define**:
- `RPCRequest`, `RPCResponse`, `RPCError` (strict JSON-RPC 2.0)
- `GameState`, `PlayerState`, `CombatState` (game domain)
- `UIComponent`, `RenderLayer` (frontend architecture)

### Phase 3: Core Game Logic (Week 3)
**Target**: Main game modules with dependency injection

```bash
# 3.1 Core game systems
web/static/js/game.js → src/game/GameState.ts
web/static/js/combat.js → src/game/CombatManager.ts
web/static/js/render.js → src/rendering/GameRenderer.ts

# 3.2 Improved architecture
src/game/StateManager.ts (centralized state management)
src/services/RPCService.ts (dependency-injected RPC client)
```

**Architectural Improvements**:
- Dependency injection for better testability
- Centralized state management with type-safe mutations
- Clear separation between game logic and rendering

### Phase 4: RPC Client & UI Integration (Week 4)
**Target**: Network layer and user interface

```bash
# 4.1 Network layer
web/static/js/rpc.js → src/network/RPCClient.ts
                    → src/network/WebSocketManager.ts

# 4.2 UI components
web/static/js/ui.js → src/ui/UIManager.ts
                   → src/ui/components/ (component modules)

# 4.3 Integration
src/main.ts (application entry point)
```

## Build Pipeline Configuration

```json
// package.json scripts
{
  "scripts": {
    "build": "tsc && esbuild src/main.ts --bundle --outfile=web/static/js/app.js",
    "watch": "concurrently \"tsc --watch\" \"esbuild src/main.ts --bundle --outfile=web/static/js/app.js --watch\"",
    "typecheck": "tsc --noEmit",
    "migrate": "node scripts/js-to-ts-converter.js"
  }
}
```

## Code Organization Improvements

### New Directory Structure
```
src/
├── types/           # Type definitions and interfaces
├── utils/           # Pure utility functions and classes
├── core/            # Base classes and shared infrastructure
├── game/            # Game logic and state management
├── rendering/       # Canvas rendering and graphics
├── network/         # RPC client and WebSocket management
├── ui/              # UI components and DOM management
├── services/        # Business logic services
└── main.ts          # Application entry point
```

### Key Architectural Patterns

1. **Dependency Injection**: Constructor injection for services and dependencies
2. **Type-Safe State Management**: Immutable state updates with TypeScript validation
3. **Component Lifecycle**: Standardized initialization, update, and cleanup patterns
4. **Service Layer**: Clear separation between UI, game logic, and network communication

## Migration Conversion Rules

### Type Safety Priorities
```typescript
// 1. Strict function signatures
function validateMove(direction: Direction, position: Position): boolean

// 2. Discriminated unions for game states
type GameState = 'menu' | 'playing' | 'combat' | 'paused'

// 3. Generic RPC methods
async rpcCall<T>(method: string, params: object): Promise<T>

// 4. Immutable state updates
updateGameState(state: Readonly<GameState>, changes: Partial<GameState>): GameState
```

### Backward Compatibility
- Maintain existing HTML integration during migration
- Preserve existing WebSocket message formats
- Keep current RPC method signatures unchanged
- Gradual conversion allows testing at each phase

## Testing Strategy

### Type-Safe Testing
```typescript
// Convert existing test files to TypeScript with proper typing
test-rpc-validation.js → __tests__/network/RPCValidation.test.ts
test-error-handling.js → __tests__/utils/ErrorHandler.test.ts
```

### Integration Testing
- Maintain existing functional tests during migration
- Add type checking to CI/CD pipeline
- Verify WebSocket integration with TypeScript client

## Success Metrics

- **Type Safety**: 100% TypeScript coverage for migrated modules
- **Build Integration**: Zero-configuration TypeScript compilation
- **Developer Experience**: IDE autocomplete and error detection
- **Runtime Compatibility**: No breaking changes to existing functionality
- **Code Organization**: Improved module separation and dependency management

## Risk Mitigation

- **Gradual Migration**: Each phase independently testable and deployable
- **Backward Compatibility**: Existing JavaScript continues working during migration
- **Rollback Plan**: Each phase can be reverted if issues arise
- **Testing Coverage**: Comprehensive testing at each migration phase

## Expected Benefits

1. **Immediate**: Type checking catches errors at compile time
2. **Short-term**: Better IDE support and developer productivity
3. **Long-term**: Improved maintainability and refactoring safety
4. **Architecture**: Cleaner separation of concerns and better testability

This migration plan transforms the codebase into a modern, type-safe, and well-organized TypeScript application while maintaining full backward compatibility and improving overall architecture.
