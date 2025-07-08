package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// TestFindUntestedFiles_EmptyDirectory tests behavior with an empty directory
func TestFindUntestedFiles_EmptyDirectory_ReturnsEmptySlice(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_empty_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	result, err := findUntestedFiles(tempDir)
	if err != nil {
		t.Errorf("findUntestedFiles() error = %v, want nil", err)
	}

	if len(result) != 0 {
		t.Errorf("findUntestedFiles() = %v, want empty slice", result)
	}
}

// TestFindUntestedFiles_NonexistentDirectory tests behavior with invalid directory
func TestFindUntestedFiles_NonexistentDirectory_ReturnsError(t *testing.T) {
	nonexistentDir := "/path/that/does/not/exist"

	result, err := findUntestedFiles(nonexistentDir)
	if err == nil {
		t.Error("findUntestedFiles() error = nil, want error for nonexistent directory")
	}

	if result != nil {
		t.Errorf("findUntestedFiles() = %v, want nil when error occurs", result)
	}
}

// TestFindUntestedFiles_OnlyTestFiles tests directory with only test files
func TestFindUntestedFiles_OnlyTestFiles_ReturnsEmptySlice(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_only_tests_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files only
	testFiles := []string{"example_test.go", "utils_test.go"}
	for _, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte("package main\n"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	result, err := findUntestedFiles(tempDir)
	if err != nil {
		t.Errorf("findUntestedFiles() error = %v, want nil", err)
	}

	if len(result) != 0 {
		t.Errorf("findUntestedFiles() = %v, want empty slice for test-only directory", result)
	}
}

// TestFindUntestedFiles_SourceWithoutTests tests source files without corresponding tests
func TestFindUntestedFiles_SourceWithoutTests_ReturnsUntestedFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_source_no_tests_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create source files without tests
	sourceFiles := []string{"handler.go", "utils.go", "config.go"}
	for _, filename := range sourceFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte("package main\n"), 0644); err != nil {
			t.Fatalf("Failed to create source file %s: %v", filename, err)
		}
	}

	result, err := findUntestedFiles(tempDir)
	if err != nil {
		t.Errorf("findUntestedFiles() error = %v, want nil", err)
	}

	expected := []string{"config.go", "handler.go", "utils.go"}
	sort.Strings(result)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("findUntestedFiles() = %v, want %v", result, expected)
	}
}

// TestFindUntestedFiles_MixedFiles tests directory with both tested and untested files
func TestFindUntestedFiles_MixedFiles_ReturnsOnlyUntested(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_mixed_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create source files
	sourceFiles := []string{"tested.go", "untested.go", "utils.go"}
	for _, filename := range sourceFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte("package main\n"), 0644); err != nil {
			t.Fatalf("Failed to create source file %s: %v", filename, err)
		}
	}

	// Create test file for only one source file
	testFile := filepath.Join(tempDir, "tested_test.go")
	if err := os.WriteFile(testFile, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := findUntestedFiles(tempDir)
	if err != nil {
		t.Errorf("findUntestedFiles() error = %v, want nil", err)
	}

	expected := []string{"untested.go", "utils.go"}
	sort.Strings(result)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("findUntestedFiles() = %v, want %v", result, expected)
	}
}

// TestFindUntestedFiles_MainGoFiles tests that main.go files are excluded
func TestFindUntestedFiles_MainGoFiles_ExcludesMainFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_main_exclusion_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create main.go and regular source files
	files := []string{"main.go", "handler.go"}
	for _, filename := range files {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte("package main\n"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	result, err := findUntestedFiles(tempDir)
	if err != nil {
		t.Errorf("findUntestedFiles() error = %v, want nil", err)
	}

	expected := []string{"handler.go"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("findUntestedFiles() = %v, want %v (main.go should be excluded)", result, expected)
	}
}

// TestFindUntestedFiles_NestedDirectories tests behavior with nested directory structure
func TestFindUntestedFiles_NestedDirectories_TraversesRecursively(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_nested_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create nested directory structure
	nestedDir := filepath.Join(tempDir, "pkg", "utils")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Create files in different levels
	files := map[string]string{
		"root.go":                  tempDir,
		"pkg/handler.go":           filepath.Join(tempDir, "pkg"),
		"pkg/utils/helper.go":      nestedDir,
		"pkg/utils/helper_test.go": nestedDir,
	}

	for filename, dir := range files {
		fullPath := filepath.Join(dir, filepath.Base(filename))
		if err := os.WriteFile(fullPath, []byte("package main\n"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	result, err := findUntestedFiles(tempDir)
	if err != nil {
		t.Errorf("findUntestedFiles() error = %v, want nil", err)
	}

	// Expected relative paths from tempDir
	expected := []string{
		filepath.Join("pkg", "handler.go"),
		"root.go",
	}
	sort.Strings(result)
	sort.Strings(expected)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("findUntestedFiles() = %v, want %v", result, expected)
	}
}

// TestFindUntestedFiles_NonGoFiles tests that non-Go files are ignored
func TestFindUntestedFiles_NonGoFiles_IgnoresNonGoFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_non_go_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create Go files and non-Go files
	files := []string{"valid.go", "README.md", "config.yaml", "script.sh", "data.json"}
	for _, filename := range files {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte("content\n"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	result, err := findUntestedFiles(tempDir)
	if err != nil {
		t.Errorf("findUntestedFiles() error = %v, want nil", err)
	}

	expected := []string{"valid.go"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("findUntestedFiles() = %v, want %v (non-Go files should be ignored)", result, expected)
	}
}

// TestFindUntestedFiles_ResultsSorted tests that results are consistently sorted
func TestFindUntestedFiles_ResultsSorted_ReturnsAlphabeticallySortedResults(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_sorting_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create files in non-alphabetical order
	files := []string{"zebra.go", "alpha.go", "beta.go", "gamma.go"}
	for _, filename := range files {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte("package main\n"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	result, err := findUntestedFiles(tempDir)
	if err != nil {
		t.Errorf("findUntestedFiles() error = %v, want nil", err)
	}

	// Check if result is sorted
	expected := []string{"alpha.go", "beta.go", "gamma.go", "zebra.go"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("findUntestedFiles() = %v, want %v (results should be sorted)", result, expected)
	}

	// Additional check: verify it's actually sorted
	if !sort.StringsAreSorted(result) {
		t.Error("findUntestedFiles() results are not sorted alphabetically")
	}
}

// TestFindUntestedFiles_EdgeCases_TableDriven tests various edge cases using table-driven approach
func TestFindUntestedFiles_EdgeCases_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		expectedCount int
		description   string
	}{
		{
			name:          "EmptyFileNames",
			files:         []string{},
			expectedCount: 0,
			description:   "No files should result in empty output",
		},
		{
			name:          "OnlyTestFiles",
			files:         []string{"example_test.go", "another_test.go"},
			expectedCount: 0,
			description:   "Only test files should result in no untested files",
		},
		{
			name:          "OnlyMainFiles",
			files:         []string{"main.go", "cmd/main.go"},
			expectedCount: 0,
			description:   "Only main.go files should be excluded",
		},
		{
			name:          "MixedScenario",
			files:         []string{"handler.go", "handler_test.go", "utils.go", "main.go", "config_test.go"},
			expectedCount: 1,
			description:   "Mixed scenario should return only untested non-main files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "test_edge_case_*")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create test files based on the scenario
			for _, filename := range tt.files {
				var filePath string
				if filepath.Dir(filename) != "." {
					// Handle nested paths
					dir := filepath.Join(tempDir, filepath.Dir(filename))
					if err := os.MkdirAll(dir, 0755); err != nil {
						t.Fatalf("Failed to create directory %s: %v", dir, err)
					}
					filePath = filepath.Join(tempDir, filename)
				} else {
					filePath = filepath.Join(tempDir, filename)
				}

				if err := os.WriteFile(filePath, []byte("package main\n"), 0644); err != nil {
					t.Fatalf("Failed to create file %s: %v", filename, err)
				}
			}

			result, err := findUntestedFiles(tempDir)
			if err != nil {
				t.Errorf("findUntestedFiles() error = %v, want nil", err)
			}

			if len(result) != tt.expectedCount {
				t.Errorf("findUntestedFiles() returned %d files, want %d files. Description: %s. Files: %v",
					len(result), tt.expectedCount, tt.description, result)
			}
		})
	}
}
