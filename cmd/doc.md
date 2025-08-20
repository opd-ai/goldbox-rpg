# Command Line Applications

This directory contains multiple executable applications that demonstrate and provide different aspects of the GoldBox RPG Engine.

## Applications

### server/
**Main Application Server**
- Primary entry point for the GoldBox RPG Engine
- Provides JSON-RPC API and WebSocket support
- Includes health endpoints and metrics monitoring
- Production-ready server with session management

**Usage:**
```bash
make run
# or
go run cmd/server/main.go
```

### dungeon-demo/
**Procedural Dungeon Generation Demo**
- Demonstrates the PCG (Procedural Content Generation) system
- Shows terrain generation algorithms in action
- Useful for testing and visualizing dungeon layouts

**Usage:**
```bash
go run cmd/dungeon-demo/main.go
```

### events-demo/
**Event System Demonstration**
- Shows the event-driven architecture in action
- Demonstrates event emission, handling, and propagation
- Useful for understanding the game's event system

**Usage:**
```bash
go run cmd/events-demo/main.go
```

### metrics-demo/
**Monitoring and Metrics Demo**
- Demonstrates Prometheus metrics integration
- Shows health check endpoints
- Performance monitoring examples

**Usage:**
```bash
go run cmd/metrics-demo/main.go
```

### validator-demo/
**Input Validation Demo**
- Demonstrates the validation framework
- Shows secure input handling patterns
- Examples of preventing injection attacks

**Usage:**
```bash
go run cmd/validator-demo/main.go
```

## Building All Applications

```bash
# Build all demos
for dir in cmd/*/; do
    if [ -f "$dir/main.go" ]; then
        echo "Building $dir..."
        go build -o "bin/$(basename $dir)" "$dir/main.go"
    fi
done
```

## Integration Testing

These applications can be used for integration testing and development:

1. **Server Testing**: Use the main server for API testing
2. **PCG Testing**: Use dungeon-demo to test content generation
3. **Event Testing**: Use events-demo to verify event system functionality
4. **Performance Testing**: Use metrics-demo to verify monitoring systems
5. **Security Testing**: Use validator-demo to test input validation

## Development Workflow

1. Start with the main server: `make run`
2. Use demos to test specific functionality
3. Monitor with metrics-demo for performance
4. Validate inputs with validator-demo for security

Last Updated: 2025-08-20