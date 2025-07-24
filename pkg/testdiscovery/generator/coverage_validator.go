// Package generator provides test coverage validation capabilities
package generator

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"goldbox-rpg/pkg/testdiscovery"
)

// CoverageValidator provides comprehensive test coverage validation
type CoverageValidator struct {
	workingDir    string
	tempDir       string
	coverageFiles map[string]string // packagePath -> coverage file path
}

// NewCoverageValidator creates a new coverage validator
func NewCoverageValidator(workingDir string) *CoverageValidator {
	return &CoverageValidator{
		workingDir:    workingDir,
		coverageFiles: make(map[string]string),
	}
}

// CoverageValidationResult represents the result of coverage validation
type CoverageValidationResult struct {
	PackagePath        string                 `json:"package_path"`
	TotalLines         int                    `json:"total_lines"`
	CoveredLines       int                    `json:"covered_lines"`
	CoveragePercentage float64                `json:"coverage_percentage"`
	MeetsTarget        bool                   `json:"meets_target"`
	Target             float64                `json:"target"`
	UncoveredFunctions []UncoveredFunction    `json:"uncovered_functions"`
	CoverageByFile     map[string]float64     `json:"coverage_by_file"`
	TestFiles          []string               `json:"test_files"`
	ValidationErrors   []string               `json:"validation_errors"`
	ValidationWarnings []string               `json:"validation_warnings"`
	TestExecutionTime  time.Duration          `json:"test_execution_time"`
	GeneratedAt        time.Time              `json:"generated_at"`
	Recommendations    []string               `json:"recommendations"`
	QualityMetrics     CoverageQualityMetrics `json:"quality_metrics"`
}

// UncoveredFunction represents a function that lacks test coverage
type UncoveredFunction struct {
	Function   string `json:"function"`
	File       string `json:"file"`
	StartLine  int    `json:"start_line"`
	EndLine    int    `json:"end_line"`
	Complexity int    `json:"complexity"`
	Priority   string `json:"priority"` // "high", "medium", "low"
}

// CoverageQualityMetrics provides additional quality assessment
type CoverageQualityMetrics struct {
	StatementCoverage   float64 `json:"statement_coverage"`
	BranchCoverage      float64 `json:"branch_coverage"`
	FunctionCoverage    float64 `json:"function_coverage"`
	TestQualityScore    float64 `json:"test_quality_score"`
	HasTableDrivenTests bool    `json:"has_table_driven_tests"`
	HasBenchmarks       bool    `json:"has_benchmarks"`
	HasExamples         bool    `json:"has_examples"`
	TestToCodeRatio     float64 `json:"test_to_code_ratio"`
}

// ValidateTestCoverage validates test coverage for a package
func (cv *CoverageValidator) ValidateTestCoverage(packagePath string, targetCoverage float64) (*CoverageValidationResult, error) {
	startTime := time.Now()

	result := &CoverageValidationResult{
		PackagePath:        packagePath,
		Target:             targetCoverage,
		GeneratedAt:        startTime,
		CoverageByFile:     make(map[string]float64),
		TestFiles:          make([]string, 0),
		ValidationErrors:   make([]string, 0),
		ValidationWarnings: make([]string, 0),
		Recommendations:    make([]string, 0),
		UncoveredFunctions: make([]UncoveredFunction, 0),
	}

	// Run tests with coverage
	coverageData, err := cv.runTestsWithCoverage(packagePath)
	if err != nil {
		result.ValidationErrors = append(result.ValidationErrors,
			fmt.Sprintf("Failed to run tests with coverage: %v", err))
		return result, nil // Return partial result
	}

	result.TestExecutionTime = time.Since(startTime)

	// Parse coverage data
	if err := cv.parseCoverageData(coverageData, result); err != nil {
		result.ValidationErrors = append(result.ValidationErrors,
			fmt.Sprintf("Failed to parse coverage data: %v", err))
		return result, nil
	}

	// Analyze coverage quality
	cv.analyzeCoverageQuality(packagePath, result)

	// Find uncovered functions
	cv.findUncoveredFunctions(packagePath, result)

	// Generate recommendations
	cv.generateRecommendations(result)

	// Check if target is met
	result.MeetsTarget = result.CoveragePercentage >= targetCoverage

	return result, nil
}

// runTestsWithCoverage runs tests and generates coverage data
func (cv *CoverageValidator) runTestsWithCoverage(packagePath string) (string, error) {
	// Create temporary coverage file
	coverageFile := filepath.Join(os.TempDir(), fmt.Sprintf("coverage_%d.out", time.Now().UnixNano()))
	cv.coverageFiles[packagePath] = coverageFile

	// Build test command
	cmd := exec.Command("go", "test", "-cover", "-coverprofile="+coverageFile, packagePath)
	cmd.Dir = cv.workingDir

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("test execution failed: %v\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// parseCoverageData parses the coverage profile data
func (cv *CoverageValidator) parseCoverageData(coverageOutput string, result *CoverageValidationResult) error {
	// Extract coverage percentage from output
	coverageRegex := regexp.MustCompile(`coverage:\s+([\d.]+)%`)
	matches := coverageRegex.FindStringSubmatch(coverageOutput)

	if len(matches) > 1 {
		coverage, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			result.CoveragePercentage = coverage
		}
	}

	// Parse detailed coverage from profile file
	coverageFile := cv.coverageFiles[result.PackagePath]
	if coverageFile != "" {
		return cv.parseDetailedCoverage(coverageFile, result)
	}

	return nil
}

// parseDetailedCoverage parses detailed coverage from profile file
func (cv *CoverageValidator) parseDetailedCoverage(coverageFile string, result *CoverageValidationResult) error {
	file, err := os.Open(coverageFile)
	if err != nil {
		return fmt.Errorf("failed to open coverage file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Skip the mode line
	if scanner.Scan() {
		// mode: set, count, atomic
	}

	fileCoverage := make(map[string]*FileCoverageData)

	for scanner.Scan() {
		line := scanner.Text()
		if err := cv.parseCoverageLine(line, fileCoverage); err != nil {
			result.ValidationWarnings = append(result.ValidationWarnings,
				fmt.Sprintf("Failed to parse coverage line: %v", err))
		}
	}

	// Calculate per-file coverage percentages
	for fileName, data := range fileCoverage {
		if data.TotalStatements > 0 {
			percentage := float64(data.CoveredStatements) / float64(data.TotalStatements) * 100
			result.CoverageByFile[fileName] = percentage
		}
	}

	// Calculate overall totals
	totalStatements := 0
	coveredStatements := 0
	for _, data := range fileCoverage {
		totalStatements += data.TotalStatements
		coveredStatements += data.CoveredStatements
	}

	result.TotalLines = totalStatements
	result.CoveredLines = coveredStatements

	if totalStatements > 0 {
		result.CoveragePercentage = float64(coveredStatements) / float64(totalStatements) * 100
	}

	return scanner.Err()
}

// FileCoverageData tracks coverage data for a single file
type FileCoverageData struct {
	TotalStatements   int
	CoveredStatements int
}

// parseCoverageLine parses a single line from coverage profile
func (cv *CoverageValidator) parseCoverageLine(line string, fileCoverage map[string]*FileCoverageData) error {
	// Coverage line format: file.go:startLine.startCol,endLine.endCol numStmts count
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return fmt.Errorf("invalid coverage line format: %s", line)
	}

	// Extract file path
	filePath := strings.Split(parts[0], ":")[0]

	// Extract number of statements
	numStmts, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid statement count: %s", parts[1])
	}

	// Extract execution count
	count, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid execution count: %s", parts[2])
	}

	// Initialize file data if not exists
	if fileCoverage[filePath] == nil {
		fileCoverage[filePath] = &FileCoverageData{}
	}

	// Update coverage data
	fileCoverage[filePath].TotalStatements += numStmts
	if count > 0 {
		fileCoverage[filePath].CoveredStatements += numStmts
	}

	return nil
}

// analyzeCoverageQuality analyzes the quality of test coverage
func (cv *CoverageValidator) analyzeCoverageQuality(packagePath string, result *CoverageValidationResult) {
	result.QualityMetrics.StatementCoverage = result.CoveragePercentage

	// Find test files
	testFiles, err := cv.findTestFiles(packagePath)
	if err == nil {
		result.TestFiles = testFiles
		result.QualityMetrics.HasTableDrivenTests = cv.hasTableDrivenTests(testFiles)
		result.QualityMetrics.HasBenchmarks = cv.hasBenchmarks(testFiles)
		result.QualityMetrics.HasExamples = cv.hasExamples(testFiles)
	}

	// Calculate test quality score
	result.QualityMetrics.TestQualityScore = cv.calculateTestQualityScore(&result.QualityMetrics)

	// Calculate test-to-code ratio
	result.QualityMetrics.TestToCodeRatio = cv.calculateTestToCodeRatio(packagePath, result.TestFiles)
}

// findTestFiles finds all test files in the package
func (cv *CoverageValidator) findTestFiles(packagePath string) ([]string, error) {
	var testFiles []string

	// Convert package path to directory path
	packageDir := filepath.Join(cv.workingDir, strings.ReplaceAll(packagePath, "/", string(filepath.Separator)))

	err := filepath.Walk(packageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), "_test.go") {
			relPath, err := filepath.Rel(cv.workingDir, path)
			if err == nil {
				testFiles = append(testFiles, relPath)
			}
		}

		return nil
	})

	return testFiles, err
}

// hasTableDrivenTests checks if test files contain table-driven tests
func (cv *CoverageValidator) hasTableDrivenTests(testFiles []string) bool {
	for _, testFile := range testFiles {
		content, err := os.ReadFile(filepath.Join(cv.workingDir, testFile))
		if err != nil {
			continue
		}

		contentStr := string(content)
		if strings.Contains(contentStr, "[]struct{") && strings.Contains(contentStr, "t.Run(") {
			return true
		}
	}
	return false
}

// hasBenchmarks checks if test files contain benchmark tests
func (cv *CoverageValidator) hasBenchmarks(testFiles []string) bool {
	for _, testFile := range testFiles {
		content, err := os.ReadFile(filepath.Join(cv.workingDir, testFile))
		if err != nil {
			continue
		}

		if strings.Contains(string(content), "func Benchmark") {
			return true
		}
	}
	return false
}

// hasExamples checks if test files contain example tests
func (cv *CoverageValidator) hasExamples(testFiles []string) bool {
	for _, testFile := range testFiles {
		content, err := os.ReadFile(filepath.Join(cv.workingDir, testFile))
		if err != nil {
			continue
		}

		if strings.Contains(string(content), "func Example") {
			return true
		}
	}
	return false
}

// calculateTestQualityScore calculates a quality score for tests
func (cv *CoverageValidator) calculateTestQualityScore(metrics *CoverageQualityMetrics) float64 {
	score := metrics.StatementCoverage // Base score from coverage

	// Bonus points for good testing practices
	if metrics.HasTableDrivenTests {
		score += 5
	}
	if metrics.HasBenchmarks {
		score += 3
	}
	if metrics.HasExamples {
		score += 2
	}

	// Bonus for good test-to-code ratio
	if metrics.TestToCodeRatio >= 0.5 && metrics.TestToCodeRatio <= 2.0 {
		score += 5
	}

	// Cap the score at 100
	if score > 100 {
		score = 100
	}

	return score
}

// calculateTestToCodeRatio calculates the ratio of test code to source code
func (cv *CoverageValidator) calculateTestToCodeRatio(packagePath string, testFiles []string) float64 {
	testLines := 0
	sourceLines := 0

	// Count test lines
	for _, testFile := range testFiles {
		lines := cv.countLinesInFile(filepath.Join(cv.workingDir, testFile))
		testLines += lines
	}

	// Count source lines
	packageDir := filepath.Join(cv.workingDir, strings.ReplaceAll(packagePath, "/", string(filepath.Separator)))
	filepath.Walk(packageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go") {
			lines := cv.countLinesInFile(path)
			sourceLines += lines
		}

		return nil
	})

	if sourceLines == 0 {
		return 0
	}

	return float64(testLines) / float64(sourceLines)
}

// countLinesInFile counts the number of lines in a file
func (cv *CoverageValidator) countLinesInFile(filePath string) int {
	file, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		lines++
	}

	return lines
}

// findUncoveredFunctions identifies functions that lack test coverage
func (cv *CoverageValidator) findUncoveredFunctions(packagePath string, result *CoverageValidationResult) {
	// This is a simplified implementation
	// In a real implementation, you would analyze the coverage profile more deeply
	// and cross-reference with AST analysis to identify specific uncovered functions

	for fileName, coverage := range result.CoverageByFile {
		if coverage < result.Target {
			// Add a placeholder uncovered function
			// Real implementation would parse the source file and identify specific functions
			uncovered := UncoveredFunction{
				Function: "UnknownFunction",
				File:     fileName,
				Priority: cv.calculateFunctionPriority(coverage),
			}
			result.UncoveredFunctions = append(result.UncoveredFunctions, uncovered)
		}
	}
}

// calculateFunctionPriority determines the priority for covering a function
func (cv *CoverageValidator) calculateFunctionPriority(coverage float64) string {
	if coverage < 30 {
		return "high"
	} else if coverage < 60 {
		return "medium"
	}
	return "low"
}

// generateRecommendations generates recommendations for improving test coverage
func (cv *CoverageValidator) generateRecommendations(result *CoverageValidationResult) {
	if result.CoveragePercentage < 50 {
		result.Recommendations = append(result.Recommendations,
			"Critical: Very low test coverage. Focus on testing core functionality first.")
	} else if result.CoveragePercentage < result.Target {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Coverage is %.1f%%, target is %.1f%%. Focus on uncovered functions.",
				result.CoveragePercentage, result.Target))
	}

	if !result.QualityMetrics.HasTableDrivenTests {
		result.Recommendations = append(result.Recommendations,
			"Consider using table-driven tests for better test organization and coverage.")
	}

	if result.QualityMetrics.TestToCodeRatio < 0.3 {
		result.Recommendations = append(result.Recommendations,
			"Low test-to-code ratio. Consider adding more comprehensive tests.")
	}

	if len(result.UncoveredFunctions) > 0 {
		highPriority := 0
		for _, fn := range result.UncoveredFunctions {
			if fn.Priority == "high" {
				highPriority++
			}
		}
		if highPriority > 0 {
			result.Recommendations = append(result.Recommendations,
				fmt.Sprintf("Focus on %d high-priority uncovered functions first.", highPriority))
		}
	}

	if result.QualityMetrics.TestQualityScore < 70 {
		result.Recommendations = append(result.Recommendations,
			"Consider improving test quality with benchmarks and examples.")
	}
}

// ValidatePackageCoverage validates coverage for an entire package
func (cv *CoverageValidator) ValidatePackageCoverage(packagePath string, targetCoverage float64) (*testdiscovery.CoverageReport, error) {
	validationResult, err := cv.ValidateTestCoverage(packagePath, targetCoverage)
	if err != nil {
		return nil, err
	}

	// Convert to CoverageReport format
	report := &testdiscovery.CoverageReport{
		PackagePath:     packagePath,
		TotalLines:      validationResult.TotalLines,
		CoveredLines:    validationResult.CoveredLines,
		CoveragePercent: validationResult.CoveragePercentage,
		CoverageByFile:  validationResult.CoverageByFile,
		TestFiles:       validationResult.TestFiles,
		GeneratedAt:     validationResult.GeneratedAt,
	}

	// Find files with insufficient coverage
	for fileName, coverage := range validationResult.CoverageByFile {
		if coverage < targetCoverage {
			report.UncoveredFiles = append(report.UncoveredFiles, fileName)
		}
	}

	return report, nil
}

// Cleanup removes temporary coverage files
func (cv *CoverageValidator) Cleanup() error {
	for _, coverageFile := range cv.coverageFiles {
		if err := os.Remove(coverageFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to cleanup coverage file %s: %v", coverageFile, err)
		}
	}
	cv.coverageFiles = make(map[string]string)
	return nil
}
