package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseConstDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []string // method names
	}{
		{
			name: "valid const block with multiple methods",
			source: `package server
const (
	MethodMove = "move"
	MethodAttack = "attack"
	MethodCastSpell = "castSpell"
)`,
			expected: []string{"move", "attack", "castSpell"},
		},
		{
			name: "const block with non-Method constants",
			source: `package server
const (
	MethodMove = "move"
	OtherConst = "other"
	MethodAttack = "attack"
)`,
			expected: []string{"move", "attack"},
		},
		{
			name:     "empty const block",
			source:   `package server`,
			expected: []string{},
		},
		{
			name: "const block with no Method prefix",
			source: `package server
const (
	Something = "value"
	Another = "value2"
)`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.source, parser.ParseComments)
			if err != nil {
				t.Fatalf("failed to parse source: %v", err)
			}

			gen := &OpenAPIGenerator{}
			var allMethods []*RPCMethodInfo

			ast.Inspect(file, func(n ast.Node) bool {
				if genDecl, ok := n.(*ast.GenDecl); ok {
					methods := gen.parseConstDeclaration(genDecl)
					allMethods = append(allMethods, methods...)
				}
				return true
			})

			if len(allMethods) != len(tt.expected) {
				t.Errorf("expected %d methods, got %d", len(tt.expected), len(allMethods))
				return
			}

			for i, method := range allMethods {
				if method.Name != tt.expected[i] {
					t.Errorf("method %d: expected %q, got %q", i, tt.expected[i], method.Name)
				}
			}
		})
	}
}

func TestParseHandlerFunc(t *testing.T) {
	tests := []struct {
		name            string
		source          string
		expectedMethod  string
		expectedHandler string
		expectedOK      bool
	}{
		{
			name: "valid handler with documentation",
			source: `package server
// handleMove processes movement requests
func (s *RPCServer) handleMove(params json.RawMessage) (interface{}, error) {
	return nil, nil
}`,
			expectedMethod:  "move",
			expectedHandler: "handleMove",
			expectedOK:      true,
		},
		{
			name: "handler without handle prefix",
			source: `package server
func (s *RPCServer) someOtherFunc() error {
	return nil
}`,
			expectedOK: false,
		},
		{
			name: "handler with camelCase method name",
			source: `package server
func (s *RPCServer) handleGetGameState() error {
	return nil
}`,
			expectedMethod:  "getGameState",
			expectedHandler: "handleGetGameState",
			expectedOK:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.source, parser.ParseComments)
			if err != nil {
				t.Fatalf("failed to parse source: %v", err)
			}

			gen := &OpenAPIGenerator{}
			var methodName, handlerFunc, description string
			var foundOK bool

			ast.Inspect(file, func(n ast.Node) bool {
				if funcDecl, ok := n.(*ast.FuncDecl); ok {
					var parseOK bool
					methodName, handlerFunc, description, parseOK = gen.parseHandlerFunc(funcDecl)
					if parseOK {
						foundOK = true
						return false
					}
				}
				return true
			})

			if foundOK != tt.expectedOK {
				t.Errorf("expected ok=%v, got ok=%v", tt.expectedOK, foundOK)
				return
			}

			if !tt.expectedOK {
				return
			}

			if methodName != tt.expectedMethod {
				t.Errorf("expected method %q, got %q", tt.expectedMethod, methodName)
			}

			if handlerFunc != tt.expectedHandler {
				t.Errorf("expected handler %q, got %q", tt.expectedHandler, handlerFunc)
			}

			if description == "" && tt.name == "valid handler with documentation" {
				t.Error("expected non-empty description for documented handler")
			}
		})
	}
}

func TestExtractMethods(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create test constants.go
	constantsContent := `package server

const (
	MethodMove = "move"
	MethodAttack = "attack"
	MethodGetSpell = "getSpell"
)
`
	constantsPath := filepath.Join(tmpDir, "constants.go")
	if err := os.WriteFile(constantsPath, []byte(constantsContent), 0o644); err != nil {
		t.Fatalf("failed to write constants.go: %v", err)
	}

	// Create test handlers.go
	handlersContent := `package server

import "encoding/json"

// handleMove processes player movement
func (s *RPCServer) handleMove(params json.RawMessage) (interface{}, error) {
	return nil, nil
}

// handleAttack processes attack actions
func (s *RPCServer) handleAttack(params json.RawMessage) (interface{}, error) {
	return nil, nil
}
`
	handlersPath := filepath.Join(tmpDir, "handlers.go")
	if err := os.WriteFile(handlersPath, []byte(handlersContent), 0o644); err != nil {
		t.Fatalf("failed to write handlers.go: %v", err)
	}

	gen := &OpenAPIGenerator{
		packagePath: tmpDir,
	}

	if err := gen.extractMethods(); err != nil {
		t.Fatalf("extractMethods failed: %v", err)
	}

	expectedMethods := map[string]bool{
		"move":     true,
		"attack":   true,
		"getSpell": true,
	}

	if len(gen.methods) != len(expectedMethods) {
		t.Errorf("expected %d methods, got %d", len(expectedMethods), len(gen.methods))
	}

	for _, method := range gen.methods {
		if !expectedMethods[method.Name] {
			t.Errorf("unexpected method %q", method.Name)
		}

		// Check that move and attack have handler functions
		if method.Name == "move" || method.Name == "attack" {
			if method.HandlerFunc == "" {
				t.Errorf("method %q should have handler function", method.Name)
			}
			if method.Description == "" {
				t.Errorf("method %q should have description", method.Name)
			}
		}
	}
}

func TestExtractMethods_MissingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	gen := &OpenAPIGenerator{
		packagePath: tmpDir,
	}

	err := gen.extractMethods()
	if err == nil {
		t.Error("expected error when constants.go is missing")
	}
}

func TestExtractMethods_InvalidSyntax(t *testing.T) {
	tmpDir := t.TempDir()

	// Create malformed constants.go
	constantsContent := `package server

const (
	MethodMove = "move
	// Missing closing quote
)
`
	constantsPath := filepath.Join(tmpDir, "constants.go")
	if err := os.WriteFile(constantsPath, []byte(constantsContent), 0o644); err != nil {
		t.Fatalf("failed to write constants.go: %v", err)
	}

	gen := &OpenAPIGenerator{
		packagePath: tmpDir,
	}

	err := gen.extractMethods()
	if err == nil {
		t.Error("expected error when parsing invalid Go syntax")
	}
}

func TestGenerateOpenAPISpec(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "openapi.yaml")

	// Create test constants.go
	constantsContent := `package server

const (
	MethodMove = "move"
	MethodAttack = "attack"
)
`
	constantsPath := filepath.Join(tmpDir, "constants.go")
	if err := os.WriteFile(constantsPath, []byte(constantsContent), 0o644); err != nil {
		t.Fatalf("failed to write constants.go: %v", err)
	}

	gen := &OpenAPIGenerator{
		packagePath: tmpDir,
	}

	if err := gen.Run(specPath); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify spec file was created
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Fatal("spec file was not created")
	}

	// Parse and validate spec
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read spec: %v", err)
	}

	var spec map[string]interface{}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		t.Fatalf("failed to parse spec YAML: %v", err)
	}

	// Verify basic structure
	if spec["openapi"] != "3.0.0" {
		t.Error("expected OpenAPI version 3.0.0")
	}

	// Verify methods are listed
	methods, ok := spec["x-rpc-methods"].([]interface{})
	if !ok || len(methods) != 2 {
		t.Errorf("expected 2 methods in x-rpc-methods, got %v", methods)
	}

	// Verify method groups exist
	if _, ok := spec["x-rpc-method-groups"].(map[string]interface{}); !ok {
		t.Error("expected x-rpc-method-groups in spec")
	}
}

func TestGenerateOpenAPISpec_PreservesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "openapi.yaml")

	// Create existing spec with custom content
	existingSpec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "Custom Title",
			"description": "Custom Description",
			"version":     "2.0.0",
		},
		"components": map[string]interface{}{
			"schemas": map[string]interface{}{
				"CustomSchema": map[string]interface{}{
					"type": "object",
				},
			},
		},
	}

	data, err := yaml.Marshal(existingSpec)
	if err != nil {
		t.Fatalf("failed to marshal existing spec: %v", err)
	}

	if err := os.WriteFile(specPath, data, 0o644); err != nil {
		t.Fatalf("failed to write existing spec: %v", err)
	}

	// Create test constants.go
	constantsContent := `package server

const (
	MethodMove = "move"
)
`
	constantsPath := filepath.Join(tmpDir, "constants.go")
	if err := os.WriteFile(constantsPath, []byte(constantsContent), 0o644); err != nil {
		t.Fatalf("failed to write constants.go: %v", err)
	}

	gen := &OpenAPIGenerator{
		packagePath: tmpDir,
	}

	if err := gen.Run(specPath); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Parse updated spec
	updatedData, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read updated spec: %v", err)
	}

	var updatedSpec map[string]interface{}
	if err := yaml.Unmarshal(updatedData, &updatedSpec); err != nil {
		t.Fatalf("failed to parse updated spec: %v", err)
	}

	// Verify custom content was preserved
	info := updatedSpec["info"].(map[string]interface{})
	if info["title"] != "Custom Title" {
		t.Error("expected custom title to be preserved")
	}

	components := updatedSpec["components"].(map[string]interface{})
	schemas := components["schemas"].(map[string]interface{})
	if _, ok := schemas["CustomSchema"]; !ok {
		t.Error("expected custom schema to be preserved")
	}

	// Verify new methods were added
	if _, ok := updatedSpec["x-rpc-methods"]; !ok {
		t.Error("expected x-rpc-methods to be added")
	}
}

func TestCategorizeMethod(t *testing.T) {
	tests := []struct {
		constName     string
		expectedGroup string
	}{
		{"MethodMove", "Game Actions"},
		{"MethodAttack", "Combat"},
		{"MethodCastSpell", "Spell Management"},
		{"MethodEquipItem", "Equipment Management"},
		{"MethodGetQuest", "Quest Management"},
		{"MethodGenerateDungeon", "Procedural Content Generation"},
		{"MethodPCGValidate", "Procedural Content Generation"},
		{"MethodGetNearestObjects", "Spatial Queries"},
		{"MethodStartCombat", "Combat"},
		{"MethodEndTurn", "Combat"},
		{"MethodJoinGame", "Game Actions"},
		{"MethodLeaveGame", "Game Actions"},
		{"MethodGetGameState", "Game State"},
		{"MethodCreateCharacter", "Game State"},
		{"MethodUseItem", "Item Management"},
		{"MethodSomethingElse", "General"},
	}

	gen := &OpenAPIGenerator{}

	for _, tt := range tests {
		t.Run(tt.constName, func(t *testing.T) {
			group := gen.categorizeMethod(tt.constName)
			if group != tt.expectedGroup {
				t.Errorf("expected group %q, got %q", tt.expectedGroup, group)
			}
		})
	}
}
