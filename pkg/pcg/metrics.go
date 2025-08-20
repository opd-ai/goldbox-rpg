package pcg

import (
	"crypto/sha256"
	"fmt"
	"math"
	"sync"
	"time"
)

// GenerationMetrics tracks performance statistics
type GenerationMetrics struct {
	mu               sync.RWMutex
	GenerationCounts map[ContentType]int64         `json:"generation_counts"`
	AverageTimings   map[ContentType]time.Duration `json:"average_timings"`
	ErrorCounts      map[ContentType]int64         `json:"error_counts"`
	CacheHits        int64                         `json:"cache_hits"`
	CacheMisses      int64                         `json:"cache_misses"`
	TotalGenerations int64                         `json:"total_generations"`
}

// ContentQualityMetrics provides comprehensive tracking of generated content quality
type ContentQualityMetrics struct {
	mu                    sync.RWMutex
	performanceMetrics    *GenerationMetrics
	validationMetrics     *ValidationMetrics
	balanceMetrics        *BalanceMetrics
	varietyMetrics        *VarietyMetrics
	consistencyMetrics    *ConsistencyMetrics
	engagementMetrics     *EngagementMetrics
	stabilityMetrics      *StabilityMetrics
	qualityThresholds     *QualityThresholds
	lastQualityAssessment time.Time
	overallQualityScore   float64
}

// VarietyMetrics tracks content uniqueness and diversity
type VarietyMetrics struct {
	mu                sync.RWMutex
	contentHashes     map[ContentType][]string      `json:"content_hashes"`
	uniquenessScores  map[ContentType]float64       `json:"uniqueness_scores"`
	diversityMetrics  map[ContentType]DiversityData `json:"diversity_metrics"`
	templateUsage     map[string]int64              `json:"template_usage"`
	lastVarietyUpdate time.Time                     `json:"last_variety_update"`
}

// DiversityData tracks specific diversity aspects per content type
type DiversityData struct {
	AttributeDistribution  map[string]int64 `json:"attribute_distribution"`
	TypeDistribution       map[string]int64 `json:"type_distribution"`
	ComplexityDistribution map[string]int64 `json:"complexity_distribution"`
	ShannonEntropy         float64          `json:"shannon_entropy"`
}

// ConsistencyMetrics tracks logical coherence and narrative consistency
type ConsistencyMetrics struct {
	mu                   sync.RWMutex
	narrativeCoherence   float64               `json:"narrative_coherence"`
	worldConsistency     float64               `json:"world_consistency"`
	factionalConsistency float64               `json:"factional_consistency"`
	temporalConsistency  float64               `json:"temporal_consistency"`
	inconsistencyCount   map[string]int64      `json:"inconsistency_count"`
	consistencyHistory   []ConsistencySnapshot `json:"consistency_history"`
	lastConsistencyCheck time.Time             `json:"last_consistency_check"`
}

// ConsistencySnapshot represents a point-in-time consistency measurement
type ConsistencySnapshot struct {
	Timestamp          time.Time          `json:"timestamp"`
	OverallScore       float64            `json:"overall_score"`
	ComponentScores    map[string]float64 `json:"component_scores"`
	InconsistencyTypes map[string]int64   `json:"inconsistency_types"`
}

// EngagementMetrics tracks player interaction and satisfaction with generated content
type EngagementMetrics struct {
	mu                   sync.RWMutex
	completionRates      map[ContentType]float64  `json:"completion_rates"`
	abandonmentRates     map[ContentType]float64  `json:"abandonment_rates"`
	retryRates           map[ContentType]float64  `json:"retry_rates"`
	playerFeedback       []PlayerFeedback         `json:"player_feedback"`
	questCompletionTimes map[string]time.Duration `json:"quest_completion_times"`
	interactionCounts    map[string]int64         `json:"interaction_counts"`
	satisfactionScores   map[ContentType]float64  `json:"satisfaction_scores"`
	lastEngagementUpdate time.Time                `json:"last_engagement_update"`
}

// PlayerFeedback represents structured player feedback data
type PlayerFeedback struct {
	Timestamp   time.Time   `json:"timestamp"`
	ContentType ContentType `json:"content_type"`
	ContentID   string      `json:"content_id"`
	Rating      int         `json:"rating"`     // 1-5 scale
	Difficulty  int         `json:"difficulty"` // 1-5 scale
	Enjoyment   int         `json:"enjoyment"`  // 1-5 scale
	Comments    string      `json:"comments"`
	SessionID   string      `json:"session_id"`
}

// StabilityMetrics tracks technical reliability and system health
type StabilityMetrics struct {
	mu                  sync.RWMutex
	errorRates          map[ContentType]float64         `json:"error_rates"`
	recoveryRates       map[ContentType]float64         `json:"recovery_rates"`
	systemHealth        float64                         `json:"system_health"`
	memoryUsage         map[ContentType]int64           `json:"memory_usage"`
	generationLatencies map[ContentType][]time.Duration `json:"generation_latencies"`
	criticalErrors      []CriticalError                 `json:"critical_errors"`
	uptime              time.Duration                   `json:"uptime"`
	systemStartTime     time.Time                       `json:"system_start_time"`
	lastStabilityCheck  time.Time                       `json:"last_stability_check"`
}

// CriticalError represents a serious system failure
type CriticalError struct {
	Timestamp    time.Time     `json:"timestamp"`
	ContentType  ContentType   `json:"content_type"`
	ErrorType    string        `json:"error_type"`
	ErrorMessage string        `json:"error_message"`
	StackTrace   string        `json:"stack_trace"`
	RecoveryTime time.Duration `json:"recovery_time"`
	Recovered    bool          `json:"recovered"`
}

// QualityThresholds defines acceptable quality levels for various metrics
type QualityThresholds struct {
	MinUniquenessScore   float64        `json:"min_uniqueness_score"`
	MinConsistencyScore  float64        `json:"min_consistency_score"`
	MinCompletionRate    float64        `json:"min_completion_rate"`
	MaxErrorRate         float64        `json:"max_error_rate"`
	MaxGenerationTime    time.Duration  `json:"max_generation_time"`
	MinSatisfactionScore float64        `json:"min_satisfaction_score"`
	MinSystemHealth      float64        `json:"min_system_health"`
	QualityWeights       QualityWeights `json:"quality_weights"`
}

// QualityWeights defines the relative importance of different quality aspects
type QualityWeights struct {
	Performance float64 `json:"performance"`
	Variety     float64 `json:"variety"`
	Consistency float64 `json:"consistency"`
	Engagement  float64 `json:"engagement"`
	Stability   float64 `json:"stability"`
}

// QualityReport provides a comprehensive assessment of content generation quality
type QualityReport struct {
	Timestamp       time.Time              `json:"timestamp"`
	OverallScore    float64                `json:"overall_score"`
	ComponentScores map[string]float64     `json:"component_scores"`
	QualityGrade    string                 `json:"quality_grade"`
	ThresholdStatus map[string]bool        `json:"threshold_status"`
	Recommendations []string               `json:"recommendations"`
	CriticalIssues  []string               `json:"critical_issues"`
	TrendAnalysis   map[string]TrendData   `json:"trend_analysis"`
	SystemSummary   map[string]interface{} `json:"system_summary"`
}

// TrendData represents quality trends over time
type TrendData struct {
	Direction  string    `json:"direction"`  // "improving", "declining", "stable"
	Magnitude  float64   `json:"magnitude"`  // How much change
	Confidence float64   `json:"confidence"` // Statistical confidence in trend
	LastChange time.Time `json:"last_change"`
}

// NewContentQualityMetrics creates a comprehensive quality metrics system
func NewContentQualityMetrics() *ContentQualityMetrics {
	return &ContentQualityMetrics{
		performanceMetrics:    NewGenerationMetrics(),
		validationMetrics:     NewValidationMetrics(),
		balanceMetrics:        NewBalanceMetrics(),
		varietyMetrics:        NewVarietyMetrics(),
		consistencyMetrics:    NewConsistencyMetrics(),
		engagementMetrics:     NewEngagementMetrics(),
		stabilityMetrics:      NewStabilityMetrics(),
		qualityThresholds:     NewDefaultQualityThresholds(),
		lastQualityAssessment: time.Now(),
		overallQualityScore:   0.0,
	}
}

// NewVarietyMetrics creates a new variety metrics tracker
func NewVarietyMetrics() *VarietyMetrics {
	return &VarietyMetrics{
		contentHashes:     make(map[ContentType][]string),
		uniquenessScores:  make(map[ContentType]float64),
		diversityMetrics:  make(map[ContentType]DiversityData),
		templateUsage:     make(map[string]int64),
		lastVarietyUpdate: time.Now(),
	}
}

// NewConsistencyMetrics creates a new consistency metrics tracker
func NewConsistencyMetrics() *ConsistencyMetrics {
	return &ConsistencyMetrics{
		narrativeCoherence:   1.0,
		worldConsistency:     1.0,
		factionalConsistency: 1.0,
		temporalConsistency:  1.0,
		inconsistencyCount:   make(map[string]int64),
		consistencyHistory:   make([]ConsistencySnapshot, 0),
		lastConsistencyCheck: time.Now(),
	}
}

// NewEngagementMetrics creates a new engagement metrics tracker
func NewEngagementMetrics() *EngagementMetrics {
	return &EngagementMetrics{
		completionRates:      make(map[ContentType]float64),
		abandonmentRates:     make(map[ContentType]float64),
		retryRates:           make(map[ContentType]float64),
		playerFeedback:       make([]PlayerFeedback, 0),
		questCompletionTimes: make(map[string]time.Duration),
		interactionCounts:    make(map[string]int64),
		satisfactionScores:   make(map[ContentType]float64),
		lastEngagementUpdate: time.Now(),
	}
}

// NewStabilityMetrics creates a new stability metrics tracker
func NewStabilityMetrics() *StabilityMetrics {
	return &StabilityMetrics{
		errorRates:          make(map[ContentType]float64),
		recoveryRates:       make(map[ContentType]float64),
		systemHealth:        1.0,
		memoryUsage:         make(map[ContentType]int64),
		generationLatencies: make(map[ContentType][]time.Duration),
		criticalErrors:      make([]CriticalError, 0),
		uptime:              0,
		systemStartTime:     time.Now(),
		lastStabilityCheck:  time.Now(),
	}
}

// NewDefaultQualityThresholds creates default quality thresholds
func NewDefaultQualityThresholds() *QualityThresholds {
	return &QualityThresholds{
		MinUniquenessScore:   0.7,
		MinConsistencyScore:  0.8,
		MinCompletionRate:    0.6,
		MaxErrorRate:         0.05,
		MaxGenerationTime:    5 * time.Second,
		MinSatisfactionScore: 3.0,
		MinSystemHealth:      0.9,
		QualityWeights: QualityWeights{
			Performance: 0.2,
			Variety:     0.2,
			Consistency: 0.25,
			Engagement:  0.2,
			Stability:   0.15,
		},
	}
}

// NewGenerationMetrics creates a new metrics tracker
func NewGenerationMetrics() *GenerationMetrics {
	return &GenerationMetrics{
		GenerationCounts: make(map[ContentType]int64),
		AverageTimings:   make(map[ContentType]time.Duration),
		ErrorCounts:      make(map[ContentType]int64),
	}
}

// RecordGeneration records a successful generation
func (gm *GenerationMetrics) RecordGeneration(contentType ContentType, duration time.Duration) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.GenerationCounts[contentType]++
	gm.TotalGenerations++

	// Update rolling average
	if current, exists := gm.AverageTimings[contentType]; exists {
		count := gm.GenerationCounts[contentType]
		gm.AverageTimings[contentType] = (current*time.Duration(count-1) + duration) / time.Duration(count)
	} else {
		gm.AverageTimings[contentType] = duration
	}
}

// RecordError records a generation error
func (gm *GenerationMetrics) RecordError(contentType ContentType) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.ErrorCounts[contentType]++
}

// RecordCacheHit records a cache hit
func (gm *GenerationMetrics) RecordCacheHit() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.CacheHits++
}

// RecordCacheMiss records a cache miss
func (gm *GenerationMetrics) RecordCacheMiss() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.CacheMisses++
}

// GetStats returns current performance statistics
func (gm *GenerationMetrics) GetStats() map[string]interface{} {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	return map[string]interface{}{
		"generation_counts": gm.GenerationCounts,
		"average_timings":   gm.AverageTimings,
		"error_counts":      gm.ErrorCounts,
		"cache_hits":        gm.CacheHits,
		"cache_misses":      gm.CacheMisses,
		"total_generations": gm.TotalGenerations,
	}
}

// GetGenerationCount returns the total generation count for a content type
func (gm *GenerationMetrics) GetGenerationCount(contentType ContentType) int64 {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	return gm.GenerationCounts[contentType]
}

// GetAverageTiming returns the average generation time for a content type
func (gm *GenerationMetrics) GetAverageTiming(contentType ContentType) time.Duration {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	return gm.AverageTimings[contentType]
}

// GetErrorCount returns the total error count for a content type
func (gm *GenerationMetrics) GetErrorCount(contentType ContentType) int64 {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	return gm.ErrorCounts[contentType]
}

// GetCacheHitRatio returns the cache hit ratio as a percentage
func (gm *GenerationMetrics) GetCacheHitRatio() float64 {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	total := gm.CacheHits + gm.CacheMisses
	if total == 0 {
		return 0.0
	}

	return float64(gm.CacheHits) / float64(total) * 100.0
}

// Reset clears all metrics data
func (gm *GenerationMetrics) Reset() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.GenerationCounts = make(map[ContentType]int64)
	gm.AverageTimings = make(map[ContentType]time.Duration)
	gm.ErrorCounts = make(map[ContentType]int64)
	gm.CacheHits = 0
	gm.CacheMisses = 0
	gm.TotalGenerations = 0
}

// ContentQualityMetrics methods

// RecordContentGeneration records a content generation event with quality assessment
func (cqm *ContentQualityMetrics) RecordContentGeneration(contentType ContentType, content interface{}, duration time.Duration, err error) {
	cqm.mu.Lock()
	defer cqm.mu.Unlock()

	// Record performance metrics
	if err != nil {
		cqm.performanceMetrics.RecordError(contentType)
		cqm.stabilityMetrics.recordError(contentType, err)
	} else {
		cqm.performanceMetrics.RecordGeneration(contentType, duration)
		cqm.stabilityMetrics.recordSuccess(contentType, duration)

		// Analyze content for variety and consistency
		cqm.varietyMetrics.analyzeContent(contentType, content)
		cqm.consistencyMetrics.validateConsistency(contentType, content)
	}
}

// RecordPlayerFeedback records player feedback for content quality assessment
func (cqm *ContentQualityMetrics) RecordPlayerFeedback(feedback PlayerFeedback) {
	cqm.mu.Lock()
	defer cqm.mu.Unlock()

	cqm.engagementMetrics.addFeedback(feedback)
	cqm.updateEngagementScores()
}

// RecordQuestCompletion records quest completion for engagement tracking
func (cqm *ContentQualityMetrics) RecordQuestCompletion(questID string, completionTime time.Duration, completed bool) {
	cqm.mu.Lock()
	defer cqm.mu.Unlock()

	cqm.engagementMetrics.recordCompletion(ContentTypeQuests, questID, completionTime, completed)
}

// RecordContentAbandonment records when players abandon content
func (cqm *ContentQualityMetrics) RecordContentAbandonment(contentType ContentType, contentID string, timeSpent time.Duration) {
	cqm.mu.Lock()
	defer cqm.mu.Unlock()

	cqm.engagementMetrics.recordAbandonment(contentType, contentID, timeSpent)
}

// GenerateQualityReport creates a comprehensive quality assessment
func (cqm *ContentQualityMetrics) GenerateQualityReport() *QualityReport {
	cqm.mu.RLock()
	defer cqm.mu.RUnlock()

	report := &QualityReport{
		Timestamp:       time.Now(),
		ComponentScores: make(map[string]float64),
		ThresholdStatus: make(map[string]bool),
		Recommendations: make([]string, 0),
		CriticalIssues:  make([]string, 0),
		TrendAnalysis:   make(map[string]TrendData),
		SystemSummary:   make(map[string]interface{}),
	}

	// Calculate component scores
	performanceScore := cqm.calculatePerformanceScore()
	varietyScore := cqm.calculateVarietyScore()
	consistencyScore := cqm.calculateConsistencyScore()
	engagementScore := cqm.calculateEngagementScore()
	stabilityScore := cqm.calculateStabilityScore()

	report.ComponentScores["performance"] = performanceScore
	report.ComponentScores["variety"] = varietyScore
	report.ComponentScores["consistency"] = consistencyScore
	report.ComponentScores["engagement"] = engagementScore
	report.ComponentScores["stability"] = stabilityScore

	// Calculate overall score using weights
	weights := cqm.qualityThresholds.QualityWeights
	report.OverallScore = performanceScore*weights.Performance +
		varietyScore*weights.Variety +
		consistencyScore*weights.Consistency +
		engagementScore*weights.Engagement +
		stabilityScore*weights.Stability

	// Determine quality grade
	report.QualityGrade = cqm.calculateQualityGrade(report.OverallScore)

	// Check threshold compliance
	report.ThresholdStatus = cqm.checkThresholdCompliance()

	// Generate recommendations and identify critical issues
	report.Recommendations = cqm.generateRecommendations(report.ComponentScores)
	report.CriticalIssues = cqm.identifyCriticalIssues(report.ComponentScores, report.ThresholdStatus)

	// Add trend analysis
	report.TrendAnalysis = cqm.analyzeTrends()

	// Add system summary
	report.SystemSummary = cqm.getSystemSummary()

	// Update overall quality score
	cqm.overallQualityScore = report.OverallScore
	cqm.lastQualityAssessment = report.Timestamp

	return report
}

// GetOverallQualityScore returns the current overall quality score
func (cqm *ContentQualityMetrics) GetOverallQualityScore() float64 {
	cqm.mu.RLock()
	defer cqm.mu.RUnlock()
	return cqm.overallQualityScore
}

// GetPerformanceMetrics returns the performance metrics instance
func (cqm *ContentQualityMetrics) GetPerformanceMetrics() *GenerationMetrics {
	return cqm.performanceMetrics
}

// GetValidationMetrics returns the validation metrics instance
func (cqm *ContentQualityMetrics) GetValidationMetrics() *ValidationMetrics {
	return cqm.validationMetrics
}

// GetBalanceMetrics returns the balance metrics instance
func (cqm *ContentQualityMetrics) GetBalanceMetrics() *BalanceMetrics {
	return cqm.balanceMetrics
}

// calculatePerformanceScore computes a performance quality score
func (cqm *ContentQualityMetrics) calculatePerformanceScore() float64 {
	stats := cqm.performanceMetrics.GetStats()

	// Base score starts at 1.0 (perfect)
	score := 1.0

	// Penalize for errors
	if totalGens, ok := stats["total_generations"].(int64); ok && totalGens > 0 {
		totalErrors := int64(0)
		if errorCounts, ok := stats["error_counts"].(map[ContentType]int64); ok {
			for _, count := range errorCounts {
				totalErrors += count
			}
		}
		errorRate := float64(totalErrors) / float64(totalGens)
		if errorRate > cqm.qualityThresholds.MaxErrorRate {
			score -= (errorRate - cqm.qualityThresholds.MaxErrorRate) * 2.0
		}
	}

	// Penalize for slow generation times
	if avgTimings, ok := stats["average_timings"].(map[ContentType]time.Duration); ok {
		for _, timing := range avgTimings {
			if timing > cqm.qualityThresholds.MaxGenerationTime {
				score -= 0.1
			}
		}
	}

	// Reward for cache efficiency
	cacheRatio := cqm.performanceMetrics.GetCacheHitRatio()
	if cacheRatio > 80.0 {
		score += 0.1
	}

	return math.Max(0.0, math.Min(1.0, score))
}

// calculateVarietyScore computes a content variety quality score
func (cqm *ContentQualityMetrics) calculateVarietyScore() float64 {
	cqm.varietyMetrics.mu.RLock()
	defer cqm.varietyMetrics.mu.RUnlock()

	if len(cqm.varietyMetrics.uniquenessScores) == 0 {
		return 1.0 // No content generated yet, assume perfect
	}

	totalScore := 0.0
	count := 0

	for _, score := range cqm.varietyMetrics.uniquenessScores {
		totalScore += score
		count++
	}

	if count == 0 {
		return 1.0
	}

	return totalScore / float64(count)
}

// calculateConsistencyScore computes a logical consistency quality score
func (cqm *ContentQualityMetrics) calculateConsistencyScore() float64 {
	cqm.consistencyMetrics.mu.RLock()
	defer cqm.consistencyMetrics.mu.RUnlock()

	// Weight different consistency aspects
	weights := map[string]float64{
		"narrative": 0.3,
		"world":     0.3,
		"factional": 0.2,
		"temporal":  0.2,
	}

	score := weights["narrative"]*cqm.consistencyMetrics.narrativeCoherence +
		weights["world"]*cqm.consistencyMetrics.worldConsistency +
		weights["factional"]*cqm.consistencyMetrics.factionalConsistency +
		weights["temporal"]*cqm.consistencyMetrics.temporalConsistency

	return math.Max(0.0, math.Min(1.0, score))
}

// calculateEngagementScore computes a player engagement quality score
func (cqm *ContentQualityMetrics) calculateEngagementScore() float64 {
	cqm.engagementMetrics.mu.RLock()
	defer cqm.engagementMetrics.mu.RUnlock()

	if len(cqm.engagementMetrics.completionRates) == 0 {
		return 1.0 // No data yet
	}

	totalCompletionRate := 0.0
	totalSatisfaction := 0.0
	count := 0

	for _, rate := range cqm.engagementMetrics.completionRates {
		totalCompletionRate += rate
		count++
	}

	satisfactionCount := 0
	for _, score := range cqm.engagementMetrics.satisfactionScores {
		totalSatisfaction += score / 5.0 // Normalize to 0-1
		satisfactionCount++
	}

	if count == 0 {
		return 1.0
	}

	completionScore := totalCompletionRate / float64(count)
	satisfactionScore := 1.0
	if satisfactionCount > 0 {
		satisfactionScore = totalSatisfaction / float64(satisfactionCount)
	}

	// Weighted combination
	return 0.6*completionScore + 0.4*satisfactionScore
}

// calculateStabilityScore computes a system stability quality score
func (cqm *ContentQualityMetrics) calculateStabilityScore() float64 {
	cqm.stabilityMetrics.mu.RLock()
	defer cqm.stabilityMetrics.mu.RUnlock()

	// Start with system health as base score
	score := cqm.stabilityMetrics.systemHealth

	// Penalize for high error rates
	totalErrorRate := 0.0
	count := 0
	for _, rate := range cqm.stabilityMetrics.errorRates {
		totalErrorRate += rate
		count++
	}

	if count > 0 {
		avgErrorRate := totalErrorRate / float64(count)
		if avgErrorRate > cqm.qualityThresholds.MaxErrorRate {
			score -= (avgErrorRate - cqm.qualityThresholds.MaxErrorRate) * 2.0
		}
	}

	// Consider critical errors
	recentCriticalErrors := 0
	cutoff := time.Now().Add(-1 * time.Hour)
	for _, err := range cqm.stabilityMetrics.criticalErrors {
		if err.Timestamp.After(cutoff) {
			recentCriticalErrors++
		}
	}

	if recentCriticalErrors > 0 {
		score -= float64(recentCriticalErrors) * 0.1
	}

	return math.Max(0.0, math.Min(1.0, score))
}

// calculateQualityGrade converts a numeric score to a letter grade
func (cqm *ContentQualityMetrics) calculateQualityGrade(score float64) string {
	if score >= 0.9 {
		return "A"
	} else if score >= 0.8 {
		return "B"
	} else if score >= 0.7 {
		return "C"
	} else if score >= 0.6 {
		return "D"
	} else {
		return "F"
	}
}

// checkThresholdCompliance checks if metrics meet quality thresholds
func (cqm *ContentQualityMetrics) checkThresholdCompliance() map[string]bool {
	status := make(map[string]bool)

	// Check performance thresholds
	status["error_rate"] = cqm.calculatePerformanceScore() >= 0.8
	status["generation_time"] = true // Checked in performance score

	// Check variety thresholds
	status["uniqueness"] = cqm.calculateVarietyScore() >= cqm.qualityThresholds.MinUniquenessScore

	// Check consistency thresholds
	status["consistency"] = cqm.calculateConsistencyScore() >= cqm.qualityThresholds.MinConsistencyScore

	// Check engagement thresholds
	status["completion_rate"] = cqm.calculateEngagementScore() >= 0.8

	// Check stability thresholds
	status["system_health"] = cqm.stabilityMetrics.systemHealth >= cqm.qualityThresholds.MinSystemHealth

	return status
}

// generateRecommendations creates actionable recommendations based on metrics
func (cqm *ContentQualityMetrics) generateRecommendations(scores map[string]float64) []string {
	recommendations := make([]string, 0)

	if scores["performance"] < 0.8 {
		recommendations = append(recommendations, "Consider optimizing generation algorithms to improve performance")
	}

	if scores["variety"] < 0.7 {
		recommendations = append(recommendations, "Increase template diversity and randomization parameters")
	}

	if scores["consistency"] < 0.8 {
		recommendations = append(recommendations, "Review validation rules and consistency checking logic")
	}

	if scores["engagement"] < 0.7 {
		recommendations = append(recommendations, "Analyze player feedback to improve content appeal")
	}

	if scores["stability"] < 0.9 {
		recommendations = append(recommendations, "Review error handling and system reliability measures")
	}

	return recommendations
}

// identifyCriticalIssues identifies issues requiring immediate attention
func (cqm *ContentQualityMetrics) identifyCriticalIssues(scores map[string]float64, thresholds map[string]bool) []string {
	issues := make([]string, 0)

	if scores["performance"] < 0.5 {
		issues = append(issues, "Critical performance degradation detected")
	}

	if scores["stability"] < 0.7 {
		issues = append(issues, "System stability below acceptable levels")
	}

	if !thresholds["system_health"] {
		issues = append(issues, "System health check failed")
	}

	return issues
}

// analyzeTrends analyzes quality trends over time
func (cqm *ContentQualityMetrics) analyzeTrends() map[string]TrendData {
	trends := make(map[string]TrendData)

	// For now, return placeholder trend data
	// In a full implementation, this would analyze historical data
	trends["overall"] = TrendData{
		Direction:  "stable",
		Magnitude:  0.0,
		Confidence: 0.5,
		LastChange: time.Now(),
	}

	return trends
}

// getSystemSummary provides a high-level system summary
func (cqm *ContentQualityMetrics) getSystemSummary() map[string]interface{} {
	summary := make(map[string]interface{})

	summary["total_generations"] = cqm.performanceMetrics.TotalGenerations
	summary["uptime"] = time.Since(cqm.stabilityMetrics.systemStartTime)
	summary["cache_hit_ratio"] = cqm.performanceMetrics.GetCacheHitRatio()
	summary["overall_quality"] = cqm.overallQualityScore
	summary["last_assessment"] = cqm.lastQualityAssessment

	return summary
}

// Helper methods for sub-metrics components

// analyzeContent analyzes generated content for variety metrics
func (vm *VarietyMetrics) analyzeContent(contentType ContentType, content interface{}) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	// Generate content hash for uniqueness tracking
	contentHash := vm.generateContentHash(content)

	if vm.contentHashes[contentType] == nil {
		vm.contentHashes[contentType] = make([]string, 0)
	}

	vm.contentHashes[contentType] = append(vm.contentHashes[contentType], contentHash)

	// Calculate uniqueness score
	vm.updateUniquenessScore(contentType)

	vm.lastVarietyUpdate = time.Now()
}

// generateContentHash creates a hash representation of content
func (vm *VarietyMetrics) generateContentHash(content interface{}) string {
	// Simple string representation for hashing
	contentStr := fmt.Sprintf("%+v", content)
	hash := sha256.Sum256([]byte(contentStr))
	return fmt.Sprintf("%x", hash)
}

// updateUniquenessScore calculates uniqueness based on hash diversity
func (vm *VarietyMetrics) updateUniquenessScore(contentType ContentType) {
	hashes := vm.contentHashes[contentType]
	if len(hashes) < 2 {
		vm.uniquenessScores[contentType] = 1.0
		return
	}

	// Count unique hashes
	uniqueHashes := make(map[string]bool)
	for _, hash := range hashes {
		uniqueHashes[hash] = true
	}

	uniquenessScore := float64(len(uniqueHashes)) / float64(len(hashes))
	vm.uniquenessScores[contentType] = uniquenessScore
}

// validateConsistency checks content for logical consistency
func (cm *ConsistencyMetrics) validateConsistency(contentType ContentType, content interface{}) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Placeholder consistency validation
	// In a full implementation, this would perform deep consistency checks

	cm.lastConsistencyCheck = time.Now()
}

// addFeedback adds player feedback to engagement metrics
func (em *EngagementMetrics) addFeedback(feedback PlayerFeedback) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.playerFeedback = append(em.playerFeedback, feedback)
	em.lastEngagementUpdate = time.Now()
}

// recordCompletion records content completion for engagement tracking
func (em *EngagementMetrics) recordCompletion(contentType ContentType, contentID string, completionTime time.Duration, completed bool) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if completed {
		if em.completionRates[contentType] == 0 {
			em.completionRates[contentType] = 1.0
		} else {
			// Update rolling average (simplified)
			em.completionRates[contentType] = (em.completionRates[contentType] + 1.0) / 2.0
		}

		if contentType == ContentTypeQuests {
			em.questCompletionTimes[contentID] = completionTime
		}
	} else {
		em.abandonmentRates[contentType]++
	}

	em.lastEngagementUpdate = time.Now()
}

// recordAbandonment records content abandonment
func (em *EngagementMetrics) recordAbandonment(contentType ContentType, contentID string, timeSpent time.Duration) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.abandonmentRates[contentType]++
	em.lastEngagementUpdate = time.Now()
}

// updateEngagementScores updates satisfaction scores based on feedback
func (cqm *ContentQualityMetrics) updateEngagementScores() {
	em := cqm.engagementMetrics
	em.mu.Lock()
	defer em.mu.Unlock()

	// Calculate satisfaction scores by content type
	contentTypeCounts := make(map[ContentType]int)
	contentTypeTotals := make(map[ContentType]float64)

	for _, feedback := range em.playerFeedback {
		contentTypeCounts[feedback.ContentType]++
		contentTypeTotals[feedback.ContentType] += float64(feedback.Rating)
	}

	for contentType, total := range contentTypeTotals {
		count := contentTypeCounts[contentType]
		if count > 0 {
			em.satisfactionScores[contentType] = total / float64(count)
		}
	}
}

// recordError records an error for stability tracking
func (sm *StabilityMetrics) recordError(contentType ContentType, err error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.errorRates[contentType] == 0 {
		sm.errorRates[contentType] = 0.1
	} else {
		sm.errorRates[contentType] = math.Min(1.0, sm.errorRates[contentType]+0.1)
	}

	sm.lastStabilityCheck = time.Now()
}

// recordSuccess records a successful operation for stability tracking
func (sm *StabilityMetrics) recordSuccess(contentType ContentType, duration time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Reduce error rate on success
	if sm.errorRates[contentType] > 0 {
		sm.errorRates[contentType] = math.Max(0.0, sm.errorRates[contentType]-0.01)
	}

	// Track latency
	if sm.generationLatencies[contentType] == nil {
		sm.generationLatencies[contentType] = make([]time.Duration, 0)
	}
	sm.generationLatencies[contentType] = append(sm.generationLatencies[contentType], duration)

	// Keep only recent latencies (last 100)
	if len(sm.generationLatencies[contentType]) > 100 {
		sm.generationLatencies[contentType] = sm.generationLatencies[contentType][1:]
	}

	sm.lastStabilityCheck = time.Now()
}

// NewBalanceMetrics creates a new balance metrics tracker (for integration)
func NewBalanceMetrics() *BalanceMetrics {
	// This should return the actual BalanceMetrics from balancer.go
	return &BalanceMetrics{
		TotalBalanceChecks:     0,
		SuccessfulBalances:     0,
		FailedBalances:         0,
		CriticalFailures:       0,
		AverageBalanceTime:     0,
		ContentTypeMetrics:     make(map[ContentType]TypeMetrics),
		ResourceUsageMetrics:   make(map[string]ResourceMetrics),
		DifficultyDistribution: make(map[int]int64),
		LastBalanceCheck:       time.Now(),
		SystemHealth:           1.0,
	}
}
