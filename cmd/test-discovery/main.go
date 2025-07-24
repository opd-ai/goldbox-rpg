// Test Discovery System - AI-powered Go test generation and coverage analysis
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"goldbox-rpg/pkg/testdiscovery"
	"goldbox-rpg/pkg/testdiscovery/analyzer"
	"goldbox-rpg/pkg/testdiscovery/generator"
	"goldbox-rpg/pkg/testdiscovery/selector"
)

// CommandLineOptions represents the CLI configuration
type CommandLineOptions struct {
	ProjectRoot      string  `json:"project_root"`
	OutputFormat     string  `json:"output_format"` // "json", "text", "html"
	MaxCandidates    int     `json:"max_candidates"`
	TargetCoverage   float64 `json:"target_coverage"`
	GenerateTests    bool    `json:"generate_tests"`
	ValidateCoverage bool    `json:"validate_coverage"`
	OutputFile       string  `json:"output_file"`
	ConfigFile       string  `json:"config_file"`
	Verbose          bool    `json:"verbose"`
	ProjectSize      string  `json:"project_size"`     // "small", "medium", "large"
	TestingGoal      string  `json:"testing_goal"`     // "coverage", "quality", "utility"
	ExcludePatterns  string  `json:"exclude_patterns"` // Comma-separated patterns
	IncludePatterns  string  `json:"include_patterns"` // Comma-separated patterns
}

// DiscoveryEngine coordinates the entire test discovery and generation process
type DiscoveryEngine struct {
	options           CommandLineOptions
	fileScanner       *analyzer.FileScanner
	metricsCollector  *analyzer.MetricsCollector
	priorityRanker    *selector.PriorityRanker
	testBuilder       *generator.TestBuilder
	coverageValidator *generator.CoverageValidator
	startTime         time.Time
}

// NewDiscoveryEngine creates a new discovery engine with the given options
func NewDiscoveryEngine(options CommandLineOptions) *DiscoveryEngine {
	return &DiscoveryEngine{
		options:           options,
		fileScanner:       analyzer.NewFileScanner(options.ProjectRoot),
		priorityRanker:    selector.NewPriorityRanker(),
		coverageValidator: generator.NewCoverageValidator(options.ProjectRoot),
		startTime:         time.Now(),
	}
}

// Execute runs the complete test discovery and generation process
func (de *DiscoveryEngine) Execute() error {
	de.logVerbose("Starting test discovery system...")

	// Phase 1: Enhanced File Discovery
	de.logVerbose("Phase 1: Discovering and analyzing Go files...")
	files, err := de.fileScanner.ScanDirectory()
	if err != nil {
		return fmt.Errorf("file discovery failed: %v", err)
	}
	de.logVerbose(fmt.Sprintf("Discovered %d Go files", len(files)))

	// Initialize metrics collector with file set from scanner
	de.metricsCollector = analyzer.NewMetricsCollector(de.fileScanner.GetFileSet())

	// Phase 2: Comprehensive Metrics Collection
	de.logVerbose("Phase 2: Collecting comprehensive metrics...")
	projectMetrics := de.metricsCollector.CollectProjectMetrics(files)
	de.logVerbose(fmt.Sprintf("Analyzed %d files, %d with tests, %d without tests",
		projectMetrics.TotalFiles, projectMetrics.FilesWithTests, projectMetrics.FilesWithoutTests))

	// Phase 3: Intelligent File Selection
	de.logVerbose("Phase 3: Ranking files for test generation...")

	// Configure criteria based on project characteristics
	if err := de.configureCriteria(projectMetrics); err != nil {
		return fmt.Errorf("criteria configuration failed: %v", err)
	}

	// Rank files
	scores := de.priorityRanker.RankFiles(files)
	de.logVerbose(fmt.Sprintf("Ranked %d files, found %d candidates", len(scores), de.countCandidates(scores)))

	// Phase 4: Test Generation (if requested)
	var generationResults []*testdiscovery.TestGenerationResult
	if de.options.GenerateTests {
		de.logVerbose("Phase 4: Generating tests for selected files...")
		generationResults, err = de.generateTests(files, scores)
		if err != nil {
			return fmt.Errorf("test generation failed: %v", err)
		}
		de.logVerbose(fmt.Sprintf("Generated tests for %d files", len(generationResults)))
	}

	// Phase 5: Coverage Validation (if requested)
	var coverageResults []*generator.CoverageValidationResult
	if de.options.ValidateCoverage {
		de.logVerbose("Phase 5: Validating test coverage...")
		coverageResults, err = de.validateCoverage(files)
		if err != nil {
			return fmt.Errorf("coverage validation failed: %v", err)
		}
		de.logVerbose(fmt.Sprintf("Validated coverage for %d packages", len(coverageResults)))
	}

	// Phase 6: Generate Report
	de.logVerbose("Phase 6: Generating analysis report...")
	report := de.generateReport(files, projectMetrics, scores, generationResults, coverageResults)

	// Output results
	if err := de.outputResults(report); err != nil {
		return fmt.Errorf("output generation failed: %v", err)
	}

	de.logVerbose(fmt.Sprintf("Test discovery completed in %v", time.Since(de.startTime)))
	return nil
}

// configureCriteria configures selection criteria based on project characteristics
func (de *DiscoveryEngine) configureCriteria(metrics *analyzer.ProjectMetrics) error {
	validator := selector.NewCriteriaValidator()

	// Determine project size if not specified
	projectSize := de.options.ProjectSize
	if projectSize == "" {
		switch {
		case metrics.TotalFiles < 50:
			projectSize = "small"
		case metrics.TotalFiles < 200:
			projectSize = "medium"
		default:
			projectSize = "large"
		}
	}

	// Get suggested criteria
	criteria, weights := validator.SuggestOptimalCriteria(projectSize, de.options.TestingGoal)

	// Apply any custom exclusion patterns
	if de.options.ExcludePatterns != "" {
		// This would be implemented to modify criteria based on patterns
		de.logVerbose(fmt.Sprintf("Applying exclusion patterns: %s", de.options.ExcludePatterns))
	}

	// Update the priority ranker
	de.priorityRanker.UpdateCriteria(criteria)
	de.priorityRanker.UpdateWeights(weights)

	return nil
}

// countCandidates counts non-excluded candidates
func (de *DiscoveryEngine) countCandidates(scores []testdiscovery.FileScore) int {
	count := 0
	for _, score := range scores {
		if !score.IsExcluded {
			count++
		}
	}
	return count
}

// generateTests generates tests for selected files
func (de *DiscoveryEngine) generateTests(files map[string]*testdiscovery.FileInfo, scores []testdiscovery.FileScore) ([]*testdiscovery.TestGenerationResult, error) {
	candidates := de.priorityRanker.GetTopCandidates(scores, de.options.MaxCandidates)
	var results []*testdiscovery.TestGenerationResult

	for i, candidate := range candidates {
		if i >= de.options.MaxCandidates {
			break
		}

		fileInfo := files[candidate.FilePath]
		if fileInfo == nil {
			continue
		}

		de.logVerbose(fmt.Sprintf("Generating tests for %s (score: %.1f)", candidate.FilePath, candidate.TotalScore))

		// Create test builder for this file's package
		testBuilder := generator.NewTestBuilder(fileInfo.PackageName)

		// Configure builder options
		options := generator.DefaultBuilderOptions()
		options.CoverageTarget = de.options.TargetCoverage
		testBuilder.SetOptions(options)

		// Generate output path
		outputPath := de.generateTestOutputPath(candidate.FilePath)

		// Generate test file
		result, err := testBuilder.GenerateTestFile(fileInfo, outputPath)
		if err != nil {
			de.logVerbose(fmt.Sprintf("Failed to generate tests for %s: %v", candidate.FilePath, err))
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

// generateTestOutputPath generates the output path for a test file
func (de *DiscoveryEngine) generateTestOutputPath(sourceFilePath string) string {
	dir := filepath.Dir(sourceFilePath)
	baseName := strings.TrimSuffix(filepath.Base(sourceFilePath), ".go")
	testFileName := baseName + "_test.go"
	return filepath.Join(dir, testFileName)
}

// validateCoverage validates test coverage for packages
func (de *DiscoveryEngine) validateCoverage(files map[string]*testdiscovery.FileInfo) ([]*generator.CoverageValidationResult, error) {
	// Group files by package
	packages := make(map[string]bool)
	for _, fileInfo := range files {
		packages[fileInfo.PackageName] = true
	}

	var results []*generator.CoverageValidationResult

	for packageName := range packages {
		de.logVerbose(fmt.Sprintf("Validating coverage for package: %s", packageName))

		result, err := de.coverageValidator.ValidateTestCoverage(packageName, de.options.TargetCoverage)
		if err != nil {
			de.logVerbose(fmt.Sprintf("Coverage validation failed for %s: %v", packageName, err))
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

// generateReport generates a comprehensive analysis report
func (de *DiscoveryEngine) generateReport(
	files map[string]*testdiscovery.FileInfo,
	metrics *analyzer.ProjectMetrics,
	scores []testdiscovery.FileScore,
	generationResults []*testdiscovery.TestGenerationResult,
	coverageResults []*generator.CoverageValidationResult,
) *AnalysisReport {

	report := &AnalysisReport{
		GeneratedAt:       time.Now(),
		ExecutionTime:     time.Since(de.startTime),
		ProjectRoot:       de.options.ProjectRoot,
		ProjectMetrics:    metrics,
		SelectionReport:   de.priorityRanker.GenerateSelectionReport(scores),
		TopCandidates:     de.priorityRanker.GetTopCandidates(scores, 10),
		GenerationResults: generationResults,
		CoverageResults:   coverageResults,
		Options:           de.options,
	}

	// Generate summary statistics
	report.Summary = de.generateSummary(report)

	// Generate recommendations
	report.Recommendations = de.generateRecommendations(report)

	return report
}

// AnalysisReport represents the complete analysis report
type AnalysisReport struct {
	GeneratedAt       time.Time                             `json:"generated_at"`
	ExecutionTime     time.Duration                         `json:"execution_time"`
	ProjectRoot       string                                `json:"project_root"`
	ProjectMetrics    *analyzer.ProjectMetrics              `json:"project_metrics"`
	SelectionReport   selector.SelectionReport              `json:"selection_report"`
	TopCandidates     []testdiscovery.FileScore             `json:"top_candidates"`
	GenerationResults []*testdiscovery.TestGenerationResult `json:"generation_results"`
	CoverageResults   []*generator.CoverageValidationResult `json:"coverage_results"`
	Summary           AnalysisSummary                       `json:"summary"`
	Recommendations   []string                              `json:"recommendations"`
	Options           CommandLineOptions                    `json:"options"`
}

// AnalysisSummary provides high-level summary statistics
type AnalysisSummary struct {
	TotalFilesAnalyzed     int     `json:"total_files_analyzed"`
	FilesWithTests         int     `json:"files_with_tests"`
	FilesWithoutTests      int     `json:"files_without_tests"`
	CandidateFiles         int     `json:"candidate_files"`
	TestsGenerated         int     `json:"tests_generated"`
	AverageCoverage        float64 `json:"average_coverage"`
	HighPriorityCandidates int     `json:"high_priority_candidates"`
	EstimatedTestLines     int     `json:"estimated_test_lines"`
}

// generateSummary generates summary statistics
func (de *DiscoveryEngine) generateSummary(report *AnalysisReport) AnalysisSummary {
	summary := AnalysisSummary{
		TotalFilesAnalyzed: report.ProjectMetrics.TotalFiles,
		FilesWithTests:     report.ProjectMetrics.FilesWithTests,
		FilesWithoutTests:  report.ProjectMetrics.FilesWithoutTests,
		CandidateFiles:     report.SelectionReport.CandidateFiles,
		TestsGenerated:     len(report.GenerationResults),
	}

	// Count high priority candidates
	for _, candidate := range report.TopCandidates {
		if candidate.TotalScore >= 80 {
			summary.HighPriorityCandidates++
		}
	}

	// Calculate average coverage from coverage results
	if len(report.CoverageResults) > 0 {
		totalCoverage := 0.0
		for _, result := range report.CoverageResults {
			totalCoverage += result.CoveragePercentage
		}
		summary.AverageCoverage = totalCoverage / float64(len(report.CoverageResults))
	}

	// Estimate test lines from generation results
	for _, result := range report.GenerationResults {
		summary.EstimatedTestLines += result.GeneratedLines
	}

	return summary
}

// generateRecommendations generates actionable recommendations
func (de *DiscoveryEngine) generateRecommendations(report *AnalysisReport) []string {
	var recommendations []string

	// Coverage-based recommendations
	testCoveragePercent := float64(report.Summary.FilesWithTests) / float64(report.Summary.TotalFilesAnalyzed) * 100
	if testCoveragePercent < 50 {
		recommendations = append(recommendations,
			fmt.Sprintf("Critical: Only %.1f%% of files have tests. Prioritize test creation for high-value files.", testCoveragePercent))
	}

	// Selection-based recommendations
	if report.Summary.CandidateFiles < 10 {
		recommendations = append(recommendations,
			"Few files meet selection criteria. Consider relaxing constraints to increase test generation opportunities.")
	}

	// Generation-based recommendations
	if de.options.GenerateTests && report.Summary.TestsGenerated == 0 {
		recommendations = append(recommendations,
			"No tests were generated. Review selection criteria and ensure target files are testable.")
	}

	// Coverage-based recommendations
	if report.Summary.AverageCoverage < de.options.TargetCoverage {
		recommendations = append(recommendations,
			fmt.Sprintf("Average coverage (%.1f%%) is below target (%.1f%%). Focus on improving test comprehensiveness.",
				report.Summary.AverageCoverage, de.options.TargetCoverage))
	}

	// High-priority recommendations
	if report.Summary.HighPriorityCandidates > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Focus on %d high-priority candidates for maximum testing impact.",
				report.Summary.HighPriorityCandidates))
	}

	return recommendations
}

// outputResults outputs the analysis results in the specified format
func (de *DiscoveryEngine) outputResults(report *AnalysisReport) error {
	switch de.options.OutputFormat {
	case "json":
		return de.outputJSON(report)
	case "text":
		return de.outputText(report)
	case "html":
		return de.outputHTML(report)
	default:
		return fmt.Errorf("unsupported output format: %s", de.options.OutputFormat)
	}
}

// outputJSON outputs results in JSON format
func (de *DiscoveryEngine) outputJSON(report *AnalysisReport) error {
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON marshaling failed: %v", err)
	}

	if de.options.OutputFile != "" {
		return os.WriteFile(de.options.OutputFile, jsonData, 0644)
	}

	fmt.Println(string(jsonData))
	return nil
}

// outputText outputs results in human-readable text format
func (de *DiscoveryEngine) outputText(report *AnalysisReport) error {
	output := de.generateTextReport(report)

	if de.options.OutputFile != "" {
		return os.WriteFile(de.options.OutputFile, []byte(output), 0644)
	}

	fmt.Print(output)
	return nil
}

// generateTextReport generates a human-readable text report
func (de *DiscoveryEngine) generateTextReport(report *AnalysisReport) string {
	var output strings.Builder

	output.WriteString("=== Go Test Coverage Discovery Report ===\n\n")
	output.WriteString(fmt.Sprintf("Generated: %s\n", report.GeneratedAt.Format(time.RFC3339)))
	output.WriteString(fmt.Sprintf("Execution Time: %v\n", report.ExecutionTime))
	output.WriteString(fmt.Sprintf("Project Root: %s\n\n", report.ProjectRoot))

	// Summary Section
	output.WriteString("=== SUMMARY ===\n")
	output.WriteString(fmt.Sprintf("Total Files Analyzed: %d\n", report.Summary.TotalFilesAnalyzed))
	output.WriteString(fmt.Sprintf("Files With Tests: %d\n", report.Summary.FilesWithTests))
	output.WriteString(fmt.Sprintf("Files Without Tests: %d\n", report.Summary.FilesWithoutTests))
	output.WriteString(fmt.Sprintf("Candidate Files: %d\n", report.Summary.CandidateFiles))
	output.WriteString(fmt.Sprintf("Tests Generated: %d\n", report.Summary.TestsGenerated))
	output.WriteString(fmt.Sprintf("High Priority Candidates: %d\n", report.Summary.HighPriorityCandidates))
	if report.Summary.AverageCoverage > 0 {
		output.WriteString(fmt.Sprintf("Average Coverage: %.1f%%\n", report.Summary.AverageCoverage))
	}
	output.WriteString("\n")

	// Top Candidates Section
	if len(report.TopCandidates) > 0 {
		output.WriteString("=== TOP CANDIDATES FOR TESTING ===\n")
		for i, candidate := range report.TopCandidates {
			if i >= 10 { // Limit to top 10
				break
			}
			output.WriteString(fmt.Sprintf("%d. %s (Score: %.1f) - %s\n",
				i+1, candidate.FilePath, candidate.TotalScore, candidate.SelectionReason))
		}
		output.WriteString("\n")
	}

	// Recommendations Section
	if len(report.Recommendations) > 0 {
		output.WriteString("=== RECOMMENDATIONS ===\n")
		for i, rec := range report.Recommendations {
			output.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
		}
		output.WriteString("\n")
	}

	// Generation Results Section
	if len(report.GenerationResults) > 0 {
		output.WriteString("=== TEST GENERATION RESULTS ===\n")
		for _, result := range report.GenerationResults {
			output.WriteString(fmt.Sprintf("File: %s\n", result.FilePath))
			output.WriteString(fmt.Sprintf("  Tests Generated: %d\n", result.TestCount))
			output.WriteString(fmt.Sprintf("  Lines Generated: %d\n", result.GeneratedLines))
			output.WriteString(fmt.Sprintf("  Success: %t\n", result.Success))
			if len(result.Warnings) > 0 {
				output.WriteString(fmt.Sprintf("  Warnings: %s\n", strings.Join(result.Warnings, ", ")))
			}
			output.WriteString("\n")
		}
	}

	return output.String()
}

// outputHTML outputs results in HTML format
func (de *DiscoveryEngine) outputHTML(report *AnalysisReport) error {
	html := de.generateHTMLReport(report)

	if de.options.OutputFile != "" {
		return os.WriteFile(de.options.OutputFile, []byte(html), 0644)
	}

	fmt.Print(html)
	return nil
}

// generateHTMLReport generates an HTML report
func (de *DiscoveryEngine) generateHTMLReport(report *AnalysisReport) string {
	// This is a simplified HTML report
	// In a full implementation, you'd use templates and more sophisticated HTML/CSS
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Go Test Coverage Discovery Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .summary { background-color: #f5f5f5; padding: 20px; border-radius: 5px; }
        .candidates { margin-top: 20px; }
        .candidate { margin: 10px 0; padding: 10px; border-left: 4px solid #007cba; }
        .recommendations { background-color: #fff3cd; padding: 15px; border-radius: 5px; margin-top: 20px; }
    </style>
</head>
<body>
    <h1>Go Test Coverage Discovery Report</h1>
    <div class="summary">
        <h2>Summary</h2>
        <p>Generated: ` + report.GeneratedAt.Format(time.RFC3339) + `</p>
        <p>Execution Time: ` + report.ExecutionTime.String() + `</p>
        <p>Total Files: ` + fmt.Sprintf("%d", report.Summary.TotalFilesAnalyzed) + `</p>
        <p>Files With Tests: ` + fmt.Sprintf("%d", report.Summary.FilesWithTests) + `</p>
        <p>Candidate Files: ` + fmt.Sprintf("%d", report.Summary.CandidateFiles) + `</p>
    </div>`

	if len(report.TopCandidates) > 0 {
		html += `
    <div class="candidates">
        <h2>Top Candidates</h2>`
		for i, candidate := range report.TopCandidates {
			if i >= 10 {
				break
			}
			html += fmt.Sprintf(`
        <div class="candidate">
            <strong>%s</strong> (Score: %.1f)<br>
            <small>%s</small>
        </div>`, candidate.FilePath, candidate.TotalScore, candidate.SelectionReason)
		}
		html += `
    </div>`
	}

	if len(report.Recommendations) > 0 {
		html += `
    <div class="recommendations">
        <h2>Recommendations</h2>
        <ul>`
		for _, rec := range report.Recommendations {
			html += fmt.Sprintf(`<li>%s</li>`, rec)
		}
		html += `
        </ul>
    </div>`
	}

	html += `
</body>
</html>`

	return html
}

// logVerbose logs verbose messages if verbose mode is enabled
func (de *DiscoveryEngine) logVerbose(message string) {
	if de.options.Verbose {
		log.Printf("[VERBOSE] %s", message)
	}
}

// main is the entry point for the test discovery CLI
func main() {
	var options CommandLineOptions

	// Define command line flags
	flag.StringVar(&options.ProjectRoot, "root", ".", "Project root directory")
	flag.StringVar(&options.OutputFormat, "format", "text", "Output format (json, text, html)")
	flag.IntVar(&options.MaxCandidates, "max", 10, "Maximum number of test candidates")
	flag.Float64Var(&options.TargetCoverage, "coverage", 80.0, "Target test coverage percentage")
	flag.BoolVar(&options.GenerateTests, "generate", false, "Generate test files")
	flag.BoolVar(&options.ValidateCoverage, "validate", false, "Validate existing test coverage")
	flag.StringVar(&options.OutputFile, "output", "", "Output file path")
	flag.StringVar(&options.ConfigFile, "config", "", "Configuration file path")
	flag.BoolVar(&options.Verbose, "verbose", false, "Enable verbose logging")
	flag.StringVar(&options.ProjectSize, "size", "", "Project size hint (small, medium, large)")
	flag.StringVar(&options.TestingGoal, "goal", "coverage", "Testing goal (coverage, quality, utility)")
	flag.StringVar(&options.ExcludePatterns, "exclude", "", "Comma-separated exclusion patterns")
	flag.StringVar(&options.IncludePatterns, "include", "", "Comma-separated inclusion patterns")

	flag.Parse()

	// Create and execute discovery engine
	engine := NewDiscoveryEngine(options)
	if err := engine.Execute(); err != nil {
		log.Fatalf("Test discovery failed: %v", err)
	}
}
