// Package analyzer provides code complexity analysis capabilities
package analyzer

import (
	"go/ast"
	"go/token"
	"goldbox-rpg/pkg/testdiscovery"
)

// ComplexityCalculator provides comprehensive code complexity analysis
type ComplexityCalculator struct {
	fileSet *token.FileSet
}

// NewComplexityCalculator creates a new complexity calculator
func NewComplexityCalculator(fileSet *token.FileSet) *ComplexityCalculator {
	return &ComplexityCalculator{
		fileSet: fileSet,
	}
}

// CalculateFileComplexity calculates comprehensive complexity metrics for a file
func (cc *ComplexityCalculator) CalculateFileComplexity(fileInfo *testdiscovery.FileInfo) ComplexityMetrics {
	metrics := ComplexityMetrics{
		FilePath: fileInfo.Path,
	}

	// Calculate function-level complexity
	for i, function := range fileInfo.ExportedFunctions {
		funcMetrics := cc.calculateFunctionComplexity(&function, fileInfo.AST)
		metrics.Functions = append(metrics.Functions, funcMetrics)

		// Update the function info with calculated complexity
		fileInfo.ExportedFunctions[i].ComplexityScore = funcMetrics.CyclomaticComplexity
		fileInfo.ExportedFunctions[i].LineCount = funcMetrics.LineCount
	}

	// Calculate overall file metrics
	metrics.TotalComplexity = cc.calculateTotalComplexity(metrics.Functions)
	metrics.AverageComplexity = cc.calculateAverageComplexity(metrics.Functions)
	metrics.MaxComplexity = cc.findMaxComplexity(metrics.Functions)
	metrics.Maintainability = cc.calculateMaintainabilityIndex(fileInfo, &metrics)

	return metrics
}

// ComplexityMetrics represents comprehensive complexity analysis results
type ComplexityMetrics struct {
	FilePath          string               `json:"file_path"`
	TotalComplexity   int                  `json:"total_complexity"`
	AverageComplexity float64              `json:"average_complexity"`
	MaxComplexity     int                  `json:"max_complexity"`
	Maintainability   float64              `json:"maintainability_index"`
	Functions         []FunctionComplexity `json:"functions"`
}

// FunctionComplexity represents complexity metrics for a single function
type FunctionComplexity struct {
	Name                 string   `json:"name"`
	CyclomaticComplexity int      `json:"cyclomatic_complexity"`
	CognitiveComplexity  int      `json:"cognitive_complexity"`
	LineCount            int      `json:"line_count"`
	ParameterCount       int      `json:"parameter_count"`
	ReturnCount          int      `json:"return_count"`
	NestingDepth         int      `json:"nesting_depth"`
	BranchCount          int      `json:"branch_count"`
	LoopCount            int      `json:"loop_count"`
	IsComplexFunction    bool     `json:"is_complex"`
	Recommendations      []string `json:"recommendations"`
}

// calculateFunctionComplexity calculates detailed complexity metrics for a function
func (cc *ComplexityCalculator) calculateFunctionComplexity(funcInfo *testdiscovery.FunctionInfo, file *ast.File) FunctionComplexity {
	complexity := FunctionComplexity{
		Name:           funcInfo.Name,
		ParameterCount: funcInfo.ParameterCount,
		ReturnCount:    funcInfo.ReturnCount,
	}

	// Find the function declaration in the AST
	funcDecl := cc.findFunctionDeclaration(file, funcInfo.Name)
	if funcDecl == nil || funcDecl.Body == nil {
		return complexity
	}

	// Calculate line count
	startPos := cc.fileSet.Position(funcDecl.Pos())
	endPos := cc.fileSet.Position(funcDecl.End())
	complexity.LineCount = endPos.Line - startPos.Line + 1

	// Analyze function body
	visitor := &complexityAnalyzer{
		cyclomatic:   1, // Base complexity
		cognitive:    0,
		nestingDepth: 0,
		maxNesting:   0,
		branches:     0,
		loops:        0,
	}

	ast.Walk(visitor, funcDecl.Body)

	complexity.CyclomaticComplexity = visitor.cyclomatic
	complexity.CognitiveComplexity = visitor.cognitive
	complexity.NestingDepth = visitor.maxNesting
	complexity.BranchCount = visitor.branches
	complexity.LoopCount = visitor.loops

	// Determine if function is complex
	complexity.IsComplexFunction = complexity.CyclomaticComplexity > 10 ||
		complexity.CognitiveComplexity > 15 ||
		complexity.NestingDepth > 4

	// Generate recommendations
	complexity.Recommendations = cc.generateRecommendations(&complexity)

	return complexity
}

// complexityAnalyzer implements ast.Visitor to analyze code complexity
type complexityAnalyzer struct {
	cyclomatic   int
	cognitive    int
	nestingDepth int
	maxNesting   int
	branches     int
	loops        int
}

func (ca *complexityAnalyzer) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.IfStmt:
		ca.cyclomatic++
		ca.cognitive += (1 + ca.nestingDepth)
		ca.branches++
		ca.nestingDepth++
		if ca.nestingDepth > ca.maxNesting {
			ca.maxNesting = ca.nestingDepth
		}

		// Visit the condition and body
		ast.Walk(&complexityAnalyzer{
			cyclomatic:   0,
			cognitive:    ca.cognitive,
			nestingDepth: ca.nestingDepth,
			maxNesting:   ca.maxNesting,
			branches:     0,
			loops:        0,
		}, n.Body)

		ca.nestingDepth--
		return nil

	case *ast.ForStmt, *ast.RangeStmt:
		ca.cyclomatic++
		ca.cognitive += (1 + ca.nestingDepth)
		ca.loops++
		ca.nestingDepth++
		if ca.nestingDepth > ca.maxNesting {
			ca.maxNesting = ca.nestingDepth
		}

		// Don't automatically visit children - we'll handle nesting
		if forStmt, ok := n.(*ast.ForStmt); ok && forStmt.Body != nil {
			childVisitor := &complexityAnalyzer{
				cyclomatic:   0,
				cognitive:    ca.cognitive,
				nestingDepth: ca.nestingDepth,
				maxNesting:   ca.maxNesting,
				branches:     0,
				loops:        0,
			}
			ast.Walk(childVisitor, forStmt.Body)
			ca.cognitive = childVisitor.cognitive
			ca.maxNesting = childVisitor.maxNesting
		}

		if rangeStmt, ok := n.(*ast.RangeStmt); ok && rangeStmt.Body != nil {
			childVisitor := &complexityAnalyzer{
				cyclomatic:   0,
				cognitive:    ca.cognitive,
				nestingDepth: ca.nestingDepth,
				maxNesting:   ca.maxNesting,
				branches:     0,
				loops:        0,
			}
			ast.Walk(childVisitor, rangeStmt.Body)
			ca.cognitive = childVisitor.cognitive
			ca.maxNesting = childVisitor.maxNesting
		}

		ca.nestingDepth--
		return nil

	case *ast.SwitchStmt:
		ca.cyclomatic++
		ca.cognitive += (1 + ca.nestingDepth)
		ca.branches++

		// Count case statements
		if n.Body != nil {
			for _, stmt := range n.Body.List {
				if _, ok := stmt.(*ast.CaseClause); ok {
					ca.cyclomatic++
				}
			}
		}

	case *ast.TypeSwitchStmt:
		ca.cyclomatic++
		ca.cognitive += (1 + ca.nestingDepth)
		ca.branches++

		// Count case statements
		if n.Body != nil {
			for _, stmt := range n.Body.List {
				if _, ok := stmt.(*ast.CaseClause); ok {
					ca.cyclomatic++
				}
			}
		}

	case *ast.CaseClause:
		// Additional complexity for each case condition
		if len(n.List) > 1 {
			ca.cyclomatic += len(n.List) - 1
		}

	case *ast.FuncLit:
		// Anonymous functions add cognitive complexity
		ca.cognitive += 1

	case *ast.CallExpr:
		// Check for recursive calls (adds complexity)
		if ident, ok := n.Fun.(*ast.Ident); ok {
			// This is a simplified check - in a real implementation,
			// you'd want to verify this is actually a recursive call
			_ = ident // Use the variable to avoid compiler warnings
		}
	}

	return ca
}

// findFunctionDeclaration finds a function declaration by name in the AST
func (cc *ComplexityCalculator) findFunctionDeclaration(file *ast.File, funcName string) *ast.FuncDecl {
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == funcName {
				return funcDecl
			}
		}
	}
	return nil
}

// calculateTotalComplexity calculates the total complexity across all functions
func (cc *ComplexityCalculator) calculateTotalComplexity(functions []FunctionComplexity) int {
	total := 0
	for _, function := range functions {
		total += function.CyclomaticComplexity
	}
	return total
}

// calculateAverageComplexity calculates the average complexity
func (cc *ComplexityCalculator) calculateAverageComplexity(functions []FunctionComplexity) float64 {
	if len(functions) == 0 {
		return 0
	}

	total := cc.calculateTotalComplexity(functions)
	return float64(total) / float64(len(functions))
}

// findMaxComplexity finds the maximum complexity among all functions
func (cc *ComplexityCalculator) findMaxComplexity(functions []FunctionComplexity) int {
	max := 0
	for _, function := range functions {
		if function.CyclomaticComplexity > max {
			max = function.CyclomaticComplexity
		}
	}
	return max
}

// calculateMaintainabilityIndex calculates the maintainability index
// Based on the Microsoft formula: MAX(0, (171 - 5.2 * ln(V) - 0.23 * G - 16.2 * ln(LOC)) * 100 / 171)
// Where V = Halstead Volume, G = Cyclomatic Complexity, LOC = Lines of Code
func (cc *ComplexityCalculator) calculateMaintainabilityIndex(fileInfo *testdiscovery.FileInfo, metrics *ComplexityMetrics) float64 {
	if len(metrics.Functions) == 0 {
		return 100.0 // Perfect maintainability for empty files
	}

	// Simplified maintainability calculation
	// In a full implementation, you'd calculate Halstead metrics
	avgComplexity := metrics.AverageComplexity
	lineCount := float64(fileInfo.LineCount)

	// Simplified formula focusing on complexity and size
	maintainability := 100.0 - (avgComplexity * 5.0) - (lineCount * 0.1)

	if maintainability < 0 {
		maintainability = 0
	}
	if maintainability > 100 {
		maintainability = 100
	}

	return maintainability
}

// generateRecommendations generates recommendations based on complexity metrics
func (cc *ComplexityCalculator) generateRecommendations(complexity *FunctionComplexity) []string {
	var recommendations []string

	if complexity.CyclomaticComplexity > 15 {
		recommendations = append(recommendations, "Consider breaking this function into smaller functions (high cyclomatic complexity)")
	}

	if complexity.CognitiveComplexity > 20 {
		recommendations = append(recommendations, "Reduce cognitive complexity by simplifying nested logic")
	}

	if complexity.NestingDepth > 4 {
		recommendations = append(recommendations, "Reduce nesting depth by using early returns or extracting functions")
	}

	if complexity.ParameterCount > 7 {
		recommendations = append(recommendations, "Consider using a struct or reducing the number of parameters")
	}

	if complexity.LineCount > 50 {
		recommendations = append(recommendations, "Function is quite long, consider breaking it into smaller functions")
	}

	if complexity.LoopCount > 3 {
		recommendations = append(recommendations, "Multiple loops detected, consider optimizing or restructuring")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Good function complexity - easy to test and maintain")
	}

	return recommendations
}

// AnalyzeTestability provides detailed testability analysis
func (cc *ComplexityCalculator) AnalyzeTestability(fileInfo *testdiscovery.FileInfo) TestabilityAnalysis {
	analysis := TestabilityAnalysis{
		FilePath:    fileInfo.Path,
		Testable:    true,
		Score:       100.0,
		Challenges:  make([]string, 0),
		Suggestions: make([]string, 0),
	}

	// Analyze dependencies
	if fileInfo.ImportCount > 5 {
		analysis.Score -= float64(fileInfo.ImportCount-5) * 3
		analysis.Challenges = append(analysis.Challenges, "High number of dependencies makes testing complex")
		analysis.Suggestions = append(analysis.Suggestions, "Consider using dependency injection to reduce coupling")
	}

	// Check for problematic dependencies
	if fileInfo.HasDatabaseAccess {
		analysis.Score -= 20
		analysis.Challenges = append(analysis.Challenges, "Database access requires mocking or test databases")
		analysis.Suggestions = append(analysis.Suggestions, "Use interfaces for database operations to enable mocking")
	}

	if fileInfo.HasNetworkAccess {
		analysis.Score -= 15
		analysis.Challenges = append(analysis.Challenges, "Network access requires HTTP mocking")
		analysis.Suggestions = append(analysis.Suggestions, "Use HTTP client interfaces and test servers")
	}

	if fileInfo.HasFileIO {
		analysis.Score -= 10
		analysis.Challenges = append(analysis.Challenges, "File I/O operations need filesystem mocking")
		analysis.Suggestions = append(analysis.Suggestions, "Use io.Reader/Writer interfaces or afero for filesystem abstraction")
	}

	// Analyze complexity
	if fileInfo.ComplexityScore > 30 {
		analysis.Score -= 15
		analysis.Challenges = append(analysis.Challenges, "High complexity makes comprehensive testing difficult")
		analysis.Suggestions = append(analysis.Suggestions, "Break complex functions into smaller, testable units")
	}

	// Check for interfaces (positive factor)
	if fileInfo.InterfaceCount > 0 {
		analysis.Score += float64(fileInfo.InterfaceCount) * 5
		analysis.Suggestions = append(analysis.Suggestions, "Good use of interfaces enables easy mocking")
	}

	// Ensure score bounds
	if analysis.Score < 0 {
		analysis.Score = 0
		analysis.Testable = false
		analysis.Challenges = append(analysis.Challenges, "File requires significant refactoring before testing")
	}
	if analysis.Score > 100 {
		analysis.Score = 100
	}

	// Determine overall testability
	if analysis.Score < 30 {
		analysis.Testable = false
	}

	return analysis
}

// TestabilityAnalysis represents the testability analysis results
type TestabilityAnalysis struct {
	FilePath    string   `json:"file_path"`
	Testable    bool     `json:"testable"`
	Score       float64  `json:"score"`
	Challenges  []string `json:"challenges"`
	Suggestions []string `json:"suggestions"`
}
