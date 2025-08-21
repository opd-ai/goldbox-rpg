# Test Coverage Analysis Scripts

This directory contains scripts to help analyze Go test coverage and identify source files that lack corresponding test files.

## Scripts

### 1. `find_untested_files.go`
A Go program that identifies Go source files without corresponding test files.

**Usage:**
```bash
go run scripts/find_untested_files.go [directory]
```

**Features:**
- Scans recursively through directories
- Excludes `main.go` files (typically don't need tests)
- Provides sorted output
- Written in Go for consistency with the project

### 2. `find_untested_files.sh`
A simple bash script that performs the same function as the Go version.

**Usage:**
```bash
./scripts/find_untested_files.sh [directory]
```

**Features:**
- Fast execution
- Color-coded output
- No dependencies beyond standard Unix tools
- Excludes `main.go` files

### 3. `analyze_test_coverage.sh`
An advanced analysis script with multiple output formats and detailed statistics.

**Usage:**
```bash
./scripts/analyze_test_coverage.sh [options] [directory]
```

**Options:**
- `-v, --verbose`: Show detailed information including file sizes and package info
- `-j, --json`: Output results in JSON format for scripting/automation
- `-e, --exclude`: Exclude patterns (comma-separated, e.g., "vendor,scripts")
- `-h, --help`: Show help message

**Examples:**
```bash
# Basic analysis
./scripts/analyze_test_coverage.sh

# Verbose output with file details
./scripts/analyze_test_coverage.sh -v

# JSON output for automation
./scripts/analyze_test_coverage.sh -j

# Exclude specific directories
./scripts/analyze_test_coverage.sh -e "vendor,scripts"

# Analyze specific directory
./scripts/analyze_test_coverage.sh ./pkg/game
```

**Features:**
- Comprehensive statistics (total files, coverage percentage)
- Multiple output formats (human-readable, JSON)
- File size and package information in verbose mode
- Exclusion patterns for vendor directories or other unwanted paths
- Color-coded output for better readability

## Makefile Integration

The following Makefile targets are available:

```bash
# Simple list of untested files
make find-untested

# Basic coverage analysis
make test-coverage

# Detailed coverage analysis
make test-coverage-verbose

# JSON output for automation
make test-coverage-json
```

## Go Testing Conventions

These scripts follow Go testing conventions:
- Test files are named `*_test.go`
- Test files should be in the same package as the source files
- `main.go` files are excluded as they typically don't have unit tests
- Each source file `example.go` should have a corresponding `example_test.go`

## Integration with CI/CD

The JSON output format makes these scripts suitable for CI/CD integration:

```bash
# Get coverage percentage for CI thresholds
coverage=$(./scripts/analyze_test_coverage.sh -j | jq '.summary.coverage_percentage')

# Fail CI if coverage is below threshold
if [ "$coverage" -lt 80 ]; then
    echo "Test coverage below 80%: $coverage%"
    exit 1
fi
```

## Current Project Status

As of the latest analysis (August 2025), the GoldBox RPG Engine has:
- **78% test coverage** (72 out of 92 source files have tests)
- **20 files** without test coverage
- **Enhanced system resilience** with circuit breaker patterns
- **Comprehensive input validation** for security
- **Procedural Content Generation** system
- **Multiple demo applications** showcasing different features
- **Integration utilities** combining validation and resilience patterns

The testing framework covers:
- Core game mechanics (`pkg/game/`) - Most files tested
- Server functionality (`pkg/server/`) - Partially tested
- PCG systems (`pkg/pcg/`) - Basic test coverage
- Resilience patterns (`pkg/resilience/`, `pkg/retry/`) - Well tested
- Validation framework (`pkg/validation/`) - Well tested
- Integration utilities (`pkg/integration/`) - Well tested

Major untested files include:
- `pkg/game/character.go` - Core character system
- `pkg/server/handlers.go` - Main API handlers  
- `pkg/server/combat.go` - Combat system
- `pkg/pcg/manager.go` - PCG coordination
- `pkg/resilience/manager.go` - Circuit breaker management

## Recommendations

1. **Prioritize integration tests**: Focus on testing the interaction between PCG, validation, and resilience systems
2. **Test edge cases**: Particularly important for validation and circuit breaker logic
3. **Performance testing**: Add benchmarks for PCG algorithms and spatial indexing
4. **Regular monitoring**: Run `make test-coverage` regularly to track progress across all packages
5. **CI integration**: Consider adding coverage checks to your continuous integration pipeline
6. **Demo applications**: Use the demo applications in `cmd/` to validate new features

## Contributing

When adding new Go source files to the project:
1. Always create a corresponding `*_test.go` file
2. Run `make find-untested` to verify your new files have tests
3. Aim for meaningful test coverage, not just file coverage
4. Use table-driven tests for multiple test cases following Go best practices
