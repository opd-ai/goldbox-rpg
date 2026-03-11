package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// RPCMethodInfo contains information about an RPC method
type RPCMethodInfo struct {
	Name        string
	ConstName   string
	HandlerFunc string
	Description string
	Group       string
}

// OpenAPIGenerator generates OpenAPI specifications from Go code
type OpenAPIGenerator struct {
	methods      []RPCMethodInfo
	packagePath  string
	specTemplate string
}

func main() {
	packagePath := flag.String("package", "pkg/server", "Package path to analyze")
	specPath := flag.String("spec", "api/openapi.yaml", "Path to OpenAPI spec file")
	flag.Parse()

	gen := &OpenAPIGenerator{
		packagePath: *packagePath,
	}

	if err := gen.Run(*specPath); err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("OpenAPI spec generation complete!")
}

// Run executes the OpenAPI generation process
func (g *OpenAPIGenerator) Run(specPath string) error {
	// Step 1: Extract RPC methods from constants.go
	if err := g.extractMethods(); err != nil {
		return fmt.Errorf("failed to extract methods: %w", err)
	}

	// Step 2: Load existing spec for template/base structure
	existingSpec, err := g.loadExistingSpec(specPath)
	if err != nil {
		log.Printf("Warning: Could not load existing spec: %v", err)
		existingSpec = g.createBaseSpec()
	}

	// Step 3: Update paths with discovered methods
	if err := g.updateSpecPaths(existingSpec); err != nil {
		return fmt.Errorf("failed to update spec paths: %w", err)
	}

	// Step 4: Write updated spec
	if err := g.writeSpec(specPath, existingSpec); err != nil {
		return fmt.Errorf("failed to write spec: %w", err)
	}

	fmt.Printf("Updated OpenAPI spec with %d RPC methods\n", len(g.methods))
	return nil
}

// extractMethods parses the Go source files to extract RPC method information
func (g *OpenAPIGenerator) extractMethods() error {
	fset := token.NewFileSet()

	// Parse constants.go to get method names
	constantsPath := filepath.Join(g.packagePath, "constants.go")
	file, err := parser.ParseFile(fset, constantsPath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse constants.go: %w", err)
	}

	// Extract method constants
	methodMap := make(map[string]*RPCMethodInfo)
	ast.Inspect(file, func(n ast.Node) bool {
		// Look for constant declarations
		if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.CONST {
			for _, spec := range genDecl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					for i, name := range valueSpec.Names {
						// Check if this is an RPC method constant
						if strings.HasPrefix(name.Name, "Method") {
							methodName := ""
							if i < len(valueSpec.Values) {
								if basicLit, ok := valueSpec.Values[i].(*ast.BasicLit); ok {
									methodName = strings.Trim(basicLit.Value, "\"")
								}
							}

							if methodName != "" {
								group := g.categorizeMethod(name.Name)
								methodMap[methodName] = &RPCMethodInfo{
									Name:      methodName,
									ConstName: name.Name,
									Group:     group,
								}
							}
						}
					}
				}
			}
		}
		return true
	})

	// Parse handlers.go to get handler function documentation
	handlersPath := filepath.Join(g.packagePath, "handlers.go")
	handlersFile, err := parser.ParseFile(fset, handlersPath, nil, parser.ParseComments)
	if err == nil {
		ast.Inspect(handlersFile, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				funcName := funcDecl.Name.Name
				if strings.HasPrefix(funcName, "handle") {
					// Extract method name from handler name (e.g., handleMove -> move)
					methodName := strings.TrimPrefix(funcName, "handle")
					methodName = strings.ToLower(methodName[:1]) + methodName[1:]

					if info, exists := methodMap[methodName]; exists {
						info.HandlerFunc = funcName
						// Extract function documentation
						if funcDecl.Doc != nil {
							info.Description = strings.TrimSpace(funcDecl.Doc.Text())
						}
					}
				}
			}
			return true
		})
	}

	// Convert map to sorted slice
	for _, info := range methodMap {
		g.methods = append(g.methods, *info)
	}
	sort.Slice(g.methods, func(i, j int) bool {
		return g.methods[i].Name < g.methods[j].Name
	})

	return nil
}

// categorizeMethod determines the category/group for a method based on its name
func (g *OpenAPIGenerator) categorizeMethod(constName string) string {
	switch {
	case strings.Contains(constName, "Quest"):
		return "Quest Management"
	case strings.Contains(constName, "Spell"):
		return "Spell Management"
	case strings.Contains(constName, "Equip"):
		return "Equipment Management"
	case strings.Contains(constName, "PCG") || strings.Contains(constName, "Generate") || strings.Contains(constName, "Validate"):
		return "Procedural Content Generation"
	case strings.Contains(constName, "Objects") || strings.Contains(constName, "Nearest"):
		return "Spatial Queries"
	case strings.Contains(constName, "Combat") || constName == "MethodAttack" || constName == "MethodEndTurn":
		return "Combat"
	case constName == "MethodMove" || constName == "MethodJoinGame" || constName == "MethodLeaveGame":
		return "Game Actions"
	case constName == "MethodGetGameState" || constName == "MethodCreateCharacter":
		return "Game State"
	case constName == "MethodUseItem":
		return "Item Management"
	default:
		return "General"
	}
}

// loadExistingSpec loads the existing OpenAPI spec file
func (g *OpenAPIGenerator) loadExistingSpec(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var spec map[string]interface{}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	return spec, nil
}

// createBaseSpec creates a minimal base OpenAPI spec structure
func (g *OpenAPIGenerator) createBaseSpec() map[string]interface{} {
	return map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "Gold Box RPG JSON-RPC API",
			"description": "Comprehensive JSON-RPC 2.0 API for the Gold Box RPG game engine",
			"version":     "1.0.0",
		},
		"servers": []interface{}{
			map[string]interface{}{
				"url":         "http://localhost:8080",
				"description": "Development server",
			},
		},
		"paths":      map[string]interface{}{},
		"components": map[string]interface{}{},
	}
}

// updateSpecPaths updates the paths section of the OpenAPI spec with discovered methods
func (g *OpenAPIGenerator) updateSpecPaths(spec map[string]interface{}) error {
	// Ensure paths exist
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		paths = make(map[string]interface{})
		spec["paths"] = paths
	}

	// Get or create the /rpc endpoint
	rpcPath, ok := paths["/rpc"].(map[string]interface{})
	if !ok {
		rpcPath = map[string]interface{}{
			"post": map[string]interface{}{
				"summary":     "JSON-RPC 2.0 Endpoint",
				"description": "Main endpoint for all JSON-RPC method calls",
			},
		}
		paths["/rpc"] = rpcPath
	}

	// Ensure components exist
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		components = make(map[string]interface{})
		spec["components"] = components
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		schemas = make(map[string]interface{})
		components["schemas"] = schemas
	}

	// Add metadata about methods
	methodList := make([]string, 0, len(g.methods))
	methodsByGroup := make(map[string][]string)

	for _, method := range g.methods {
		methodList = append(methodList, method.Name)
		methodsByGroup[method.Group] = append(methodsByGroup[method.Group], method.Name)
	}

	// Add custom metadata to spec
	spec["x-rpc-methods"] = methodList
	spec["x-rpc-method-groups"] = methodsByGroup
	spec["x-rpc-base-url"] = "/rpc"
	spec["x-rpc-protocol"] = "JSON-RPC 2.0"

	return nil
}

// writeSpec writes the OpenAPI spec to a file
func (g *OpenAPIGenerator) writeSpec(path string, spec map[string]interface{}) error {
	data, err := yaml.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to marshal spec: %w", err)
	}

	// Add header comment
	header := []byte(`# OpenAPI Specification for Gold Box RPG JSON-RPC API
# 
# This file is partially auto-generated from Go source code.
# The x-rpc-methods and x-rpc-method-groups sections are generated automatically.
# To regenerate: make openapi-gen
#
# Manual edits to request/response schemas and descriptions are preserved.
#
`)

	fullData := append(header, data...)

	if err := os.WriteFile(path, fullData, 0o644); err != nil {
		return fmt.Errorf("failed to write spec file: %w", err)
	}

	return nil
}
