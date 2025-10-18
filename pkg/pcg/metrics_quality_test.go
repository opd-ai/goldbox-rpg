package pcg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Tests for ContentQualityMetrics

func TestNewContentQualityMetrics(t *testing.T) {
	cqm := NewContentQualityMetrics()

	assert.NotNil(t, cqm)
	assert.NotNil(t, cqm.performanceMetrics)
	assert.NotNil(t, cqm.validationMetrics)
	assert.NotNil(t, cqm.balanceMetrics)
	assert.NotNil(t, cqm.varietyMetrics)
	assert.NotNil(t, cqm.consistencyMetrics)
	assert.NotNil(t, cqm.engagementMetrics)
	assert.NotNil(t, cqm.stabilityMetrics)
	assert.NotNil(t, cqm.qualityThresholds)
	assert.Equal(t, 0.0, cqm.overallQualityScore)
}

func TestRecordContentGeneration(t *testing.T) {
	cqm := NewContentQualityMetrics()

	// Test successful generation
	testContent := "test content"
	duration := 100 * time.Millisecond

	cqm.RecordContentGeneration(ContentTypeTerrain, testContent, duration, nil)

	// Verify performance metrics were updated
	assert.Equal(t, int64(1), cqm.performanceMetrics.TotalGenerations)
	assert.Equal(t, int64(1), cqm.performanceMetrics.GetGenerationCount(ContentTypeTerrain))

	// Test error generation
	testError := assert.AnError
	cqm.RecordContentGeneration(ContentTypeItems, testContent, duration, testError)

	// Verify error was recorded
	assert.Equal(t, int64(1), cqm.performanceMetrics.GetErrorCount(ContentTypeItems))
}

func TestRecordPlayerFeedback(t *testing.T) {
	cqm := NewContentQualityMetrics()

	feedback := PlayerFeedback{
		Timestamp:   time.Now(),
		ContentType: ContentTypeQuests,
		ContentID:   "quest-001",
		Rating:      5,
		Difficulty:  3,
		Enjoyment:   5,
		Comments:    "Great quest!",
		SessionID:   "session-123",
	}

	cqm.RecordPlayerFeedback(feedback)

	// Verify feedback was recorded
	cqm.engagementMetrics.mu.RLock()
	assert.Len(t, cqm.engagementMetrics.PlayerFeedback, 1)
	assert.Equal(t, feedback, cqm.engagementMetrics.PlayerFeedback[0])
	cqm.engagementMetrics.mu.RUnlock()
}

func TestRecordQuestCompletion(t *testing.T) {
	cqm := NewContentQualityMetrics()

	questID := "quest-001"
	completionTime := 30 * time.Minute

	// Test successful completion
	cqm.RecordQuestCompletion(questID, completionTime, true)

	cqm.engagementMetrics.mu.RLock()
	assert.Contains(t, cqm.engagementMetrics.QuestCompletionTimes, questID)
	assert.Equal(t, completionTime, cqm.engagementMetrics.QuestCompletionTimes[questID])
	cqm.engagementMetrics.mu.RUnlock()

	// Test failed completion (abandonment)
	cqm.RecordQuestCompletion("quest-002", completionTime, false)

	cqm.engagementMetrics.mu.RLock()
	assert.Greater(t, cqm.engagementMetrics.AbandonmentRates[ContentTypeQuests], 0.0)
	cqm.engagementMetrics.mu.RUnlock()
}

func TestGenerateQualityReport(t *testing.T) {
	cqm := NewContentQualityMetrics()

	// Add some test data
	cqm.RecordContentGeneration(ContentTypeTerrain, "test", 100*time.Millisecond, nil)
	cqm.RecordContentGeneration(ContentTypeQuests, "test quest", 200*time.Millisecond, nil)

	feedback := PlayerFeedback{
		Timestamp:   time.Now(),
		ContentType: ContentTypeQuests,
		ContentID:   "quest-001",
		Rating:      4,
		Difficulty:  3,
		Enjoyment:   4,
		SessionID:   "session-123",
	}
	cqm.RecordPlayerFeedback(feedback)

	report := cqm.GenerateQualityReport()

	assert.NotNil(t, report)
	assert.NotEmpty(t, report.Timestamp)
	assert.Contains(t, report.ComponentScores, "performance")
	assert.Contains(t, report.ComponentScores, "variety")
	assert.Contains(t, report.ComponentScores, "consistency")
	assert.Contains(t, report.ComponentScores, "engagement")
	assert.Contains(t, report.ComponentScores, "stability")
	assert.Greater(t, report.OverallScore, 0.0)
	assert.NotEmpty(t, report.QualityGrade)
	assert.NotNil(t, report.ThresholdStatus)
	assert.NotNil(t, report.SystemSummary)
}

func TestQualityScoreCalculation(t *testing.T) {
	cqm := NewContentQualityMetrics()

	// Test performance score calculation
	perfScore := cqm.calculatePerformanceScore()
	assert.Equal(t, 1.0, perfScore) // Should be perfect with no data

	// Add some successful generations
	cqm.RecordContentGeneration(ContentTypeTerrain, "test", 100*time.Millisecond, nil)
	cqm.RecordContentGeneration(ContentTypeQuests, "test", 50*time.Millisecond, nil)

	perfScore = cqm.calculatePerformanceScore()
	assert.Greater(t, perfScore, 0.8) // Should still be high

	// Test variety score calculation
	varietyScore := cqm.calculateVarietyScore()
	assert.Equal(t, 1.0, varietyScore) // Should be perfect with limited data

	// Test consistency score calculation
	consistencyScore := cqm.calculateConsistencyScore()
	assert.Equal(t, 1.0, consistencyScore) // Should be perfect initially

	// Test engagement score calculation
	engagementScore := cqm.calculateEngagementScore()
	assert.Equal(t, 1.0, engagementScore) // Should be perfect with no data

	// Test stability score calculation
	stabilityScore := cqm.calculateStabilityScore()
	assert.Equal(t, 1.0, stabilityScore) // Should be perfect initially
}

func TestQualityGradeCalculation(t *testing.T) {
	cqm := NewContentQualityMetrics()

	tests := []struct {
		score float64
		grade string
	}{
		{0.95, "A"},
		{0.85, "B"},
		{0.75, "C"},
		{0.65, "D"},
		{0.45, "F"},
	}

	for _, test := range tests {
		grade := cqm.calculateQualityGrade(test.score)
		assert.Equal(t, test.grade, grade, "Score %.2f should get grade %s", test.score, test.grade)
	}
}

func TestVarietyMetrics(t *testing.T) {
	vm := NewVarietyMetrics()

	assert.NotNil(t, vm)
	assert.NotNil(t, vm.ContentHashes)
	assert.NotNil(t, vm.UniquenessScores)
	assert.NotNil(t, vm.DiversityMetrics)
	assert.NotNil(t, vm.TemplateUsage)

	// Test content analysis
	vm.analyzeContent(ContentTypeTerrain, "test content 1")
	vm.analyzeContent(ContentTypeTerrain, "test content 2")
	vm.analyzeContent(ContentTypeTerrain, "test content 1") // Duplicate

	// Should have recorded the content
	assert.Len(t, vm.ContentHashes[ContentTypeTerrain], 3)

	// Uniqueness score should be 2/3 (2 unique out of 3 total)
	expectedUniqueness := 2.0 / 3.0
	assert.InDelta(t, expectedUniqueness, vm.UniquenessScores[ContentTypeTerrain], 0.01)
}

func TestConsistencyMetrics(t *testing.T) {
	cm := NewConsistencyMetrics()

	assert.NotNil(t, cm)
	assert.Equal(t, 1.0, cm.NarrativeCoherence)
	assert.Equal(t, 1.0, cm.WorldConsistency)
	assert.Equal(t, 1.0, cm.FactionalConsistency)
	assert.Equal(t, 1.0, cm.TemporalConsistency)
	assert.NotNil(t, cm.InconsistencyCount)
	assert.NotNil(t, cm.ConsistencyHistory)

	// Test consistency validation
	cm.validateConsistency(ContentTypeQuests, "test quest")

	// Should have updated the last check time
	assert.True(t, cm.LastConsistencyCheck.After(time.Now().Add(-time.Second)))
}

func TestEngagementMetrics(t *testing.T) {
	em := NewEngagementMetrics()

	assert.NotNil(t, em)
	assert.NotNil(t, em.CompletionRates)
	assert.NotNil(t, em.AbandonmentRates)
	assert.NotNil(t, em.RetryRates)
	assert.NotNil(t, em.PlayerFeedback)
	assert.NotNil(t, em.QuestCompletionTimes)
	assert.NotNil(t, em.InteractionCounts)
	assert.NotNil(t, em.SatisfactionScores)

	// Test feedback addition
	feedback := PlayerFeedback{
		Timestamp:   time.Now(),
		ContentType: ContentTypeQuests,
		ContentID:   "quest-001",
		Rating:      4,
		SessionID:   "session-123",
	}

	em.addFeedback(feedback)
	assert.Len(t, em.PlayerFeedback, 1)
	assert.Equal(t, feedback, em.PlayerFeedback[0])

	// Test completion recording
	em.recordCompletion(ContentTypeQuests, "quest-001", 30*time.Minute, true)
	assert.Greater(t, em.CompletionRates[ContentTypeQuests], 0.0)
	assert.Contains(t, em.QuestCompletionTimes, "quest-001")

	// Test abandonment recording
	em.recordAbandonment(ContentTypeQuests, "quest-002", 10*time.Minute)
	assert.Greater(t, em.AbandonmentRates[ContentTypeQuests], 0.0)
}

func TestStabilityMetrics(t *testing.T) {
	sm := NewStabilityMetrics()

	assert.NotNil(t, sm)
	assert.NotNil(t, sm.ErrorRates)
	assert.NotNil(t, sm.RecoveryRates)
	assert.Equal(t, 1.0, sm.SystemHealth)
	assert.NotNil(t, sm.MemoryUsage)
	assert.NotNil(t, sm.GenerationLatencies)
	assert.NotNil(t, sm.CriticalErrors)

	// Test error recording
	testError := assert.AnError
	sm.recordError(ContentTypeTerrain, testError)
	assert.Greater(t, sm.ErrorRates[ContentTypeTerrain], 0.0)

	// Test success recording
	sm.recordSuccess(ContentTypeTerrain, 100*time.Millisecond)
	assert.Less(t, sm.ErrorRates[ContentTypeTerrain], 0.1) // Should have decreased

	// Test latency tracking
	assert.Len(t, sm.GenerationLatencies[ContentTypeTerrain], 1)
	assert.Equal(t, 100*time.Millisecond, sm.GenerationLatencies[ContentTypeTerrain][0])
}

func TestQualityThresholds(t *testing.T) {
	thresholds := NewDefaultQualityThresholds()

	assert.NotNil(t, thresholds)
	assert.Greater(t, thresholds.MinUniquenessScore, 0.0)
	assert.Greater(t, thresholds.MinConsistencyScore, 0.0)
	assert.Greater(t, thresholds.MinCompletionRate, 0.0)
	assert.Greater(t, thresholds.MaxErrorRate, 0.0)
	assert.Greater(t, thresholds.MaxGenerationTime, time.Duration(0))
	assert.Greater(t, thresholds.MinSatisfactionScore, 0.0)
	assert.Greater(t, thresholds.MinSystemHealth, 0.0)

	// Test quality weights sum to 1.0
	weights := thresholds.QualityWeights
	totalWeight := weights.Performance + weights.Variety + weights.Consistency + weights.Engagement + weights.Stability
	assert.InDelta(t, 1.0, totalWeight, 0.01)
}

func TestContentHashGeneration(t *testing.T) {
	vm := NewVarietyMetrics()

	// Test that different content generates different hashes
	hash1 := vm.generateContentHash("content 1")
	hash2 := vm.generateContentHash("content 2")
	hash3 := vm.generateContentHash("content 1") // Same as hash1

	assert.NotEqual(t, hash1, hash2)
	assert.Equal(t, hash1, hash3)
	assert.NotEmpty(t, hash1)
	assert.NotEmpty(t, hash2)
}

func TestConcurrentQualityMetrics(t *testing.T) {
	cqm := NewContentQualityMetrics()

	// Test concurrent access to quality metrics
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				// Record various operations
				content := "test content"
				duration := time.Duration(j) * time.Millisecond

				cqm.RecordContentGeneration(ContentTypeTerrain, content, duration, nil)

				feedback := PlayerFeedback{
					Timestamp:   time.Now(),
					ContentType: ContentTypeQuests,
					ContentID:   "quest-001",
					Rating:      4,
					SessionID:   "session-123",
				}
				cqm.RecordPlayerFeedback(feedback)

				cqm.RecordQuestCompletion("quest-001", 30*time.Minute, true)

				// Generate quality report
				report := cqm.GenerateQualityReport()
				assert.NotNil(t, report)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state
	assert.Equal(t, int64(500), cqm.performanceMetrics.TotalGenerations)
	assert.Greater(t, cqm.GetOverallQualityScore(), 0.0)
}

func BenchmarkContentGeneration(b *testing.B) {
	cqm := NewContentQualityMetrics()
	content := "benchmark test content"
	duration := 100 * time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cqm.RecordContentGeneration(ContentTypeTerrain, content, duration, nil)
	}
}

func BenchmarkQualityReportGeneration(b *testing.B) {
	cqm := NewContentQualityMetrics()

	// Setup some data
	for i := 0; i < 100; i++ {
		cqm.RecordContentGeneration(ContentTypeTerrain, "test", 100*time.Millisecond, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		report := cqm.GenerateQualityReport()
		_ = report
	}
}
