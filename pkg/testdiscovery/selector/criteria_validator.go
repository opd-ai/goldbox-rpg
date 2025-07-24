// Package selector provides validation capabilities for selection criteria
package selector

import (
	"errors"
	"fmt"
	"math"
)

// CriteriaValidator provides validation for selection criteria and scoring weights
type CriteriaValidator struct{}

// NewCriteriaValidator creates a new criteria validator
func NewCriteriaValidator() *CriteriaValidator {
	return &CriteriaValidator{}
}

// ValidationResult represents the result of criteria validation
type ValidationResult struct {
	IsValid     bool     `json:"is_valid"`
	Errors      []string `json:"errors"`
	Warnings    []string `json:"warnings"`
	Suggestions []string `json:"suggestions"`
}

// ValidateSelectionCriteria validates selection criteria for correctness and reasonableness
func (cv *CriteriaValidator) ValidateSelectionCriteria(criteria SelectionCriteria) ValidationResult {
	result := ValidationResult{
		IsValid:     true,
		Errors:      make([]string, 0),
		Warnings:    make([]string, 0),
		Suggestions: make([]string, 0),
	}

	// Validate dependencies
	if criteria.MaxDependencies < 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "MaxDependencies cannot be negative")
	} else if criteria.MaxDependencies == 0 {
		result.Warnings = append(result.Warnings, "MaxDependencies of 0 will exclude all files with imports")
	} else if criteria.MaxDependencies > 20 {
		result.Warnings = append(result.Warnings, "MaxDependencies > 20 may allow files that are difficult to test")
	}

	// Validate complexity range
	if criteria.MinComplexity < 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "MinComplexity cannot be negative")
	}

	if criteria.MaxComplexity < 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "MaxComplexity cannot be negative")
	}

	if criteria.MinComplexity >= criteria.MaxComplexity {
		result.IsValid = false
		result.Errors = append(result.Errors, "MinComplexity must be less than MaxComplexity")
	}

	if criteria.MaxComplexity > 50 {
		result.Warnings = append(result.Warnings, "MaxComplexity > 50 may include very complex files that are hard to test")
	}

	if criteria.MinComplexity > 10 {
		result.Warnings = append(result.Warnings, "MinComplexity > 10 may exclude moderately complex files worth testing")
	}

	// Validate size range
	if criteria.MinSize < 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "MinSize cannot be negative")
	}

	if criteria.MaxSize < 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "MaxSize cannot be negative")
	}

	if criteria.MinSize >= criteria.MaxSize {
		result.IsValid = false
		result.Errors = append(result.Errors, "MinSize must be less than MaxSize")
	}

	if criteria.MaxSize > 1000 {
		result.Warnings = append(result.Warnings, "MaxSize > 1000 lines may include very large files that are hard to test comprehensively")
	}

	if criteria.MinSize < 5 {
		result.Warnings = append(result.Warnings, "MinSize < 5 lines may include trivial files not worth testing")
	}

	// Validate testability score
	if criteria.MinTestabilityScore < 0 || criteria.MinTestabilityScore > 100 {
		result.IsValid = false
		result.Errors = append(result.Errors, "MinTestabilityScore must be between 0 and 100")
	}

	if criteria.MinTestabilityScore > 80 {
		result.Warnings = append(result.Warnings, "MinTestabilityScore > 80 may be too restrictive")
	}

	// Validate file size
	if criteria.MaxFileSize < 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "MaxFileSize cannot be negative")
	}

	if criteria.MaxFileSize > 1000000 { // 1MB
		result.Warnings = append(result.Warnings, "MaxFileSize > 1MB may include very large files")
	}

	// Validate mocking complexity
	if criteria.MaxMockingComplexity < 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "MaxMockingComplexity cannot be negative")
	}

	// Generate suggestions based on criteria
	cv.generateCriteriaSuggestions(&result, criteria)

	return result
}

// ValidateScoringWeights validates scoring weights for correctness
func (cv *CriteriaValidator) ValidateScoringWeights(weights ScoringWeights) ValidationResult {
	result := ValidationResult{
		IsValid:     true,
		Errors:      make([]string, 0),
		Warnings:    make([]string, 0),
		Suggestions: make([]string, 0),
	}

	// Validate individual weights are non-negative
	if weights.DependencyWeight < 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "DependencyWeight cannot be negative")
	}

	if weights.ComplexityWeight < 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "ComplexityWeight cannot be negative")
	}

	if weights.SizeWeight < 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "SizeWeight cannot be negative")
	}

	if weights.TestabilityWeight < 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "TestabilityWeight cannot be negative")
	}

	if weights.UtilityWeight < 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "UtilityWeight cannot be negative")
	}

	// Calculate total weight
	totalWeight := weights.DependencyWeight + weights.ComplexityWeight +
		weights.SizeWeight + weights.TestabilityWeight + weights.UtilityWeight

	// Validate total weight
	if math.Abs(totalWeight) < 0.001 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Total weight cannot be zero")
	} else if math.Abs(totalWeight-1.0) > 0.01 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Weights sum to %.3f instead of 1.0 - will be normalized", totalWeight))
	}

	// Check for reasonable weight distribution
	maxWeight := math.Max(math.Max(weights.DependencyWeight, weights.ComplexityWeight),
		math.Max(math.Max(weights.SizeWeight, weights.TestabilityWeight), weights.UtilityWeight))

	if maxWeight > 0.6 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("One weight (%.3f) dominates - consider more balanced distribution", maxWeight))
	}

	// Generate weight suggestions
	cv.generateWeightSuggestions(&result, weights)

	return result
}

// ValidateCombinedCriteria validates criteria and weights together for consistency
func (cv *CriteriaValidator) ValidateCombinedCriteria(criteria SelectionCriteria, weights ScoringWeights) ValidationResult {
	// First validate individually
	criteriaResult := cv.ValidateSelectionCriteria(criteria)
	weightsResult := cv.ValidateScoringWeights(weights)

	// Combine results
	result := ValidationResult{
		IsValid:     criteriaResult.IsValid && weightsResult.IsValid,
		Errors:      append(criteriaResult.Errors, weightsResult.Errors...),
		Warnings:    append(criteriaResult.Warnings, weightsResult.Warnings...),
		Suggestions: append(criteriaResult.Suggestions, weightsResult.Suggestions...),
	}

	// Add combined validation checks
	cv.validateCombinedLogic(&result, criteria, weights)

	return result
}

// generateCriteriaSuggestions generates suggestions for improving criteria
func (cv *CriteriaValidator) generateCriteriaSuggestions(result *ValidationResult, criteria SelectionCriteria) {
	// Suggest optimal dependency range
	if criteria.MaxDependencies < 3 {
		result.Suggestions = append(result.Suggestions, "Consider allowing 3-5 dependencies for better file selection")
	} else if criteria.MaxDependencies > 8 {
		result.Suggestions = append(result.Suggestions, "Consider limiting dependencies to 5-8 for better testability")
	}

	// Suggest optimal complexity range
	if criteria.MaxComplexity-criteria.MinComplexity < 5 {
		result.Suggestions = append(result.Suggestions, "Consider widening complexity range for better selection variety")
	}

	// Suggest optimal size range
	if criteria.MaxSize-criteria.MinSize < 50 {
		result.Suggestions = append(result.Suggestions, "Consider widening size range (e.g., 50-300 lines)")
	}

	// Suggest enabling utility preference
	if !criteria.PreferUtilityFiles {
		result.Suggestions = append(result.Suggestions, "Consider enabling PreferUtilityFiles for better test ROI")
	}

	// Suggest reasonable exclusions
	if !criteria.ExcludeDatabaseIO && !criteria.ExcludeNetworkIO {
		result.Suggestions = append(result.Suggestions, "Consider excluding database/network I/O for easier testing")
	}
}

// generateWeightSuggestions generates suggestions for improving weights
func (cv *CriteriaValidator) generateWeightSuggestions(result *ValidationResult, weights ScoringWeights) {
	// Suggest balanced approach
	if weights.DependencyWeight < 0.2 {
		result.Suggestions = append(result.Suggestions, "Consider increasing DependencyWeight (recommended: 0.25-0.35)")
	}

	if weights.ComplexityWeight < 0.15 {
		result.Suggestions = append(result.Suggestions, "Consider increasing ComplexityWeight (recommended: 0.20-0.30)")
	}

	if weights.TestabilityWeight < 0.1 {
		result.Suggestions = append(result.Suggestions, "Consider increasing TestabilityWeight (recommended: 0.15-0.25)")
	}

	// Suggest specific improvements based on common patterns
	if weights.UtilityWeight > weights.ComplexityWeight {
		result.Suggestions = append(result.Suggestions, "ComplexityWeight typically should be higher than UtilityWeight")
	}
}

// validateCombinedLogic validates the interaction between criteria and weights
func (cv *CriteriaValidator) validateCombinedLogic(result *ValidationResult, criteria SelectionCriteria, weights ScoringWeights) {
	// Check for conflicting priorities
	if criteria.ExcludeNetworkIO && criteria.ExcludeDatabaseIO && criteria.ExcludeFileIO &&
		weights.TestabilityWeight < 0.2 {
		result.Warnings = append(result.Warnings,
			"High exclusion criteria with low testability weight may result in limited file selection")
	}

	// Check for overly restrictive criteria
	restrictiveCount := 0
	if criteria.MaxDependencies < 3 {
		restrictiveCount++
	}
	if criteria.MaxComplexity < 10 {
		restrictiveCount++
	}
	if criteria.MaxSize < 100 {
		restrictiveCount++
	}
	if criteria.MinTestabilityScore > 70 {
		restrictiveCount++
	}

	if restrictiveCount >= 3 {
		result.Warnings = append(result.Warnings,
			"Criteria may be too restrictive - consider relaxing some constraints")
	}

	// Check for weight-criteria alignment
	if weights.DependencyWeight > 0.4 && criteria.MaxDependencies > 10 {
		result.Warnings = append(result.Warnings,
			"High dependency weight conflicts with high max dependencies")
	}
}

// SuggestOptimalCriteria suggests optimal criteria based on project characteristics
func (cv *CriteriaValidator) SuggestOptimalCriteria(projectSize string, testingGoal string) (SelectionCriteria, ScoringWeights) {
	var criteria SelectionCriteria
	var weights ScoringWeights

	switch projectSize {
	case "small": // < 50 files
		criteria = SelectionCriteria{
			MaxDependencies:      4,
			MinComplexity:        1.0,
			MaxComplexity:        25.0,
			MinSize:              5,
			MaxSize:              500,
			RequireInterfaces:    false,
			ExcludeNetworkIO:     false,
			ExcludeDatabaseIO:    false,
			ExcludeFileIO:        false,
			MaxMockingComplexity: 5,
			PreferUtilityFiles:   true,
			MinTestabilityScore:  20.0,
			MaxFileSize:          100000,
		}

	case "medium": // 50-200 files
		criteria = DefaultSelectionCriteria()

	case "large": // 200+ files
		criteria = SelectionCriteria{
			MaxDependencies:      3,
			MinComplexity:        3.0,
			MaxComplexity:        15.0,
			MinSize:              20,
			MaxSize:              200,
			RequireInterfaces:    true,
			ExcludeNetworkIO:     true,
			ExcludeDatabaseIO:    true,
			ExcludeFileIO:        true,
			MaxMockingComplexity: 2,
			PreferUtilityFiles:   true,
			MinTestabilityScore:  50.0,
			MaxFileSize:          50000,
		}

	default:
		criteria = DefaultSelectionCriteria()
	}

	switch testingGoal {
	case "coverage": // Maximize test coverage
		weights = ScoringWeights{
			DependencyWeight:  0.40,
			ComplexityWeight:  0.20,
			SizeWeight:        0.20,
			TestabilityWeight: 0.15,
			UtilityWeight:     0.05,
		}

	case "quality": // Focus on code quality and maintainability
		weights = ScoringWeights{
			DependencyWeight:  0.25,
			ComplexityWeight:  0.35,
			SizeWeight:        0.15,
			TestabilityWeight: 0.20,
			UtilityWeight:     0.05,
		}

	case "utility": // Focus on testing reusable components
		weights = ScoringWeights{
			DependencyWeight:  0.20,
			ComplexityWeight:  0.20,
			SizeWeight:        0.15,
			TestabilityWeight: 0.15,
			UtilityWeight:     0.30,
		}

	default:
		weights = DefaultScoringWeights()
	}

	return criteria, weights
}

// OptimizeCriteria automatically optimizes criteria based on analysis results
func (cv *CriteriaValidator) OptimizeCriteria(criteria SelectionCriteria, weights ScoringWeights,
	candidateCount int, excludedCount int) (SelectionCriteria, ScoringWeights, error) {

	if candidateCount < 0 || excludedCount < 0 {
		return criteria, weights, errors.New("candidate and excluded counts must be non-negative")
	}

	totalFiles := candidateCount + excludedCount
	if totalFiles == 0 {
		return criteria, weights, errors.New("no files analyzed")
	}

	selectionRate := float64(candidateCount) / float64(totalFiles)

	// Adjust criteria based on selection rate
	optimizedCriteria := criteria
	optimizedWeights := weights

	if selectionRate < 0.1 { // Too restrictive
		optimizedCriteria.MaxDependencies += 2
		optimizedCriteria.MaxComplexity += 5
		optimizedCriteria.MaxSize += 50
		optimizedCriteria.MinTestabilityScore -= 10

		if optimizedCriteria.MinTestabilityScore < 0 {
			optimizedCriteria.MinTestabilityScore = 0
		}

	} else if selectionRate > 0.7 { // Too permissive
		optimizedCriteria.MaxDependencies -= 1
		optimizedCriteria.MaxComplexity -= 3
		optimizedCriteria.MaxSize -= 30
		optimizedCriteria.MinTestabilityScore += 10

		if optimizedCriteria.MaxDependencies < 1 {
			optimizedCriteria.MaxDependencies = 1
		}
		if optimizedCriteria.MinTestabilityScore > 90 {
			optimizedCriteria.MinTestabilityScore = 90
		}
	}

	// Validate optimized criteria
	validation := cv.ValidateCombinedCriteria(optimizedCriteria, optimizedWeights)
	if !validation.IsValid {
		return criteria, weights, fmt.Errorf("optimization resulted in invalid criteria: %v", validation.Errors)
	}

	return optimizedCriteria, optimizedWeights, nil
}
