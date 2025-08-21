package server

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// Test_Handler_Registration_Coverage ensures that every RPC method constant
// has a corresponding case in the handleMethod switch statement
func Test_Handler_Registration_Coverage(t *testing.T) {
	// Parse the constants file to extract RPC method constants
	constFileSet := token.NewFileSet()
	constAST, err := parser.ParseFile(constFileSet, "constants.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse constants.go: %v", err)
	}

	// Parse the server file to extract switch cases
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

	// Extract switch cases from server.go handleMethod function
	switchCases := make(map[string]bool)
	ast.Inspect(serverAST, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Name.Name == "handleMethod" {
			ast.Inspect(funcDecl, func(n ast.Node) bool {
				if _, ok := n.(*ast.TypeSwitchStmt); ok {
					return true // Skip type switches
				}
				if switchStmt, ok := n.(*ast.SwitchStmt); ok {
					for _, stmt := range switchStmt.Body.List {
						if caseClause, ok := stmt.(*ast.CaseClause); ok {
							for _, expr := range caseClause.List {
								if ident, ok := expr.(*ast.Ident); ok {
									if strings.HasPrefix(ident.Name, "Method") {
										switchCases[ident.Name] = true
									}
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
		if !switchCases[methodName] {
			missing = append(missing, methodName)
		}
	}

	// Find extra registrations
	extra := []string{}
	for caseName := range switchCases {
		if !methodConstants[caseName] {
			extra = append(extra, caseName)
		}
	}

	t.Logf("Found %d method constants and %d switch cases", len(methodConstants), len(switchCases))

	if len(missing) > 0 {
		t.Errorf("❌ Missing switch cases for %d methods: %v", len(missing), missing)
	}

	if len(extra) > 0 {
		t.Logf("⚠️  Extra switch cases not in constants: %v", extra)
	}

	if len(missing) == 0 && len(extra) == 0 {
		t.Logf("✅ All %d RPC method constants have corresponding switch cases", len(methodConstants))
	}
}
