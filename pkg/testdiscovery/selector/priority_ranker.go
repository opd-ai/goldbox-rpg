// Package selector provides intelligent file selection algorithms for test generation
package selector

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"goldbox-rpg/pkg/testdiscovery"
)

// PriorityRanker implements intelligent file prioritization for test generation
type PriorityRanker struct {
	criteria SelectionCriteria
	weights  ScoringWeights
}

// SelectionCriteria defines the criteria for file selection
type SelectionCriteria struct {
	MaxDependencies      int     // Maximum allowed dependencies
	MinComplexity        float64 // Minimum complexity threshold
	MaxComplexity        float64 // Maximum complexity threshold
	MinSize              int     // Minimum file size in lines
	MaxSize              int     // Maximum file size in lines
	RequireInterfaces    bool    // Prefer files with interfaces
	ExcludeNetworkIO     bool    // Exclude files with network I/O
	ExcludeDatabaseIO    bool    // Exclude files with database I/O
	ExcludeFileIO        bool    // Exclude files with file I/O
	MaxMockingComplexity int     // Maximum acceptable mocking complexity
	PreferUtilityFiles   bool    // Prioritize utility/helper files
	MinTestabilityScore  float64 // Minimum testability score
	MaxFileSize          int64   // Maximum file size in bytes
}

// ScoringWeights defines the weights for different scoring factors
type ScoringWeights struct {
	DependencyWeight  float64 // Weight for dependency score (default: 0.30)
	ComplexityWeight  float64 // Weight for complexity score (default: 0.25)
	SizeWeight        float64 // Weight for size score (default: 0.20)
	TestabilityWeight float64 // Weight for testability score (default: 0.15)
	UtilityWeight     float64 // Weight for utility score (default: 0.10)
}

// NewPriorityRanker creates a new priority ranker with default settings
func NewPriorityRanker() *PriorityRanker {
	return &PriorityRanker{
		criteria: DefaultSelectionCriteria(),
		weights:  DefaultScoringWeights(),
	}
}

// NewPriorityRankerWithCriteria creates a priority ranker with custom criteria
func NewPriorityRankerWithCriteria(criteria SelectionCriteria, weights ScoringWeights) *PriorityRanker {
	// Validate weights sum to 1.0
	totalWeight := weights.DependencyWeight + weights.ComplexityWeight +
		weights.SizeWeight + weights.TestabilityWeight + weights.UtilityWeight

	if math.Abs(totalWeight-1.0) > 0.01 {
		// Normalize weights if they don't sum to 1.0
		weights.DependencyWeight /= totalWeight
		weights.ComplexityWeight /= totalWeight
		weights.SizeWeight /= totalWeight
		weights.TestabilityWeight /= totalWeight
		weights.UtilityWeight /= totalWeight
	}

	return &PriorityRanker{
		criteria: criteria,
		weights:  weights,
	}
}

// DefaultSelectionCriteria returns sensible default selection criteria
func DefaultSelectionCriteria() SelectionCriteria {
	return SelectionCriteria{
		MaxDependencies:      5,
		MinComplexity:        2.0,
		MaxComplexity:        20.0,
		MinSize:              10,
		MaxSize:              300,
		RequireInterfaces:    false,
		ExcludeNetworkIO:     true,
		ExcludeDatabaseIO:    true,
		ExcludeFileIO:        false,
		MaxMockingComplexity: 3,
		PreferUtilityFiles:   true,
		MinTestabilityScore:  30.0,
		MaxFileSize:          50000, // 50KB
	}
}

// DefaultScoringWeights returns the default scoring weights
func DefaultScoringWeights() ScoringWeights {
	return ScoringWeights{
		DependencyWeight:  0.30,
		ComplexityWeight:  0.25,
		SizeWeight:        0.20,
		TestabilityWeight: 0.15,
		UtilityWeight:     0.10,
	}
}

// RankFiles ranks files based on their suitability for test generation
func (pr *PriorityRanker) RankFiles(files map[string]*testdiscovery.FileInfo) []testdiscovery.FileScore {
	var scores []testdiscovery.FileScore

	for path, fileInfo := range files {
		// Skip files that already have tests
		if fileInfo.HasTests {
			continue
		}

		// Calculate score for this file
		score := pr.calculateFileScore(path, fileInfo)
		scores = append(scores, score)
	}

	// Sort by total score (descending)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].TotalScore > scores[j].TotalScore
	})

	return scores
}

// calculateFileScore calculates a comprehensive score for a file
func (pr *PriorityRanker) calculateFileScore(path string, fileInfo *testdiscovery.FileInfo) testdiscovery.FileScore {
	score := testdiscovery.FileScore{
		FilePath: path,
	}

	// Check exclusion criteria first
	if pr.shouldExcludeFile(fileInfo) {
		score.IsExcluded = true
		score.ExclusionReason = pr.getExclusionReason(fileInfo)
		return score
	}

	// Calculate individual scores
	score.DependencyScore = pr.calculateDependencyScore(fileInfo)
	score.ComplexityScore = pr.calculateComplexityScore(fileInfo)
	score.SizeScore = pr.calculateSizeScore(fileInfo)
	score.TestabilityScore = pr.calculateTestabilityScore(fileInfo)
	score.UtilityScore = pr.calculateUtilityScore(fileInfo)

	// Calculate weighted total score
	score.TotalScore = (score.DependencyScore * pr.weights.DependencyWeight) +
		(score.ComplexityScore * pr.weights.ComplexityWeight) +
		(score.SizeScore * pr.weights.SizeWeight) +
		(score.TestabilityScore * pr.weights.TestabilityWeight) +
		(score.UtilityScore * pr.weights.UtilityWeight)

	// Generate selection reason
	score.SelectionReason = pr.generateSelectionReason(&score)

	return score
}

// shouldExcludeFile determines if a file should be excluded from testing
func (pr *PriorityRanker) shouldExcludeFile(fileInfo *testdiscovery.FileInfo) bool {
	// Exclude generated files
	if fileInfo.IsGenerated {
		return true
	}

	// Exclude files with too many dependencies
	if fileInfo.ImportCount > pr.criteria.MaxDependencies {
		return true
	}

	// Exclude files based on I/O criteria
	if pr.criteria.ExcludeNetworkIO && fileInfo.HasNetworkAccess {
		return true
	}

	if pr.criteria.ExcludeDatabaseIO && fileInfo.HasDatabaseAccess {
		return true
	}

	if pr.criteria.ExcludeFileIO && fileInfo.HasFileIO {
		return true
	}

	// Exclude files with too low testability
	if fileInfo.TestabilityScore < pr.criteria.MinTestabilityScore {
		return true
	}

	// Exclude files outside size range
	if fileInfo.LineCount < pr.criteria.MinSize || fileInfo.LineCount > pr.criteria.MaxSize {
		return true
	}

	// Exclude files outside complexity range
	if fileInfo.ComplexityScore < pr.criteria.MinComplexity || fileInfo.ComplexityScore > pr.criteria.MaxComplexity {
		return true
	}

	// Exclude files that are too large
	if fileInfo.Size > pr.criteria.MaxFileSize {
		return true
	}

	return false
}

// getExclusionReason returns the reason why a file was excluded
func (pr *PriorityRanker) getExclusionReason(fileInfo *testdiscovery.FileInfo) string {
	if fileInfo.IsGenerated {
		return "Generated file - should not be tested"
	}

	if fileInfo.ImportCount > pr.criteria.MaxDependencies {
		return fmt.Sprintf("Too many dependencies (%d > %d)", fileInfo.ImportCount, pr.criteria.MaxDependencies)
	}

	if pr.criteria.ExcludeNetworkIO && fileInfo.HasNetworkAccess {
		return "Network I/O operations make testing complex"
	}

	if pr.criteria.ExcludeDatabaseIO && fileInfo.HasDatabaseAccess {
		return "Database operations require complex mocking"
	}

	if pr.criteria.ExcludeFileIO && fileInfo.HasFileIO {
		return "File I/O operations complicate testing"
	}

	if fileInfo.TestabilityScore < pr.criteria.MinTestabilityScore {
		return fmt.Sprintf("Low testability score (%.1f < %.1f)", fileInfo.TestabilityScore, pr.criteria.MinTestabilityScore)
	}

	if fileInfo.LineCount < pr.criteria.MinSize {
		return fmt.Sprintf("File too small (%d < %d lines)", fileInfo.LineCount, pr.criteria.MinSize)
	}

	if fileInfo.LineCount > pr.criteria.MaxSize {
		return fmt.Sprintf("File too large (%d > %d lines)", fileInfo.LineCount, pr.criteria.MaxSize)
	}

	if fileInfo.ComplexityScore < pr.criteria.MinComplexity {
		return fmt.Sprintf("Complexity too low (%.1f < %.1f)", fileInfo.ComplexityScore, pr.criteria.MinComplexity)
	}

	if fileInfo.ComplexityScore > pr.criteria.MaxComplexity {
		return fmt.Sprintf("Complexity too high (%.1f > %.1f)", fileInfo.ComplexityScore, pr.criteria.MaxComplexity)
	}

	if fileInfo.Size > pr.criteria.MaxFileSize {
		return fmt.Sprintf("File size too large (%d > %d bytes)", fileInfo.Size, pr.criteria.MaxFileSize)
	}

	return "Unknown exclusion reason"
}

// calculateDependencyScore calculates score based on dependency count (fewer = better)
func (pr *PriorityRanker) calculateDependencyScore(fileInfo *testdiscovery.FileInfo) float64 {
	// Ideal dependency count is 1-3
	optimal := 2.0
	actual := float64(fileInfo.ImportCount)

	if actual <= optimal {
		return 100.0
	}

	// Decrease score as dependencies increase
	penalty := (actual - optimal) * 10
	score := 100.0 - penalty

	if score < 0 {
		score = 0
	}

	return score
}

// calculateComplexityScore calculates score based on complexity (moderate = better)
func (pr *PriorityRanker) calculateComplexityScore(fileInfo *testdiscovery.FileInfo) float64 {
	// Optimal complexity range is 5-15
	minOptimal := 5.0
	maxOptimal := 15.0
	actual := fileInfo.ComplexityScore

	if actual >= minOptimal && actual <= maxOptimal {
		return 100.0
	}

	var penalty float64
	if actual < minOptimal {
		// Too simple - less interesting to test
		penalty = (minOptimal - actual) * 5
	} else {
		// Too complex - harder to test comprehensively
		penalty = (actual - maxOptimal) * 3
	}

	score := 100.0 - penalty
	if score < 0 {
		score = 0
	}

	return score
}

// calculateSizeScore calculates score based on file size (optimal range = better)
func (pr *PriorityRanker) calculateSizeScore(fileInfo *testdiscovery.FileInfo) float64 {
	// Optimal size range is 50-200 lines
	minOptimal := 50
	maxOptimal := 200
	actual := fileInfo.LineCount

	if actual >= minOptimal && actual <= maxOptimal {
		return 100.0
	}

	var penalty float64
	if actual < minOptimal {
		// Too small - might not be worth testing
		penalty = float64(minOptimal-actual) * 1.0
	} else {
		// Too large - harder to achieve good coverage
		penalty = float64(actual-maxOptimal) * 0.3
	}

	score := 100.0 - penalty
	if score < 0 {
		score = 0
	}

	return score
}

// calculateTestabilityScore uses the existing testability score
func (pr *PriorityRanker) calculateTestabilityScore(fileInfo *testdiscovery.FileInfo) float64 {
	return fileInfo.TestabilityScore
}

// calculateUtilityScore calculates score based on utility/reusability factors
func (pr *PriorityRanker) calculateUtilityScore(fileInfo *testdiscovery.FileInfo) float64 {
	score := 50.0 // Base score

	// Bonus points for pure functions and data structures
	if pr.isUtilityFile(fileInfo) {
		score += 30
	}

	// Bonus for interfaces (enable testing of other code)
	score += float64(fileInfo.InterfaceCount) * 5

	// Bonus for exported functions (more likely to be reused)
	score += float64(fileInfo.FunctionCount) * 2

	// Penalty for complex mocking requirements
	if fileInfo.RequiresMocking {
		score -= 20
	}

	// Bonus for structs with methods (testable units)
	score += float64(fileInfo.MethodCount) * 1.5

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}

// isUtilityFile determines if a file is a utility/helper file
func (pr *PriorityRanker) isUtilityFile(fileInfo *testdiscovery.FileInfo) bool {
	path := strings.ToLower(fileInfo.Path)

	// Check for common utility file patterns
	utilityPatterns := []string{
		"util", "utils", "helper", "helpers", "common", "shared",
		"constants", "types", "errors", "validation", "convert",
	}

	for _, pattern := range utilityPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}

	// Check if file has mostly pure functions (no I/O, no complex dependencies)
	return !fileInfo.HasNetworkAccess &&
		!fileInfo.HasDatabaseAccess &&
		!fileInfo.HasFileIO &&
		fileInfo.ImportCount <= 3
}

// generateSelectionReason generates a human-readable reason for the selection score
func (pr *PriorityRanker) generateSelectionReason(score *testdiscovery.FileScore) string {
	var reasons []string

	// Analyze individual scores
	if score.DependencyScore >= 80 {
		reasons = append(reasons, "minimal dependencies")
	} else if score.DependencyScore >= 60 {
		reasons = append(reasons, "moderate dependencies")
	} else {
		reasons = append(reasons, "many dependencies")
	}

	if score.ComplexityScore >= 80 {
		reasons = append(reasons, "optimal complexity")
	} else if score.ComplexityScore >= 60 {
		reasons = append(reasons, "acceptable complexity")
	} else {
		reasons = append(reasons, "challenging complexity")
	}

	if score.SizeScore >= 80 {
		reasons = append(reasons, "good size")
	} else if score.SizeScore >= 60 {
		reasons = append(reasons, "acceptable size")
	} else {
		reasons = append(reasons, "size concerns")
	}

	if score.TestabilityScore >= 80 {
		reasons = append(reasons, "highly testable")
	} else if score.TestabilityScore >= 60 {
		reasons = append(reasons, "moderately testable")
	} else {
		reasons = append(reasons, "testing challenges")
	}

	if score.UtilityScore >= 80 {
		reasons = append(reasons, "high utility value")
	} else if score.UtilityScore >= 60 {
		reasons = append(reasons, "moderate utility")
	} else {
		reasons = append(reasons, "limited utility")
	}

	// Create a concise reason string
	if len(reasons) == 0 {
		return "Standard candidate for testing"
	}

	return strings.Join(reasons, ", ")
}

// GetTopCandidates returns the top N candidates for test generation
func (pr *PriorityRanker) GetTopCandidates(scores []testdiscovery.FileScore, count int) []testdiscovery.FileScore {
	// Filter out excluded files
	var candidates []testdiscovery.FileScore
	for _, score := range scores {
		if !score.IsExcluded {
			candidates = append(candidates, score)
		}
	}

	// Return top N candidates
	if len(candidates) < count {
		count = len(candidates)
	}

	return candidates[:count]
}

// UpdateCriteria updates the selection criteria
func (pr *PriorityRanker) UpdateCriteria(criteria SelectionCriteria) {
	pr.criteria = criteria
}

// UpdateWeights updates the scoring weights
func (pr *PriorityRanker) UpdateWeights(weights ScoringWeights) {
	pr.weights = weights
}

// GetCriteria returns the current selection criteria
func (pr *PriorityRanker) GetCriteria() SelectionCriteria {
	return pr.criteria
}

// GetWeights returns the current scoring weights
func (pr *PriorityRanker) GetWeights() ScoringWeights {
	return pr.weights
}

// GenerateSelectionReport generates a detailed report about the selection process
func (pr *PriorityRanker) GenerateSelectionReport(scores []testdiscovery.FileScore) SelectionReport {
	report := SelectionReport{
		GeneratedAt:      time.Now(),
		TotalFiles:       len(scores),
		Criteria:         pr.criteria,
		Weights:          pr.weights,
		ExclusionReasons: make(map[string]int),
	}

	// Categorize files
	for _, score := range scores {
		if score.IsExcluded {
			report.ExcludedFiles++
			report.ExclusionReasons[score.ExclusionReason]++
		} else {
			report.CandidateFiles++
			if score.TotalScore >= 80 {
				report.HighPriorityFiles++
			} else if score.TotalScore >= 60 {
				report.MediumPriorityFiles++
			} else {
				report.LowPriorityFiles++
			}
		}

		// Track score statistics
		if score.TotalScore > report.MaxScore {
			report.MaxScore = score.TotalScore
			report.BestCandidate = score.FilePath
		}

		if !score.IsExcluded && (report.MinScore == 0 || score.TotalScore < report.MinScore) {
			report.MinScore = score.TotalScore
		}

		report.TotalScore += score.TotalScore
	}

	// Calculate average
	if report.CandidateFiles > 0 {
		report.AverageScore = report.TotalScore / float64(report.CandidateFiles)
	}

	return report
}

// SelectionReport provides detailed information about the selection process
type SelectionReport struct {
	GeneratedAt         time.Time         `json:"generated_at"`
	TotalFiles          int               `json:"total_files"`
	CandidateFiles      int               `json:"candidate_files"`
	ExcludedFiles       int               `json:"excluded_files"`
	HighPriorityFiles   int               `json:"high_priority_files"`
	MediumPriorityFiles int               `json:"medium_priority_files"`
	LowPriorityFiles    int               `json:"low_priority_files"`
	MaxScore            float64           `json:"max_score"`
	MinScore            float64           `json:"min_score"`
	AverageScore        float64           `json:"average_score"`
	TotalScore          float64           `json:"total_score"`
	BestCandidate       string            `json:"best_candidate"`
	ExclusionReasons    map[string]int    `json:"exclusion_reasons"`
	Criteria            SelectionCriteria `json:"criteria"`
	Weights             ScoringWeights    `json:"weights"`
}
