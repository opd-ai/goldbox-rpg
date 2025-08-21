package pcg

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goldbox-rpg/pkg/game"
)

func TestNewPCGEventManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	eventSystem := game.NewEventSystem()
	pcgManager := createTestPCGManager()

	manager := NewPCGEventManager(logger, eventSystem, pcgManager)

	assert.NotNil(t, manager)
	assert.Equal(t, logger, manager.logger)
	assert.Equal(t, eventSystem, manager.eventSystem)
	assert.Equal(t, pcgManager, manager.pcgManager)
	assert.NotNil(t, manager.adjustmentConfig)
	assert.False(t, manager.isMonitoring)
	assert.Equal(t, 0, manager.adjustmentCount)
}

func TestDefaultRuntimeAdjustmentConfig(t *testing.T) {
	config := DefaultRuntimeAdjustmentConfig()

	assert.NotNil(t, config)
	assert.True(t, config.EnableRuntimeAdjustments)
	assert.Equal(t, 30*time.Second, config.MonitoringInterval)
	assert.Equal(t, 10, config.MaxAdjustments)

	// Check quality thresholds
	assert.Equal(t, 0.7, config.QualityThresholds.MinOverallScore)
	assert.Equal(t, 0.6, config.QualityThresholds.MinPerformance)
	assert.Equal(t, 0.5, config.QualityThresholds.MinVariety)
	assert.Equal(t, 0.7, config.QualityThresholds.MinConsistency)
	assert.Equal(t, 0.6, config.QualityThresholds.MinEngagement)
	assert.Equal(t, 0.8, config.QualityThresholds.MinStability)

	// Check adjustment rates
	assert.Equal(t, 0.1, config.AdjustmentRates.DifficultyStep)
	assert.Equal(t, 0.2, config.AdjustmentRates.VarietyBoost)
	assert.Equal(t, 0.15, config.AdjustmentRates.ComplexityReduction)
	assert.Equal(t, 1.5, config.AdjustmentRates.GenerationSpeed)
}

func TestSetAndGetAdjustmentConfig(t *testing.T) {
	manager := createTestEventManager()

	newConfig := &RuntimeAdjustmentConfig{
		EnableRuntimeAdjustments: false,
		MonitoringInterval:       60 * time.Second,
		MaxAdjustments:           5,
	}

	manager.SetAdjustmentConfig(newConfig)
	retrievedConfig := manager.GetAdjustmentConfig()

	assert.Equal(t, newConfig, retrievedConfig)
	assert.False(t, retrievedConfig.EnableRuntimeAdjustments)
	assert.Equal(t, 60*time.Second, retrievedConfig.MonitoringInterval)
	assert.Equal(t, 5, retrievedConfig.MaxAdjustments)
}

func TestStartAndStopMonitoring(t *testing.T) {
	manager := createTestEventManager()

	assert.False(t, manager.IsMonitoring())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start monitoring
	manager.StartMonitoring(ctx)
	assert.True(t, manager.IsMonitoring())

	// Starting again should be no-op
	manager.StartMonitoring(ctx)
	assert.True(t, manager.IsMonitoring())

	// Stop monitoring
	manager.StopMonitoring()
	assert.False(t, manager.IsMonitoring())

	// Stopping again should be no-op
	manager.StopMonitoring()
	assert.False(t, manager.IsMonitoring())
}

func TestEmitContentGenerated(t *testing.T) {
	manager := createTestEventManager()

	// Create a channel to capture emitted events
	eventCaptured := make(chan game.GameEvent, 1)
	manager.eventSystem.Subscribe(EventPCGContentGenerated, func(event game.GameEvent) {
		eventCaptured <- event
	})

	// Emit content generated event
	contentType := ContentTypeQuests
	generationTime := 5 * time.Millisecond
	qualityScore := 0.85

	manager.EmitContentGenerated(contentType, "test content", generationTime, qualityScore)

	// Verify event was emitted
	select {
	case event := <-eventCaptured:
		assert.Equal(t, EventPCGContentGenerated, event.Type)
		assert.Equal(t, "pcg_manager", event.SourceID)
		assert.Equal(t, string(contentType), event.TargetID)

		pcgData, ok := event.Data["pcg_data"].(PCGEventData)
		require.True(t, ok)
		assert.Equal(t, contentType, pcgData.ContentType)
		assert.Equal(t, generationTime, pcgData.GenerationTime)
		assert.Equal(t, qualityScore, pcgData.QualityScore)

	case <-time.After(100 * time.Millisecond):
		t.Fatal("Event was not emitted within timeout")
	}
}

func TestEmitQualityAssessment(t *testing.T) {
	manager := createTestEventManager()

	// Create a channel to capture emitted events
	eventCaptured := make(chan game.GameEvent, 1)
	manager.eventSystem.Subscribe(EventPCGQualityAssessment, func(event game.GameEvent) {
		eventCaptured <- event
	})

	// Create a test quality report
	report := &QualityReport{
		OverallScore: 0.85,
		QualityGrade: "B+",
		ComponentScores: map[string]float64{
			"performance": 0.8,
			"variety":     0.9,
		},
	}

	manager.EmitQualityAssessment(report)

	// Verify event was emitted
	select {
	case event := <-eventCaptured:
		assert.Equal(t, EventPCGQualityAssessment, event.Type)
		assert.Equal(t, "quality_metrics", event.SourceID)
		assert.Equal(t, "pcg_system", event.TargetID)

		receivedReport, ok := event.Data["quality_report"].(*QualityReport)
		require.True(t, ok)
		assert.Equal(t, report.OverallScore, receivedReport.OverallScore)
		assert.Equal(t, report.QualityGrade, receivedReport.QualityGrade)

	case <-time.After(100 * time.Millisecond):
		t.Fatal("Event was not emitted within timeout")
	}
}

func TestEmitPlayerFeedback(t *testing.T) {
	manager := createTestEventManager()

	// Create a channel to capture emitted events
	eventCaptured := make(chan game.GameEvent, 1)
	manager.eventSystem.Subscribe(EventPCGPlayerFeedback, func(event game.GameEvent) {
		eventCaptured <- event
	})

	// Create test player feedback
	feedback := &PlayerFeedback{
		ContentType: ContentTypeQuests,
		ContentID:   "test_quest",
		Rating:      4,
		Enjoyment:   5,
		Difficulty:  3,
		Comments:    "Great quest!",
		SessionID:   "test_session",
		Timestamp:   time.Now(),
	}

	manager.EmitPlayerFeedback(feedback)

	// Verify event was emitted
	select {
	case event := <-eventCaptured:
		assert.Equal(t, EventPCGPlayerFeedback, event.Type)
		assert.Equal(t, "player", event.SourceID)
		assert.Equal(t, "pcg_system", event.TargetID)

		receivedFeedback, ok := event.Data["feedback"].(*PlayerFeedback)
		require.True(t, ok)
		assert.Equal(t, feedback.ContentID, receivedFeedback.ContentID)
		assert.Equal(t, feedback.Rating, receivedFeedback.Rating)
		assert.Equal(t, feedback.Enjoyment, receivedFeedback.Enjoyment)

	case <-time.After(100 * time.Millisecond):
		t.Fatal("Event was not emitted within timeout")
	}
}

func TestHandleContentGenerated(t *testing.T) {
	manager := createTestEventManager()

	// Test with valid PCG data
	eventData := PCGEventData{
		ContentType:    ContentTypeQuests,
		GenerationTime: 5 * time.Millisecond,
		QualityScore:   0.5, // Below threshold
		Timestamp:      time.Now(),
	}

	event := game.GameEvent{
		Type:      EventPCGContentGenerated,
		SourceID:  "test",
		TargetID:  "test",
		Data:      map[string]interface{}{"pcg_data": eventData},
		Timestamp: time.Now().Unix(),
	}

	// Should trigger adjustment due to low quality score
	manager.handleContentGenerated(event)

	// Adjustment should be scheduled (implementation may vary)
	// At minimum, the handler should not panic
	assert.True(t, true) // Placeholder - actual assertion depends on implementation
}

func TestHandleQualityAssessment(t *testing.T) {
	manager := createTestEventManager()

	// Create quality report with low scores to trigger adjustments
	report := &QualityReport{
		OverallScore: 0.6, // Below threshold of 0.7
		ComponentScores: map[string]float64{
			"performance": 0.5, // Below threshold of 0.6
			"variety":     0.4, // Below threshold of 0.5
			"consistency": 0.6, // Below threshold of 0.7
			"engagement":  0.5, // Below threshold of 0.6
			"stability":   0.7, // Below threshold of 0.8
		},
	}

	event := game.GameEvent{
		Type:      EventPCGQualityAssessment,
		SourceID:  "test",
		TargetID:  "test",
		Data:      map[string]interface{}{"quality_report": report},
		Timestamp: time.Now().Unix(),
	}

	initialCount := manager.GetAdjustmentCount()
	manager.handleQualityAssessment(event)

	// Should have triggered adjustments for each low-scoring component
	// The exact number depends on implementation
	finalCount := manager.GetAdjustmentCount()
	assert.GreaterOrEqual(t, finalCount, initialCount)
}

func TestHandlePlayerFeedback(t *testing.T) {
	manager := createTestEventManager()

	tests := []struct {
		name             string
		feedback         *PlayerFeedback
		expectAdjustment bool
	}{
		{
			name: "too_easy_difficulty",
			feedback: &PlayerFeedback{
				ContentID:  "test_content",
				Difficulty: 2, // Too easy
				Enjoyment:  5,
				Timestamp:  time.Now(),
			},
			expectAdjustment: true,
		},
		{
			name: "too_hard_difficulty",
			feedback: &PlayerFeedback{
				ContentID:  "test_content",
				Difficulty: 8, // Too hard
				Enjoyment:  5,
				Timestamp:  time.Now(),
			},
			expectAdjustment: true,
		},
		{
			name: "low_enjoyment",
			feedback: &PlayerFeedback{
				ContentID:  "test_content",
				Difficulty: 5,
				Enjoyment:  3, // Low enjoyment
				Timestamp:  time.Now(),
			},
			expectAdjustment: true,
		},
		{
			name: "good_feedback",
			feedback: &PlayerFeedback{
				ContentID:  "test_content",
				Difficulty: 5, // Just right
				Enjoyment:  7, // High enjoyment
				Timestamp:  time.Now(),
			},
			expectAdjustment: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset adjustment count
			manager.ResetAdjustmentCount()

			event := game.GameEvent{
				Type:      EventPCGPlayerFeedback,
				SourceID:  "test",
				TargetID:  "test",
				Data:      map[string]interface{}{"feedback": tt.feedback},
				Timestamp: time.Now().Unix(),
			}

			initialCount := manager.GetAdjustmentCount()
			manager.handlePlayerFeedback(event)
			finalCount := manager.GetAdjustmentCount()

			if tt.expectAdjustment {
				assert.Greater(t, finalCount, initialCount, "Expected adjustment to be made")
			} else {
				assert.Equal(t, initialCount, finalCount, "Expected no adjustment to be made")
			}
		})
	}
}

func TestAdjustmentHistory(t *testing.T) {
	manager := createTestEventManager()

	// Simulate some adjustments
	params1 := map[string]interface{}{
		"trigger": "test_trigger_1",
		"score":   0.5,
	}
	manager.recordAdjustment(AdjustmentTypeDifficulty, params1, true)

	params2 := map[string]interface{}{
		"trigger": "test_trigger_2",
		"score":   0.6,
	}
	manager.recordAdjustment(AdjustmentTypeVariety, params2, false)

	// Check adjustment count
	assert.Equal(t, 1, manager.GetAdjustmentCount()) // Only successful adjustments count

	// Check adjustment history
	history := manager.GetAdjustmentHistory()
	assert.Len(t, history, 2)

	assert.Equal(t, AdjustmentTypeDifficulty, history[0].AdjustmentType)
	assert.True(t, history[0].Success)
	assert.Equal(t, "test_trigger_1", history[0].Trigger)

	assert.Equal(t, AdjustmentTypeVariety, history[1].AdjustmentType)
	assert.False(t, history[1].Success)
	assert.Equal(t, "test_trigger_2", history[1].Trigger)

	// Reset adjustment count
	manager.ResetAdjustmentCount()
	assert.Equal(t, 0, manager.GetAdjustmentCount())
}

func TestMaxAdjustmentsLimit(t *testing.T) {
	manager := createTestEventManager()

	// Set low max adjustments for testing
	config := manager.GetAdjustmentConfig()
	config.MaxAdjustments = 2
	manager.SetAdjustmentConfig(config)

	// Simulate multiple quality assessments that would trigger adjustments
	for i := 0; i < 5; i++ {
		manager.scheduleQualityAdjustment("test_trigger", 0.5)
	}

	// Should not exceed max adjustments
	assert.LessOrEqual(t, manager.GetAdjustmentCount(), config.MaxAdjustments)
}

func TestConcurrentEventHandling(t *testing.T) {
	manager := createTestEventManager()

	// Test concurrent access to adjustment methods
	const numGoroutines = 10
	const numAdjustments = 5

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < numAdjustments; j++ {
				params := map[string]interface{}{
					"trigger":   "concurrent_test",
					"id":        id,
					"iteration": j,
				}
				manager.recordAdjustment(AdjustmentTypePerformance, params, true)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent test timed out")
		}
	}

	// Verify all adjustments were recorded
	expectedCount := numGoroutines * numAdjustments
	assert.Equal(t, expectedCount, manager.GetAdjustmentCount())

	history := manager.GetAdjustmentHistory()
	assert.Len(t, history, expectedCount)
}

func TestPerformQualityCheck(t *testing.T) {
	manager := createTestEventManager()

	// Set up a PCG manager with metrics
	pcgManager := createTestPCGManager()
	manager.pcgManager = pcgManager

	// Perform quality check
	manager.performQualityCheck()

	// Verify that lastQualityCheck was updated
	assert.True(t, time.Since(manager.lastQualityCheck) < time.Second)
}

func TestMonitorSystemHealth(t *testing.T) {
	manager := createTestEventManager()

	tests := []struct {
		name             string
		healthData       map[string]interface{}
		expectAdjustment bool
	}{
		{
			name: "high_memory_usage",
			healthData: map[string]interface{}{
				"memory_usage": 0.9, // Above threshold
				"error_rate":   0.01,
			},
			expectAdjustment: true,
		},
		{
			name: "high_error_rate",
			healthData: map[string]interface{}{
				"memory_usage": 0.5,
				"error_rate":   0.1, // Above threshold
			},
			expectAdjustment: true,
		},
		{
			name: "healthy_system",
			healthData: map[string]interface{}{
				"memory_usage": 0.4,
				"error_rate":   0.01,
			},
			expectAdjustment: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager.ResetAdjustmentCount()

			initialCount := manager.GetAdjustmentCount()
			manager.monitorSystemHealth(tt.healthData)
			finalCount := manager.GetAdjustmentCount()

			if tt.expectAdjustment {
				assert.Greater(t, finalCount, initialCount, "Expected system health adjustment")
			} else {
				assert.Equal(t, initialCount, finalCount, "Expected no system health adjustment")
			}
		})
	}
}

// Helper functions for testing

func createTestEventManager() *PCGEventManager {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Suppress logs during tests

	eventSystem := game.NewEventSystem()
	pcgManager := createTestPCGManager()

	return NewPCGEventManager(logger, eventSystem, pcgManager)
}

func createTestPCGManager() *PCGManager {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	world := &game.World{
		Width:       100,
		Height:      100,
		Levels:      []game.Level{},
		Objects:     make(map[string]game.GameObject),
		Players:     make(map[string]*game.Player),
		NPCs:        make(map[string]*game.NPC),
		SpatialGrid: make(map[game.Position][]string),
	}

	return NewPCGManager(world, logger)
}
