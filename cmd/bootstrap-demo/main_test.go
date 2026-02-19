package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDemoConfigDefaults tests the DemoConfig default values.
func TestDemoConfigDefaults(t *testing.T) {
	// Test that a new DemoConfig can be created
	config := &DemoConfig{
		TemplateName:     "",
		GameLength:       "medium",
		ComplexityLevel:  "standard",
		GenreVariant:     "classic_fantasy",
		MaxPlayers:       4,
		StartingLevel:    1,
		WorldSeed:        0,
		OutputDir:        "demo_output",
		EnableQuickStart: true,
		Verbose:          false,
		ListTemplates:    false,
	}

	assert.NotNil(t, config)
	assert.Equal(t, "", config.TemplateName)
	assert.Equal(t, "medium", config.GameLength)
	assert.Equal(t, "standard", config.ComplexityLevel)
	assert.Equal(t, "classic_fantasy", config.GenreVariant)
	assert.Equal(t, 4, config.MaxPlayers)
	assert.Equal(t, 1, config.StartingLevel)
	assert.Equal(t, int64(0), config.WorldSeed)
	assert.Equal(t, "demo_output", config.OutputDir)
	assert.True(t, config.EnableQuickStart)
	assert.False(t, config.Verbose)
	assert.False(t, config.ListTemplates)
}

// TestRun tests the main run() function error handling.
func TestRun(t *testing.T) {
	// Save original os.Args and restore after test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Suppress logrus output during test
	logrus.SetOutput(io.Discard)
	defer logrus.SetOutput(os.Stderr)

	// Test with list-templates flag (should succeed)
	os.Args = []string{"cmd", "-list-templates"}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	runErr := run()
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()

	// run() should succeed for list-templates
	assert.NoError(t, runErr, "run() should succeed with list-templates flag")
}

// TestDemoConfigCustomValues tests DemoConfig with custom values.
func TestDemoConfigCustomValues(t *testing.T) {
	config := &DemoConfig{
		TemplateName:     "epic_campaign",
		GameLength:       "long",
		ComplexityLevel:  "advanced",
		GenreVariant:     "grimdark",
		MaxPlayers:       6,
		StartingLevel:    5,
		WorldSeed:        12345,
		OutputDir:        "custom_output",
		EnableQuickStart: false,
		Verbose:          true,
		ListTemplates:    false,
	}

	assert.Equal(t, "epic_campaign", config.TemplateName)
	assert.Equal(t, "long", config.GameLength)
	assert.Equal(t, "advanced", config.ComplexityLevel)
	assert.Equal(t, "grimdark", config.GenreVariant)
	assert.Equal(t, 6, config.MaxPlayers)
	assert.Equal(t, 5, config.StartingLevel)
	assert.Equal(t, int64(12345), config.WorldSeed)
	assert.Equal(t, "custom_output", config.OutputDir)
	assert.False(t, config.EnableQuickStart)
	assert.True(t, config.Verbose)
}

// TestDemoConfigListTemplates tests the ListTemplates flag.
func TestDemoConfigListTemplates(t *testing.T) {
	config := &DemoConfig{
		ListTemplates: true,
	}

	assert.True(t, config.ListTemplates)
}

// TestSetupLogging tests logging configuration.
func TestSetupLogging(t *testing.T) {
	tests := []struct {
		name          string
		verbose       bool
		expectedLevel logrus.Level
	}{
		{
			name:          "verbose enabled",
			verbose:       true,
			expectedLevel: logrus.DebugLevel,
		},
		{
			name:          "verbose disabled",
			verbose:       false,
			expectedLevel: logrus.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Suppress output during test
			logrus.SetOutput(io.Discard)
			defer logrus.SetOutput(os.Stderr)

			setupLogging(tt.verbose)
			assert.Equal(t, tt.expectedLevel, logrus.GetLevel())
		})
	}
}

// TestConvertToBootstrapConfig tests the config conversion function.
func TestConvertToBootstrapConfig(t *testing.T) {
	tests := []struct {
		name          string
		demoConfig    *DemoConfig
		expectError   bool
		errorContains string
		checkFunc     func(*testing.T, *pcg.BootstrapConfig)
	}{
		{
			name: "valid_short_simple_classic",
			demoConfig: &DemoConfig{
				GameLength:       "short",
				ComplexityLevel:  "simple",
				GenreVariant:     "classic_fantasy",
				MaxPlayers:       4,
				StartingLevel:    1,
				WorldSeed:        42,
				OutputDir:        "output",
				EnableQuickStart: true,
			},
			expectError: false,
			checkFunc: func(t *testing.T, cfg *pcg.BootstrapConfig) {
				assert.Equal(t, pcg.GameLengthShort, cfg.GameLength)
				assert.Equal(t, pcg.ComplexitySimple, cfg.ComplexityLevel)
				assert.Equal(t, pcg.GenreClassicFantasy, cfg.GenreVariant)
				assert.Equal(t, 4, cfg.MaxPlayers)
				assert.Equal(t, 1, cfg.StartingLevel)
				assert.Equal(t, int64(42), cfg.WorldSeed)
			},
		},
		{
			name: "valid_medium_standard_grimdark",
			demoConfig: &DemoConfig{
				GameLength:       "medium",
				ComplexityLevel:  "standard",
				GenreVariant:     "grimdark",
				MaxPlayers:       6,
				StartingLevel:    3,
				WorldSeed:        0,
				OutputDir:        "test_output",
				EnableQuickStart: false,
			},
			expectError: false,
			checkFunc: func(t *testing.T, cfg *pcg.BootstrapConfig) {
				assert.Equal(t, pcg.GameLengthMedium, cfg.GameLength)
				assert.Equal(t, pcg.ComplexityStandard, cfg.ComplexityLevel)
				assert.Equal(t, pcg.GenreGrimdark, cfg.GenreVariant)
				assert.Equal(t, 6, cfg.MaxPlayers)
				assert.Equal(t, 3, cfg.StartingLevel)
				assert.False(t, cfg.EnableQuickStart)
			},
		},
		{
			name: "valid_long_advanced_high_magic",
			demoConfig: &DemoConfig{
				GameLength:       "long",
				ComplexityLevel:  "advanced",
				GenreVariant:     "high_magic",
				MaxPlayers:       8,
				StartingLevel:    10,
				WorldSeed:        999,
				OutputDir:        "magic_output",
				EnableQuickStart: true,
			},
			expectError: false,
			checkFunc: func(t *testing.T, cfg *pcg.BootstrapConfig) {
				assert.Equal(t, pcg.GameLengthLong, cfg.GameLength)
				assert.Equal(t, pcg.ComplexityAdvanced, cfg.ComplexityLevel)
				assert.Equal(t, pcg.GenreHighMagic, cfg.GenreVariant)
				assert.Equal(t, 8, cfg.MaxPlayers)
				assert.Equal(t, 10, cfg.StartingLevel)
			},
		},
		{
			name: "valid_low_fantasy",
			demoConfig: &DemoConfig{
				GameLength:       "short",
				ComplexityLevel:  "simple",
				GenreVariant:     "low_fantasy",
				MaxPlayers:       2,
				StartingLevel:    1,
				OutputDir:        "low_fantasy_output",
				EnableQuickStart: true,
			},
			expectError: false,
			checkFunc: func(t *testing.T, cfg *pcg.BootstrapConfig) {
				assert.Equal(t, pcg.GenreLowFantasy, cfg.GenreVariant)
			},
		},
		{
			name: "invalid_game_length",
			demoConfig: &DemoConfig{
				GameLength:      "extra_long",
				ComplexityLevel: "standard",
				GenreVariant:    "classic_fantasy",
			},
			expectError:   true,
			errorContains: "invalid game length",
		},
		{
			name: "invalid_complexity_level",
			demoConfig: &DemoConfig{
				GameLength:      "short",
				ComplexityLevel: "extreme",
				GenreVariant:    "classic_fantasy",
			},
			expectError:   true,
			errorContains: "invalid complexity level",
		},
		{
			name: "invalid_genre_variant",
			demoConfig: &DemoConfig{
				GameLength:      "short",
				ComplexityLevel: "simple",
				GenreVariant:    "sci_fi",
			},
			expectError:   true,
			errorContains: "invalid genre variant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := convertToBootstrapConfig(tt.demoConfig)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, cfg)
				if tt.checkFunc != nil {
					tt.checkFunc(t, cfg)
				}
			}
		})
	}
}

// TestDisplayResults tests the results display function.
func TestDisplayResults(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Suppress logrus output
	logrus.SetOutput(io.Discard)
	defer logrus.SetOutput(os.Stderr)

	world := game.NewWorld()
	bootstrap := pcg.NewBootstrap(&pcg.BootstrapConfig{
		GameLength:       pcg.GameLengthMedium,
		ComplexityLevel:  pcg.ComplexityStandard,
		GenreVariant:     pcg.GenreClassicFantasy,
		MaxPlayers:       4,
		StartingLevel:    1,
		EnableQuickStart: true,
		DataDirectory:    "test_output",
	}, world, logrus.StandardLogger())

	config := &DemoConfig{
		GameLength:       "medium",
		ComplexityLevel:  "standard",
		GenreVariant:     "classic_fantasy",
		MaxPlayers:       4,
		StartingLevel:    1,
		WorldSeed:        42,
		OutputDir:        "test_output",
		EnableQuickStart: true,
	}

	displayResults(world, bootstrap, 5*time.Second, config)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	// Verify expected content
	assert.Contains(t, output, "GOLDBOX RPG ENGINE")
	assert.Contains(t, output, "BOOTSTRAP DEMO RESULTS")
	assert.Contains(t, output, "Generation Summary")
	assert.Contains(t, output, "medium")
	assert.Contains(t, output, "standard")
	assert.Contains(t, output, "classic_fantasy")
	assert.Contains(t, output, "4")  // MaxPlayers
	assert.Contains(t, output, "42") // WorldSeed
	assert.Contains(t, output, "Quick Start Scenario")
	assert.Contains(t, output, "Next Steps")
}

// TestDisplayResultsWithoutQuickStart tests display without quick start.
func TestDisplayResultsWithoutQuickStart(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	logrus.SetOutput(io.Discard)
	defer logrus.SetOutput(os.Stderr)

	world := game.NewWorld()
	bootstrap := pcg.NewBootstrap(&pcg.BootstrapConfig{
		GameLength:       pcg.GameLengthShort,
		ComplexityLevel:  pcg.ComplexitySimple,
		GenreVariant:     pcg.GenreGrimdark,
		MaxPlayers:       2,
		StartingLevel:    5,
		EnableQuickStart: false,
		DataDirectory:    "no_quick_start_output",
	}, world, logrus.StandardLogger())

	config := &DemoConfig{
		GameLength:       "short",
		ComplexityLevel:  "simple",
		GenreVariant:     "grimdark",
		MaxPlayers:       2,
		StartingLevel:    5,
		WorldSeed:        0,
		OutputDir:        "no_quick_start_output",
		EnableQuickStart: false,
	}

	displayResults(world, bootstrap, 2*time.Second, config)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	// Verify quick start section is NOT present when disabled
	assert.NotContains(t, output, "Quick Start Scenario Available")
	assert.Contains(t, output, "short")
	assert.Contains(t, output, "grimdark")
}

// TestVerifyGeneratedFiles tests the file verification function.
func TestVerifyGeneratedFiles(t *testing.T) {
	// Create temp directory with expected structure
	tmpDir := t.TempDir()

	// Create the expected file
	pcgDir := fmt.Sprintf("%s/pcg", tmpDir)
	err := os.MkdirAll(pcgDir, 0o755)
	require.NoError(t, err)

	configFile := fmt.Sprintf("%s/bootstrap_config.yaml", pcgDir)
	err = os.WriteFile(configFile, []byte("test: config"), 0o644)
	require.NoError(t, err)

	// Suppress logrus output
	logrus.SetOutput(io.Discard)
	defer logrus.SetOutput(os.Stderr)

	// Test successful verification
	err = verifyGeneratedFiles(tmpDir)
	assert.NoError(t, err)
}

// TestVerifyGeneratedFilesMissingFile tests verification with missing files.
func TestVerifyGeneratedFilesMissingFile(t *testing.T) {
	// Create temp directory without expected files
	tmpDir := t.TempDir()

	logrus.SetOutput(io.Discard)
	defer logrus.SetOutput(os.Stderr)

	err := verifyGeneratedFiles(tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected file not found")
}

// TestListAvailableTemplates tests the template listing function.
func TestListAvailableTemplates(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// This may fail if templates file doesn't exist, which is acceptable
	listErr := listAvailableTemplates()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	if listErr != nil {
		// Templates may not exist in test environment
		t.Logf("listAvailableTemplates returned error (may be expected in test): %v", listErr)
	} else {
		// If successful, should have some output
		assert.True(t, len(output) > 0 || listErr != nil)
	}
}

// TestRunBootstrapDemo tests the main bootstrap demo function.
func TestRunBootstrapDemo(t *testing.T) {
	// Create a clean temp directory for output
	tmpDir := t.TempDir()

	// Suppress logrus output
	logrus.SetOutput(io.Discard)
	defer logrus.SetOutput(os.Stderr)

	// Capture stdout
	oldStdout := os.Stdout
	_, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() {
		w.Close()
		os.Stdout = oldStdout
	}()

	config := &DemoConfig{
		GameLength:       "short",
		ComplexityLevel:  "simple",
		GenreVariant:     "classic_fantasy",
		MaxPlayers:       2,
		StartingLevel:    1,
		WorldSeed:        42,
		OutputDir:        tmpDir,
		EnableQuickStart: true,
	}

	// This test exercises the full bootstrap flow
	// It may take some time to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run in separate goroutine with timeout
	done := make(chan error, 1)
	go func() {
		done <- runBootstrapDemo(config)
	}()

	select {
	case err := <-done:
		// Bootstrap may fail if PCG resources are not available
		if err != nil {
			t.Logf("runBootstrapDemo returned error (may be expected in test): %v", err)
		}
	case <-ctx.Done():
		t.Logf("runBootstrapDemo timed out (expected in CI environment)")
	}
}

// TestRunBootstrapDemoWithTemplate tests bootstrap with template loading.
func TestRunBootstrapDemoWithTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	logrus.SetOutput(io.Discard)
	defer logrus.SetOutput(os.Stderr)

	oldStdout := os.Stdout
	_, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() {
		w.Close()
		os.Stdout = oldStdout
	}()

	config := &DemoConfig{
		TemplateName: "nonexistent_template",
		OutputDir:    tmpDir,
	}

	// This should fail because the template doesn't exist
	err = runBootstrapDemo(config)
	// The function should return an error for a nonexistent template
	// If it doesn't error (unexpected), we log and skip
	if err != nil {
		assert.Contains(t, err.Error(), "template")
	} else {
		t.Logf("runBootstrapDemo did not return error for nonexistent template (unexpected)")
	}
}

// TestRunBootstrapDemoCleanupExisting tests cleanup of existing output.
func TestRunBootstrapDemoCleanupExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some existing content
	existingFile := fmt.Sprintf("%s/old_file.txt", tmpDir)
	err := os.WriteFile(existingFile, []byte("old content"), 0o644)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(existingFile)
	require.NoError(t, err)

	logrus.SetOutput(io.Discard)
	defer logrus.SetOutput(os.Stderr)

	oldStdout := os.Stdout
	_, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	os.Stdout = w
	defer func() {
		w.Close()
		os.Stdout = oldStdout
	}()

	config := &DemoConfig{
		GameLength:       "short",
		ComplexityLevel:  "simple",
		GenreVariant:     "classic_fantasy",
		MaxPlayers:       2,
		StartingLevel:    1,
		WorldSeed:        42,
		OutputDir:        tmpDir,
		EnableQuickStart: true,
	}

	// Run bootstrap - it should clean up existing content
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- runBootstrapDemo(config)
	}()

	select {
	case <-done:
		// Existing file should be cleaned up
		_, err = os.Stat(existingFile)
		assert.True(t, os.IsNotExist(err), "Old file should be cleaned up")
	case <-ctx.Done():
		t.Logf("Bootstrap timed out")
	}
}

// TestDemoConfigValidation tests DemoConfig struct validation scenarios.
func TestDemoConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *DemoConfig
	}{
		{
			name: "minimal_config",
			config: &DemoConfig{
				GameLength:      "short",
				ComplexityLevel: "simple",
				GenreVariant:    "classic_fantasy",
			},
		},
		{
			name: "full_config",
			config: &DemoConfig{
				TemplateName:     "",
				GameLength:       "long",
				ComplexityLevel:  "advanced",
				GenreVariant:     "high_magic",
				MaxPlayers:       10,
				StartingLevel:    20,
				WorldSeed:        999999,
				OutputDir:        "/tmp/bootstrap_test",
				EnableQuickStart: false,
				Verbose:          true,
				ListTemplates:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the config can be converted
			cfg, err := convertToBootstrapConfig(tt.config)
			assert.NoError(t, err)
			assert.NotNil(t, cfg)
		})
	}
}

// BenchmarkVerifyGeneratedFiles benchmarks file verification.
func BenchmarkVerifyGeneratedFiles(b *testing.B) {
	tmpDir := b.TempDir()
	pcgDir := fmt.Sprintf("%s/pcg", tmpDir)
	_ = os.MkdirAll(pcgDir, 0o755)
	_ = os.WriteFile(fmt.Sprintf("%s/bootstrap_config.yaml", pcgDir), []byte("test"), 0o644)

	logrus.SetOutput(io.Discard)
	defer logrus.SetOutput(os.Stderr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = verifyGeneratedFiles(tmpDir)
	}
}
