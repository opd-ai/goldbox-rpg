// Package testdiscovery provides comprehensive test coverage discovery and generation
// for Go projects, implementing systematic file analysis and intelligent test creation.
package testdiscovery

import (
	"go/ast"
	"go/token"
	"time"
)

// FileInfo represents detailed information about a Go source file
type FileInfo struct {
	Path              string            // Relative path from project root
	AbsolutePath      string            // Absolute file path
	PackageName       string            // Go package name
	Size              int64             // File size in bytes
	LineCount         int               // Total lines of code
	ImportCount       int               // Number of import statements
	ImportDepth       int               // Maximum depth in import chain
	FunctionCount     int               // Total exported functions
	MethodCount       int               // Total methods on exported types
	InterfaceCount    int               // Number of interface definitions
	StructCount       int               // Number of struct definitions
	ComplexityScore   float64           // Cyclomatic complexity score
	TestabilityScore  float64           // Calculated testability score
	HasTests          bool              // Whether test file exists
	TestPath          string            // Path to corresponding test file
	Dependencies      []string          // List of imported packages
	ExportedFunctions []FunctionInfo    // Details of exported functions
	ExportedTypes     []TypeInfo        // Details of exported types
	Imports           map[string]string // Import alias -> package mapping
	AST               *ast.File         // Parsed AST (cached)
	FileSet           *token.FileSet    // Token file set for AST
	LastModified      time.Time         // File modification time
	IsGenerated       bool              // Whether file is generated
	HasDatabaseAccess bool              // Contains database operations
	HasNetworkAccess  bool              // Contains network operations
	HasFileIO         bool              // Contains file I/O operations
	RequiresMocking   bool              // Needs complex mocking setup
}

// FunctionInfo represents information about a function or method
type FunctionInfo struct {
	Name            string          // Function name
	Receiver        string          // Receiver type (for methods)
	IsExported      bool            // Whether function is exported
	IsVariadic      bool            // Whether function accepts variadic args
	ParameterCount  int             // Number of parameters
	ReturnCount     int             // Number of return values
	ReturnsError    bool            // Whether returns error
	ComplexityScore int             // Cyclomatic complexity
	LineCount       int             // Lines of code in function
	StartPos        token.Pos       // Starting position in file
	EndPos          token.Pos       // Ending position in file
	HasIfStatements bool            // Contains conditional logic
	HasLoops        bool            // Contains loops
	HasPanicRecover bool            // Uses panic/recover
	CallsExternal   bool            // Calls external packages
	Parameters      []ParameterInfo // Parameter details
	ReturnTypes     []string        // Return type names
	Documentation   string          // Function documentation
}

// ParameterInfo represents function parameter information
type ParameterInfo struct {
	Name      string // Parameter name
	Type      string // Parameter type
	IsPointer bool   // Whether parameter is pointer
	IsSlice   bool   // Whether parameter is slice
	IsMap     bool   // Whether parameter is map
}

// TypeInfo represents information about exported types
type TypeInfo struct {
	Name        string         // Type name
	Kind        string         // Type kind (struct, interface, etc.)
	Methods     []FunctionInfo // Associated methods
	Fields      []FieldInfo    // Struct fields (if applicable)
	IsInterface bool           // Whether type is interface
	IsStruct    bool           // Whether type is struct
	IsExported  bool           // Whether type is exported
}

// FieldInfo represents struct field information
type FieldInfo struct {
	Name       string // Field name
	Type       string // Field type
	IsExported bool   // Whether field is exported
	Tags       string // Struct tags
}

// FileScore represents the scoring metrics for file prioritization
type FileScore struct {
	FilePath         string  // File being scored
	DependencyScore  float64 // Score based on dependency count (0.3 weight)
	ComplexityScore  float64 // Score based on code complexity (0.25 weight)
	SizeScore        float64 // Score based on file size (0.2 weight)
	TestabilityScore float64 // Score based on testability factors (0.15 weight)
	UtilityScore     float64 // Score based on utility/reusability (0.1 weight)
	TotalScore       float64 // Weighted total score
	SelectionReason  string  // Explanation for score
	IsExcluded       bool    // Whether file should be excluded
	ExclusionReason  string  // Reason for exclusion
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
}

// AnalysisResult contains the complete analysis results
type AnalysisResult struct {
	ProjectRoot       string               // Project root directory
	TotalFiles        int                  // Total Go files analyzed
	FilesWithTests    int                  // Files that have tests
	FilesWithoutTests int                  // Files missing tests
	AnalyzedFiles     map[string]*FileInfo // All analyzed files
	ScoredFiles       []FileScore          // Files with selection scores
	SelectedFiles     []string             // Files selected for test generation
	ExcludedFiles     []string             // Files excluded from testing
	AnalysisTime      time.Duration        // Time taken for analysis
	Recommendations   []string             // Analysis recommendations
	Statistics        AnalysisStatistics   // Statistical summary
}

// AnalysisStatistics provides statistical summary of the analysis
type AnalysisStatistics struct {
	AverageComplexity    float64 // Average complexity score
	AverageDependencies  float64 // Average dependency count
	AverageFileSize      float64 // Average file size
	MostComplexFile      string  // File with highest complexity
	LargestFile          string  // Largest file by lines
	MostDependentFile    string  // File with most dependencies
	RecommendedTestFiles int     // Number of files recommended for testing
	EstimatedTestLines   int     // Estimated lines of test code needed
}

// TestGenerationRequest represents a request for test generation
type TestGenerationRequest struct {
	FilePath          string                 // File to generate tests for
	OutputPath        string                 // Where to write test file
	PackageName       string                 // Test package name
	Coverage          float64                // Target coverage percentage
	IncludeExamples   bool                   // Whether to include example tests
	IncludeBenchmarks bool                   // Whether to include benchmarks
	MockStrategy      string                 // Mocking strategy (auto, manual, none)
	Options           map[string]interface{} // Additional options
}

// TestGenerationResult represents the result of test generation
type TestGenerationResult struct {
	FilePath       string        // Generated test file path
	TestCount      int           // Number of tests generated
	Coverage       float64       // Achieved coverage percentage
	GeneratedLines int           // Lines of test code generated
	Warnings       []string      // Generation warnings
	Errors         []string      // Generation errors
	Duration       time.Duration // Time taken to generate
	Success        bool          // Whether generation was successful
}

// CoverageReport represents coverage analysis results
type CoverageReport struct {
	PackagePath     string             // Package being analyzed
	TotalLines      int                // Total lines in package
	CoveredLines    int                // Lines covered by tests
	CoveragePercent float64            // Coverage percentage
	UncoveredFiles  []string           // Files with insufficient coverage
	CoverageByFile  map[string]float64 // Coverage percentage per file
	TestFiles       []string           // Associated test files
	GeneratedAt     time.Time          // When report was generated
}

// MockingInfo represents information needed for generating mocks
type MockingInfo struct {
	InterfaceName  string   // Interface to mock
	Package        string   // Package containing interface
	Methods        []string // Methods to implement
	ImportPaths    []string // Required import paths
	GenerationType string   // Type of mock to generate
}
