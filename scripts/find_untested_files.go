package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// fileInfo holds information about discovered Go files
type fileInfo struct {
	sourceFiles map[string]bool
	testFiles   map[string]bool
}

// findUntestedFiles finds Go source files that don't have corresponding test files
func findUntestedFiles(rootDir string) ([]string, error) {
	files, err := collectGoFiles(rootDir)
	if err != nil {
		return nil, err
	}

	untestedFiles := identifyUntestedFiles(files)
	sort.Strings(untestedFiles)

	return untestedFiles, nil
}

// collectGoFiles walks through the directory tree and categorizes Go files
func collectGoFiles(rootDir string) (*fileInfo, error) {
	files := &fileInfo{
		sourceFiles: make(map[string]bool),
		testFiles:   make(map[string]bool),
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if shouldSkipFile(info, path) {
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		classifyGoFile(files, relPath, info.Name())
		return nil
	})

	return files, err
}

// shouldSkipFile determines if a file should be skipped during processing
func shouldSkipFile(info os.FileInfo, path string) bool {
	return info.IsDir() || !strings.HasSuffix(path, ".go")
}

// classifyGoFile categorizes a Go file as either source or test file
func classifyGoFile(files *fileInfo, relPath, fileName string) {
	if strings.HasSuffix(fileName, "_test.go") {
		registerTestFile(files, relPath, fileName)
	} else {
		files.sourceFiles[relPath] = true
	}
}

// registerTestFile processes a test file and maps it to its corresponding source file
func registerTestFile(files *fileInfo, relPath, fileName string) {
	baseName := strings.TrimSuffix(fileName, "_test.go")
	dir := filepath.Dir(relPath)
	baseFile := filepath.Join(dir, baseName+".go")
	files.testFiles[baseFile] = true
}

// identifyUntestedFiles finds source files without corresponding test files
func identifyUntestedFiles(files *fileInfo) []string {
	var untestedFiles []string

	for sourceFile := range files.sourceFiles {
		if shouldIncludeInUntested(sourceFile, files.testFiles) {
			untestedFiles = append(untestedFiles, sourceFile)
		}
	}

	return untestedFiles
}

// shouldIncludeInUntested determines if a source file should be included in untested files
func shouldIncludeInUntested(sourceFile string, testFiles map[string]bool) bool {
	// Skip main.go files as they typically don't have tests
	if strings.HasSuffix(sourceFile, "main.go") {
		return false
	}
	return !testFiles[sourceFile]
}

func main() {
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	untestedFiles, err := findUntestedFiles(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(untestedFiles) == 0 {
		fmt.Println("All Go source files have corresponding test files!")
		return
	}

	fmt.Printf("Found %d Go source files without test files:\n\n", len(untestedFiles))

	for _, file := range untestedFiles {
		fmt.Println(file)
	}

	fmt.Printf("\nSummary: %d files need test coverage\n", len(untestedFiles))
}
