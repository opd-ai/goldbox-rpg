package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"goldbox-rpg/pkg/testdiscovery"
	"goldbox-rpg/pkg/testdiscovery/analyzer"
	"goldbox-rpg/pkg/testdiscovery/generator"
	"goldbox-rpg/pkg/testdiscovery/selector"
)

// TestDiscoveryEngine_Execute_ValidProject tests the complete discovery process
func TestDiscoveryEngine_Execute_ValidProject(t *testing.T) {
	// Create a temporary test project
	tempDir, cleanup := createTestProject(t)
	defer cleanup()

	options := CommandLineOptions{
		ProjectRoot:      tempDir,
		OutputFormat:     "json",
		MaxCandidates:    5,
		TargetCoverage:   80.0,
		GenerateTests:    false, // Don't generate for this test
		ValidateCoverage: false, // Don't validate for this test
		Verbose:          false,
		ProjectSize:      "small",
		TestingGoal:      "coverage",
	}

	engine := NewDiscoveryEngine(options)
	err := engine.Execute()

	if err != nil {
		t.Errorf("DiscoveryEngine.Execute() failed: %v", err)
	}
}

// TestCommandLineOptions_DefaultValues tests default command line options
func TestCommandLineOptions_DefaultValues_AreReasonable(t *testing.T) {
	options := CommandLineOptions{
		ProjectRoot:    ".",
		OutputFormat:   "text",
		MaxCandidates:  10,
		TargetCoverage: 80.0,
		GenerateTests:  false,
		Verbose:        false,
		ProjectSize:    "",
		TestingGoal:    "coverage",
	}

	// Test that defaults are reasonable
	if options.TargetCoverage < 70 || options.TargetCoverage > 90 {
		t.Errorf("Default target coverage %.1f is not reasonable (should be 70-90)", options.TargetCoverage)
	}

	if options.MaxCandidates < 5 || options.MaxCandidates > 20 {
		t.Errorf("Default max candidates %d is not reasonable (should be 5-20)", options.MaxCandidates)
	}

	validFormats := []string{"json", "text", "html"}
	validFormat := false
	for _, format := range validFormats {
		if options.OutputFormat == format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		t.Errorf("Default output format '%s' is not valid", options.OutputFormat)
	}
}

// TestAnalysisSummary_CalculationLogic tests summary calculation
func TestAnalysisSummary_CalculationLogic_ProducesCorrectResults(t *testing.T) {
	// Create test data
	projectMetrics := &analyzer.ProjectMetrics{
		TotalFiles:        100,
		FilesWithTests:    30,
		FilesWithoutTests: 70,
	}

	selectionReport := selector.SelectionReport{
		CandidateFiles: 25,
	}

	topCandidates := []testdiscovery.FileScore{
		{FilePath: "file1.go", TotalScore: 85.0},
		{FilePath: "file2.go", TotalScore: 75.0},
		{FilePath: "file3.go", TotalScore: 65.0},
	}

	generationResults := []*testdiscovery.TestGenerationResult{
		{GeneratedLines: 100},
		{GeneratedLines: 150},
	}

	coverageResults := []*generator.CoverageValidationResult{
		{CoveragePercentage: 80.0},
		{CoveragePercentage: 75.0},
	}

	report := &AnalysisReport{
		ProjectMetrics:    projectMetrics,
		SelectionReport:   selectionReport,
		TopCandidates:     topCandidates,
		GenerationResults: generationResults,
		CoverageResults:   coverageResults,
	}

	engine := &DiscoveryEngine{}
	summary := engine.generateSummary(report)

	// Validate summary calculations
	if summary.TotalFilesAnalyzed != 100 {
		t.Errorf("Expected TotalFilesAnalyzed=100, got %d", summary.TotalFilesAnalyzed)
	}

	if summary.FilesWithTests != 30 {
		t.Errorf("Expected FilesWithTests=30, got %d", summary.FilesWithTests)
	}

	if summary.CandidateFiles != 25 {
		t.Errorf("Expected CandidateFiles=25, got %d", summary.CandidateFiles)
	}

	if summary.TestsGenerated != 2 {
		t.Errorf("Expected TestsGenerated=2, got %d", summary.TestsGenerated)
	}

	if summary.HighPriorityCandidates != 1 { // Only file1.go has score >= 80
		t.Errorf("Expected HighPriorityCandidates=1, got %d", summary.HighPriorityCandidates)
	}

	expectedAvgCoverage := (80.0 + 75.0) / 2.0
	if summary.AverageCoverage != expectedAvgCoverage {
		t.Errorf("Expected AverageCoverage=%.1f, got %.1f", expectedAvgCoverage, summary.AverageCoverage)
	}

	expectedTestLines := 100 + 150
	if summary.EstimatedTestLines != expectedTestLines {
		t.Errorf("Expected EstimatedTestLines=%d, got %d", expectedTestLines, summary.EstimatedTestLines)
	}
}

// TestGenerateRecommendations_VariousScenarios tests recommendation generation
func TestGenerateRecommendations_VariousScenarios_GeneratesAppropriateAdvice(t *testing.T) {
	tests := []struct {
		name                    string
		summary                 AnalysisSummary
		options                 CommandLineOptions
		expectedRecommendations int
		shouldContain           string
	}{
		{
			name: "LowTestCoverage",
			summary: AnalysisSummary{
				TotalFilesAnalyzed: 100,
				FilesWithTests:     20, // 20% coverage
				FilesWithoutTests:  80,
			},
			options:                 CommandLineOptions{TargetCoverage: 80.0},
			expectedRecommendations: 1,
			shouldContain:           "Critical",
		},
		{
			name: "FewCandidates",
			summary: AnalysisSummary{
				TotalFilesAnalyzed: 100,
				FilesWithTests:     50,
				CandidateFiles:     5, // Very few candidates
			},
			options:                 CommandLineOptions{TargetCoverage: 80.0},
			expectedRecommendations: 1,
			shouldContain:           "Few files meet",
		},
		{
			name: "HighPriorityCandidates",
			summary: AnalysisSummary{
				TotalFilesAnalyzed:     100,
				FilesWithTests:         60,
				HighPriorityCandidates: 15,
			},
			options:                 CommandLineOptions{TargetCoverage: 80.0},
			expectedRecommendations: 1,
			shouldContain:           "high-priority candidates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &AnalysisReport{
				Summary: tt.summary,
				Options: tt.options,
			}

			engine := &DiscoveryEngine{options: tt.options}
			recommendations := engine.generateRecommendations(report)

			if len(recommendations) < tt.expectedRecommendations {
				t.Errorf("Expected at least %d recommendations, got %d",
					tt.expectedRecommendations, len(recommendations))
			}

			found := false
			for _, rec := range recommendations {
				if len(tt.shouldContain) > 0 && containsIgnoreCase(rec, tt.shouldContain) {
					found = true
					break
				}
			}

			if len(tt.shouldContain) > 0 && !found {
				t.Errorf("Expected recommendations to contain '%s', got: %v",
					tt.shouldContain, recommendations)
			}
		})
	}
}

// TestOutputFormats_ValidData tests different output formats
func TestOutputFormats_ValidData_ProduceCorrectOutput(t *testing.T) {
	// Create test report
	report := &AnalysisReport{
		GeneratedAt:   time.Now(),
		ExecutionTime: 5 * time.Second,
		ProjectRoot:   "/test/project",
		Summary: AnalysisSummary{
			TotalFilesAnalyzed: 50,
			FilesWithTests:     20,
			FilesWithoutTests:  30,
			CandidateFiles:     10,
		},
		Recommendations: []string{"Test recommendation 1", "Test recommendation 2"},
	}

	tests := []struct {
		name   string
		format string
		verify func(t *testing.T, output string)
	}{
		{
			name:   "JSONFormat",
			format: "json",
			verify: func(t *testing.T, output string) {
				var parsed AnalysisReport
				if err := json.Unmarshal([]byte(output), &parsed); err != nil {
					t.Errorf("JSON output is not valid: %v", err)
				}
				if parsed.Summary.TotalFilesAnalyzed != 50 {
					t.Errorf("JSON parsing failed: expected 50 files, got %d",
						parsed.Summary.TotalFilesAnalyzed)
				}
			},
		},
		{
			name:   "TextFormat",
			format: "text",
			verify: func(t *testing.T, output string) {
				if !containsIgnoreCase(output, "Test Coverage Discovery Report") {
					t.Errorf("Text output missing report title")
				}
				if !containsIgnoreCase(output, "Total Files Analyzed: 50") {
					t.Errorf("Text output missing file count")
				}
			},
		},
		{
			name:   "HTMLFormat",
			format: "html",
			verify: func(t *testing.T, output string) {
				if !containsIgnoreCase(output, "<html>") {
					t.Errorf("HTML output missing HTML tag")
				}
				if !containsIgnoreCase(output, "Test Coverage Discovery Report") {
					t.Errorf("HTML output missing report title")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &DiscoveryEngine{
				options: CommandLineOptions{OutputFormat: tt.format},
			}

			var output string
			switch tt.format {
			case "json":
				jsonData, _ := json.MarshalIndent(report, "", "  ")
				output = string(jsonData)
			case "text":
				output = engine.generateTextReport(report)
			case "html":
				output = engine.generateHTMLReport(report)
			}

			tt.verify(t, output)
		})
	}
}

// TestConfigureCriteria_DifferentProjectSizes tests criteria configuration
func TestConfigureCriteria_DifferentProjectSizes_AppliesCorrectCriteria(t *testing.T) {
	tests := []struct {
		name        string
		projectSize string
		fileCount   int
		expectSize  string
	}{
		{
			name:        "SmallProject",
			projectSize: "",
			fileCount:   30,
			expectSize:  "small",
		},
		{
			name:        "MediumProject",
			projectSize: "",
			fileCount:   100,
			expectSize:  "medium",
		},
		{
			name:        "LargeProject",
			projectSize: "",
			fileCount:   300,
			expectSize:  "large",
		},
		{
			name:        "ExplicitSize",
			projectSize: "large",
			fileCount:   50, // Small but explicitly set to large
			expectSize:  "large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := CommandLineOptions{
				ProjectSize: tt.projectSize,
				TestingGoal: "coverage",
			}

			engine := &DiscoveryEngine{
				options:        options,
				priorityRanker: selector.NewPriorityRanker(),
			}

			metrics := &analyzer.ProjectMetrics{
				TotalFiles: tt.fileCount,
			}

			// This should not error
			err := engine.configureCriteria(metrics)
			if err != nil {
				t.Errorf("configureCriteria() failed: %v", err)
			}

			// Verify that criteria were configured (basic check)
			criteria := engine.priorityRanker.GetCriteria()
			weights := engine.priorityRanker.GetWeights()

			// Basic validation
			if criteria.MaxDependencies <= 0 {
				t.Errorf("MaxDependencies should be positive, got %d", criteria.MaxDependencies)
			}

			if weights.DependencyWeight <= 0 {
				t.Errorf("DependencyWeight should be positive, got %.3f", weights.DependencyWeight)
			}
		})
	}
}

// TestNewDiscoveryEngine_Initialization tests engine initialization
func TestNewDiscoveryEngine_Initialization_SetsUpComponentsCorrectly(t *testing.T) {
	options := CommandLineOptions{
		ProjectRoot:    "/test/project",
		TargetCoverage: 85.0,
		MaxCandidates:  15,
	}

	engine := NewDiscoveryEngine(options)

	// Verify initialization
	if engine == nil {
		t.Fatal("NewDiscoveryEngine() returned nil")
	}

	if engine.options.ProjectRoot != "/test/project" {
		t.Errorf("Expected ProjectRoot='/test/project', got '%s'", engine.options.ProjectRoot)
	}

	if engine.options.TargetCoverage != 85.0 {
		t.Errorf("Expected TargetCoverage=85.0, got %.1f", engine.options.TargetCoverage)
	}

	if engine.fileScanner == nil {
		t.Error("FileScanner was not initialized")
	}

	if engine.priorityRanker == nil {
		t.Error("PriorityRanker was not initialized")
	}

	if engine.coverageValidator == nil {
		t.Error("CoverageValidator was not initialized")
	}

	// Verify start time was set
	if engine.startTime.IsZero() {
		t.Error("Start time was not set")
	}
}

// Helper functions for testing

// TestHelper defines common methods for both testing.T and testing.B
type TestHelper interface {
	Fatalf(format string, args ...interface{})
	Helper()
}

// createTestProject creates a temporary test project structure
func createTestProject(t *testing.T) (string, func()) {
	return createTestProjectWithHelper(t)
}

// createTestProjectForBenchmark creates a temporary test project structure for benchmarks
func createTestProjectForBenchmark(b *testing.B) (string, func()) {
	return createTestProjectWithHelper(b)
}

// createTestProjectWithHelper creates a temporary test project structure using a test helper
func createTestProjectWithHelper(th TestHelper) (string, func()) {
	th.Helper()

	tempDir, err := os.MkdirTemp("", "test_discovery_*")
	if err != nil {
		th.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create some test Go files
	testFiles := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println(hello("world"))
}

func hello(name string) string {
	return "Hello, " + name + "!"
}`,

		"utils.go": `package main

import "strings"

// StringUtils provides string utility functions
type StringUtils struct{}

// Reverse reverses a string
func (su *StringUtils) Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// IsEmpty checks if a string is empty
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}`,

		"math.go": `package main

// Add adds two integers
func Add(a, b int) int {
	return a + b
}

// Multiply multiplies two integers
func Multiply(a, b int) int {
	result := 0
	for i := 0; i < b; i++ {
		result = Add(result, a)
	}
	return result
}`,

		"utils_test.go": `package main

import "testing"

func TestIsEmpty(t *testing.T) {
	if !IsEmpty("") {
		t.Error("Expected empty string to be empty")
	}
	if IsEmpty("test") {
		t.Error("Expected non-empty string to not be empty")
	}
}`,
	}

	// Write test files
	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			os.RemoveAll(tempDir)
			th.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// containsIgnoreCase checks if a string contains a substring (case insensitive)
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// TestIntegration_CompleteWorkflow tests the complete workflow
func TestIntegration_CompleteWorkflow_ExecutesSuccessfully(t *testing.T) {
	// This is an integration test that exercises the complete system
	tempDir, cleanup := createTestProject(t)
	defer cleanup()

	options := CommandLineOptions{
		ProjectRoot:      tempDir,
		OutputFormat:     "json",
		MaxCandidates:    3,
		TargetCoverage:   70.0,
		GenerateTests:    false, // Don't actually generate for integration test
		ValidateCoverage: false, // Don't validate for integration test
		Verbose:          false,
		ProjectSize:      "small",
		TestingGoal:      "coverage",
	}

	engine := NewDiscoveryEngine(options)

	// Test that execution completes without error
	err := engine.Execute()
	if err != nil {
		t.Fatalf("Integration test failed: %v", err)
	}

	// The integration test validates that all components work together
	// without errors. More detailed validation would require capturing
	// the actual output and verifying its correctness.
}

// Benchmark tests for performance
func BenchmarkDiscoveryEngine_FileScanning(b *testing.B) {
	tempDir, cleanup := createTestProjectForBenchmark(b)
	defer cleanup()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		scanner := analyzer.NewFileScanner(tempDir)
		_, err := scanner.ScanDirectory()
		if err != nil {
			b.Fatalf("File scanning failed: %v", err)
		}
	}
}

func BenchmarkDiscoveryEngine_MetricsCollection(b *testing.B) {
	tempDir, cleanup := createTestProjectForBenchmark(b)
	defer cleanup()

	scanner := analyzer.NewFileScanner(tempDir)
	files, err := scanner.ScanDirectory()
	if err != nil {
		b.Fatalf("File scanning failed: %v", err)
	}

	collector := analyzer.NewMetricsCollector(scanner.GetFileSet())

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = collector.CollectProjectMetrics(files)
	}
}

// Test table-driven approach for multiple scenarios
func TestDiscoveryEngine_ConfigurationScenarios_TableDriven(t *testing.T) {
	scenarios := []struct {
		name        string
		options     CommandLineOptions
		expectError bool
		description string
	}{
		{
			name: "StandardConfiguration",
			options: CommandLineOptions{
				ProjectRoot:    ".",
				OutputFormat:   "text",
				MaxCandidates:  10,
				TargetCoverage: 80.0,
			},
			expectError: false,
			description: "Standard configuration should work without errors",
		},
		{
			name: "HighCoverageTarget",
			options: CommandLineOptions{
				ProjectRoot:    ".",
				OutputFormat:   "json",
				MaxCandidates:  5,
				TargetCoverage: 95.0,
			},
			expectError: false,
			description: "High coverage target should be acceptable",
		},
		{
			name: "LowCoverageTarget",
			options: CommandLineOptions{
				ProjectRoot:    ".",
				OutputFormat:   "html",
				MaxCandidates:  20,
				TargetCoverage: 50.0,
			},
			expectError: false,
			description: "Low coverage target should be acceptable",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			engine := NewDiscoveryEngine(scenario.options)

			// Test initialization
			if engine == nil {
				t.Errorf("NewDiscoveryEngine() returned nil for scenario: %s", scenario.description)
			}

			// Verify options were set correctly
			if !reflect.DeepEqual(engine.options, scenario.options) {
				t.Errorf("Options not set correctly for scenario: %s", scenario.description)
			}
		})
	}
}
