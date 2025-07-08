# TypeScript Migration Plan for GoldBox RPG Engine

## Overview

This document outlines the strategic plan for migrating the GoldBox RPG Engine's JavaScript client code to TypeScript for enhanced type safety and development experience.

## Migration Benefits

### Immediate Benefits
- **Compile-time Error Detection**: Catch type-related errors before runtime
- **Enhanced IDE Support**: Better autocomplete, refactoring, and navigation
- **Self-Documenting Code**: Type annotations serve as documentation
- **Improved Maintainability**: Easier to understand and modify complex code

### Long-term Benefits
- **Reduced Debugging Time**: Fewer type-related runtime errors
- **Better Refactoring Safety**: Compiler ensures type consistency during changes
- **Team Development**: Easier onboarding with explicit type contracts
- **API Evolution**: Safe interface changes with compiler validation

## Phased Migration Strategy

### Phase 1: Foundation Setup (Week 1)
**Goal**: Establish TypeScript tooling and build pipeline

#### Tasks:
1. **Install TypeScript Dependencies**
   ```bash
   npm install --save-dev typescript @types/node
   npm install --save-dev @typescript-eslint/parser @typescript-eslint/eslint-plugin
   ```

2. **Create TypeScript Configuration** (`tsconfig.json`)
   ```json
   {
     "compilerOptions": {
       "target": "ES2020",
       "module": "ES2020",
       "moduleResolution": "node",
       "lib": ["ES2020", "DOM"],
       "outDir": "./dist",
       "rootDir": "./web/static/js",
       "strict": true,
       "esModuleInterop": true,
       "skipLibCheck": true,
       "forceConsistentCasingInFileNames": true,
       "declaration": true,
       "declarationMap": true,
       "sourceMap": true
     },
     "include": ["web/static/js/**/*"],
     "exclude": ["node_modules", "dist", "**/*.test.ts"]
   }
   ```

3. **Update Build Pipeline**
   - Add TypeScript compilation to Makefile
   - Configure development server for TypeScript
   - Set up incremental compilation for development

### Phase 2: Type Definitions and Utilities (Week 2)
**Goal**: Create foundational types and convert utility classes

#### Files to Convert:
1. **`error-handler.js` → `error-handler.ts`**
   ```typescript
   interface ErrorMetadata {
     context?: string;
     timestamp?: number;
     userAgent?: string;
     [key: string]: any;
   }

   class ErrorHandler {
     private context: string;
     private eventEmitter?: EventEmitter;
     private userMessageCallback?: (message: string) => void;
     
     constructor(
       context: string, 
       eventEmitter?: EventEmitter, 
       userMessageCallback?: (message: string) => void
     ) { /* implementation */ }
   }
   ```

2. **`logger.js` → `logger.ts`**
   ```typescript
   type LogLevel = 'debug' | 'info' | 'warn' | 'error' | 'group';
   
   interface LoggerConfig {
     isDevelopment: boolean;
     enabledLevels: LogLevel[];
     hostname: string;
     port: string;
   }
   ```

3. **Create Game State Type Definitions** (`types.ts`)
   ```typescript
   interface PlayerState {
     id: string;
     name: string;
     position: Position;
     health: number;
     maxHealth: number;
     level: number;
   }

   interface WorldState {
     map: GameMap;
     players: { [id: string]: PlayerState };
     objects: GameObject[];
   }

   interface CombatState {
     active: boolean;
     currentTurn?: string;
     initiative: InitiativeEntry[];
   }

   interface GameState {
     player: PlayerState | null;
     world: WorldState | null;
     combat: CombatState | null;
   }
   ```

### Phase 3: Core Game Logic (Week 3-4)
**Goal**: Convert main game management classes

#### Priority Order:
1. **EventEmitter Base Class** - Foundation for all other classes
2. **GameState Class** - Central state management
3. **GameRenderer Class** - Rendering system
4. **UIManager Class** - User interface management
5. **CombatManager Class** - Combat system

#### Example TypeScript Conversion:
```typescript
// game.ts
class GameState extends EventEmitter {
  private rpc: RPCClient;
  private player: PlayerState | null = null;
  private world: WorldState | null = null;
  private combat: CombatState | null = null;
  private lastUpdate: number = 0;
  private updateInterval: number = 100;
  private initialized: boolean = false;
  private updating: boolean = false;

  constructor(rpcClient: RPCClient) {
    super();
    this.rpc = rpcClient;
  }

  async initialize(): Promise<void> { /* implementation */ }
  
  async move(direction: MoveDirection): Promise<void> { /* implementation */ }
  
  handleStateUpdate(state: Partial<GameState>): void { /* implementation */ }
}
```

### Phase 4: RPC Client with Enhanced Type Safety (Week 5)
**Goal**: Add comprehensive typing to network communication

#### Features:
1. **Generic RPC Methods**
   ```typescript
   interface RPCRequest<T = any> {
     jsonrpc: '2.0';
     method: string;
     params?: T;
     id: number;
   }

   interface RPCResponse<T = any> {
     jsonrpc: '2.0';
     result?: T;
     error?: RPCError;
     id: number | null;
   }

   class RPCClient extends EventEmitter {
     async request<TParams, TResult>(
       method: string, 
       params?: TParams
     ): Promise<TResult> { /* implementation */ }
   }
   ```

2. **Method-Specific Type Safety**
   ```typescript
   interface MoveParams {
     direction: 'n' | 's' | 'e' | 'w' | 'ne' | 'nw' | 'se' | 'sw';
   }

   interface AttackParams {
     target_id: string;
     weapon_id: string;
   }

   interface SpellParams {
     spell_id: string;
     target_id?: string;
     position?: Position;
   }
   ```

### Phase 5: Testing and Validation (Week 6)
**Goal**: Ensure type safety and fix any issues

#### Tasks:
1. **Compile All Files**: `npx tsc --noEmit` for type checking
2. **Runtime Testing**: Verify all functionality works correctly
3. **Type Coverage Analysis**: Ensure comprehensive typing
4. **Performance Testing**: Verify no performance regressions

## Implementation Guidelines

### Code Style Standards
- Use `interface` for object shapes, `type` for unions/intersections
- Prefer `readonly` for immutable properties
- Use strict null checks (`strictNullChecks: true`)
- Avoid `any` type - use `unknown` or specific types
- Use type guards for runtime type validation

### Error Handling
```typescript
// Type-safe error handling
type RPCError = {
  code: number;
  message: string;
  data?: unknown;
};

function isRPCError(error: unknown): error is RPCError {
  return typeof error === 'object' && 
         error !== null && 
         'code' in error && 
         'message' in error;
}
```

### Event System Typing
```typescript
interface GameEvents {
  'stateChanged': (state: GameState) => void;
  'error': (error: Error) => void;
  'playerMoved': (player: PlayerState) => void;
}

class TypedEventEmitter<T> {
  on<K extends keyof T>(event: K, listener: T[K]): this;
  emit<K extends keyof T>(event: K, ...args: Parameters<T[K]>): boolean;
}
```

## Rollback Strategy

If issues arise during migration:

1. **Incremental Rollback**: Convert problematic files back to JavaScript
2. **Parallel Development**: Maintain JavaScript versions alongside TypeScript
3. **Gradual Adoption**: Use `allowJs: true` for mixed JavaScript/TypeScript
4. **Type Stubs**: Create `.d.ts` files for complex conversions

## Success Metrics

- [ ] Zero TypeScript compilation errors
- [ ] All unit tests pass
- [ ] No runtime regressions
- [ ] Improved development experience (IDE support)
- [ ] Reduced type-related bugs in future development

## Resources and References

- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [TypeScript Migration Guide](https://www.typescriptlang.org/docs/handbook/migrating-from-javascript.html)
- [ESLint TypeScript Integration](https://typescript-eslint.io/)
- [VS Code TypeScript Support](https://code.visualstudio.com/docs/languages/typescript)

---

**Note**: This migration plan is designed for gradual implementation to minimize disruption to ongoing development. Each phase can be implemented independently, allowing for flexible scheduling and risk management.
