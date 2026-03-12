package server

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// Test_Handler_Registration_Coverage ensures that every RPC method constant
// has a corresponding handler in the registerMethodHandlers function
func Test_Handler_Registration_Coverage(t *testing.T) {
	// Parse the constants file to extract RPC method constants
	constFileSet := token.NewFileSet()
	constAST, err := parser.ParseFile(constFileSet, "constants.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse constants.go: %v", err)
	}

	// Parse the server file to extract registry assignments
	serverFileSet := token.NewFileSet()
	serverAST, err := parser.ParseFile(serverFileSet, "server.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse server.go: %v", err)
	}

	// Extract method constants from constants.go
	methodConstants := make(map[string]bool)
	ast.Inspect(constAST, func(n ast.Node) bool {
		if valueSpec, ok := n.(*ast.ValueSpec); ok {
			for _, name := range valueSpec.Names {
				if strings.HasPrefix(name.Name, "Method") {
					methodConstants[name.Name] = true
				}
			}
		}
		return true
	})

	// Extract registry assignments from server.go registerMethodHandlers function
	registryAssignments := make(map[string]bool)
	ast.Inspect(serverAST, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Name.Name == "registerMethodHandlers" {
			ast.Inspect(funcDecl, func(n ast.Node) bool {
				if assignStmt, ok := n.(*ast.AssignStmt); ok {
					// Look for s.methodRegistry[MethodXxx] = s.handleXxx
					for _, lhs := range assignStmt.Lhs {
						if indexExpr, ok := lhs.(*ast.IndexExpr); ok {
							if ident, ok := indexExpr.Index.(*ast.Ident); ok {
								if strings.HasPrefix(ident.Name, "Method") {
									registryAssignments[ident.Name] = true
								}
							}
						}
					}
				}
				return true
			})
		}
		return true
	})

	// Find missing registrations
	missing := []string{}
	for methodName := range methodConstants {
		if !registryAssignments[methodName] {
			missing = append(missing, methodName)
		}
	}

	// Find extra registrations
	extra := []string{}
	for regName := range registryAssignments {
		if !methodConstants[regName] {
			extra = append(extra, regName)
		}
	}

	t.Logf("Found %d method constants and %d registry assignments", len(methodConstants), len(registryAssignments))

	if len(missing) > 0 {
		t.Errorf("❌ Missing registry assignments for %d methods: %v", len(missing), missing)
	}

	if len(extra) > 0 {
		t.Logf("⚠️  Extra registry assignments not in constants: %v", extra)
	}

	if len(missing) == 0 && len(extra) == 0 {
		t.Logf("✅ All %d RPC method constants have corresponding registry assignments", len(methodConstants))
	}
}
