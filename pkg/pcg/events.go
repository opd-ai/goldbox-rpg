package pcg

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"goldbox-rpg/pkg/game"
)

// PCG-specific event types for runtime adjustments
const (
	// EventPCGContentGenerated is emitted when PCG content is generated
	EventPCGContentGenerated game.EventType = iota + 1000
	// EventPCGQualityAssessment is emitted when content quality is assessed
	EventPCGQualityAssessment
	// EventPCGPlayerFeedback is emitted when player feedback is received
	EventPCGPlayerFeedback
	// EventPCGDifficultyAdjustment is emitted when difficulty should be adjusted
	EventPCGDifficultyAdjustment
	// EventPCGContentRequest is emitted when new content is needed
	EventPCGContentRequest
	// EventPCGSystemHealth is emitted for system health monitoring
	EventPCGSystemHealth
)

// PCGEventData contains common data structures for PCG events
type PCGEventData struct {
	ContentType      ContentType            `json:"content_type"`
	GenerationTime   time.Duration          `json:"generation_time"`
	QualityScore     float64               `json:"quality_score"`
	PlayerFeedback   *PlayerFeedback       `json:"player_feedback,omitempty"`
	AdjustmentParams map[string]interface{} `json:"adjustment_params,omitempty"`
	Timestamp        time.Time             `json:"timestamp"`
}

// RuntimeAdjustmentConfig defines configuration for runtime PCG adjustments
type RuntimeAdjustmentConfig struct {
	// EnableRuntimeAdjustments enables/disables the runtime adjustment system
	EnableRuntimeAdjustments bool `yaml:"enable_runtime_adjustments"`
	
	// QualityThresholds define when adjustments should be made
	QualityThresholds struct {
		MinOverallScore    float64 `yaml:"min_overall_score"`    // Below this, make adjustments
		MinPerformance     float64 `yaml:"min_performance"`      // Performance threshold
		MinVariety         float64 `yaml:"min_variety"`          // Variety threshold
		MinConsistency     float64 `yaml:"min_consistency"`      // Consistency threshold
		MinEngagement      float64 `yaml:"min_engagement"`       // Engagement threshold
		MinStability       float64 `yaml:"min_stability"`        // Stability threshold
	} `yaml:"quality_thresholds"`
	
	// AdjustmentRates control how aggressively adjustments are made
	AdjustmentRates struct {
		DifficultyStep      float64 `yaml:"difficulty_step"`       // How much to adjust difficulty
		VarietyBoost        float64 `yaml:"variety_boost"`         // How much to boost variety
		ComplexityReduction float64 `yaml:"complexity_reduction"`  // How much to reduce complexity
		GenerationSpeed     float64 `yaml:"generation_speed"`      // Speed adjustment factor
	} `yaml:"adjustment_rates"`
	
	// Monitoring settings
	MonitoringInterval time.Duration `yaml:"monitoring_interval"` // How often to check quality
	MaxAdjustments     int           `yaml:"max_adjustments"`     // Max adjustments per session
}

// PCGEventManager manages PCG-specific events and runtime adjustments
type PCGEventManager struct {
	mu                    sync.RWMutex
	logger                *logrus.Logger
	eventSystem           *game.EventSystem
	adjustmentConfig      *RuntimeAdjustmentConfig
	pcgManager            *PCGManager
	adjustmentHistory     []AdjustmentRecord
	lastQualityCheck      time.Time
	adjustmentCount       int
	isMonitoring          bool
	monitoringStop        chan bool
}

// AdjustmentRecord tracks when and why adjustments were made
type AdjustmentRecord struct {
	Timestamp       time.Time                `json:"timestamp"`
	Trigger         string                   `json:"trigger"`           // What triggered the adjustment
	QualityBefore   float64                  `json:"quality_before"`    // Quality score before adjustment
	QualityAfter    float64                  `json:"quality_after"`     // Quality score after adjustment
	AdjustmentType  AdjustmentType           `json:"adjustment_type"`   // Type of adjustment made
	Parameters      map[string]interface{}   `json:"parameters"`        // Adjustment parameters
	Success         bool                     `json:"success"`           // Whether adjustment was successful
}

// AdjustmentType defines the type of runtime adjustment
type AdjustmentType string

const (
	AdjustmentTypeDifficulty  AdjustmentType = "difficulty"
	AdjustmentTypeVariety     AdjustmentType = "variety"
	AdjustmentTypeComplexity  AdjustmentType = "complexity"
	AdjustmentTypePerformance AdjustmentType = "performance"
)

// NewPCGEventManager creates a new PCG event manager with runtime adjustment capabilities
func NewPCGEventManager(logger *logrus.Logger, eventSystem *game.EventSystem, pcgManager *PCGManager) *PCGEventManager {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}
	
	if eventSystem == nil {
		eventSystem = game.NewEventSystem()
	}
	
	manager := &PCGEventManager{
		logger:            logger,
		eventSystem:       eventSystem,
		pcgManager:        pcgManager,
		adjustmentHistory: make([]AdjustmentRecord, 0),
		lastQualityCheck:  time.Now(),
		monitoringStop:    make(chan bool),
		adjustmentConfig:  DefaultRuntimeAdjustmentConfig(),
	}
	
	manager.registerEventHandlers()
	return manager
}

// DefaultRuntimeAdjustmentConfig returns default configuration for runtime adjustments
func DefaultRuntimeAdjustmentConfig() *RuntimeAdjustmentConfig {
	config := &RuntimeAdjustmentConfig{
		EnableRuntimeAdjustments: true,
		MonitoringInterval:       30 * time.Second,
		MaxAdjustments:          10,
	}
	
	// Set quality thresholds
	config.QualityThresholds.MinOverallScore = 0.7
	config.QualityThresholds.MinPerformance = 0.6
	config.QualityThresholds.MinVariety = 0.5
	config.QualityThresholds.MinConsistency = 0.7
	config.QualityThresholds.MinEngagement = 0.6
	config.QualityThresholds.MinStability = 0.8
	
	// Set adjustment rates
	config.AdjustmentRates.DifficultyStep = 0.1
	config.AdjustmentRates.VarietyBoost = 0.2
	config.AdjustmentRates.ComplexityReduction = 0.15
	config.AdjustmentRates.GenerationSpeed = 1.5
	
	return config
}

// SetAdjustmentConfig updates the runtime adjustment configuration
func (em *PCGEventManager) SetAdjustmentConfig(config *RuntimeAdjustmentConfig) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.adjustmentConfig = config
}

// GetAdjustmentConfig returns the current runtime adjustment configuration
func (em *PCGEventManager) GetAdjustmentConfig() *RuntimeAdjustmentConfig {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.adjustmentConfig
}

// StartMonitoring begins the runtime quality monitoring and adjustment system
func (em *PCGEventManager) StartMonitoring(ctx context.Context) {
	em.mu.Lock()
	if em.isMonitoring {
		em.mu.Unlock()
		return
	}
	em.isMonitoring = true
	em.mu.Unlock()
	
	go em.monitoringLoop(ctx)
	em.logger.Info("PCG runtime monitoring started")
}

// StopMonitoring stops the runtime quality monitoring system
func (em *PCGEventManager) StopMonitoring() {
	em.mu.Lock()
	if !em.isMonitoring {
		em.mu.Unlock()
		return
	}
	em.isMonitoring = false
	em.mu.Unlock()
	
	select {
	case em.monitoringStop <- true:
	default:
	}
	
	em.logger.Info("PCG runtime monitoring stopped")
}

// registerEventHandlers sets up event handlers for PCG events
func (em *PCGEventManager) registerEventHandlers() {
	// Handle content generation events
	em.eventSystem.Subscribe(EventPCGContentGenerated, em.handleContentGenerated)
	
	// Handle quality assessment events  
	em.eventSystem.Subscribe(EventPCGQualityAssessment, em.handleQualityAssessment)
	
	// Handle player feedback events
	em.eventSystem.Subscribe(EventPCGPlayerFeedback, em.handlePlayerFeedback)
	
	// Handle difficulty adjustment requests
	em.eventSystem.Subscribe(EventPCGDifficultyAdjustment, em.handleDifficultyAdjustment)
	
	// Handle content requests
	em.eventSystem.Subscribe(EventPCGContentRequest, em.handleContentRequest)
	
	// Handle system health events
	em.eventSystem.Subscribe(EventPCGSystemHealth, em.handleSystemHealth)
	
	em.logger.Debug("PCG event handlers registered")
}

// EmitContentGenerated emits an event when PCG content is generated
func (em *PCGEventManager) EmitContentGenerated(contentType ContentType, content interface{}, generationTime time.Duration, qualityScore float64) {
	eventData := PCGEventData{
		ContentType:    contentType,
		GenerationTime: generationTime,
		QualityScore:   qualityScore,
		Timestamp:      time.Now(),
	}
	
	event := game.GameEvent{
		Type:      EventPCGContentGenerated,
		SourceID:  "pcg_manager",
		TargetID:  string(contentType),
		Data:      map[string]interface{}{"pcg_data": eventData},
		Timestamp: time.Now().Unix(),
	}
	
	em.eventSystem.Emit(event)
}

// EmitQualityAssessment emits an event when content quality is assessed
func (em *PCGEventManager) EmitQualityAssessment(qualityReport *QualityReport) {
	eventData := PCGEventData{
		QualityScore: qualityReport.OverallScore,
		Timestamp:    time.Now(),
	}
	
	event := game.GameEvent{
		Type:      EventPCGQualityAssessment,
		SourceID:  "quality_metrics",
		TargetID:  "pcg_system",
		Data:      map[string]interface{}{"quality_report": qualityReport, "pcg_data": eventData},
		Timestamp: time.Now().Unix(),
	}
	
	em.eventSystem.Emit(event)
}

// EmitPlayerFeedback emits an event when player feedback is received
func (em *PCGEventManager) EmitPlayerFeedback(feedback *PlayerFeedback) {
	eventData := PCGEventData{
		PlayerFeedback: feedback,
		Timestamp:      time.Now(),
	}
	
	event := game.GameEvent{
		Type:      EventPCGPlayerFeedback,
		SourceID:  "player",
		TargetID:  "pcg_system",
		Data:      map[string]interface{}{"feedback": feedback, "pcg_data": eventData},
		Timestamp: time.Now().Unix(),
	}
	
	em.eventSystem.Emit(event)
}

// Event handler implementations
func (em *PCGEventManager) handleContentGenerated(event game.GameEvent) {
	pcgData, ok := event.Data["pcg_data"].(PCGEventData)
	if !ok {
		em.logger.Error("Invalid PCG data in content generated event")
		return
	}
	
	em.logger.WithFields(logrus.Fields{
		"content_type":    pcgData.ContentType,
		"generation_time": pcgData.GenerationTime,
		"quality_score":   pcgData.QualityScore,
	}).Debug("PCG content generated")
	
	// Check if quality is below threshold and adjustment is needed
	if em.adjustmentConfig.EnableRuntimeAdjustments && 
	   pcgData.QualityScore < em.adjustmentConfig.QualityThresholds.MinOverallScore {
		em.scheduleQualityAdjustment("low_content_quality", pcgData.QualityScore)
	}
}

func (em *PCGEventManager) handleQualityAssessment(event game.GameEvent) {
	qualityReport, ok := event.Data["quality_report"].(*QualityReport)
	if !ok {
		em.logger.Error("Invalid quality report in quality assessment event")
		return
	}
	
	em.logger.WithFields(logrus.Fields{
		"overall_score":    qualityReport.OverallScore,
		"quality_grade":    qualityReport.QualityGrade,
		"component_scores": qualityReport.ComponentScores,
	}).Debug("Quality assessment received")
	
	// Check if adjustments are needed based on quality thresholds
	if em.adjustmentConfig.EnableRuntimeAdjustments {
		em.evaluateQualityThresholds(qualityReport)
	}
}

func (em *PCGEventManager) handlePlayerFeedback(event game.GameEvent) {
	feedback, ok := event.Data["feedback"].(*PlayerFeedback)
	if !ok {
		em.logger.Error("Invalid player feedback in feedback event")
		return
	}
	
	em.logger.WithFields(logrus.Fields{
		"rating":     feedback.Rating,
		"enjoyment":  feedback.Enjoyment,
		"difficulty": feedback.Difficulty,
	}).Debug("Player feedback received")
	
	// Adjust based on player feedback
	if em.adjustmentConfig.EnableRuntimeAdjustments {
		em.adjustBasedOnFeedback(feedback)
	}
}

func (em *PCGEventManager) handleDifficultyAdjustment(event game.GameEvent) {
	em.logger.Debug("Difficulty adjustment event received")
	
	adjustmentParams, ok := event.Data["adjustment_params"].(map[string]interface{})
	if !ok {
		em.logger.Error("Invalid adjustment parameters in difficulty adjustment event")
		return
	}
	
	em.applyDifficultyAdjustment(adjustmentParams)
}

func (em *PCGEventManager) handleContentRequest(event game.GameEvent) {
	em.logger.Debug("Content request event received")
	
	// Handle dynamic content requests based on gameplay needs
	contentType, ok := event.Data["content_type"].(ContentType)
	if !ok {
		em.logger.Error("Invalid content type in content request event")
		return
	}
	
	em.handleDynamicContentRequest(contentType, event.Data)
}

func (em *PCGEventManager) handleSystemHealth(event game.GameEvent) {
	em.logger.Debug("System health event received")
	
	// Monitor system health and adjust generation parameters if needed
	healthData, ok := event.Data["health_data"].(map[string]interface{})
	if !ok {
		em.logger.Error("Invalid health data in system health event")
		return
	}
	
	em.monitorSystemHealth(healthData)
}

// Monitoring and adjustment implementation
func (em *PCGEventManager) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(em.adjustmentConfig.MonitoringInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-em.monitoringStop:
			return
		case <-ticker.C:
			em.performQualityCheck()
		}
	}
}

func (em *PCGEventManager) performQualityCheck() {
	if em.pcgManager == nil {
		return
	}
	
	// Generate quality report
	report := em.pcgManager.GenerateQualityReport()
	if report == nil {
		em.logger.Error("Failed to generate quality report during monitoring")
		return
	}
	
	// Emit quality assessment event
	em.EmitQualityAssessment(report)
	
	em.mu.Lock()
	em.lastQualityCheck = time.Now()
	em.mu.Unlock()
}

func (em *PCGEventManager) evaluateQualityThresholds(report *QualityReport) {
	thresholds := em.adjustmentConfig.QualityThresholds
	
	// Check overall score
	if report.OverallScore < thresholds.MinOverallScore {
		em.scheduleQualityAdjustment("low_overall_quality", report.OverallScore)
		return
	}
	
	// Check individual component scores
	if score, exists := report.ComponentScores["performance"]; exists && score < thresholds.MinPerformance {
		em.schedulePerformanceAdjustment("low_performance", score)
	}
	
	if score, exists := report.ComponentScores["variety"]; exists && score < thresholds.MinVariety {
		em.scheduleVarietyAdjustment("low_variety", score)
	}
	
	if score, exists := report.ComponentScores["consistency"]; exists && score < thresholds.MinConsistency {
		em.scheduleConsistencyAdjustment("low_consistency", score)
	}
	
	if score, exists := report.ComponentScores["engagement"]; exists && score < thresholds.MinEngagement {
		em.scheduleEngagementAdjustment("low_engagement", score)
	}
	
	if score, exists := report.ComponentScores["stability"]; exists && score < thresholds.MinStability {
		em.scheduleStabilityAdjustment("low_stability", score)
	}
}

func (em *PCGEventManager) scheduleQualityAdjustment(trigger string, currentScore float64) {
	if em.adjustmentCount >= em.adjustmentConfig.MaxAdjustments {
		em.logger.Warn("Maximum adjustments reached, skipping quality adjustment")
		return
	}
	
	em.logger.WithFields(logrus.Fields{
		"trigger":       trigger,
		"current_score": currentScore,
	}).Info("Scheduling quality adjustment")
	
	// Implement adjustment logic based on trigger type
	adjustmentParams := map[string]interface{}{
		"trigger": trigger,
		"score":   currentScore,
		"type":    "general_quality",
	}
	
	em.applyGeneralQualityAdjustment(adjustmentParams)
}

func (em *PCGEventManager) schedulePerformanceAdjustment(trigger string, score float64) {
	adjustmentParams := map[string]interface{}{
		"trigger": trigger,
		"score":   score,
		"type":    "performance",
	}
	em.applyPerformanceAdjustment(adjustmentParams)
}

func (em *PCGEventManager) scheduleVarietyAdjustment(trigger string, score float64) {
	adjustmentParams := map[string]interface{}{
		"trigger": trigger,
		"score":   score,
		"type":    "variety",
	}
	em.applyVarietyAdjustment(adjustmentParams)
}

func (em *PCGEventManager) scheduleConsistencyAdjustment(trigger string, score float64) {
	adjustmentParams := map[string]interface{}{
		"trigger": trigger,
		"score":   score,
		"type":    "consistency",
	}
	em.applyConsistencyAdjustment(adjustmentParams)
}

func (em *PCGEventManager) scheduleEngagementAdjustment(trigger string, score float64) {
	adjustmentParams := map[string]interface{}{
		"trigger": trigger,
		"score":   score,
		"type":    "engagement",
	}
	em.applyEngagementAdjustment(adjustmentParams)
}

func (em *PCGEventManager) scheduleStabilityAdjustment(trigger string, score float64) {
	adjustmentParams := map[string]interface{}{
		"trigger": trigger,
		"score":   score,
		"type":    "stability",
	}
	em.applyStabilityAdjustment(adjustmentParams)
}

// Adjustment implementation methods
func (em *PCGEventManager) applyGeneralQualityAdjustment(params map[string]interface{}) {
	em.logger.Info("Applying general quality adjustment")
	
	// Record the adjustment
	record := AdjustmentRecord{
		Timestamp:      time.Now(),
		Trigger:        params["trigger"].(string),
		QualityBefore:  params["score"].(float64),
		AdjustmentType: AdjustmentTypeComplexity,
		Parameters:     params,
	}
	
	// Apply the adjustment to PCG Manager (implementation depends on specific needs)
	success := em.adjustPCGParameters(params)
	record.Success = success
	
	if success {
		em.mu.Lock()
		em.adjustmentCount++
		em.adjustmentHistory = append(em.adjustmentHistory, record)
		em.mu.Unlock()
		
		em.logger.Info("General quality adjustment applied successfully")
	} else {
		em.logger.Error("Failed to apply general quality adjustment")
	}
}

func (em *PCGEventManager) applyPerformanceAdjustment(params map[string]interface{}) {
	em.logger.Info("Applying performance adjustment")
	// Implement performance-specific adjustments
	// e.g., reduce generation complexity, increase caching, etc.
	em.recordAdjustment(AdjustmentTypePerformance, params, true)
}

func (em *PCGEventManager) applyVarietyAdjustment(params map[string]interface{}) {
	em.logger.Info("Applying variety adjustment")
	// Implement variety-specific adjustments  
	// e.g., increase randomness, expand content pools, etc.
	em.recordAdjustment(AdjustmentTypeVariety, params, true)
}

func (em *PCGEventManager) applyConsistencyAdjustment(params map[string]interface{}) {
	em.logger.Info("Applying consistency adjustment")
	// Implement consistency-specific adjustments
	// e.g., strengthen validation rules, improve coherence checks, etc.
	em.recordAdjustment(AdjustmentTypeComplexity, params, true)
}

func (em *PCGEventManager) applyEngagementAdjustment(params map[string]interface{}) {
	em.logger.Info("Applying engagement adjustment")
	// Implement engagement-specific adjustments
	// e.g., adjust difficulty curves, add more dynamic content, etc.
	em.recordAdjustment(AdjustmentTypeDifficulty, params, true)
}

func (em *PCGEventManager) applyStabilityAdjustment(params map[string]interface{}) {
	em.logger.Info("Applying stability adjustment")
	// Implement stability-specific adjustments
	// e.g., add error handling, improve fallback mechanisms, etc.
	em.recordAdjustment(AdjustmentTypePerformance, params, true)
}

func (em *PCGEventManager) applyDifficultyAdjustment(params map[string]interface{}) {
	em.logger.Info("Applying difficulty adjustment")
	// Implement difficulty-specific adjustments
	em.recordAdjustment(AdjustmentTypeDifficulty, params, true)
}

func (em *PCGEventManager) adjustBasedOnFeedback(feedback *PlayerFeedback) {
	// Adjust generation parameters based on player feedback
	if feedback.Difficulty < 3 { // Too easy
		params := map[string]interface{}{
			"trigger":         "player_feedback_easy",
			"difficulty_up":   em.adjustmentConfig.AdjustmentRates.DifficultyStep,
			"feedback_rating": feedback.Difficulty,
		}
		em.applyDifficultyAdjustment(params)
	} else if feedback.Difficulty > 7 { // Too hard
		params := map[string]interface{}{
			"trigger":           "player_feedback_hard",
			"difficulty_down":   em.adjustmentConfig.AdjustmentRates.DifficultyStep,
			"feedback_rating":   feedback.Difficulty,
		}
		em.applyDifficultyAdjustment(params)
	}
	
	if feedback.Enjoyment < 4 { // Low enjoyment
		params := map[string]interface{}{
			"trigger":          "player_feedback_low_enjoyment",
			"variety_boost":    em.adjustmentConfig.AdjustmentRates.VarietyBoost,
			"feedback_rating":  feedback.Enjoyment,
		}
		em.applyVarietyAdjustment(params)
	}
}

func (em *PCGEventManager) handleDynamicContentRequest(contentType ContentType, eventData map[string]interface{}) {
	em.logger.WithField("content_type", contentType).Info("Handling dynamic content request")
	
	// Implement dynamic content generation based on real-time needs
	// This could trigger immediate content generation with adjusted parameters
	params := map[string]interface{}{
		"trigger":      "dynamic_request",
		"content_type": contentType,
		"urgency":      eventData["urgency"],
	}
	
	em.recordAdjustment(AdjustmentTypePerformance, params, true)
}

func (em *PCGEventManager) monitorSystemHealth(healthData map[string]interface{}) {
	em.logger.Info("Monitoring system health")
	
	// Monitor system metrics and adjust if needed
	if memoryUsage, ok := healthData["memory_usage"].(float64); ok && memoryUsage > 0.8 {
		params := map[string]interface{}{
			"trigger":       "high_memory_usage",
			"memory_usage":  memoryUsage,
		}
		em.applyPerformanceAdjustment(params)
	}
	
	if errorRate, ok := healthData["error_rate"].(float64); ok && errorRate > 0.05 {
		params := map[string]interface{}{
			"trigger":    "high_error_rate",
			"error_rate": errorRate,
		}
		em.applyStabilityAdjustment(params)
	}
}

// Helper methods
func (em *PCGEventManager) adjustPCGParameters(params map[string]interface{}) bool {
	// This method would interface with the PCGManager to adjust generation parameters
	// Implementation would depend on specific parameter adjustment needs
	em.logger.WithField("params", params).Debug("Adjusting PCG parameters")
	return true // Placeholder - actual implementation would vary
}

func (em *PCGEventManager) recordAdjustment(adjustmentType AdjustmentType, params map[string]interface{}, success bool) {
	record := AdjustmentRecord{
		Timestamp:      time.Now(),
		Trigger:        params["trigger"].(string),
		AdjustmentType: adjustmentType,
		Parameters:     params,
		Success:        success,
	}
	
	em.mu.Lock()
	if success {
		em.adjustmentCount++
	}
	em.adjustmentHistory = append(em.adjustmentHistory, record)
	em.mu.Unlock()
}

// GetAdjustmentHistory returns the history of runtime adjustments
func (em *PCGEventManager) GetAdjustmentHistory() []AdjustmentRecord {
	em.mu.RLock()
	defer em.mu.RUnlock()
	
	history := make([]AdjustmentRecord, len(em.adjustmentHistory))
	copy(history, em.adjustmentHistory)
	return history
}

// GetAdjustmentCount returns the current number of adjustments made
func (em *PCGEventManager) GetAdjustmentCount() int {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.adjustmentCount
}

// ResetAdjustmentCount resets the adjustment counter
func (em *PCGEventManager) ResetAdjustmentCount() {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.adjustmentCount = 0
}

// IsMonitoring returns whether the system is currently monitoring
func (em *PCGEventManager) IsMonitoring() bool {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.isMonitoring
}
