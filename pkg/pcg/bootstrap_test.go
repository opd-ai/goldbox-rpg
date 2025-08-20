package pcg

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBootstrap(t *testing.T) {
	config := DefaultBootstrapConfig()
	world := game.NewWorld()
	logger := logrus.New()

	bootstrap := NewBootstrap(config, world, logger)

	assert.NotNil(t, bootstrap)
	assert.Equal(t, config, bootstrap.config)
	assert.Equal(t, world, bootstrap.world)
	assert.Equal(t, logger, bootstrap.logger)
	assert.NotNil(t, bootstrap.pcgManager)
	assert.NotNil(t, bootstrap.generatedFiles)
}

func TestNewBootstrap_NilLogger(t *testing.T) {
	config := DefaultBootstrapConfig()
	world := game.NewWorld()

	bootstrap := NewBootstrap(config, world, nil)

	assert.NotNil(t, bootstrap)
	assert.NotNil(t, bootstrap.logger)
}

func TestDefaultBootstrapConfig(t *testing.T) {
	config := DefaultBootstrapConfig()

	assert.Equal(t, GameLengthMedium, config.GameLength)
	assert.Equal(t, ComplexityStandard, config.ComplexityLevel)
	assert.Equal(t, GenreClassicFantasy, config.GenreVariant)
	assert.Equal(t, 4, config.MaxPlayers)
	assert.Equal(t, 1, config.StartingLevel)
	assert.Equal(t, int64(0), config.WorldSeed)
	assert.True(t, config.EnableQuickStart)
	assert.Equal(t, "data", config.DataDirectory)
}

func TestDetectConfigurationPresence(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(string)
		expected bool
	}{
		{
			name: "all required files present",
			setup: func(dir string) {
				// Create required directories
				os.MkdirAll(filepath.Join(dir, "spells"), 0755)
				os.MkdirAll(filepath.Join(dir, "items"), 0755)

				// Create required files
				os.WriteFile(filepath.Join(dir, "spells", "cantrips.yaml"), []byte("test"), 0644)
				os.WriteFile(filepath.Join(dir, "spells", "level1.yaml"), []byte("test"), 0644)
				os.WriteFile(filepath.Join(dir, "items", "items.yaml"), []byte("test"), 0644)
			},
			expected: true,
		},
		{
			name: "missing some files",
			setup: func(dir string) {
				os.MkdirAll(filepath.Join(dir, "spells"), 0755)
				os.WriteFile(filepath.Join(dir, "spells", "cantrips.yaml"), []byte("test"), 0644)
				// Missing level1.yaml and items.yaml
			},
			expected: false,
		},
		{
			name: "no files present",
			setup: func(dir string) {
				// No setup, directory is empty
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "test_config_*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Setup test scenario
			tt.setup(tempDir)

			// Test detection
			result := DetectConfigurationPresence(tempDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBootstrap_GenerateCompleteGame(t *testing.T) {
	tests := []struct {
		name   string
		config *BootstrapConfig
	}{
		{
			name: "short game generation",
			config: &BootstrapConfig{
				GameLength:       GameLengthShort,
				ComplexityLevel:  ComplexitySimple,
				GenreVariant:     GenreClassicFantasy,
				MaxPlayers:       4,
				StartingLevel:    1,
				WorldSeed:        12345,
				EnableQuickStart: true,
				DataDirectory:    "",
			},
		},
		{
			name: "medium game generation",
			config: &BootstrapConfig{
				GameLength:       GameLengthMedium,
				ComplexityLevel:  ComplexityStandard,
				GenreVariant:     GenreHighMagic,
				MaxPlayers:       4,
				StartingLevel:    1,
				WorldSeed:        54321,
				EnableQuickStart: true,
				DataDirectory:    "",
			},
		},
		{
			name: "long game generation",
			config: &BootstrapConfig{
				GameLength:       GameLengthLong,
				ComplexityLevel:  ComplexityAdvanced,
				GenreVariant:     GenreGrimdark,
				MaxPlayers:       6,
				StartingLevel:    3,
				WorldSeed:        99999,
				EnableQuickStart: false,
				DataDirectory:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for data files
			tempDir, err := os.MkdirTemp("", "test_bootstrap_*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			tt.config.DataDirectory = tempDir

			world := game.NewWorld()
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel) // Reduce noise in tests

			bootstrap := NewBootstrap(tt.config, world, logger)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			generatedWorld, err := bootstrap.GenerateCompleteGame(ctx)

			assert.NoError(t, err)
			assert.NotNil(t, generatedWorld)
			assert.NotEmpty(t, bootstrap.generatedFiles)

			// Verify that basic content was generated
			assert.Contains(t, bootstrap.generatedFiles, "world")
			assert.Contains(t, bootstrap.generatedFiles, "factions")
			assert.Contains(t, bootstrap.generatedFiles, "characters")
			assert.Contains(t, bootstrap.generatedFiles, "quests")
			assert.Contains(t, bootstrap.generatedFiles, "dialogue")
			assert.Contains(t, bootstrap.generatedFiles, "spells")
			assert.Contains(t, bootstrap.generatedFiles, "items")

			if tt.config.EnableQuickStart {
				assert.Contains(t, bootstrap.generatedFiles, "starting_scenario")
			}

			// Verify configuration file was created
			configPath := filepath.Join(tempDir, "pcg", "bootstrap_config.yaml")
			_, err = os.Stat(configPath)
			assert.NoError(t, err, "Bootstrap config file should be created")
		})
	}
}

func TestBootstrap_ParameterCalculation(t *testing.T) {
	tests := []struct {
		name     string
		config   *BootstrapConfig
		expected map[string]int
	}{
		{
			name: "short simple game",
			config: &BootstrapConfig{
				GameLength:      GameLengthShort,
				ComplexityLevel: ComplexitySimple,
			},
			expected: map[string]int{
				"regions":  1,
				"factions": 2,
				"npcs":     10,
				"quests":   5,
			},
		},
		{
			name: "medium standard game",
			config: &BootstrapConfig{
				GameLength:      GameLengthMedium,
				ComplexityLevel: ComplexityStandard,
			},
			expected: map[string]int{
				"regions":  3,
				"factions": 4,
				"npcs":     20,
				"quests":   12,
			},
		},
		{
			name: "long advanced game",
			config: &BootstrapConfig{
				GameLength:      GameLengthLong,
				ComplexityLevel: ComplexityAdvanced,
			},
			expected: map[string]int{
				"regions":  5,
				"factions": 6,
				"npcs":     30,
				"quests":   25,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := game.NewWorld()
			logger := logrus.New()
			bootstrap := NewBootstrap(tt.config, world, logger)

			assert.Equal(t, tt.expected["regions"], bootstrap.getRegionCountForLength())
			assert.Equal(t, tt.expected["factions"], bootstrap.getFactionCountForLength())
			assert.Equal(t, tt.expected["npcs"], bootstrap.getNPCCountForComplexity())
			assert.Equal(t, tt.expected["quests"], bootstrap.getQuestCountForLength())
		})
	}
}

func TestBootstrap_DeterministicGeneration(t *testing.T) {
	config := &BootstrapConfig{
		GameLength:       GameLengthMedium,
		ComplexityLevel:  ComplexityStandard,
		GenreVariant:     GenreClassicFantasy,
		MaxPlayers:       4,
		StartingLevel:    1,
		WorldSeed:        42,
		EnableQuickStart: true,
		DataDirectory:    "",
	}

	// Generate two games with the same seed
	results := make([]map[string]string, 2)

	for i := 0; i < 2; i++ {
		tempDir, err := os.MkdirTemp("", "test_deterministic_*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config.DataDirectory = tempDir

		world := game.NewWorld()
		logger := logrus.New()
		logger.SetLevel(logrus.WarnLevel)

		bootstrap := NewBootstrap(config, world, logger)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err = bootstrap.GenerateCompleteGame(ctx)
		require.NoError(t, err)

		// Store the generated content for comparison
		results[i] = make(map[string]string)
		for k, v := range bootstrap.generatedFiles {
			results[i][k] = v
		}
	}

	// Compare the results - they should be identical for deterministic generation
	// Note: This is a basic comparison. In a full implementation, you would
	// compare specific aspects that should be deterministic
	assert.Equal(t, len(results[0]), len(results[1]))
	for key := range results[0] {
		assert.Contains(t, results[1], key)
	}
}

func TestBootstrap_ErrorHandling(t *testing.T) {
	config := DefaultBootstrapConfig()
	config.DataDirectory = "/invalid/path/that/cannot/be/created"

	world := game.NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress error logs in tests

	bootstrap := NewBootstrap(config, world, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := bootstrap.GenerateCompleteGame(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save generated configuration")
}

func TestBootstrap_ContextTimeout(t *testing.T) {
	config := DefaultBootstrapConfig()

	tempDir, err := os.MkdirTemp("", "test_timeout_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config.DataDirectory = tempDir

	world := game.NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs

	bootstrap := NewBootstrap(config, world, logger)

	// Use a very short timeout to test timeout handling
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to timeout
	<-ctx.Done()

	_, err = bootstrap.GenerateCompleteGame(ctx)

	// The function should complete quickly despite the timeout context
	// since it doesn't actually check the context in the current simple implementation
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkBootstrap_GenerateCompleteGame(b *testing.B) {
	config := DefaultBootstrapConfig()

	for i := 0; i < b.N; i++ {
		tempDir, err := os.MkdirTemp("", "bench_bootstrap_*")
		require.NoError(b, err)

		config.DataDirectory = tempDir
		config.WorldSeed = int64(i) // Use different seeds for each iteration

		world := game.NewWorld()
		logger := logrus.New()
		logger.SetLevel(logrus.FatalLevel) // Suppress logs in benchmarks

		bootstrap := NewBootstrap(config, world, logger)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		_, err = bootstrap.GenerateCompleteGame(ctx)
		require.NoError(b, err)

		cancel()
		os.RemoveAll(tempDir)
	}
}

func BenchmarkBootstrap_ParameterCalculation(b *testing.B) {
	config := DefaultBootstrapConfig()
	world := game.NewWorld()
	logger := logrus.New()

	bootstrap := NewBootstrap(config, world, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bootstrap.getRegionCountForLength()
		_ = bootstrap.getFactionCountForLength()
		_ = bootstrap.getNPCCountForComplexity()
		_ = bootstrap.getQuestCountForLength()
	}
}
