package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// findUntestedFiles finds Go source files that don't have corresponding test files
func findUntestedFiles(rootDir string) ([]string, error) {
	sourceFiles := make(map[string]bool)
	testFiles := make(map[string]bool)
	var untestedFiles []string

	// Walk through the directory tree
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Get relative path from root directory
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		fileName := info.Name()

		if strings.HasSuffix(fileName, "_test.go") {
			// This is a test file - extract the base name
			baseName := strings.TrimSuffix(fileName, "_test.go")
			dir := filepath.Dir(relPath)
			baseFile := filepath.Join(dir, baseName+".go")
			testFiles[baseFile] = true
		} else {
			// This is a source file
			sourceFiles[relPath] = true
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Find source files without corresponding test files
	for sourceFile := range sourceFiles {
		if !testFiles[sourceFile] {
			// Skip main.go files as they typically don't have tests
			if strings.HasSuffix(sourceFile, "main.go") {
				continue
			}
			untestedFiles = append(untestedFiles, sourceFile)
		}
	}

	// Sort the results for consistent output
	sort.Strings(untestedFiles)

	return untestedFiles, nil
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
