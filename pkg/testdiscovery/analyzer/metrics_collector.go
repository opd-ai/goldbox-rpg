// Package analyzer provides comprehensive metrics collection for Go source code analysis
package analyzer

import (
	"fmt"
	"go/token"
	"sort"
	"time"

	"goldbox-rpg/pkg/testdiscovery"
)

// MetricsCollector provides comprehensive code metrics collection
type MetricsCollector struct {
	fileSet             *token.FileSet
	complexityCalc      *ComplexityCalculator
	collectionStartTime time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(fileSet *token.FileSet) *MetricsCollector {
	return &MetricsCollector{
		fileSet:        fileSet,
		complexityCalc: NewComplexityCalculator(fileSet),
	}
}

// CollectProjectMetrics collects comprehensive metrics for all files in a project
func (mc *MetricsCollector) CollectProjectMetrics(files map[string]*testdiscovery.FileInfo) *ProjectMetrics {
	mc.collectionStartTime = time.Now()

	metrics := &ProjectMetrics{
		ProjectRoot:    ".", // This should be set by the caller
		TotalFiles:     len(files),
		CollectedAt:    mc.collectionStartTime,
		FileMetrics:    make(map[string]*FileMetrics),
		PackageMetrics: make(map[string]*PackageMetrics),
	}

	// Collect metrics for each file
	for path, fileInfo := range files {
		fileMetrics := mc.collectFileMetrics(fileInfo)
		metrics.FileMetrics[path] = fileMetrics

		// Update totals
		metrics.TotalLines += fileMetrics.LineCount
		metrics.TotalFunctions += fileMetrics.FunctionCount
		metrics.TotalComplexity += fileMetrics.TotalComplexity

		if fileInfo.HasTests {
			metrics.FilesWithTests++
		} else {
			metrics.FilesWithoutTests++
		}
	}

	// Collect package-level metrics
	mc.collectPackageMetrics(files, metrics)

	// Calculate aggregated statistics
	mc.calculateAggregatedStatistics(metrics)

	// Generate recommendations
	metrics.Recommendations = mc.generateProjectRecommendations(metrics)

	metrics.CollectionDuration = time.Since(mc.collectionStartTime)
	return metrics
}

// ProjectMetrics represents comprehensive project-level metrics
type ProjectMetrics struct {
	ProjectRoot        string                     `json:"project_root"`
	TotalFiles         int                        `json:"total_files"`
	FilesWithTests     int                        `json:"files_with_tests"`
	FilesWithoutTests  int                        `json:"files_without_tests"`
	TotalLines         int                        `json:"total_lines"`
	TotalFunctions     int                        `json:"total_functions"`
	TotalComplexity    int                        `json:"total_complexity"`
	AverageComplexity  float64                    `json:"average_complexity"`
	TestCoverage       float64                    `json:"test_coverage_percentage"`
	FileMetrics        map[string]*FileMetrics    `json:"file_metrics"`
	PackageMetrics     map[string]*PackageMetrics `json:"package_metrics"`
	TopComplexFiles    []string                   `json:"top_complex_files"`
	LargestFiles       []string                   `json:"largest_files"`
	MostDependentFiles []string                   `json:"most_dependent_files"`
	Recommendations    []string                   `json:"recommendations"`
	CollectedAt        time.Time                  `json:"collected_at"`
	CollectionDuration time.Duration              `json:"collection_duration"`
}

// FileMetrics represents comprehensive file-level metrics
type FileMetrics struct {
	Path                 string              `json:"path"`
	Package              string              `json:"package"`
	LineCount            int                 `json:"line_count"`
	FunctionCount        int                 `json:"function_count"`
	MethodCount          int                 `json:"method_count"`
	TypeCount            int                 `json:"type_count"`
	InterfaceCount       int                 `json:"interface_count"`
	StructCount          int                 `json:"struct_count"`
	ImportCount          int                 `json:"import_count"`
	TotalComplexity      int                 `json:"total_complexity"`
	AverageComplexity    float64             `json:"average_complexity"`
	MaxComplexity        int                 `json:"max_complexity"`
	TestabilityScore     float64             `json:"testability_score"`
	MaintainabilityIndex float64             `json:"maintainability_index"`
	HasTests             bool                `json:"has_tests"`
	TestPath             string              `json:"test_path"`
	IsGenerated          bool                `json:"is_generated"`
	RequiresMocking      bool                `json:"requires_mocking"`
	ProblematicElements  []string            `json:"problematic_elements"`
	ComplexityBreakdown  ComplexityMetrics   `json:"complexity_breakdown"`
	TestabilityAnalysis  TestabilityAnalysis `json:"testability_analysis"`
	QualityScore         float64             `json:"quality_score"`
	Recommendations      []string            `json:"recommendations"`
}

// PackageMetrics represents package-level metrics
type PackageMetrics struct {
	Name              string       `json:"name"`
	Path              string       `json:"path"`
	FileCount         int          `json:"file_count"`
	TotalLines        int          `json:"total_lines"`
	TotalFunctions    int          `json:"total_functions"`
	TotalComplexity   int          `json:"total_complexity"`
	AverageComplexity float64      `json:"average_complexity"`
	TestCoverage      float64      `json:"test_coverage"`
	FilesWithTests    int          `json:"files_with_tests"`
	FilesWithoutTests int          `json:"files_without_tests"`
	MostComplexFile   string       `json:"most_complex_file"`
	LargestFile       string       `json:"largest_file"`
	Dependencies      []string     `json:"dependencies"`
	Exports           []ExportInfo `json:"exports"`
	QualityGrade      string       `json:"quality_grade"`
}

// ExportInfo represents information about exported symbols
type ExportInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // function, type, constant, variable
	Package  string `json:"package"`
	FilePath string `json:"file_path"`
}

// collectFileMetrics collects comprehensive metrics for a single file
func (mc *MetricsCollector) collectFileMetrics(fileInfo *testdiscovery.FileInfo) *FileMetrics {
	metrics := &FileMetrics{
		Path:                fileInfo.Path,
		Package:             fileInfo.PackageName,
		LineCount:           fileInfo.LineCount,
		FunctionCount:       fileInfo.FunctionCount,
		MethodCount:         fileInfo.MethodCount,
		TypeCount:           len(fileInfo.ExportedTypes),
		InterfaceCount:      fileInfo.InterfaceCount,
		StructCount:         fileInfo.StructCount,
		ImportCount:         fileInfo.ImportCount,
		TestabilityScore:    fileInfo.TestabilityScore,
		HasTests:            fileInfo.HasTests,
		TestPath:            fileInfo.TestPath,
		IsGenerated:         fileInfo.IsGenerated,
		RequiresMocking:     fileInfo.RequiresMocking,
		ProblematicElements: make([]string, 0),
	}

	// Collect complexity metrics
	metrics.ComplexityBreakdown = mc.complexityCalc.CalculateFileComplexity(fileInfo)
	metrics.TotalComplexity = metrics.ComplexityBreakdown.TotalComplexity
	metrics.AverageComplexity = metrics.ComplexityBreakdown.AverageComplexity
	metrics.MaxComplexity = metrics.ComplexityBreakdown.MaxComplexity
	metrics.MaintainabilityIndex = metrics.ComplexityBreakdown.Maintainability

	// Collect testability analysis
	metrics.TestabilityAnalysis = mc.complexityCalc.AnalyzeTestability(fileInfo)

	// Identify problematic elements
	mc.identifyProblematicElements(fileInfo, metrics)

	// Calculate quality score
	metrics.QualityScore = mc.calculateQualityScore(metrics)

	// Generate file-specific recommendations
	metrics.Recommendations = mc.generateFileRecommendations(metrics)

	return metrics
}

// identifyProblematicElements identifies elements that make testing difficult
func (mc *MetricsCollector) identifyProblematicElements(fileInfo *testdiscovery.FileInfo, metrics *FileMetrics) {
	if fileInfo.HasDatabaseAccess {
		metrics.ProblematicElements = append(metrics.ProblematicElements, "database_access")
	}

	if fileInfo.HasNetworkAccess {
		metrics.ProblematicElements = append(metrics.ProblematicElements, "network_access")
	}

	if fileInfo.HasFileIO {
		metrics.ProblematicElements = append(metrics.ProblematicElements, "file_io")
	}

	if metrics.MaxComplexity > 15 {
		metrics.ProblematicElements = append(metrics.ProblematicElements, "high_complexity")
	}

	if fileInfo.ImportCount > 10 {
		metrics.ProblematicElements = append(metrics.ProblematicElements, "excessive_dependencies")
	}

	// Check for functions with high parameter counts
	for _, function := range fileInfo.ExportedFunctions {
		if function.ParameterCount > 5 {
			metrics.ProblematicElements = append(metrics.ProblematicElements,
				fmt.Sprintf("high_parameter_count_%s", function.Name))
		}
	}
}

// calculateQualityScore calculates an overall quality score for the file
func (mc *MetricsCollector) calculateQualityScore(metrics *FileMetrics) float64 {
	score := 100.0

	// Penalty for high complexity
	if metrics.AverageComplexity > 10 {
		score -= (metrics.AverageComplexity - 10) * 3
	}

	// Penalty for low maintainability
	if metrics.MaintainabilityIndex < 70 {
		score -= (70 - metrics.MaintainabilityIndex) * 0.5
	}

	// Penalty for problematic elements
	score -= float64(len(metrics.ProblematicElements)) * 5

	// Bonus for having tests
	if metrics.HasTests {
		score += 10
	}

	// Penalty for being generated (should not be tested)
	if metrics.IsGenerated {
		score -= 50
	}

	// Bonus for good testability
	score += (metrics.TestabilityScore - 50) * 0.2

	// Ensure bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// generateFileRecommendations generates recommendations for improving the file
func (mc *MetricsCollector) generateFileRecommendations(metrics *FileMetrics) []string {
	var recommendations []string

	if !metrics.HasTests {
		if metrics.QualityScore > 70 {
			recommendations = append(recommendations, "High priority for test creation - good testability")
		} else if metrics.QualityScore > 40 {
			recommendations = append(recommendations, "Medium priority for test creation - some refactoring may be needed")
		} else {
			recommendations = append(recommendations, "Low priority for test creation - significant refactoring required")
		}
	}

	if metrics.AverageComplexity > 15 {
		recommendations = append(recommendations, "Consider breaking down complex functions")
	}

	if metrics.MaintainabilityIndex < 50 {
		recommendations = append(recommendations, "Improve maintainability by reducing complexity and size")
	}

	if len(metrics.ProblematicElements) > 3 {
		recommendations = append(recommendations, "Address multiple testing challenges through refactoring")
	}

	if metrics.RequiresMocking {
		recommendations = append(recommendations, "Design interfaces to enable easier mocking")
	}

	return recommendations
}

// collectPackageMetrics collects package-level metrics
func (mc *MetricsCollector) collectPackageMetrics(files map[string]*testdiscovery.FileInfo, projectMetrics *ProjectMetrics) {
	packageFiles := make(map[string][]*testdiscovery.FileInfo)

	// Group files by package
	for _, fileInfo := range files {
		packageFiles[fileInfo.PackageName] = append(packageFiles[fileInfo.PackageName], fileInfo)
	}

	// Calculate metrics for each package
	for packageName, pkgFiles := range packageFiles {
		pkgMetrics := &PackageMetrics{
			Name:         packageName,
			FileCount:    len(pkgFiles),
			Dependencies: make([]string, 0),
			Exports:      make([]ExportInfo, 0),
		}

		// Aggregate file metrics
		var maxComplexityFile string
		var largestFile string
		maxComplexity := 0
		maxSize := 0

		for _, fileInfo := range pkgFiles {
			pkgMetrics.TotalLines += fileInfo.LineCount
			pkgMetrics.TotalFunctions += fileInfo.FunctionCount + fileInfo.MethodCount
			pkgMetrics.TotalComplexity += int(fileInfo.ComplexityScore)

			if fileInfo.HasTests {
				pkgMetrics.FilesWithTests++
			} else {
				pkgMetrics.FilesWithoutTests++
			}

			// Track most complex file
			if int(fileInfo.ComplexityScore) > maxComplexity {
				maxComplexity = int(fileInfo.ComplexityScore)
				maxComplexityFile = fileInfo.Path
			}

			// Track largest file
			if fileInfo.LineCount > maxSize {
				maxSize = fileInfo.LineCount
				largestFile = fileInfo.Path
			}

			// Collect dependencies
			for _, dep := range fileInfo.Dependencies {
				if !mc.containsString(pkgMetrics.Dependencies, dep) {
					pkgMetrics.Dependencies = append(pkgMetrics.Dependencies, dep)
				}
			}

			// Collect exports
			for _, function := range fileInfo.ExportedFunctions {
				export := ExportInfo{
					Name:     function.Name,
					Type:     "function",
					Package:  packageName,
					FilePath: fileInfo.Path,
				}
				pkgMetrics.Exports = append(pkgMetrics.Exports, export)
			}

			for _, typeInfo := range fileInfo.ExportedTypes {
				export := ExportInfo{
					Name:     typeInfo.Name,
					Type:     typeInfo.Kind,
					Package:  packageName,
					FilePath: fileInfo.Path,
				}
				pkgMetrics.Exports = append(pkgMetrics.Exports, export)
			}
		}

		// Calculate averages and derived metrics
		if pkgMetrics.FileCount > 0 {
			pkgMetrics.AverageComplexity = float64(pkgMetrics.TotalComplexity) / float64(pkgMetrics.FileCount)
			pkgMetrics.TestCoverage = float64(pkgMetrics.FilesWithTests) / float64(pkgMetrics.FileCount) * 100
		}

		pkgMetrics.MostComplexFile = maxComplexityFile
		pkgMetrics.LargestFile = largestFile
		pkgMetrics.QualityGrade = mc.calculateQualityGrade(pkgMetrics)

		projectMetrics.PackageMetrics[packageName] = pkgMetrics
	}
}

// calculateQualityGrade calculates a quality grade for the package
func (mc *MetricsCollector) calculateQualityGrade(pkgMetrics *PackageMetrics) string {
	score := 100.0

	// Factor in test coverage
	score = score * (pkgMetrics.TestCoverage / 100.0)

	// Factor in complexity
	if pkgMetrics.AverageComplexity > 10 {
		score -= (pkgMetrics.AverageComplexity - 10) * 5
	}

	// Assign grade
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

// calculateAggregatedStatistics calculates project-level statistics
func (mc *MetricsCollector) calculateAggregatedStatistics(metrics *ProjectMetrics) {
	if metrics.TotalFiles > 0 {
		metrics.TestCoverage = float64(metrics.FilesWithTests) / float64(metrics.TotalFiles) * 100
		metrics.AverageComplexity = float64(metrics.TotalComplexity) / float64(metrics.TotalFiles)
	}

	// Find top complex files
	type fileComplexity struct {
		path       string
		complexity int
	}

	var complexities []fileComplexity
	for path, fileMetrics := range metrics.FileMetrics {
		complexities = append(complexities, fileComplexity{
			path:       path,
			complexity: fileMetrics.TotalComplexity,
		})
	}

	sort.Slice(complexities, func(i, j int) bool {
		return complexities[i].complexity > complexities[j].complexity
	})

	// Get top 5 most complex files
	count := 5
	if len(complexities) < count {
		count = len(complexities)
	}
	for i := 0; i < count; i++ {
		metrics.TopComplexFiles = append(metrics.TopComplexFiles, complexities[i].path)
	}

	// Find largest files
	var sizes []fileComplexity
	for path, fileMetrics := range metrics.FileMetrics {
		sizes = append(sizes, fileComplexity{
			path:       path,
			complexity: fileMetrics.LineCount,
		})
	}

	sort.Slice(sizes, func(i, j int) bool {
		return sizes[i].complexity > sizes[j].complexity
	})

	// Get top 5 largest files
	count = 5
	if len(sizes) < count {
		count = len(sizes)
	}
	for i := 0; i < count; i++ {
		metrics.LargestFiles = append(metrics.LargestFiles, sizes[i].path)
	}

	// Find most dependent files
	var dependencies []fileComplexity
	for path, fileMetrics := range metrics.FileMetrics {
		dependencies = append(dependencies, fileComplexity{
			path:       path,
			complexity: fileMetrics.ImportCount,
		})
	}

	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].complexity > dependencies[j].complexity
	})

	// Get top 5 most dependent files
	count = 5
	if len(dependencies) < count {
		count = len(dependencies)
	}
	for i := 0; i < count; i++ {
		metrics.MostDependentFiles = append(metrics.MostDependentFiles, dependencies[i].path)
	}
}

// generateProjectRecommendations generates project-level recommendations
func (mc *MetricsCollector) generateProjectRecommendations(metrics *ProjectMetrics) []string {
	var recommendations []string

	coveragePercent := metrics.TestCoverage
	switch {
	case coveragePercent < 30:
		recommendations = append(recommendations, "Critical: Very low test coverage - prioritize test creation")
	case coveragePercent < 60:
		recommendations = append(recommendations, "Warning: Low test coverage - increase testing efforts")
	case coveragePercent < 80:
		recommendations = append(recommendations, "Good test coverage - aim for 80%+ for better confidence")
	default:
		recommendations = append(recommendations, "Excellent test coverage - maintain current testing practices")
	}

	avgComplexity := metrics.AverageComplexity
	if avgComplexity > 15 {
		recommendations = append(recommendations, "High average complexity - consider refactoring complex functions")
	}

	// Package-specific recommendations
	highComplexityPackages := 0
	lowCoveragePackages := 0
	for _, pkgMetrics := range metrics.PackageMetrics {
		if pkgMetrics.AverageComplexity > 15 {
			highComplexityPackages++
		}
		if pkgMetrics.TestCoverage < 60 {
			lowCoveragePackages++
		}
	}

	if highComplexityPackages > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("%d package(s) have high complexity - focus refactoring efforts", highComplexityPackages))
	}

	if lowCoveragePackages > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("%d package(s) have low test coverage - prioritize test creation", lowCoveragePackages))
	}

	return recommendations
}

// containsString checks if a slice contains a string
func (mc *MetricsCollector) containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
