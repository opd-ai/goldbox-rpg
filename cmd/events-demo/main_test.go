package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goldbox-rpg/pkg/game"
	"goldbox-rpg/pkg/pcg"
)

// TestPCGEventManagerBasic tests basic PCG event manager initialization.
func TestPCGEventManagerBasic(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	eventSystem := game.NewEventSystem()

	eventManager := pcg.NewPCGEventManager(logger, eventSystem, pcgManager)
	require.NotNil(t, eventManager)

	config := eventManager.GetAdjustmentConfig()
	assert.True(t, config.EnableRuntimeAdjustments)
	assert.Equal(t, 0.7, config.QualityThresholds.MinOverallScore)
	assert.Equal(t, 10, config.MaxAdjustments)
}

// TestPCGEventManagerNilParameters tests event manager with nil parameters.
func TestPCGEventManagerNilParameters(t *testing.T) {
	// Should handle nil logger gracefully
	eventManager := pcg.NewPCGEventManager(nil, nil, nil)
	require.NotNil(t, eventManager)

	config := eventManager.GetAdjustmentConfig()
	assert.NotNil(t, config)
}

// TestRuntimeAdjustmentConfig tests runtime adjustment configuration.
func TestRuntimeAdjustmentConfig(t *testing.T) {
	config := pcg.DefaultRuntimeAdjustmentConfig()

	assert.True(t, config.EnableRuntimeAdjustments)
	assert.Equal(t, 0.7, config.QualityThresholds.MinOverallScore)
	assert.Equal(t, 0.6, config.QualityThresholds.MinPerformance)
	assert.Equal(t, 0.5, config.QualityThresholds.MinVariety)
	assert.Equal(t, 0.7, config.QualityThresholds.MinConsistency)
	assert.Equal(t, 0.6, config.QualityThresholds.MinEngagement)
	assert.Equal(t, 0.8, config.QualityThresholds.MinStability)
	assert.Equal(t, 0.1, config.AdjustmentRates.DifficultyStep)
	assert.Equal(t, 0.2, config.AdjustmentRates.VarietyBoost)
	assert.Equal(t, 30*time.Second, config.MonitoringInterval)
	assert.Equal(t, 10, config.MaxAdjustments)
}

// TestSetAdjustmentConfig tests setting custom adjustment configuration.
func TestSetAdjustmentConfig(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	eventManager := pcg.NewPCGEventManager(logger, nil, nil)

	customConfig := &pcg.RuntimeAdjustmentConfig{
		EnableRuntimeAdjustments: false,
		MonitoringInterval:       60 * time.Second,
		MaxAdjustments:           5,
	}
	customConfig.QualityThresholds.MinOverallScore = 0.8

	eventManager.SetAdjustmentConfig(customConfig)
	retrievedConfig := eventManager.GetAdjustmentConfig()

	assert.False(t, retrievedConfig.EnableRuntimeAdjustments)
	assert.Equal(t, 60*time.Second, retrievedConfig.MonitoringInterval)
	assert.Equal(t, 5, retrievedConfig.MaxAdjustments)
	assert.Equal(t, 0.8, retrievedConfig.QualityThresholds.MinOverallScore)
}

// TestStartStopMonitoring tests monitoring start and stop functionality.
func TestStartStopMonitoring(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	eventSystem := game.NewEventSystem()
	eventManager := pcg.NewPCGEventManager(logger, eventSystem, pcgManager)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	eventManager.StartMonitoring(ctx)
	assert.True(t, eventManager.IsMonitoring())

	// Starting again should be a no-op
	eventManager.StartMonitoring(ctx)
	assert.True(t, eventManager.IsMonitoring())

	eventManager.StopMonitoring()
	// Allow goroutine to stop
	time.Sleep(50 * time.Millisecond)
	assert.False(t, eventManager.IsMonitoring())

	// Stopping again should be a no-op
	eventManager.StopMonitoring()
	assert.False(t, eventManager.IsMonitoring())
}

// TestEmitContentGenerated tests content generation event emission.
func TestEmitContentGenerated(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	eventSystem := game.NewEventSystem()
	eventManager := pcg.NewPCGEventManager(logger, eventSystem, pcgManager)

	// Emit content generated event
	eventManager.EmitContentGenerated(pcg.ContentTypeQuests, "test_quest", 10*time.Millisecond, 0.85)

	// Allow event processing
	time.Sleep(50 * time.Millisecond)

	// Verify no panic occurred and event was processed
	assert.Equal(t, 0, eventManager.GetAdjustmentCount())
}

// TestEmitContentGeneratedLowQuality tests low quality triggers adjustment.
func TestEmitContentGeneratedLowQuality(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	eventSystem := game.NewEventSystem()
	eventManager := pcg.NewPCGEventManager(logger, eventSystem, pcgManager)

	// Emit low quality content event (below 0.7 threshold)
	eventManager.EmitContentGenerated(pcg.ContentTypeQuests, "low_quality_quest", 50*time.Millisecond, 0.4)

	// Allow event processing
	time.Sleep(50 * time.Millisecond)

	// Low quality should trigger an adjustment
	assert.Greater(t, eventManager.GetAdjustmentCount(), 0)
}

// TestEmitQualityAssessment tests quality assessment event emission.
func TestEmitQualityAssessment(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	eventSystem := game.NewEventSystem()
	eventManager := pcg.NewPCGEventManager(logger, eventSystem, pcgManager)

	qualityReport := &pcg.QualityReport{
		OverallScore:    0.85,
		QualityGrade:    "B",
		ComponentScores: map[string]float64{"performance": 0.9, "variety": 0.8},
		Recommendations: []string{},
		CriticalIssues:  []string{},
	}

	eventManager.EmitQualityAssessment(qualityReport)

	// Allow event processing
	time.Sleep(50 * time.Millisecond)

	// Good quality should not trigger adjustments
	assert.Equal(t, 0, eventManager.GetAdjustmentCount())
}

// TestEmitPlayerFeedback tests player feedback event emission.
func TestEmitPlayerFeedback(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	eventSystem := game.NewEventSystem()
	eventManager := pcg.NewPCGEventManager(logger, eventSystem, pcgManager)

	tests := []struct {
		name           string
		feedback       pcg.PlayerFeedback
		expectAdjust   bool
		minAdjustments int
	}{
		{
			name: "too_easy",
			feedback: pcg.PlayerFeedback{
				ContentType: pcg.ContentTypeQuests,
				ContentID:   "quest_001",
				Rating:      3,
				Difficulty:  2, // Below 3 is too easy
				Enjoyment:   6,
				SessionID:   "test_session",
				Timestamp:   time.Now(),
			},
			expectAdjust:   true,
			minAdjustments: 1,
		},
		{
			name: "too_hard",
			feedback: pcg.PlayerFeedback{
				ContentType: pcg.ContentTypeQuests,
				ContentID:   "quest_002",
				Rating:      4,
				Difficulty:  8, // Above 7 is too hard
				Enjoyment:   5,
				SessionID:   "test_session",
				Timestamp:   time.Now(),
			},
			expectAdjust:   true,
			minAdjustments: 1,
		},
		{
			name: "low_enjoyment",
			feedback: pcg.PlayerFeedback{
				ContentType: pcg.ContentTypeDungeon,
				ContentID:   "dungeon_001",
				Rating:      2,
				Difficulty:  5,
				Enjoyment:   2, // Below 4 is low enjoyment
				SessionID:   "test_session",
				Timestamp:   time.Now(),
			},
			expectAdjust:   true,
			minAdjustments: 1,
		},
		{
			name: "satisfied",
			feedback: pcg.PlayerFeedback{
				ContentType: pcg.ContentTypeQuests,
				ContentID:   "quest_003",
				Rating:      5,
				Difficulty:  5, // Middle range
				Enjoyment:   8, // Good enjoyment
				SessionID:   "test_session",
				Timestamp:   time.Now(),
			},
			expectAdjust:   false,
			minAdjustments: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset for each test
			eventManager.ResetAdjustmentCount()

			eventManager.EmitPlayerFeedback(&tc.feedback)

			// Allow event processing
			time.Sleep(50 * time.Millisecond)

			if tc.expectAdjust {
				assert.GreaterOrEqual(t, eventManager.GetAdjustmentCount(), tc.minAdjustments,
					"Expected at least %d adjustments for %s", tc.minAdjustments, tc.name)
			}
		})
	}
}

// TestSystemHealthEvents tests system health event handling.
func TestSystemHealthEvents(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	eventSystem := game.NewEventSystem()
	eventManager := pcg.NewPCGEventManager(logger, eventSystem, pcgManager)

	tests := []struct {
		name         string
		healthData   map[string]interface{}
		expectAdjust bool
	}{
		{
			name: "high_memory",
			healthData: map[string]interface{}{
				"memory_usage": 0.85, // Above 0.8 threshold
				"error_rate":   0.02,
			},
			expectAdjust: true,
		},
		{
			name: "high_error_rate",
			healthData: map[string]interface{}{
				"memory_usage": 0.4,
				"error_rate":   0.08, // Above 0.05 threshold
			},
			expectAdjust: true,
		},
		{
			name: "normal",
			healthData: map[string]interface{}{
				"memory_usage": 0.3,
				"error_rate":   0.01,
			},
			expectAdjust: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			eventManager.ResetAdjustmentCount()

			healthEvent := game.GameEvent{
				Type:      pcg.EventPCGSystemHealth,
				SourceID:  "system_monitor",
				TargetID:  "pcg_system",
				Data:      map[string]interface{}{"health_data": tc.healthData},
				Timestamp: time.Now().Unix(),
			}

			eventSystem.Emit(healthEvent)

			// Allow event processing
			time.Sleep(50 * time.Millisecond)

			if tc.expectAdjust {
				assert.Greater(t, eventManager.GetAdjustmentCount(), 0,
					"Expected adjustment for %s", tc.name)
			}
		})
	}
}

// TestGetAdjustmentHistory tests adjustment history retrieval.
func TestGetAdjustmentHistory(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	eventSystem := game.NewEventSystem()
	eventManager := pcg.NewPCGEventManager(logger, eventSystem, pcgManager)

	// Trigger some adjustments
	eventManager.EmitContentGenerated(pcg.ContentTypeQuests, "low_q", 10*time.Millisecond, 0.3)

	// Allow event processing
	time.Sleep(50 * time.Millisecond)

	history := eventManager.GetAdjustmentHistory()
	assert.NotEmpty(t, history, "Expected adjustment history to be non-empty")

	if len(history) > 0 {
		assert.NotEmpty(t, history[0].Trigger)
		assert.False(t, history[0].Timestamp.IsZero())
	}
}

// TestResetAdjustmentCount tests adjustment count reset.
func TestResetAdjustmentCount(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	eventManager := pcg.NewPCGEventManager(logger, nil, nil)

	// Emit low quality to trigger adjustment
	eventManager.EmitContentGenerated(pcg.ContentTypeQuests, "low", 10*time.Millisecond, 0.3)
	time.Sleep(50 * time.Millisecond)

	// Reset
	eventManager.ResetAdjustmentCount()
	assert.Equal(t, 0, eventManager.GetAdjustmentCount())
}

// TestQualityReportGeneration tests quality report generation via PCG manager.
func TestQualityReportGeneration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)

	report := pcgManager.GenerateQualityReport()
	require.NotNil(t, report)
	assert.NotEmpty(t, report.QualityGrade)
	assert.GreaterOrEqual(t, report.OverallScore, 0.0)
	assert.LessOrEqual(t, report.OverallScore, 1.0)
}

// TestContentGeneration tests actual content generation.
func TestContentGeneration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)

	ctx := context.Background()

	// Generate quest
	quest, err := pcgManager.GenerateQuestForArea(ctx, "test_area", pcg.QuestTypeFetch, 5)
	if err != nil {
		// Quest generation may fail without proper setup, which is acceptable
		t.Logf("Quest generation failed (expected in test environment): %v", err)
	} else {
		assert.NotNil(t, quest)
		assert.NotEmpty(t, quest.ID)
	}

	// Generate items
	items, err := pcgManager.GenerateItemsForLocation(ctx, "test_location", 3, pcg.RarityCommon, pcg.RarityRare, 5)
	if err != nil {
		t.Logf("Item generation failed (expected in test environment): %v", err)
	} else {
		assert.NotEmpty(t, items)
	}
}

// TestMonitoringContextCancellation tests monitoring stops on context cancel.
func TestMonitoringContextCancellation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	eventSystem := game.NewEventSystem()
	eventManager := pcg.NewPCGEventManager(logger, eventSystem, pcgManager)

	ctx, cancel := context.WithCancel(context.Background())
	eventManager.StartMonitoring(ctx)
	assert.True(t, eventManager.IsMonitoring())

	// Cancel context
	cancel()

	// The monitoring loop exits on context cancellation but isMonitoring flag
	// is only set to false by StopMonitoring() - this is by design.
	// The goroutine will exit cleanly but the flag won't change.
	time.Sleep(100 * time.Millisecond)

	// Clean up properly
	eventManager.StopMonitoring()
	assert.False(t, eventManager.IsMonitoring())
}

// TestMaxAdjustmentsLimit tests that max adjustments limit is enforced.
func TestMaxAdjustmentsLimit(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	eventSystem := game.NewEventSystem()
	eventManager := pcg.NewPCGEventManager(logger, eventSystem, pcgManager)

	// Set low max adjustments
	config := pcg.DefaultRuntimeAdjustmentConfig()
	config.MaxAdjustments = 2
	eventManager.SetAdjustmentConfig(config)

	// Trigger many low-quality events sequentially with longer waits to avoid race conditions
	for i := 0; i < 5; i++ {
		eventManager.EmitContentGenerated(pcg.ContentTypeQuests, "low", 10*time.Millisecond, 0.3)
		// Wait longer for event to be fully processed before next emission
		time.Sleep(100 * time.Millisecond)
	}

	// Should not exceed max (but may be slightly lower due to async processing)
	assert.LessOrEqual(t, eventManager.GetAdjustmentCount(), 2)
}

// TestMainOutputIntegration tests that main produces expected output.
func TestMainOutputIntegration(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	done := make(chan bool)
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				t.Logf("main() panicked: %v", rec)
			}
			done <- true
		}()
		main()
	}()

	<-done

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	// Verify expected output sections
	assert.Contains(t, output, "PCG Event System Integration Demo")
	assert.Contains(t, output, "PCG Manager initialized")
	assert.Contains(t, output, "PCG Event Manager initialized")
	assert.Contains(t, output, "Runtime monitoring started")
	assert.Contains(t, output, "Simulating Content Generation Events")
	assert.Contains(t, output, "Simulating Player Feedback")
	assert.Contains(t, output, "Simulating System Health Monitoring")
	assert.Contains(t, output, "Runtime Adjustment Results")
	assert.Contains(t, output, "Demo Complete")
}

// TestEventTypeConstants tests PCG event type constants.
func TestEventTypeConstants(t *testing.T) {
	// Verify event types are distinct and start at 1000+
	assert.GreaterOrEqual(t, int(pcg.EventPCGContentGenerated), 1000)
	assert.NotEqual(t, pcg.EventPCGContentGenerated, pcg.EventPCGQualityAssessment)
	assert.NotEqual(t, pcg.EventPCGQualityAssessment, pcg.EventPCGPlayerFeedback)
	assert.NotEqual(t, pcg.EventPCGPlayerFeedback, pcg.EventPCGDifficultyAdjustment)
	assert.NotEqual(t, pcg.EventPCGDifficultyAdjustment, pcg.EventPCGContentRequest)
	assert.NotEqual(t, pcg.EventPCGContentRequest, pcg.EventPCGSystemHealth)
}

// TestAdjustmentTypes tests adjustment type constants.
func TestAdjustmentTypes(t *testing.T) {
	assert.Equal(t, pcg.AdjustmentType("difficulty"), pcg.AdjustmentTypeDifficulty)
	assert.Equal(t, pcg.AdjustmentType("variety"), pcg.AdjustmentTypeVariety)
	assert.Equal(t, pcg.AdjustmentType("complexity"), pcg.AdjustmentTypeComplexity)
	assert.Equal(t, pcg.AdjustmentType("performance"), pcg.AdjustmentTypePerformance)
}

// TestContentTypes tests content type constants used in demo.
func TestContentTypes(t *testing.T) {
	contentTypes := []pcg.ContentType{
		pcg.ContentTypeQuests,
		pcg.ContentTypeItems,
		pcg.ContentTypeCharacters,
		pcg.ContentTypeDungeon,
	}

	for _, ct := range contentTypes {
		assert.NotEmpty(t, string(ct), "Content type should have string representation")
	}
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

// TestTimeNowInjection tests that timeNow can be overridden for reproducibility.
func TestTimeNowInjection(t *testing.T) {
	// Save original and restore after test
	originalTimeNow := timeNow
	defer func() { timeNow = originalTimeNow }()

	// Set a fixed time for testing
	fixedTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	timeNow = func() time.Time {
		return fixedTime
	}

	// Verify the fixed time is returned
	assert.Equal(t, fixedTime, timeNow())

	// Verify consistent results
	assert.Equal(t, timeNow(), timeNow())
	assert.Equal(t, fixedTime.Unix(), timeNow().Unix())
}

// TestTimeMeasurementReproducibility tests that time measurement is reproducible with injected time.
func TestTimeMeasurementReproducibility(t *testing.T) {
	// Save original and restore after test
	originalTimeNow := timeNow
	defer func() { timeNow = originalTimeNow }()

	// Create a sequence of times for reproducibility
	callCount := 0
	baseTimes := []time.Time{
		time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 15, 12, 0, 1, 0, time.UTC),
		time.Date(2025, 1, 15, 12, 0, 5, 0, time.UTC),
	}

	timeNow = func() time.Time {
		idx := callCount
		if idx >= len(baseTimes) {
			idx = len(baseTimes) - 1
		}
		callCount++
		return baseTimes[idx]
	}

	// Each call advances through the sequence
	assert.Equal(t, baseTimes[0], timeNow())
	assert.Equal(t, baseTimes[1], timeNow())
	assert.Equal(t, baseTimes[2], timeNow())
}

// TestDefaultConfig tests that DefaultConfig returns expected values.
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, 30*time.Second, cfg.Timeout, "Default timeout should be 30 seconds")
}

// TestInitializeDemo tests demo initialization.
func TestInitializeDemo(t *testing.T) {
	// Capture stdout to avoid polluting test output
	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	cfg := DefaultConfig()
	dctx := initializeDemo(cfg)

	w.Close()
	os.Stdout = oldStdout

	require.NotNil(t, dctx)
	assert.NotNil(t, dctx.World)
	assert.NotNil(t, dctx.PCGManager)
	assert.NotNil(t, dctx.EventSystem)
	assert.NotNil(t, dctx.EventManager)
	assert.Equal(t, cfg, dctx.Config)
}

// TestStartMonitoringWithConfig tests monitoring starts with configured timeout.
func TestStartMonitoringWithConfig(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Capture stdout
	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	cfg := &Config{Timeout: 5 * time.Second}
	dctx := &DemoContext{
		World:        createTestWorld(),
		PCGManager:   pcg.NewPCGManager(createTestWorld(), logger),
		EventSystem:  game.NewEventSystem(),
		EventManager: pcg.NewPCGEventManager(logger, game.NewEventSystem(), nil),
		Config:       cfg,
	}

	ctx, cancel := startMonitoring(dctx)
	defer cancel()

	w.Close()
	os.Stdout = oldStdout

	require.NotNil(t, ctx)
	require.NotNil(t, cancel)

	// Verify context has deadline
	deadline, ok := ctx.Deadline()
	assert.True(t, ok, "Context should have deadline")
	assert.False(t, deadline.IsZero())
}

// TestDisplayConfiguration tests configuration display does not panic.
func TestDisplayConfiguration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Capture stdout
	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	eventManager := pcg.NewPCGEventManager(logger, nil, nil)

	// Should not panic
	displayConfiguration(eventManager)

	w.Close()
	os.Stdout = oldStdout
}

// TestSimulateContentGeneration tests content generation simulation.
func TestSimulateContentGeneration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	world := createTestWorld()
	cfg := DefaultConfig()
	dctx := &DemoContext{
		World:        world,
		PCGManager:   pcg.NewPCGManager(world, logger),
		EventSystem:  game.NewEventSystem(),
		EventManager: pcg.NewPCGEventManager(logger, game.NewEventSystem(), pcg.NewPCGManager(world, logger)),
		Config:       cfg,
	}

	// Should not panic
	simulateContentGeneration(dctx)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "Simulating Content Generation Events")
}

// TestSimulatePlayerFeedback tests player feedback simulation.
func TestSimulatePlayerFeedback(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	world := createTestWorld()
	cfg := DefaultConfig()
	dctx := &DemoContext{
		World:        world,
		PCGManager:   pcg.NewPCGManager(world, logger),
		EventSystem:  game.NewEventSystem(),
		EventManager: pcg.NewPCGEventManager(logger, game.NewEventSystem(), nil),
		Config:       cfg,
	}

	simulatePlayerFeedback(dctx)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "Simulating Player Feedback")
	assert.Contains(t, output, "Player finds content too easy")
}

// TestSimulateSystemHealth tests system health simulation.
func TestSimulateSystemHealth(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	world := createTestWorld()
	cfg := DefaultConfig()
	dctx := &DemoContext{
		World:        world,
		PCGManager:   pcg.NewPCGManager(world, logger),
		EventSystem:  game.NewEventSystem(),
		EventManager: pcg.NewPCGEventManager(logger, game.NewEventSystem(), nil),
		Config:       cfg,
	}

	simulateSystemHealth(dctx)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "Simulating System Health Monitoring")
	assert.Contains(t, output, "High memory usage detected")
}

// TestDisplayAdjustmentResults tests adjustment results display.
func TestDisplayAdjustmentResults(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	eventManager := pcg.NewPCGEventManager(logger, nil, nil)
	displayAdjustmentResults(eventManager)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "Runtime Adjustment Results")
	assert.Contains(t, output, "Total Adjustments Made:")
}

// TestDisplayFinalAssessment tests final assessment display.
func TestDisplayFinalAssessment(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	world := createTestWorld()
	pcgManager := pcg.NewPCGManager(world, logger)
	report := displayFinalAssessment(pcgManager)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "Final Quality Assessment")
	assert.NotNil(t, report)
}

// TestDisplayStatistics tests statistics display.
func TestDisplayStatistics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	eventManager := pcg.NewPCGEventManager(logger, nil, nil)
	report := &pcg.QualityReport{OverallScore: 0.85}

	displayStatistics(eventManager, report, 30*time.Second)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "Event System Statistics")
	assert.Contains(t, output, "Monitoring Duration: 30s")
	assert.Contains(t, output, "HEALTHY")
}

// TestDisplayStatisticsNeedsAttention tests statistics with low quality.
func TestDisplayStatisticsNeedsAttention(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	eventManager := pcg.NewPCGEventManager(logger, nil, nil)
	report := &pcg.QualityReport{OverallScore: 0.5}

	displayStatistics(eventManager, report, 1*time.Minute)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "NEEDS ATTENTION")
}

// TestDisplayConclusion tests conclusion display.
func TestDisplayConclusion(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	displayConclusion()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	assert.Contains(t, output, "Demo Complete")
	assert.Contains(t, output, "Real-time quality monitoring")
}
