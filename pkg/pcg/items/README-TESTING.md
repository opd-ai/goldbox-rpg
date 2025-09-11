# PCG Items Testing Guide

## Dynamic Path Resolution for Template Files

The PCG items package includes integration tests that load YAML template files from the project's data directory. To ensure these tests work across different development environments and CI systems, we use dynamic path resolution instead of hardcoded absolute paths.

### Implementation Pattern

```go
// Get the project root directory dynamically
_, filename, _, ok := runtime.Caller(0)
if !ok {
    t.Fatal("Failed to get current file path")
}
projectRoot := filepath.Join(filepath.Dir(filename), "..", "..", "..")
templatesPath := filepath.Join(projectRoot, "data", "pcg", "items", "templates.yaml")

// Load the template file using the resolved path
err := registry.LoadFromFile(templatesPath)
```

### Why This Approach?

1. **Environment Independence**: Tests work in any development environment without modification
2. **CI Compatibility**: No hardcoded paths that break in containerized CI environments  
3. **Cross-Platform**: Uses `filepath.Join()` for proper path construction on all operating systems
4. **Maintainability**: Automatically adapts if project structure changes
5. **Go Best Practices**: Uses standard library functions (`runtime.Caller`, `filepath.Join`)

### Files Using This Pattern

- `example_templates_test.go` - Tests loading of custom YAML templates
- `generator_integration_test.go` - Integration test for template-based generation

### Template File Location

The template files are located at:
```
/workspaces/goldbox-rpg/data/pcg/items/templates.yaml
```

This file contains example custom item templates used by the integration tests to verify that:
- Custom templates can be loaded from external YAML files
- Generated items inherit properties from custom templates
- Template validation works correctly

### Error Handling

The dynamic path resolution includes proper error handling:
- Checks if `runtime.Caller(0)` succeeds
- Validates file existence before attempting to load
- Provides descriptive error messages for debugging

### Performance Impact

The dynamic path resolution has minimal overhead (<1ms) and only affects test execution, not production code performance.
