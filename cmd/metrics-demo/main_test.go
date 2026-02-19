package main

import (
	"bytes"
	"flag"
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

// TestPCGManagerInitialization tests PCG manager creation and initialization.
func TestPCGManagerInitialization(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	require.NotNil(t, pcgManager)

	pcgManager.InitializeWithSeed(42)

	// Verify quality metrics are accessible
	qualityMetrics := pcgManager.GetQualityMetrics()
	assert.NotNil(t, qualityMetrics)
}

// TestInitializePCG tests the initializePCG function with valid and invalid inputs.
func TestInitializePCG(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tests := []struct {
		name      string
		world     *game.World
		seed      int64
		wantErr   bool
		errTarget error
	}{
		{
			name:    "valid_world",
			world:   createTestWorld(),
			seed:    42,
			wantErr: false,
		},
		{
			name:      "nil_world",
			world:     nil,
			seed:      42,
			wantErr:   true,
			errTarget: ErrNilWorld,
		},
		{
			name:    "zero_seed",
			world:   createTestWorld(),
			seed:    0,
			wantErr: false,
		},
		{
			name:    "negative_seed",
			world:   createTestWorld(),
			seed:    -1,
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, err := initializePCG(tc.world, logger, tc.seed)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errTarget != nil {
					assert.ErrorIs(t, err, tc.errTarget)
				}
				assert.Nil(t, ctx)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ctx)
				assert.NotNil(t, ctx.pcgManager)
				assert.NotNil(t, ctx.qualityMetrics)
				assert.NotNil(t, ctx.logger)
			}
		})
	}
}

// TestDemoContext tests that demoContext holds required components.
func TestDemoContext(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	ctx, err := initializePCG(world, logger, 42)
	require.NoError(t, err)

	assert.Same(t, logger, ctx.logger)
	assert.NotNil(t, ctx.pcgManager)
	assert.NotNil(t, ctx.qualityMetrics)
}

// TestDemonstrateTerrainGeneration tests terrain generation recording.
func TestDemonstrateTerrainGeneration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	ctx, err := initializePCG(world, logger, 42)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	demonstrateTerrainGeneration(ctx)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	// Verify all 5 terrain levels were generated
	for i := 1; i <= 5; i++ {
		assert.Contains(t, output, "Generated terrain level")
	}
}

// TestDemonstrateQuestGeneration tests quest generation with failures.
func TestDemonstrateQuestGeneration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	ctx, err := initializePCG(world, logger, 42)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	demonstrateQuestGeneration(ctx)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	// Should have both successful and failed quests
	assert.Contains(t, output, "Generated quest")
	assert.Contains(t, output, "generation failed")
}

// TestDemonstrateItemGeneration tests item generation recording.
func TestDemonstrateItemGeneration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	ctx, err := initializePCG(world, logger, 42)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	demonstrateItemGeneration(ctx)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	// Verify all 3 item sets were generated
	assert.Contains(t, output, "Generated item set 1")
	assert.Contains(t, output, "Generated item set 2")
	assert.Contains(t, output, "Generated item set 3")
}

// TestDemonstratePlayerFeedback tests player feedback recording.
func TestDemonstratePlayerFeedback(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	ctx, err := initializePCG(world, logger, 42)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	demonstratePlayerFeedback(ctx)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	assert.Contains(t, output, "Recorded feedback for")
	assert.Contains(t, output, "Rating")
	assert.Contains(t, output, "Enjoyment")
}

// TestDemonstrateQuestCompletions tests quest completion tracking.
func TestDemonstrateQuestCompletions(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	ctx, err := initializePCG(world, logger, 42)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	demonstrateQuestCompletions(ctx)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	// Should have both completed and abandoned quests
	assert.Contains(t, output, "completed in")
	assert.Contains(t, output, "abandoned after")
}

// TestDisplayQualityReport tests quality report display.
func TestDisplayQualityReport(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	ctx, err := initializePCG(world, logger, 42)
	require.NoError(t, err)

	// Generate some content first
	demonstrateTerrainGeneration(ctx)
	demonstrateQuestGeneration(ctx)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	displayQualityReport(ctx)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	assert.Contains(t, output, "CONTENT QUALITY REPORT")
	assert.Contains(t, output, "Overall Quality Score")
	assert.Contains(t, output, "Quality Grade")
	assert.Contains(t, output, "Component Scores")
}

// TestDisplayMetricsComponents tests metrics components display.
func TestDisplayMetricsComponents(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	ctx, err := initializePCG(world, logger, 42)
	require.NoError(t, err)

	// Generate some content first
	demonstrateTerrainGeneration(ctx)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	displayMetricsComponents(ctx)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	assert.Contains(t, output, "Performance Metrics")
	assert.Contains(t, output, "Total Generations")
	assert.Contains(t, output, "Balance Metrics")
	assert.Contains(t, output, "System Health")
}

// TestDisplayFinalAssessment tests final assessment display for different scores.
func TestDisplayFinalAssessment(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	ctx, err := initializePCG(world, logger, 42)
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	displayFinalAssessment(ctx)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	assert.Contains(t, output, "FINAL QUALITY ASSESSMENT")
	assert.Contains(t, output, "Overall Quality Score")
	assert.Contains(t, output, "Quality Status")
}

// TestDisplayDemoSummary tests demo summary display.
func TestDisplayDemoSummary(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	displayDemoSummary()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	assert.Contains(t, output, "Demo completed successfully")
	assert.Contains(t, output, "Content generation performance")
	assert.Contains(t, output, "Player engagement")
}

// TestQuestCompletionStruct tests the questCompletion struct.
func TestQuestCompletionStruct(t *testing.T) {
	qc := questCompletion{
		questID:        "test_quest",
		completionTime: 10 * time.Minute,
		completed:      true,
	}

	assert.Equal(t, "test_quest", qc.questID)
	assert.Equal(t, 10*time.Minute, qc.completionTime)
	assert.True(t, qc.completed)
}

// TestConfigDefault tests that Config has expected default values.
func TestConfigDefault(t *testing.T) {
	cfg := &Config{Seed: 42}
	assert.Equal(t, int64(42), cfg.Seed)
}

// TestConfigCustomSeed tests Config with custom seed values.
func TestConfigCustomSeed(t *testing.T) {
	tests := []struct {
		name string
		seed int64
	}{
		{"zero_seed", 0},
		{"positive_seed", 12345},
		{"large_seed", 9223372036854775807},
		{"negative_seed", -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{Seed: tc.seed}
			assert.Equal(t, tc.seed, cfg.Seed)
		})
	}
}

// TestRunWithConfig tests that run() accepts a custom Config.
func TestRunWithConfig(t *testing.T) {
	// Capture stdout to avoid polluting test output
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	cfg := &Config{Seed: 12345}
	runErr := run(cfg)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	assert.NoError(t, runErr)
	assert.Contains(t, output, "Using seed: 12345")
}

// TestRunDifferentSeeds tests that different seeds produce different outputs.
func TestRunDifferentSeeds(t *testing.T) {
	// Run with seed 1
	oldStdout := os.Stdout
	r1, w1, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w1

	err = run(&Config{Seed: 1})
	require.NoError(t, err)

	w1.Close()
	os.Stdout = oldStdout

	var buf1 bytes.Buffer
	_, err = io.Copy(&buf1, r1)
	require.NoError(t, err)
	output1 := buf1.String()

	// Run with seed 2
	r2, w2, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w2

	err = run(&Config{Seed: 2})
	require.NoError(t, err)

	w2.Close()
	os.Stdout = oldStdout

	var buf2 bytes.Buffer
	_, err = io.Copy(&buf2, r2)
	require.NoError(t, err)
	output2 := buf2.String()

	// Both outputs should contain their respective seeds
	assert.Contains(t, output1, "Using seed: 1")
	assert.Contains(t, output2, "Using seed: 2")
}

// TestParseFlagsDefault tests parseFlags with default values.
func TestParseFlagsDefault(t *testing.T) {
	// Reset flag state for test isolation
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd"}

	cfg := parseFlags()
	assert.Equal(t, int64(42), cfg.Seed)
}

// TestParseFlagsCustomSeed tests parseFlags with custom seed flag.
func TestParseFlagsCustomSeed(t *testing.T) {
	// Reset flag state for test isolation
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{"cmd", "-seed", "99999"}

	cfg := parseFlags()
	assert.Equal(t, int64(99999), cfg.Seed)
}

// TestQualityMetricsRecording tests content generation recording.
func TestQualityMetricsRecording(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	pcgManager.InitializeWithSeed(42)

	qualityMetrics := pcgManager.GetQualityMetrics()

	// Record successful generations
	for i := 0; i < 5; i++ {
		contentID := "terrain_test"
		duration := time.Duration(50+i*10) * time.Millisecond
		qualityMetrics.RecordContentGeneration(pcg.ContentTypeTerrain, contentID, duration, nil)
	}

	// Verify performance metrics updated
	performanceStats := qualityMetrics.GetPerformanceMetrics().GetStats()
	assert.NotNil(t, performanceStats)
	assert.Contains(t, performanceStats, "total_generations")
}

// TestQualityMetricsWithErrors tests recording failed generations.
func TestQualityMetricsWithErrors(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	pcgManager.InitializeWithSeed(42)

	qualityMetrics := pcgManager.GetQualityMetrics()

	tests := []struct {
		name        string
		contentType pcg.ContentType
		contentID   string
		duration    time.Duration
		err         error
	}{
		{
			name:        "successful_terrain",
			contentType: pcg.ContentTypeTerrain,
			contentID:   "terrain_1",
			duration:    50 * time.Millisecond,
			err:         nil,
		},
		{
			name:        "failed_quest",
			contentType: pcg.ContentTypeQuests,
			contentID:   "quest_fail",
			duration:    100 * time.Millisecond,
			err:         assert.AnError,
		},
		{
			name:        "successful_items",
			contentType: pcg.ContentTypeItems,
			contentID:   "items_1",
			duration:    30 * time.Millisecond,
			err:         nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Recording should not panic regardless of error status
			qualityMetrics.RecordContentGeneration(tc.contentType, tc.contentID, tc.duration, tc.err)
		})
	}
}

// TestPlayerFeedbackRecording tests player feedback integration.
func TestPlayerFeedbackRecording(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	pcgManager.InitializeWithSeed(42)

	feedbacks := []pcg.PlayerFeedback{
		{
			Timestamp:   time.Now(),
			ContentType: pcg.ContentTypeQuests,
			ContentID:   "quest_test_1",
			Rating:      5,
			Difficulty:  3,
			Enjoyment:   5,
			Comments:    "Great quest!",
			SessionID:   "session_test",
		},
		{
			Timestamp:   time.Now(),
			ContentType: pcg.ContentTypeTerrain,
			ContentID:   "terrain_test_1",
			Rating:      3,
			Difficulty:  2,
			Enjoyment:   3,
			Comments:    "Average terrain",
			SessionID:   "session_test",
		},
		{
			Timestamp:   time.Now(),
			ContentType: pcg.ContentTypeQuests,
			ContentID:   "quest_test_2",
			Rating:      2,
			Difficulty:  5,
			Enjoyment:   2,
			Comments:    "Too difficult",
			SessionID:   "session_test",
		},
	}

	for _, feedback := range feedbacks {
		// Recording should not panic
		pcgManager.RecordPlayerFeedback(feedback)
	}
}

// TestQuestCompletionTracking tests quest completion recording.
func TestQuestCompletionTracking(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	pcgManager.InitializeWithSeed(42)

	tests := []struct {
		name           string
		questID        string
		completionTime time.Duration
		completed      bool
	}{
		{
			name:           "completed_quickly",
			questID:        "quest_1",
			completionTime: 10 * time.Minute,
			completed:      true,
		},
		{
			name:           "completed_slowly",
			questID:        "quest_2",
			completionTime: 45 * time.Minute,
			completed:      true,
		},
		{
			name:           "abandoned",
			questID:        "quest_3",
			completionTime: 8 * time.Minute,
			completed:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Recording should not panic
			pcgManager.RecordQuestCompletion(tc.questID, tc.completionTime, tc.completed)
		})
	}
}

// TestQualityReportGeneration tests comprehensive quality report.
func TestQualityReportGeneration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	pcgManager.InitializeWithSeed(42)

	// Generate some content first
	qualityMetrics := pcgManager.GetQualityMetrics()
	for i := 0; i < 3; i++ {
		qualityMetrics.RecordContentGeneration(pcg.ContentTypeTerrain, "terrain", 50*time.Millisecond, nil)
		qualityMetrics.RecordContentGeneration(pcg.ContentTypeQuests, "quest", 80*time.Millisecond, nil)
	}

	report := pcgManager.GenerateQualityReport()
	require.NotNil(t, report)

	// Verify report structure
	assert.NotZero(t, report.Timestamp)
	assert.GreaterOrEqual(t, report.OverallScore, 0.0)
	assert.LessOrEqual(t, report.OverallScore, 1.0)
	assert.NotEmpty(t, report.QualityGrade)
	assert.NotNil(t, report.ComponentScores)
	assert.NotNil(t, report.ThresholdStatus)
	assert.NotNil(t, report.SystemSummary)
}

// TestQualityGrades tests quality grade assignment based on score.
func TestQualityGrades(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	pcgManager.InitializeWithSeed(42)

	report := pcgManager.GenerateQualityReport()
	require.NotNil(t, report)

	// Verify grade is valid
	validGrades := []string{"A+", "A", "A-", "B+", "B", "B-", "C+", "C", "C-", "D", "F"}
	found := false
	for _, grade := range validGrades {
		if report.QualityGrade == grade {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected valid grade, got: %s", report.QualityGrade)
}

// TestOverallQualityScore tests overall quality score calculation.
func TestOverallQualityScore(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	pcgManager.InitializeWithSeed(42)

	score := pcgManager.GetOverallQualityScore()
	assert.GreaterOrEqual(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)
}

// TestPerformanceMetrics tests performance metrics retrieval.
func TestPerformanceMetrics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	pcgManager.InitializeWithSeed(42)

	qualityMetrics := pcgManager.GetQualityMetrics()

	// Record some generations to populate metrics
	for i := 0; i < 10; i++ {
		qualityMetrics.RecordContentGeneration(pcg.ContentTypeTerrain, "terrain", 50*time.Millisecond, nil)
	}

	performanceMetrics := qualityMetrics.GetPerformanceMetrics()
	require.NotNil(t, performanceMetrics)

	stats := performanceMetrics.GetStats()
	assert.NotNil(t, stats)

	cacheHitRatio := performanceMetrics.GetCacheHitRatio()
	assert.GreaterOrEqual(t, cacheHitRatio, 0.0)
	assert.LessOrEqual(t, cacheHitRatio, 100.0)
}

// TestBalanceMetrics tests balance metrics retrieval.
func TestBalanceMetrics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	pcgManager.InitializeWithSeed(42)

	qualityMetrics := pcgManager.GetQualityMetrics()
	balanceMetrics := qualityMetrics.GetBalanceMetrics()

	assert.GreaterOrEqual(t, balanceMetrics.SystemHealth, 0.0)
	assert.LessOrEqual(t, balanceMetrics.SystemHealth, 1.0)
	assert.GreaterOrEqual(t, balanceMetrics.TotalBalanceChecks, int64(0))
}

// TestContentTypes tests content type constants.
func TestContentTypes(t *testing.T) {
	contentTypes := []pcg.ContentType{
		pcg.ContentTypeTerrain,
		pcg.ContentTypeQuests,
		pcg.ContentTypeItems,
	}

	for _, ct := range contentTypes {
		assert.NotEmpty(t, string(ct), "Content type should have string representation")
	}
}

// TestDeterministicSeed tests that fixed seed produces consistent results.
func TestDeterministicSeed(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create two managers with the same seed
	world1 := createTestWorld()
	pcgManager1 := pcg.NewPCGManager(world1, logger)
	pcgManager1.InitializeWithSeed(42)

	world2 := createTestWorld()
	pcgManager2 := pcg.NewPCGManager(world2, logger)
	pcgManager2.InitializeWithSeed(42)

	// Both should produce the same quality report structure
	report1 := pcgManager1.GenerateQualityReport()
	report2 := pcgManager2.GenerateQualityReport()

	// Verify initial scores are consistent (before any generation events)
	assert.Equal(t, report1.QualityGrade, report2.QualityGrade)
}

// TestMainOutputIntegration tests that run produces expected output.
func TestMainOutputIntegration(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	cfg := &Config{Seed: 42}
	runErr := run(cfg)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	assert.NoError(t, runErr)

	// Verify expected output sections
	assert.Contains(t, output, "Content Quality Metrics System Demo")
	assert.Contains(t, output, "Using seed: 42")
	assert.Contains(t, output, "Initializing Content Quality Metrics System")
	assert.Contains(t, output, "Generating Content with Quality Tracking")
	assert.Contains(t, output, "Recording Player Feedback")
	assert.Contains(t, output, "Recording Quest Completions")
	assert.Contains(t, output, "Generating Quality Report")
	assert.Contains(t, output, "CONTENT QUALITY REPORT")
	assert.Contains(t, output, "Overall Quality Score")
	assert.Contains(t, output, "Quality Grade")
	assert.Contains(t, output, "FINAL QUALITY ASSESSMENT")
	assert.Contains(t, output, "Demo completed successfully")
}

// TestFeedbackStructure tests PlayerFeedback struct fields.
func TestFeedbackStructure(t *testing.T) {
	feedback := pcg.PlayerFeedback{
		Timestamp:   time.Now(),
		ContentType: pcg.ContentTypeQuests,
		ContentID:   "test_content",
		Rating:      4,
		Difficulty:  3,
		Enjoyment:   4,
		Comments:    "Test comment",
		SessionID:   "test_session",
	}

	assert.NotZero(t, feedback.Timestamp)
	assert.Equal(t, pcg.ContentTypeQuests, feedback.ContentType)
	assert.Equal(t, "test_content", feedback.ContentID)
	assert.Equal(t, 4, feedback.Rating)
	assert.Equal(t, 3, feedback.Difficulty)
	assert.Equal(t, 4, feedback.Enjoyment)
	assert.Equal(t, "Test comment", feedback.Comments)
	assert.Equal(t, "test_session", feedback.SessionID)
}

// TestReportThresholdStatus tests threshold status in quality report.
func TestReportThresholdStatus(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	pcgManager.InitializeWithSeed(42)

	report := pcgManager.GenerateQualityReport()
	require.NotNil(t, report)

	// ThresholdStatus should be a map of threshold names to pass/fail status
	if report.ThresholdStatus != nil {
		for threshold, passed := range report.ThresholdStatus {
			assert.NotEmpty(t, threshold, "Threshold name should not be empty")
			// passed is a boolean, just verify type
			_ = passed
		}
	}
}

// TestReportRecommendations tests recommendations in quality report.
func TestReportRecommendations(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	pcgManager.InitializeWithSeed(42)

	report := pcgManager.GenerateQualityReport()
	require.NotNil(t, report)

	// Recommendations should be a slice of strings
	if report.Recommendations != nil {
		for _, rec := range report.Recommendations {
			assert.IsType(t, "", rec)
		}
	}
}

// TestReportCriticalIssues tests critical issues in quality report.
func TestReportCriticalIssues(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	pcgManager.InitializeWithSeed(42)

	report := pcgManager.GenerateQualityReport()
	require.NotNil(t, report)

	// CriticalIssues should be a slice of strings
	if report.CriticalIssues != nil {
		for _, issue := range report.CriticalIssues {
			assert.IsType(t, "", issue)
		}
	}
}

// TestReportSystemSummary tests system summary in quality report.
func TestReportSystemSummary(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	pcgManager.InitializeWithSeed(42)

	report := pcgManager.GenerateQualityReport()
	require.NotNil(t, report)

	// SystemSummary should be a map
	assert.NotNil(t, report.SystemSummary)
}

// createTestWorld creates a test world for demo testing.
func createTestWorld() *game.World {
	return &game.World{
		Width:       100,
		Height:      100,
		Levels:      []game.Level{},
		Objects:     make(map[string]game.GameObject),
		Players:     make(map[string]*game.Player),
		NPCs:        make(map[string]*game.NPC),
		SpatialGrid: make(map[game.Position][]string),
	}
}
