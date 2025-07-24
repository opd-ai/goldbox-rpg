// Package analyzer provides comprehensive Go source code analysis capabilities
// for test discovery and generation systems.
package analyzer

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"goldbox-rpg/pkg/testdiscovery"
)

// FileScanner provides comprehensive file discovery and analysis capabilities
type FileScanner struct {
	rootDir           string
	fileSet           *token.FileSet
	exclusionPatterns []*regexp.Regexp
	gitignorePatterns []string
	cache             map[string]*testdiscovery.FileInfo
	maxCacheAge       time.Duration
}

// NewFileScanner creates a new file scanner with default exclusion patterns
func NewFileScanner(rootDir string) *FileScanner {
	scanner := &FileScanner{
		rootDir:     rootDir,
		fileSet:     token.NewFileSet(),
		cache:       make(map[string]*testdiscovery.FileInfo),
		maxCacheAge: 5 * time.Minute,
	}

	// Initialize default exclusion patterns
	scanner.initializeExclusionPatterns()
	scanner.loadGitignorePatterns()

	return scanner
}

// initializeExclusionPatterns sets up default file exclusion patterns
func (fs *FileScanner) initializeExclusionPatterns() {
	patterns := []string{
		`.*_test\.go$`,        // Test files
		`.*/vendor/.*`,        // Vendor directory
		`.*/\.git/.*`,         // Git directory
		`.*/testdata/.*`,      // Test data directory
		`.*/node_modules/.*`,  // Node modules
		`.*\.pb\.go$`,         // Protobuf generated files
		`.*\.gen\.go$`,        // Generated files
		`.*_generated\.go$`,   // Generated files
		`.*/mock.*\.go$`,      // Mock files
		`.*/mocks/.*\.go$`,    // Mock directories
		`.*bindata\.go$`,      // Binary data files
		`.*/cmd/.*/main\.go$`, // Main files in cmd directories
		`.*/examples/.*`,      // Example directories
		`.*/demo/.*`,          // Demo directories
	}

	fs.exclusionPatterns = make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		fs.exclusionPatterns[i] = regexp.MustCompile(pattern)
	}
}

// loadGitignorePatterns loads exclusion patterns from .gitignore
func (fs *FileScanner) loadGitignorePatterns() {
	gitignorePath := filepath.Join(fs.rootDir, ".gitignore")
	file, err := os.Open(gitignorePath)
	if err != nil {
		return // .gitignore doesn't exist or can't be read
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			fs.gitignorePatterns = append(fs.gitignorePatterns, line)
		}
	}
}

// ScanDirectory performs comprehensive directory scanning and analysis
func (fs *FileScanner) ScanDirectory() (map[string]*testdiscovery.FileInfo, error) {
	result := make(map[string]*testdiscovery.FileInfo)

	err := filepath.Walk(fs.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Check exclusion patterns
		if fs.shouldExcludeFile(path) {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(fs.rootDir, path)
		if err != nil {
			return err
		}

		// Check cache first
		if cachedInfo, exists := fs.getCachedFileInfo(path, info.ModTime()); exists {
			result[relPath] = cachedInfo
			return nil
		}

		// Analyze file
		fileInfo, err := fs.analyzeFile(path, relPath, info)
		if err != nil {
			return fmt.Errorf("failed to analyze file %s: %w", path, err)
		}

		// Cache the result
		fs.cache[path] = fileInfo
		result[relPath] = fileInfo

		return nil
	})

	return result, err
}

// shouldExcludeFile checks if a file should be excluded from analysis
func (fs *FileScanner) shouldExcludeFile(path string) bool {
	// Check regex patterns
	for _, pattern := range fs.exclusionPatterns {
		if pattern.MatchString(path) {
			return true
		}
	}

	// Check gitignore patterns (simplified)
	for _, pattern := range fs.gitignorePatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
	}

	return false
}

// getCachedFileInfo retrieves cached file info if still valid
func (fs *FileScanner) getCachedFileInfo(path string, modTime time.Time) (*testdiscovery.FileInfo, bool) {
	if info, exists := fs.cache[path]; exists {
		if time.Since(info.LastModified) < fs.maxCacheAge && info.LastModified.Equal(modTime) {
			return info, true
		}
	}
	return nil, false
}

// analyzeFile performs comprehensive analysis of a single Go file
func (fs *FileScanner) analyzeFile(absPath, relPath string, info os.FileInfo) (*testdiscovery.FileInfo, error) {
	// Parse the file
	src, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	file, err := parser.ParseFile(fs.fileSet, absPath, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Initialize file info
	fileInfo := &testdiscovery.FileInfo{
		Path:              relPath,
		AbsolutePath:      absPath,
		PackageName:       file.Name.Name,
		Size:              info.Size(),
		LastModified:      info.ModTime(),
		AST:               file,
		FileSet:           fs.fileSet,
		Dependencies:      make([]string, 0),
		ExportedFunctions: make([]testdiscovery.FunctionInfo, 0),
		ExportedTypes:     make([]testdiscovery.TypeInfo, 0),
		Imports:           make(map[string]string),
	}

	// Count lines
	fileInfo.LineCount = fs.countLines(string(src))

	// Check if file is generated
	fileInfo.IsGenerated = fs.isGeneratedFile(string(src))

	// Analyze imports
	fs.analyzeImports(file, fileInfo)

	// Analyze declarations
	fs.analyzeDeclarations(file, fileInfo)

	// Calculate complexity and testability scores
	fileInfo.ComplexityScore = fs.calculateComplexityScore(fileInfo)
	fileInfo.TestabilityScore = fs.calculateTestabilityScore(fileInfo)

	// Check for test file existence
	fs.checkForTestFile(fileInfo)

	// Check for problematic dependencies
	fs.analyzeProblematicDependencies(fileInfo)

	return fileInfo, nil
}

// countLines counts the number of lines in source code
func (fs *FileScanner) countLines(src string) int {
	return strings.Count(src, "\n") + 1
}

// isGeneratedFile checks if a file is automatically generated
func (fs *FileScanner) isGeneratedFile(src string) bool {
	generatedPatterns := []string{
		"// Code generated by",
		"// This file was automatically generated",
		"// AUTO-GENERATED FILE",
		"// autogenerated",
		"DO NOT EDIT",
	}

	srcLines := strings.Split(src, "\n")
	// Check first 10 lines for generation markers
	maxLines := len(srcLines)
	if maxLines > 10 {
		maxLines = 10
	}

	for i := 0; i < maxLines; i++ {
		line := strings.TrimSpace(srcLines[i])
		for _, pattern := range generatedPatterns {
			if strings.Contains(line, pattern) {
				return true
			}
		}
	}

	return false
}

// analyzeImports analyzes import statements and builds dependency information
func (fs *FileScanner) analyzeImports(file *ast.File, fileInfo *testdiscovery.FileInfo) {
	for _, importSpec := range file.Imports {
		importPath := strings.Trim(importSpec.Path.Value, `"`)

		// Skip standard library for dependency counting
		if !fs.isStandardLibrary(importPath) {
			fileInfo.Dependencies = append(fileInfo.Dependencies, importPath)
			fileInfo.ImportCount++
		}

		// Store import mapping
		alias := ""
		if importSpec.Name != nil {
			alias = importSpec.Name.Name
		}
		fileInfo.Imports[alias] = importPath
	}
}

// isStandardLibrary checks if an import is from Go's standard library
func (fs *FileScanner) isStandardLibrary(importPath string) bool {
	// Simple heuristic: standard library packages don't contain dots
	// More sophisticated checking could use go/build.Default.IsStdLibrary
	return !strings.Contains(importPath, ".") ||
		strings.HasPrefix(importPath, "golang.org/x/")
}

// analyzeDeclarations analyzes top-level declarations (functions, types, etc.)
func (fs *FileScanner) analyzeDeclarations(file *ast.File, fileInfo *testdiscovery.FileInfo) {
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			fs.analyzeFunctionDecl(d, fileInfo)
		case *ast.GenDecl:
			fs.analyzeGenDecl(d, fileInfo)
		}
	}
}

// analyzeFunctionDecl analyzes function declarations
func (fs *FileScanner) analyzeFunctionDecl(funcDecl *ast.FuncDecl, fileInfo *testdiscovery.FileInfo) {
	if !funcDecl.Name.IsExported() {
		return
	}

	funcInfo := testdiscovery.FunctionInfo{
		Name:       funcDecl.Name.Name,
		IsExported: true,
		StartPos:   funcDecl.Pos(),
		EndPos:     funcDecl.End(),
	}

	// Analyze receiver (for methods)
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		if starExpr, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
			if ident, ok := starExpr.X.(*ast.Ident); ok {
				funcInfo.Receiver = ident.Name
			}
		} else if ident, ok := funcDecl.Recv.List[0].Type.(*ast.Ident); ok {
			funcInfo.Receiver = ident.Name
		}
	}

	// Analyze parameters
	if funcDecl.Type.Params != nil {
		funcInfo.ParameterCount = len(funcDecl.Type.Params.List)
		fs.analyzeParameters(funcDecl.Type.Params, &funcInfo)
	}

	// Analyze return values
	if funcDecl.Type.Results != nil {
		funcInfo.ReturnCount = len(funcDecl.Type.Results.List)
		fs.analyzeReturnTypes(funcDecl.Type.Results, &funcInfo)
	}

	// Analyze function body
	if funcDecl.Body != nil {
		fs.analyzeFunctionBody(funcDecl.Body, &funcInfo)
	}

	// Add to appropriate collection
	if funcInfo.Receiver != "" {
		fileInfo.MethodCount++
	} else {
		fileInfo.FunctionCount++
	}

	fileInfo.ExportedFunctions = append(fileInfo.ExportedFunctions, funcInfo)
}

// analyzeParameters analyzes function parameters
func (fs *FileScanner) analyzeParameters(fieldList *ast.FieldList, funcInfo *testdiscovery.FunctionInfo) {
	funcInfo.Parameters = make([]testdiscovery.ParameterInfo, 0)

	for _, field := range fieldList.List {
		paramType := fs.typeToString(field.Type)

		// Handle variadic parameters
		if strings.HasPrefix(paramType, "...") {
			funcInfo.IsVariadic = true
		}

		// Create parameter info for each name (or unnamed parameter)
		if len(field.Names) == 0 {
			// Unnamed parameter
			param := testdiscovery.ParameterInfo{
				Type:      paramType,
				IsPointer: strings.HasPrefix(paramType, "*"),
				IsSlice:   strings.HasPrefix(paramType, "[]"),
				IsMap:     strings.HasPrefix(paramType, "map["),
			}
			funcInfo.Parameters = append(funcInfo.Parameters, param)
		} else {
			// Named parameters
			for _, name := range field.Names {
				param := testdiscovery.ParameterInfo{
					Name:      name.Name,
					Type:      paramType,
					IsPointer: strings.HasPrefix(paramType, "*"),
					IsSlice:   strings.HasPrefix(paramType, "[]"),
					IsMap:     strings.HasPrefix(paramType, "map["),
				}
				funcInfo.Parameters = append(funcInfo.Parameters, param)
			}
		}
	}
}

// analyzeReturnTypes analyzes function return types
func (fs *FileScanner) analyzeReturnTypes(fieldList *ast.FieldList, funcInfo *testdiscovery.FunctionInfo) {
	funcInfo.ReturnTypes = make([]string, 0)

	for _, field := range fieldList.List {
		returnType := fs.typeToString(field.Type)
		funcInfo.ReturnTypes = append(funcInfo.ReturnTypes, returnType)

		// Check if returns error
		if returnType == "error" {
			funcInfo.ReturnsError = true
		}
	}
}

// analyzeFunctionBody analyzes the function body for complexity metrics
func (fs *FileScanner) analyzeFunctionBody(body *ast.BlockStmt, funcInfo *testdiscovery.FunctionInfo) {
	visitor := &complexityVisitor{funcInfo: funcInfo}
	ast.Walk(visitor, body)
}

// complexityVisitor is used to walk the AST and calculate complexity metrics
type complexityVisitor struct {
	funcInfo *testdiscovery.FunctionInfo
}

func (v *complexityVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.IfStmt:
		v.funcInfo.HasIfStatements = true
		v.funcInfo.ComplexityScore++
	case *ast.ForStmt, *ast.RangeStmt:
		v.funcInfo.HasLoops = true
		v.funcInfo.ComplexityScore++
	case *ast.TypeSwitchStmt, *ast.SwitchStmt:
		v.funcInfo.ComplexityScore++
	case *ast.CallExpr:
		if ident, ok := n.Fun.(*ast.Ident); ok {
			if ident.Name == "panic" || ident.Name == "recover" {
				v.funcInfo.HasPanicRecover = true
			}
		}
	}
	return v
}

// analyzeGenDecl analyzes general declarations (types, constants, variables)
func (fs *FileScanner) analyzeGenDecl(genDecl *ast.GenDecl, fileInfo *testdiscovery.FileInfo) {
	for _, spec := range genDecl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			if s.Name.IsExported() {
				fs.analyzeTypeSpec(s, fileInfo)
			}
		}
	}
}

// analyzeTypeSpec analyzes type specifications
func (fs *FileScanner) analyzeTypeSpec(typeSpec *ast.TypeSpec, fileInfo *testdiscovery.FileInfo) {
	typeInfo := testdiscovery.TypeInfo{
		Name:       typeSpec.Name.Name,
		IsExported: true,
		Methods:    make([]testdiscovery.FunctionInfo, 0),
		Fields:     make([]testdiscovery.FieldInfo, 0),
	}

	switch t := typeSpec.Type.(type) {
	case *ast.InterfaceType:
		typeInfo.Kind = "interface"
		typeInfo.IsInterface = true
		fileInfo.InterfaceCount++

		// Analyze interface methods
		if t.Methods != nil {
			for _, method := range t.Methods.List {
				if len(method.Names) > 0 {
					// This is a method
					funcInfo := testdiscovery.FunctionInfo{
						Name:       method.Names[0].Name,
						IsExported: method.Names[0].IsExported(),
					}
					typeInfo.Methods = append(typeInfo.Methods, funcInfo)
				}
			}
		}

	case *ast.StructType:
		typeInfo.Kind = "struct"
		typeInfo.IsStruct = true
		fileInfo.StructCount++

		// Analyze struct fields
		if t.Fields != nil {
			for _, field := range t.Fields.List {
				fieldType := fs.typeToString(field.Type)

				if len(field.Names) == 0 {
					// Embedded field
					fieldInfo := testdiscovery.FieldInfo{
						Type:       fieldType,
						IsExported: true, // Embedded fields are always exported if the type is
					}
					typeInfo.Fields = append(typeInfo.Fields, fieldInfo)
				} else {
					// Named fields
					for _, name := range field.Names {
						fieldInfo := testdiscovery.FieldInfo{
							Name:       name.Name,
							Type:       fieldType,
							IsExported: name.IsExported(),
						}

						// Extract struct tags
						if field.Tag != nil {
							fieldInfo.Tags = field.Tag.Value
						}

						typeInfo.Fields = append(typeInfo.Fields, fieldInfo)
					}
				}
			}
		}

	default:
		typeInfo.Kind = "other"
	}

	fileInfo.ExportedTypes = append(fileInfo.ExportedTypes, typeInfo)
}

// typeToString converts an AST type to its string representation
func (fs *FileScanner) typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + fs.typeToString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + fs.typeToString(t.Elt)
		}
		return "[" + fs.typeToString(t.Len) + "]" + fs.typeToString(t.Elt)
	case *ast.MapType:
		return "map[" + fs.typeToString(t.Key) + "]" + fs.typeToString(t.Value)
	case *ast.ChanType:
		return "chan " + fs.typeToString(t.Value)
	case *ast.FuncType:
		return "func"
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.SelectorExpr:
		return fs.typeToString(t.X) + "." + t.Sel.Name
	case *ast.Ellipsis:
		return "..." + fs.typeToString(t.Elt)
	default:
		return "unknown"
	}
}

// calculateComplexityScore calculates a complexity score for the file
func (fs *FileScanner) calculateComplexityScore(fileInfo *testdiscovery.FileInfo) float64 {
	score := 0.0

	// Base score from function count
	score += float64(fileInfo.FunctionCount + fileInfo.MethodCount)

	// Add complexity from individual functions
	for _, function := range fileInfo.ExportedFunctions {
		score += float64(function.ComplexityScore)
	}

	// Normalize by file size
	if fileInfo.LineCount > 0 {
		score = score / float64(fileInfo.LineCount) * 100
	}

	return score
}

// calculateTestabilityScore calculates how testable a file is
func (fs *FileScanner) calculateTestabilityScore(fileInfo *testdiscovery.FileInfo) float64 {
	score := 100.0 // Start with perfect score

	// Penalty for excessive dependencies
	if fileInfo.ImportCount > 5 {
		score -= float64(fileInfo.ImportCount-5) * 5
	}

	// Bonus for interfaces (easier to mock)
	score += float64(fileInfo.InterfaceCount) * 10

	// Penalty for complex external dependencies
	for _, dep := range fileInfo.Dependencies {
		if strings.Contains(dep, "database") ||
			strings.Contains(dep, "sql") ||
			strings.Contains(dep, "net/http") ||
			strings.Contains(dep, "os") {
			score -= 15
		}
	}

	// Penalty for very high or very low complexity
	if fileInfo.ComplexityScore > 50 || fileInfo.ComplexityScore < 5 {
		score -= 20
	}

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// checkForTestFile checks if a corresponding test file exists
func (fs *FileScanner) checkForTestFile(fileInfo *testdiscovery.FileInfo) {
	// Convert source file path to test file path
	dir := filepath.Dir(fileInfo.AbsolutePath)
	baseName := strings.TrimSuffix(filepath.Base(fileInfo.AbsolutePath), ".go")
	testFileName := baseName + "_test.go"
	testPath := filepath.Join(dir, testFileName)

	if _, err := os.Stat(testPath); err == nil {
		fileInfo.HasTests = true
		if relTestPath, err := filepath.Rel(fs.rootDir, testPath); err == nil {
			fileInfo.TestPath = relTestPath
		}
	}
}

// GetFileSet returns the token.FileSet used by the scanner
func (fs *FileScanner) GetFileSet() *token.FileSet {
	return fs.fileSet
}

// analyzeProblematicDependencies identifies dependencies that make testing difficult
func (fs *FileScanner) analyzeProblematicDependencies(fileInfo *testdiscovery.FileInfo) {
	for _, dep := range fileInfo.Dependencies {
		switch {
		case strings.Contains(dep, "database/sql") ||
			strings.Contains(dep, "gorm") ||
			strings.Contains(dep, "mongo"):
			fileInfo.HasDatabaseAccess = true
			fileInfo.RequiresMocking = true

		case strings.Contains(dep, "net/http") ||
			strings.Contains(dep, "net/url") ||
			strings.Contains(dep, "http"):
			fileInfo.HasNetworkAccess = true
			fileInfo.RequiresMocking = true

		case strings.Contains(dep, "os") ||
			strings.Contains(dep, "io/ioutil") ||
			strings.Contains(dep, "bufio"):
			fileInfo.HasFileIO = true
		}
	}
}
