package server

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"goldbox-rpg/pkg/validation"
)

// TestHandlerValidatorParity ensures every RPC method constant has a corresponding
// validator registered in the validation package. This test prevents silent feature
// breakage where new RPC methods are added without validation.
func TestHandlerValidatorParity(t *testing.T) {
	// Parse the constants file to extract RPC method constants
	fset := token.NewFileSet()
	constAST, err := parser.ParseFile(fset, "constants.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse constants.go: %v", err)
	}

	// Extract method constants from constants.go
	methodConstants := make(map[string]string) // Map: MethodName -> method string value
	ast.Inspect(constAST, func(n ast.Node) bool {
		if valueSpec, ok := n.(*ast.ValueSpec); ok {
			for i, name := range valueSpec.Names {
				if strings.HasPrefix(name.Name, "Method") {
					// Extract the string value from the constant
					if i < len(valueSpec.Values) {
						if basicLit, ok := valueSpec.Values[i].(*ast.BasicLit); ok {
							// Remove quotes from string literal
							val := strings.Trim(basicLit.Value, "\"")
							methodConstants[name.Name] = val
						}
					}
				}
			}
		}
		return true
	})

	// Create a validator instance to check which methods have validators
	validator := validation.NewInputValidator(1024 * 1024)

	// Check each method constant has a validator
	missing := []string{}
	for constName, methodName := range methodConstants {
		if !validator.HasValidator(methodName) {
			missing = append(missing, constName+" ("+methodName+")")
		}
	}

	t.Logf("Found %d RPC method constants in constants.go", len(methodConstants))

	if len(missing) > 0 {
		t.Errorf("❌ Missing validators for %d methods:\n%s", len(missing), strings.Join(missing, "\n"))
	} else {
		t.Logf("✅ All %d RPC method constants have corresponding validators", len(methodConstants))
	}
}
