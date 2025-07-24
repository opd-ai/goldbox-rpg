// Package generator provides comprehensive test generation capabilities
package generator

import (
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"

	"goldbox-rpg/pkg/testdiscovery"
)

// TestBuilder provides comprehensive test file generation capabilities
type TestBuilder struct {
	fileSet       *token.FileSet
	packageName   string
	imports       map[string]string
	testFunctions []TestFunction
	options       BuilderOptions
}

// BuilderOptions configures test generation behavior
type BuilderOptions struct {
	IncludeExamples      bool    // Generate example tests
	IncludeBenchmarks    bool    // Generate benchmark tests
	IncludeTableDriven   bool    // Use table-driven test patterns
	CoverageTarget       float64 // Target coverage percentage
	MockStrategy         string  // Mocking strategy: "auto", "manual", "none"
	GenerateHelpers      bool    // Generate test helper functions
	IncludeSetupTeardown bool    // Include setup/teardown functions
	UseTestify           bool    // Use testify/assert library
	MaxTestsPerFunction  int     // Maximum tests per function
	GenerateSubtests     bool    // Use t.Run for subtests
}

// TestFunction represents a generated test function
type TestFunction struct {
	Name          string     // Test function name
	TargetFunc    string     // Function being tested
	TestCases     []TestCase // Individual test cases
	SetupCode     string     // Setup code
	TeardownCode  string     // Teardown code
	Imports       []string   // Required imports
	MockSetup     string     // Mock setup code
	Documentation string     // Test documentation
	IsTableDriven bool       // Whether this uses table-driven pattern
	IsBenchmark   bool       // Whether this is a benchmark
}

// TestCase represents an individual test case
type TestCase struct {
	Name           string       // Test case name
	Description    string       // Test case description
	Inputs         []TestInput  // Input parameters
	ExpectedOutput []TestOutput // Expected outputs
	ExpectedError  string       // Expected error (if any)
	Setup          string       // Case-specific setup
	Assertions     []string     // Assertion statements
	IsErrorCase    bool         // Whether this tests error conditions
	Tags           []string     // Test tags (e.g., "unit", "integration")
}

// TestInput represents test input parameters
type TestInput struct {
	Name      string      // Parameter name
	Type      string      // Parameter type
	Value     interface{} // Parameter value
	IsPointer bool        // Whether this is a pointer
	IsMock    bool        // Whether this is a mock object
}

// TestOutput represents expected test outputs
type TestOutput struct {
	Name    string      // Return value name
	Type    string      // Return value type
	Value   interface{} // Expected value
	Matcher string      // Assertion matcher to use
}

// NewTestBuilder creates a new test builder
func NewTestBuilder(packageName string) *TestBuilder {
	return &TestBuilder{
		fileSet:       token.NewFileSet(),
		packageName:   packageName,
		imports:       make(map[string]string),
		testFunctions: make([]TestFunction, 0),
		options:       DefaultBuilderOptions(),
	}
}

// DefaultBuilderOptions returns sensible default options
func DefaultBuilderOptions() BuilderOptions {
	return BuilderOptions{
		IncludeExamples:      true,
		IncludeBenchmarks:    false,
		IncludeTableDriven:   true,
		CoverageTarget:       80.0,
		MockStrategy:         "auto",
		GenerateHelpers:      true,
		IncludeSetupTeardown: true,
		UseTestify:           true,
		MaxTestsPerFunction:  5,
		GenerateSubtests:     true,
	}
}

// SetOptions configures the test builder options
func (tb *TestBuilder) SetOptions(options BuilderOptions) {
	tb.options = options
}

// GenerateTestFile generates a complete test file for the given source file
func (tb *TestBuilder) GenerateTestFile(fileInfo *testdiscovery.FileInfo, outputPath string) (*testdiscovery.TestGenerationResult, error) {
	startTime := time.Now()

	result := &testdiscovery.TestGenerationResult{
		FilePath: outputPath,
		Success:  false,
	}

	// Parse the source file to understand structure
	if fileInfo.AST == nil {
		return result, fmt.Errorf("source file AST not available")
	}

	// Generate test functions for each exported function
	for _, funcInfo := range fileInfo.ExportedFunctions {
		testFunc, err := tb.generateTestFunction(&funcInfo, fileInfo)
		if err != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Failed to generate test for %s: %v", funcInfo.Name, err))
			continue
		}

		tb.testFunctions = append(tb.testFunctions, *testFunc)
		result.TestCount++
	}

	// Generate the test file content
	testContent, err := tb.generateFileContent(fileInfo)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to generate test content: %v", err))
		return result, err
	}

	// Write the test file
	if err := tb.writeTestFile(outputPath, testContent); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to write test file: %v", err))
		return result, err
	}

	// Calculate generated lines
	result.GeneratedLines = strings.Count(testContent, "\n")
	result.Duration = time.Since(startTime)
	result.Success = true

	// Add informational messages
	if result.TestCount == 0 {
		result.Warnings = append(result.Warnings, "No tests generated - file may not have testable functions")
	}

	return result, nil
}

// generateTestFunction generates a test function for a specific function
func (tb *TestBuilder) generateTestFunction(funcInfo *testdiscovery.FunctionInfo, fileInfo *testdiscovery.FileInfo) (*TestFunction, error) {
	testFunc := &TestFunction{
		Name:          fmt.Sprintf("Test%s", funcInfo.Name),
		TargetFunc:    funcInfo.Name,
		TestCases:     make([]TestCase, 0),
		Imports:       make([]string, 0),
		IsTableDriven: tb.options.IncludeTableDriven && tb.shouldUseTableDriven(funcInfo),
	}

	// Generate documentation
	testFunc.Documentation = fmt.Sprintf("Test%s tests the %s function", funcInfo.Name, funcInfo.Name)

	// Generate test cases
	testCases := tb.generateTestCases(funcInfo, fileInfo)
	testFunc.TestCases = testCases

	// Add setup code if needed
	if tb.options.IncludeSetupTeardown {
		testFunc.SetupCode = tb.generateSetupCode(funcInfo, fileInfo)
		testFunc.TeardownCode = tb.generateTeardownCode(funcInfo, fileInfo)
	}

	// Add mock setup if needed
	if tb.requiresMocking(funcInfo, fileInfo) {
		mockCode, err := tb.generateMockSetup(funcInfo, fileInfo)
		if err == nil {
			testFunc.MockSetup = mockCode
		}
	}

	// Determine required imports
	testFunc.Imports = tb.determineRequiredImports(testFunc)

	return testFunc, nil
}

// shouldUseTableDriven determines if table-driven tests should be used
func (tb *TestBuilder) shouldUseTableDriven(funcInfo *testdiscovery.FunctionInfo) bool {
	// Use table-driven tests for functions with multiple parameters or complex logic
	return funcInfo.ParameterCount > 1 || funcInfo.ComplexityScore > 5
}

// generateTestCases generates test cases for a function
func (tb *TestBuilder) generateTestCases(funcInfo *testdiscovery.FunctionInfo, fileInfo *testdiscovery.FileInfo) []TestCase {
	var testCases []TestCase

	// Generate happy path test case
	happyCase := tb.generateHappyPathCase(funcInfo)
	testCases = append(testCases, happyCase)

	// Generate edge cases
	edgeCases := tb.generateEdgeCases(funcInfo)
	testCases = append(testCases, edgeCases...)

	// Generate error cases if function returns error
	if funcInfo.ReturnsError {
		errorCases := tb.generateErrorCases(funcInfo)
		testCases = append(testCases, errorCases...)
	}

	// Limit the number of test cases
	if len(testCases) > tb.options.MaxTestsPerFunction {
		testCases = testCases[:tb.options.MaxTestsPerFunction]
	}

	return testCases
}

// generateHappyPathCase generates the main success case
func (tb *TestBuilder) generateHappyPathCase(funcInfo *testdiscovery.FunctionInfo) TestCase {
	testCase := TestCase{
		Name:        "ValidInput",
		Description: "Test with valid input parameters",
		Inputs:      tb.generateValidInputs(funcInfo),
		Assertions:  make([]string, 0),
	}

	// Generate expected outputs
	testCase.ExpectedOutput = tb.generateExpectedOutputs(funcInfo, testCase.Inputs)

	// Generate assertions
	testCase.Assertions = tb.generateAssertions(funcInfo, &testCase)

	return testCase
}

// generateEdgeCases generates edge case tests
func (tb *TestBuilder) generateEdgeCases(funcInfo *testdiscovery.FunctionInfo) []TestCase {
	var edgeCases []TestCase

	// Generate cases based on parameter types
	for i, param := range funcInfo.Parameters {
		edgeCase := tb.generateParameterEdgeCase(funcInfo, i, param)
		if edgeCase != nil {
			edgeCases = append(edgeCases, *edgeCase)
		}
	}

	return edgeCases
}

// generateErrorCases generates error condition tests
func (tb *TestBuilder) generateErrorCases(funcInfo *testdiscovery.FunctionInfo) []TestCase {
	var errorCases []TestCase

	// Generate nil parameter cases for pointer parameters
	for i, param := range funcInfo.Parameters {
		if param.IsPointer {
			errorCase := TestCase{
				Name:        fmt.Sprintf("NilParameter%d", i+1),
				Description: fmt.Sprintf("Test with nil %s parameter", param.Name),
				Inputs:      tb.generateNilInputCase(funcInfo, i),
				IsErrorCase: true,
			}
			errorCase.Assertions = []string{
				"assert.Error(t, err, \"Expected error for nil parameter\")",
			}
			errorCases = append(errorCases, errorCase)
		}
	}

	// Generate invalid input cases
	invalidCase := TestCase{
		Name:        "InvalidInput",
		Description: "Test with invalid input values",
		Inputs:      tb.generateInvalidInputs(funcInfo),
		IsErrorCase: true,
	}
	invalidCase.Assertions = []string{
		"assert.Error(t, err, \"Expected error for invalid input\")",
	}
	errorCases = append(errorCases, invalidCase)

	return errorCases
}

// generateValidInputs generates valid input parameters for testing
func (tb *TestBuilder) generateValidInputs(funcInfo *testdiscovery.FunctionInfo) []TestInput {
	var inputs []TestInput

	for _, param := range funcInfo.Parameters {
		input := TestInput{
			Name:      param.Name,
			Type:      param.Type,
			IsPointer: param.IsPointer,
		}

		// Generate appropriate test value based on type
		input.Value = tb.generateTestValue(param.Type, false)
		inputs = append(inputs, input)
	}

	return inputs
}

// generateInvalidInputs generates invalid input parameters for error testing
func (tb *TestBuilder) generateInvalidInputs(funcInfo *testdiscovery.FunctionInfo) []TestInput {
	var inputs []TestInput

	for _, param := range funcInfo.Parameters {
		input := TestInput{
			Name:      param.Name,
			Type:      param.Type,
			IsPointer: param.IsPointer,
		}

		// Generate invalid test value based on type
		input.Value = tb.generateTestValue(param.Type, true)
		inputs = append(inputs, input)
	}

	return inputs
}

// generateNilInputCase generates test inputs with one parameter set to nil
func (tb *TestBuilder) generateNilInputCase(funcInfo *testdiscovery.FunctionInfo, nilIndex int) []TestInput {
	inputs := tb.generateValidInputs(funcInfo)

	if nilIndex < len(inputs) && inputs[nilIndex].IsPointer {
		inputs[nilIndex].Value = "nil"
	}

	return inputs
}

// generateParameterEdgeCase generates edge case for a specific parameter
func (tb *TestBuilder) generateParameterEdgeCase(funcInfo *testdiscovery.FunctionInfo, paramIndex int, param testdiscovery.ParameterInfo) *TestCase {
	// Only generate edge cases for certain types
	if !tb.hasEdgeCases(param.Type) {
		return nil
	}

	testCase := &TestCase{
		Name:        fmt.Sprintf("EdgeCase_%s", strings.Title(param.Name)),
		Description: fmt.Sprintf("Test edge case for %s parameter", param.Name),
		Inputs:      tb.generateValidInputs(funcInfo),
	}

	// Replace the specific parameter with edge case value
	if paramIndex < len(testCase.Inputs) {
		testCase.Inputs[paramIndex].Value = tb.generateEdgeCaseValue(param.Type)
	}

	testCase.ExpectedOutput = tb.generateExpectedOutputs(funcInfo, testCase.Inputs)
	testCase.Assertions = tb.generateAssertions(funcInfo, testCase)

	return testCase
}

// generateTestValue generates appropriate test values based on type
func (tb *TestBuilder) generateTestValue(paramType string, invalid bool) interface{} {
	switch {
	case strings.Contains(paramType, "string"):
		if invalid {
			return `""`
		}
		return `"test_string"`

	case strings.Contains(paramType, "int"):
		if invalid {
			return "-1"
		}
		return "42"

	case strings.Contains(paramType, "float"):
		if invalid {
			return "-1.0"
		}
		return "3.14"

	case strings.Contains(paramType, "bool"):
		return "true"

	case strings.Contains(paramType, "[]"):
		if invalid {
			return "nil"
		}
		return fmt.Sprintf("[]%s{}", strings.TrimPrefix(paramType, "[]"))

	case strings.Contains(paramType, "map"):
		if invalid {
			return "nil"
		}
		return fmt.Sprintf("make(%s)", paramType)

	default:
		if invalid {
			return "nil"
		}
		return fmt.Sprintf("%s{}", paramType)
	}
}

// hasEdgeCases determines if a type has meaningful edge cases
func (tb *TestBuilder) hasEdgeCases(paramType string) bool {
	edgeCaseTypes := []string{"int", "float", "string", "slice", "[]"}

	for _, edgeType := range edgeCaseTypes {
		if strings.Contains(paramType, edgeType) {
			return true
		}
	}

	return false
}

// generateEdgeCaseValue generates edge case values for specific types
func (tb *TestBuilder) generateEdgeCaseValue(paramType string) interface{} {
	switch {
	case strings.Contains(paramType, "string"):
		return `""`
	case strings.Contains(paramType, "int"):
		return "0"
	case strings.Contains(paramType, "float"):
		return "0.0"
	case strings.Contains(paramType, "[]"):
		return "nil"
	default:
		return "nil"
	}
}

// generateExpectedOutputs generates expected return values
func (tb *TestBuilder) generateExpectedOutputs(funcInfo *testdiscovery.FunctionInfo, inputs []TestInput) []TestOutput {
	var outputs []TestOutput

	for i, returnType := range funcInfo.ReturnTypes {
		output := TestOutput{
			Name: fmt.Sprintf("result%d", i),
			Type: returnType,
		}

		if returnType == "error" {
			output.Value = "nil"
			output.Matcher = "NoError"
		} else {
			output.Value = tb.generateExpectedValue(returnType, inputs)
			output.Matcher = "Equal"
		}

		outputs = append(outputs, output)
	}

	return outputs
}

// generateExpectedValue generates expected return values based on inputs
func (tb *TestBuilder) generateExpectedValue(returnType string, inputs []TestInput) interface{} {
	// This is a simplified implementation
	// In a real implementation, you might use symbolic execution or heuristics
	switch {
	case strings.Contains(returnType, "string"):
		return `"expected_result"`
	case strings.Contains(returnType, "int"):
		return "expectedResult"
	case strings.Contains(returnType, "bool"):
		return "true"
	default:
		return "expectedResult"
	}
}

// generateAssertions generates assertion statements for test cases
func (tb *TestBuilder) generateAssertions(funcInfo *testdiscovery.FunctionInfo, testCase *TestCase) []string {
	var assertions []string

	if testCase.IsErrorCase {
		assertions = append(assertions, "assert.Error(t, err)")
		return assertions
	}

	// Generate assertions for each expected output
	for i, output := range testCase.ExpectedOutput {
		var assertion string

		if tb.options.UseTestify {
			switch output.Matcher {
			case "NoError":
				assertion = "assert.NoError(t, err)"
			case "Equal":
				if i == 0 && len(testCase.ExpectedOutput) == 1 {
					assertion = fmt.Sprintf("assert.Equal(t, %v, result)", output.Value)
				} else {
					assertion = fmt.Sprintf("assert.Equal(t, %v, %s)", output.Value, output.Name)
				}
			default:
				assertion = fmt.Sprintf("assert.Equal(t, %v, result)", output.Value)
			}
		} else {
			// Use standard testing assertions
			if output.Type == "error" {
				assertion = "if err != nil { t.Errorf(\"Unexpected error: %v\", err) }"
			} else {
				assertion = fmt.Sprintf("if result != %v { t.Errorf(\"Expected %v, got %%v\", result) }",
					output.Value, output.Value)
			}
		}

		assertions = append(assertions, assertion)
	}

	return assertions
}

// requiresMocking determines if the function requires mocking
func (tb *TestBuilder) requiresMocking(funcInfo *testdiscovery.FunctionInfo, fileInfo *testdiscovery.FileInfo) bool {
	return fileInfo.HasDatabaseAccess || fileInfo.HasNetworkAccess || len(fileInfo.Dependencies) > 3
}

// generateMockSetup generates mock setup code
func (tb *TestBuilder) generateMockSetup(funcInfo *testdiscovery.FunctionInfo, fileInfo *testdiscovery.FileInfo) (string, error) {
	if tb.options.MockStrategy == "none" {
		return "", nil
	}

	// This is a simplified mock generation
	// In a real implementation, you'd analyze dependencies and generate appropriate mocks
	var mockSetup strings.Builder

	mockSetup.WriteString("// Mock setup\n")
	mockSetup.WriteString("// TODO: Generate appropriate mocks for dependencies\n")

	return mockSetup.String(), nil
}

// generateSetupCode generates test setup code
func (tb *TestBuilder) generateSetupCode(funcInfo *testdiscovery.FunctionInfo, fileInfo *testdiscovery.FileInfo) string {
	return "// Test setup\n// TODO: Add any necessary setup code"
}

// generateTeardownCode generates test teardown code
func (tb *TestBuilder) generateTeardownCode(funcInfo *testdiscovery.FunctionInfo, fileInfo *testdiscovery.FileInfo) string {
	return "// Test teardown\n// TODO: Add any necessary cleanup code"
}

// determineRequiredImports determines what imports are needed for the test
func (tb *TestBuilder) determineRequiredImports(testFunc *TestFunction) []string {
	imports := []string{"testing"}

	if tb.options.UseTestify {
		imports = append(imports, "github.com/stretchr/testify/assert")
	}

	// Add other imports based on test content
	// This is simplified - a real implementation would analyze the generated code

	return imports
}

// generateFileContent generates the complete test file content
func (tb *TestBuilder) generateFileContent(fileInfo *testdiscovery.FileInfo) (string, error) {
	var content strings.Builder

	// Package declaration
	content.WriteString(fmt.Sprintf("package %s\n\n", tb.packageName))

	// Imports
	imports := tb.collectAllImports()
	if len(imports) > 0 {
		content.WriteString("import (\n")
		for _, imp := range imports {
			content.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
		}
		content.WriteString(")\n\n")
	}

	// Generate test functions
	for _, testFunc := range tb.testFunctions {
		funcContent, err := tb.generateTestFunctionCode(&testFunc)
		if err != nil {
			return "", fmt.Errorf("failed to generate test function %s: %v", testFunc.Name, err)
		}
		content.WriteString(funcContent)
		content.WriteString("\n\n")
	}

	return content.String(), nil
}

// collectAllImports collects all required imports
func (tb *TestBuilder) collectAllImports() []string {
	importSet := make(map[string]bool)

	for _, testFunc := range tb.testFunctions {
		for _, imp := range testFunc.Imports {
			importSet[imp] = true
		}
	}

	var imports []string
	for imp := range importSet {
		imports = append(imports, imp)
	}

	return imports
}

// generateTestFunctionCode generates the code for a single test function
func (tb *TestBuilder) generateTestFunctionCode(testFunc *TestFunction) (string, error) {
	var code strings.Builder

	// Function documentation
	code.WriteString(fmt.Sprintf("// %s\n", testFunc.Documentation))

	// Function signature
	code.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", testFunc.Name))

	// Setup code
	if testFunc.SetupCode != "" {
		code.WriteString("\t" + strings.ReplaceAll(testFunc.SetupCode, "\n", "\n\t") + "\n\n")
	}

	// Mock setup
	if testFunc.MockSetup != "" {
		code.WriteString("\t" + strings.ReplaceAll(testFunc.MockSetup, "\n", "\n\t") + "\n\n")
	}

	// Generate test cases
	if testFunc.IsTableDriven {
		caseCode, err := tb.generateTableDrivenTest(testFunc)
		if err != nil {
			return "", err
		}
		code.WriteString("\t" + strings.ReplaceAll(caseCode, "\n", "\n\t"))
	} else {
		caseCode, err := tb.generateIndividualTests(testFunc)
		if err != nil {
			return "", err
		}
		code.WriteString("\t" + strings.ReplaceAll(caseCode, "\n", "\n\t"))
	}

	// Teardown code
	if testFunc.TeardownCode != "" {
		code.WriteString("\n\t" + strings.ReplaceAll(testFunc.TeardownCode, "\n", "\n\t"))
	}

	code.WriteString("\n}")

	return code.String(), nil
}

// generateTableDrivenTest generates table-driven test code
func (tb *TestBuilder) generateTableDrivenTest(testFunc *TestFunction) (string, error) {
	var code strings.Builder

	// Table definition
	code.WriteString("tests := []struct {\n")
	code.WriteString("\tname string\n")

	// Add fields for inputs and expected outputs
	if len(testFunc.TestCases) > 0 {
		firstCase := testFunc.TestCases[0]
		for _, input := range firstCase.Inputs {
			code.WriteString(fmt.Sprintf("\t%s %s\n", input.Name, input.Type))
		}
		for _, output := range firstCase.ExpectedOutput {
			code.WriteString(fmt.Sprintf("\twant%s %s\n", strings.Title(output.Name), output.Type))
		}
		if firstCase.IsErrorCase {
			code.WriteString("\twantErr bool\n")
		}
	}

	code.WriteString("}{\n")

	// Table data
	for _, testCase := range testFunc.TestCases {
		code.WriteString(fmt.Sprintf("\t{\n\t\tname: \"%s\",\n", testCase.Name))

		for _, input := range testCase.Inputs {
			code.WriteString(fmt.Sprintf("\t\t%s: %v,\n", input.Name, input.Value))
		}

		for _, output := range testCase.ExpectedOutput {
			code.WriteString(fmt.Sprintf("\t\twant%s: %v,\n", strings.Title(output.Name), output.Value))
		}

		if testCase.IsErrorCase {
			code.WriteString("\t\twantErr: true,\n")
		}

		code.WriteString("\t},\n")
	}

	code.WriteString("}\n\n")

	// Test execution loop
	code.WriteString("for _, tt := range tests {\n")
	code.WriteString("\tt.Run(tt.name, func(t *testing.T) {\n")

	// Function call - this is simplified
	code.WriteString(fmt.Sprintf("\t\tgot, err := %s(", testFunc.TargetFunc))
	if len(testFunc.TestCases) > 0 {
		firstCase := testFunc.TestCases[0]
		for i, input := range firstCase.Inputs {
			if i > 0 {
				code.WriteString(", ")
			}
			code.WriteString(fmt.Sprintf("tt.%s", input.Name))
		}
	}
	code.WriteString(")\n")

	// Assertions
	code.WriteString("\t\tif (err != nil) != tt.wantErr {\n")
	code.WriteString("\t\t\tt.Errorf(\"error = %v, wantErr %v\", err, tt.wantErr)\n")
	code.WriteString("\t\t\treturn\n")
	code.WriteString("\t\t}\n")

	if tb.options.UseTestify {
		code.WriteString("\t\tassert.Equal(t, tt.wantResult, got)\n")
	} else {
		code.WriteString("\t\tif got != tt.wantResult {\n")
		code.WriteString("\t\t\tt.Errorf(\"got %v, want %v\", got, tt.wantResult)\n")
		code.WriteString("\t\t}\n")
	}

	code.WriteString("\t})\n")
	code.WriteString("}\n")

	return code.String(), nil
}

// generateIndividualTests generates individual test functions
func (tb *TestBuilder) generateIndividualTests(testFunc *TestFunction) (string, error) {
	var code strings.Builder

	for _, testCase := range testFunc.TestCases {
		code.WriteString(fmt.Sprintf("// %s\n", testCase.Description))

		// Generate function call
		code.WriteString(fmt.Sprintf("result, err := %s(", testFunc.TargetFunc))
		for i, input := range testCase.Inputs {
			if i > 0 {
				code.WriteString(", ")
			}
			code.WriteString(fmt.Sprintf("%v", input.Value))
		}
		code.WriteString(")\n\n")

		// Generate assertions
		for _, assertion := range testCase.Assertions {
			code.WriteString(assertion + "\n")
		}

		code.WriteString("\n")
	}

	return code.String(), nil
}

// writeTestFile writes the generated test content to a file
func (tb *TestBuilder) writeTestFile(outputPath, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Format the Go code
	formattedContent, err := tb.formatGoCode(content)
	if err != nil {
		// If formatting fails, use the original content
		formattedContent = []byte(content)
	}

	// Write to file
	if err := os.WriteFile(outputPath, formattedContent, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %v", outputPath, err)
	}

	return nil
}

// formatGoCode formats Go code using go/format
func (tb *TestBuilder) formatGoCode(content string) ([]byte, error) {
	// Parse the code first to check for syntax errors
	_, err := parser.ParseFile(tb.fileSet, "", content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("code parsing failed: %v", err)
	}

	// Format the code
	formatted, err := format.Source([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("code formatting failed: %v", err)
	}

	return formatted, nil
}
