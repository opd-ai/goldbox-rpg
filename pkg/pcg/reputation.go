package pcg

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ReputationSystem manages player standing with factions
// Tracks reputation scores, provides modification methods, and calculates effects
// on gameplay elements like dialogue, prices, and quest availability
type ReputationSystem struct {
	mu                 sync.RWMutex                  `yaml:"-"`
	PlayerReputations  map[string]*PlayerReputation  `yaml:"player_reputations"`  // Player ID -> Reputation data
	FactionReputations map[string]*FactionReputation `yaml:"faction_reputations"` // Faction ID -> Reputation settings
	ReputationHistory  []*ReputationEvent            `yaml:"reputation_history"`  // Historical reputation changes
	ReputationEffects  map[string]*ReputationEffect  `yaml:"reputation_effects"`  // Active reputation-based effects
	GlobalModifiers    map[string]float64            `yaml:"global_modifiers"`    // Global reputation modifiers
	DecaySettings      *ReputationDecaySettings      `yaml:"decay_settings"`      // Reputation decay configuration
	logger             *logrus.Logger                `yaml:"-"`
}

// PlayerReputation tracks a player's standing with all factions
type PlayerReputation struct {
	PlayerID         string                      `yaml:"player_id"`
	FactionStandings map[string]*FactionStanding `yaml:"faction_standings"` // Faction ID -> Standing
	TotalReputation  int64                       `yaml:"total_reputation"`  // Sum of all faction standings
	LastUpdated      time.Time                   `yaml:"last_updated"`
	ReputationRank   ReputationRank              `yaml:"reputation_rank"` // Overall reputation tier
	Properties       map[string]interface{}      `yaml:"properties"`
}

// FactionStanding represents a player's standing with a specific faction
type FactionStanding struct {
	FactionID       string                 `yaml:"faction_id"`
	ReputationScore int64                  `yaml:"reputation_score"` // Raw reputation points (-10000 to +10000)
	ReputationLevel ReputationLevel        `yaml:"reputation_level"` // Derived reputation tier
	FirstContact    time.Time              `yaml:"first_contact"`    // When player first encountered faction
	LastInteraction time.Time              `yaml:"last_interaction"` // Most recent reputation change
	ActionCount     int                    `yaml:"action_count"`     // Number of reputation-affecting actions
	MaxReached      int64                  `yaml:"max_reached"`      // Highest reputation ever achieved
	MinReached      int64                  `yaml:"min_reached"`      // Lowest reputation ever reached
	IsLocked        bool                   `yaml:"is_locked"`        // Whether reputation changes are disabled
	LockReason      string                 `yaml:"lock_reason"`      // Why reputation is locked (if applicable)
	Properties      map[string]interface{} `yaml:"properties"`
}

// FactionReputation defines reputation settings for a faction
type FactionReputation struct {
	FactionID       string                 `yaml:"faction_id"`
	Name            string                 `yaml:"name"`
	BaseAttitude    int64                  `yaml:"base_attitude"`    // Starting reputation for new players
	AttitudeRange   *ReputationRange       `yaml:"attitude_range"`   // Min/max reputation limits
	DecayRate       float64                `yaml:"decay_rate"`       // How fast reputation decays over time
	GainMultiplier  float64                `yaml:"gain_multiplier"`  // Modifier for positive reputation gains
	LossMultiplier  float64                `yaml:"loss_multiplier"`  // Modifier for negative reputation losses
	UnlockThreshold int64                  `yaml:"unlock_threshold"` // Reputation needed to unlock special content
	QuestMultiplier float64                `yaml:"quest_multiplier"` // Modifier for quest-based reputation changes
	MemoryDuration  time.Duration          `yaml:"memory_duration"`  // How long faction remembers actions
	AlliedFactions  []string               `yaml:"allied_factions"`  // Factions that share reputation changes
	EnemyFactions   []string               `yaml:"enemy_factions"`   // Factions with inverse reputation changes
	Properties      map[string]interface{} `yaml:"properties"`
}

// ReputationEvent records a historical reputation change
type ReputationEvent struct {
	ID            string                 `yaml:"id"`
	PlayerID      string                 `yaml:"player_id"`
	FactionID     string                 `yaml:"faction_id"`
	Change        int64                  `yaml:"change"`      // Reputation point change
	Reason        string                 `yaml:"reason"`      // Why reputation changed
	ActionType    ReputationActionType   `yaml:"action_type"` // Category of action
	Location      string                 `yaml:"location"`    // Where the action occurred
	QuestID       string                 `yaml:"quest_id"`    // Associated quest (if applicable)
	Timestamp     time.Time              `yaml:"timestamp"`
	PreviousScore int64                  `yaml:"previous_score"` // Reputation before change
	NewScore      int64                  `yaml:"new_score"`      // Reputation after change
	PreviousLevel ReputationLevel        `yaml:"previous_level"` // Level before change
	NewLevel      ReputationLevel        `yaml:"new_level"`      // Level after change
	Properties    map[string]interface{} `yaml:"properties"`
}

// ReputationEffect represents an active effect based on reputation level
type ReputationEffect struct {
	ID             string                 `yaml:"id"`
	PlayerID       string                 `yaml:"player_id"`
	FactionID      string                 `yaml:"faction_id"`
	EffectType     ReputationEffectType   `yaml:"effect_type"`
	Magnitude      float64                `yaml:"magnitude"`   // Strength of the effect
	Description    string                 `yaml:"description"` // Human-readable description
	IsActive       bool                   `yaml:"is_active"`
	StartTime      time.Time              `yaml:"start_time"`
	ExpirationTime *time.Time             `yaml:"expiration_time"` // Nil for permanent effects
	Properties     map[string]interface{} `yaml:"properties"`
}

// ReputationDecaySettings controls how reputation changes over time
type ReputationDecaySettings struct {
	EnableDecay        bool          `yaml:"enable_decay"`        // Whether reputation decays over time
	DecayInterval      time.Duration `yaml:"decay_interval"`      // How often decay is applied
	BaseDecayRate      float64       `yaml:"base_decay_rate"`     // Base percentage of reputation lost per interval
	MinDecayThreshold  int64         `yaml:"min_decay_threshold"` // Reputation won't decay below this value
	MaxDecayThreshold  int64         `yaml:"max_decay_threshold"` // Reputation won't decay above this value
	PositiveDecayRate  float64       `yaml:"positive_decay_rate"` // Decay rate for positive reputation
	NegativeDecayRate  float64       `yaml:"negative_decay_rate"` // Decay rate for negative reputation
	ActivityProtection time.Duration `yaml:"activity_protection"` // No decay for this period after activity
}

// ReputationRange defines min/max reputation limits
type ReputationRange struct {
	Min int64 `yaml:"min"`
	Max int64 `yaml:"max"`
}

// NewReputationSystem creates a new reputation management system
func NewReputationSystem(logger *logrus.Logger) *ReputationSystem {
	if logger == nil {
		logger = logrus.New()
	}

	return &ReputationSystem{
		PlayerReputations:  make(map[string]*PlayerReputation),
		FactionReputations: make(map[string]*FactionReputation),
		ReputationHistory:  make([]*ReputationEvent, 0),
		ReputationEffects:  make(map[string]*ReputationEffect),
		GlobalModifiers:    make(map[string]float64),
		DecaySettings: &ReputationDecaySettings{
			EnableDecay:        true,
			DecayInterval:      24 * time.Hour, // Daily decay
			BaseDecayRate:      0.01,           // 1% decay per day
			MinDecayThreshold:  -5000,          // Don't decay below -5000
			MaxDecayThreshold:  5000,           // Don't decay above +5000
			PositiveDecayRate:  0.005,          // Positive reputation decays slower
			NegativeDecayRate:  0.02,           // Negative reputation decays faster
			ActivityProtection: 72 * time.Hour, // 3 days protection after activity
		},
		logger: logger,
	}
}

// InitializePlayerReputation creates initial reputation data for a new player
func (rs *ReputationSystem) InitializePlayerReputation(playerID string, factionSystem *GeneratedFactionSystem) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if _, exists := rs.PlayerReputations[playerID]; exists {
		return fmt.Errorf("player reputation already exists: %s", playerID)
	}

	playerRep := &PlayerReputation{
		PlayerID:         playerID,
		FactionStandings: make(map[string]*FactionStanding),
		TotalReputation:  0,
		LastUpdated:      time.Now(),
		ReputationRank:   ReputationRankNeutral,
		Properties:       make(map[string]interface{}),
	}

	// Initialize standing with each faction
	for _, faction := range factionSystem.Factions {
		factionRep := rs.getFactionReputation(faction.ID)

		standing := &FactionStanding{
			FactionID:       faction.ID,
			ReputationScore: factionRep.BaseAttitude,
			ReputationLevel: rs.calculateReputationLevel(factionRep.BaseAttitude),
			FirstContact:    time.Now(),
			LastInteraction: time.Now(),
			ActionCount:     0,
			MaxReached:      factionRep.BaseAttitude,
			MinReached:      factionRep.BaseAttitude,
			IsLocked:        false,
			Properties:      make(map[string]interface{}),
		}

		playerRep.FactionStandings[faction.ID] = standing
		playerRep.TotalReputation += factionRep.BaseAttitude
	}

	playerRep.ReputationRank = rs.calculateOverallRank(playerRep.TotalReputation, len(factionSystem.Factions))
	rs.PlayerReputations[playerID] = playerRep

	rs.logger.WithFields(logrus.Fields{
		"player_id":        playerID,
		"faction_count":    len(factionSystem.Factions),
		"total_reputation": playerRep.TotalReputation,
		"reputation_rank":  playerRep.ReputationRank,
	}).Info("initialized player reputation")

	return nil
}

// ModifyReputation changes a player's reputation with a faction
func (rs *ReputationSystem) ModifyReputation(playerID, factionID string, change int64, reason string, actionType ReputationActionType) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	playerRep, standing, err := rs.validateReputationAccess(playerID, factionID)
	if err != nil {
		return err
	}

	finalChange := rs.calculateFinalChange(factionID, change)
	previousScore := standing.ReputationScore
	previousLevel := standing.ReputationLevel

	newScore := rs.applyScoreChange(factionID, standing, finalChange)
	rs.updateStanding(standing, newScore, playerRep, previousScore)

	event := rs.createReputationEvent(playerID, factionID, finalChange, reason, actionType, previousScore, newScore, previousLevel, standing.ReputationLevel)
	rs.ReputationHistory = append(rs.ReputationHistory, event)

	rs.applyFactionInfluenceUnsafe(playerID, factionID, finalChange)
	rs.updateReputationEffectsUnsafe(playerID, factionID)

	rs.logger.WithFields(logrus.Fields{
		"player_id":   playerID,
		"faction_id":  factionID,
		"change":      finalChange,
		"new_score":   newScore,
		"new_level":   standing.ReputationLevel,
		"action_type": actionType,
		"reason":      reason,
	}).Info("reputation modified")

	return nil
}

// validateReputationAccess checks player and faction existence and lock status
func (rs *ReputationSystem) validateReputationAccess(playerID, factionID string) (*PlayerReputation, *FactionStanding, error) {
	playerRep, exists := rs.PlayerReputations[playerID]
	if !exists {
		return nil, nil, fmt.Errorf("player reputation not found: %s", playerID)
	}

	standing, exists := playerRep.FactionStandings[factionID]
	if !exists {
		return nil, nil, fmt.Errorf("faction standing not found: player=%s, faction=%s", playerID, factionID)
	}

	if standing.IsLocked {
		rs.logger.WithFields(logrus.Fields{
			"player_id":   playerID,
			"faction_id":  factionID,
			"lock_reason": standing.LockReason,
		}).Warn("attempted to modify locked reputation")
		return nil, nil, fmt.Errorf("reputation locked: %s", standing.LockReason)
	}

	return playerRep, standing, nil
}

// calculateFinalChange applies faction-specific and global modifiers to reputation change
func (rs *ReputationSystem) calculateFinalChange(factionID string, change int64) int64 {
	factionRep := rs.getFactionReputation(factionID)
	finalChange := change

	if change > 0 {
		finalChange = int64(float64(change) * factionRep.GainMultiplier)
	} else {
		finalChange = int64(float64(change) * factionRep.LossMultiplier)
	}

	if globalMod, exists := rs.GlobalModifiers["reputation_gain"]; exists && change > 0 {
		finalChange = int64(float64(finalChange) * globalMod)
	}
	if globalMod, exists := rs.GlobalModifiers["reputation_loss"]; exists && change < 0 {
		finalChange = int64(float64(finalChange) * globalMod)
	}

	return finalChange
}

// applyScoreChange applies reputation change with range constraints
func (rs *ReputationSystem) applyScoreChange(factionID string, standing *FactionStanding, finalChange int64) int64 {
	factionRep := rs.getFactionReputation(factionID)
	newScore := standing.ReputationScore + finalChange

	if factionRep.AttitudeRange != nil {
		if newScore < factionRep.AttitudeRange.Min {
			newScore = factionRep.AttitudeRange.Min
		}
		if newScore > factionRep.AttitudeRange.Max {
			newScore = factionRep.AttitudeRange.Max
		}
	} else {
		if newScore < -10000 {
			newScore = -10000
		}
		if newScore > 10000 {
			newScore = 10000
		}
	}

	return newScore
}

// updateStanding updates faction standing and player reputation with new values
func (rs *ReputationSystem) updateStanding(standing *FactionStanding, newScore int64, playerRep *PlayerReputation, previousScore int64) {
	standing.ReputationScore = newScore
	standing.ReputationLevel = rs.calculateReputationLevel(newScore)
	standing.LastInteraction = time.Now()
	standing.ActionCount++

	if newScore > standing.MaxReached {
		standing.MaxReached = newScore
	}
	if newScore < standing.MinReached {
		standing.MinReached = newScore
	}

	playerRep.TotalReputation += (newScore - previousScore)
	playerRep.LastUpdated = time.Now()
	playerRep.ReputationRank = rs.calculateOverallRank(playerRep.TotalReputation, len(playerRep.FactionStandings))
}

// createReputationEvent creates a reputation event record
func (rs *ReputationSystem) createReputationEvent(playerID, factionID string, finalChange int64, reason string, actionType ReputationActionType, previousScore, newScore int64, previousLevel, newLevel ReputationLevel) *ReputationEvent {
	return &ReputationEvent{
		ID:            fmt.Sprintf("rep_%d_%s_%s", time.Now().UnixNano(), playerID, factionID),
		PlayerID:      playerID,
		FactionID:     factionID,
		Change:        finalChange,
		Reason:        reason,
		ActionType:    actionType,
		Timestamp:     time.Now(),
		PreviousScore: previousScore,
		NewScore:      newScore,
		PreviousLevel: previousLevel,
		NewLevel:      newLevel,
		Properties:    make(map[string]interface{}),
	}
}

// GetReputation returns a player's current reputation with a faction
func (rs *ReputationSystem) GetReputation(playerID, factionID string) (*FactionStanding, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	playerRep, exists := rs.PlayerReputations[playerID]
	if !exists {
		return nil, fmt.Errorf("player reputation not found: %s", playerID)
	}

	standing, exists := playerRep.FactionStandings[factionID]
	if !exists {
		return nil, fmt.Errorf("faction standing not found: player=%s, faction=%s", playerID, factionID)
	}

	// Return a copy to prevent external modification
	standingCopy := *standing
	return &standingCopy, nil
}

// GetPlayerReputation returns all reputation data for a player
func (rs *ReputationSystem) GetPlayerReputation(playerID string) (*PlayerReputation, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	playerRep, exists := rs.PlayerReputations[playerID]
	if !exists {
		return nil, fmt.Errorf("player reputation not found: %s", playerID)
	}

	// Return a deep copy to prevent external modification
	repCopy := &PlayerReputation{
		PlayerID:         playerRep.PlayerID,
		FactionStandings: make(map[string]*FactionStanding),
		TotalReputation:  playerRep.TotalReputation,
		LastUpdated:      playerRep.LastUpdated,
		ReputationRank:   playerRep.ReputationRank,
		Properties:       make(map[string]interface{}),
	}

	for id, standing := range playerRep.FactionStandings {
		standingCopy := *standing
		repCopy.FactionStandings[id] = &standingCopy
	}

	for key, value := range playerRep.Properties {
		repCopy.Properties[key] = value
	}

	return repCopy, nil
}

// CalculateEffect calculates the magnitude of a reputation effect
func (rs *ReputationSystem) CalculateEffect(playerID, factionID string, effectType ReputationEffectType) (float64, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	return rs.calculateEffectUnsafe(playerID, factionID, effectType)
}

// calculateEffectUnsafe calculates effect without acquiring locks (must be called with lock held)
func (rs *ReputationSystem) calculateEffectUnsafe(playerID, factionID string, effectType ReputationEffectType) (float64, error) {
	playerRep, exists := rs.PlayerReputations[playerID]
	if !exists {
		return 0.0, fmt.Errorf("player reputation not found: %s", playerID)
	}

	standing, exists := playerRep.FactionStandings[factionID]
	if !exists {
		return 0.0, fmt.Errorf("faction standing not found: player=%s, faction=%s", playerID, factionID)
	}

	// Base effect calculation based on reputation level
	var baseEffect float64
	switch standing.ReputationLevel {
	case ReputationLevelRevered:
		baseEffect = 0.5
	case ReputationLevelExalted:
		baseEffect = 0.4
	case ReputationLevelHonored:
		baseEffect = 0.3
	case ReputationLevelFriendly:
		baseEffect = 0.15
	case ReputationLevelNeutral:
		baseEffect = 0.0
	case ReputationLevelUnfriendly:
		baseEffect = -0.15
	case ReputationLevelHostile:
		baseEffect = -0.3
	case ReputationLevelHated:
		baseEffect = -0.4
	case ReputationLevelDespised:
		baseEffect = -0.5
	default:
		baseEffect = 0.0
	}

	// Apply effect type modifiers
	switch effectType {
	case ReputationEffectPriceDiscount, ReputationEffectPricePenalty:
		// Price effects are more gradual
		baseEffect *= 0.5
	case ReputationEffectQuestReward:
		// Quest rewards scale more dramatically
		baseEffect *= 1.5
	case ReputationEffectCombatAssistance, ReputationEffectCombatHostility:
		// Combat effects are binary but based on threshold
		if baseEffect > 0.2 {
			baseEffect = 1.0
		} else if baseEffect < -0.2 {
			baseEffect = -1.0
		} else {
			baseEffect = 0.0
		}
	}

	// Apply global modifiers if they exist
	if modifier, exists := rs.GlobalModifiers[string(effectType)]; exists {
		baseEffect *= modifier
	}

	return baseEffect, nil
}

// GetReputationHistory returns reputation events for a player and faction
func (rs *ReputationSystem) GetReputationHistory(playerID, factionID string, limit int) ([]*ReputationEvent, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	events := make([]*ReputationEvent, 0)
	count := 0

	// Iterate backwards through history to get most recent events first
	for i := len(rs.ReputationHistory) - 1; i >= 0 && count < limit; i-- {
		event := rs.ReputationHistory[i]
		if event.PlayerID == playerID && (factionID == "" || event.FactionID == factionID) {
			// Return a copy to prevent external modification
			eventCopy := *event
			events = append(events, &eventCopy)
			count++
		}
	}

	return events, nil
}

// ApplyDecay applies time-based reputation decay according to settings
func (rs *ReputationSystem) ApplyDecay() error {
	if rs.DecaySettings == nil || !rs.DecaySettings.EnableDecay {
		return nil
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()

	now := time.Now()
	decayApplied := 0

	for playerID, playerRep := range rs.PlayerReputations {
		decayApplied += rs.applyDecayToPlayer(playerID, playerRep, now)
		playerRep.ReputationRank = rs.calculateOverallRank(playerRep.TotalReputation, len(playerRep.FactionStandings))
		playerRep.LastUpdated = now
	}

	if decayApplied > 0 {
		rs.logger.WithFields(logrus.Fields{
			"decay_applied": decayApplied,
		}).Info("reputation decay applied")
	}

	return nil
}

// applyDecayToPlayer processes decay for all faction standings of a single player.
func (rs *ReputationSystem) applyDecayToPlayer(playerID string, playerRep *PlayerReputation, now time.Time) int {
	count := 0
	for factionID, standing := range playerRep.FactionStandings {
		if rs.applyDecayToStanding(playerID, factionID, standing, playerRep, now) {
			count++
		}
	}
	return count
}

// applyDecayToStanding applies decay to a single faction standing if eligible.
func (rs *ReputationSystem) applyDecayToStanding(playerID, factionID string, standing *FactionStanding, playerRep *PlayerReputation, now time.Time) bool {
	if !rs.isEligibleForDecay(standing, now) {
		return false
	}

	decayAmount := rs.calculateDecayAmount(standing)
	if decayAmount == 0 {
		return false
	}

	previousScore := standing.ReputationScore
	previousLevel := standing.ReputationLevel
	standing.ReputationScore -= decayAmount
	standing.ReputationLevel = rs.calculateReputationLevel(standing.ReputationScore)

	playerRep.TotalReputation += (standing.ReputationScore - previousScore)

	rs.recordDecayEvent(playerID, factionID, decayAmount, previousScore, standing.ReputationScore, previousLevel, standing.ReputationLevel, now)
	return true
}

// isEligibleForDecay checks if a standing meets all criteria for decay application.
func (rs *ReputationSystem) isEligibleForDecay(standing *FactionStanding, now time.Time) bool {
	timeSinceActivity := now.Sub(standing.LastInteraction)
	if timeSinceActivity < rs.DecaySettings.ActivityProtection {
		return false
	}
	return standing.ReputationScore >= rs.DecaySettings.MinDecayThreshold &&
		standing.ReputationScore <= rs.DecaySettings.MaxDecayThreshold
}

// calculateDecayAmount determines the decay value based on current score and decay rate.
func (rs *ReputationSystem) calculateDecayAmount(standing *FactionStanding) int64 {
	decayRate := rs.DecaySettings.NegativeDecayRate
	if standing.ReputationScore > 0 {
		decayRate = rs.DecaySettings.PositiveDecayRate
	}
	return int64(float64(standing.ReputationScore) * decayRate)
}

// recordDecayEvent creates and stores a reputation event for the decay.
func (rs *ReputationSystem) recordDecayEvent(playerID, factionID string, decayAmount, previousScore, newScore int64, previousLevel, newLevel ReputationLevel, now time.Time) {
	rs.ReputationHistory = append(rs.ReputationHistory, &ReputationEvent{
		ID:            fmt.Sprintf("decay_%d_%s_%s", now.UnixNano(), playerID, factionID),
		PlayerID:      playerID,
		FactionID:     factionID,
		Change:        -decayAmount,
		Reason:        "time decay",
		ActionType:    "decay",
		Timestamp:     now,
		PreviousScore: previousScore,
		NewScore:      newScore,
		PreviousLevel: previousLevel,
		NewLevel:      newLevel,
		Properties:    make(map[string]interface{}),
	})
}

// Helper methods

// getFactionReputation returns faction reputation settings, creating defaults if needed
func (rs *ReputationSystem) getFactionReputation(factionID string) *FactionReputation {
	if factionRep, exists := rs.FactionReputations[factionID]; exists {
		return factionRep
	}

	// Create default faction reputation settings
	baseDecayRate := 0.01
	if rs.DecaySettings != nil {
		baseDecayRate = rs.DecaySettings.BaseDecayRate
	}
	defaultRep := &FactionReputation{
		FactionID:       factionID,
		Name:            factionID,
		BaseAttitude:    0,
		AttitudeRange:   &ReputationRange{Min: -10000, Max: 10000},
		DecayRate:       baseDecayRate,
		GainMultiplier:  1.0,
		LossMultiplier:  1.0,
		UnlockThreshold: 2500,
		QuestMultiplier: 1.0,
		MemoryDuration:  30 * 24 * time.Hour, // 30 days
		AlliedFactions:  make([]string, 0),
		EnemyFactions:   make([]string, 0),
		Properties:      make(map[string]interface{}),
	}

	rs.FactionReputations[factionID] = defaultRep
	return defaultRep
}

// calculateReputationLevel determines reputation tier from raw score
func (rs *ReputationSystem) calculateReputationLevel(score int64) ReputationLevel {
	switch {
	case score >= 7501:
		return ReputationLevelRevered
	case score >= 5001:
		return ReputationLevelExalted
	case score >= 2501:
		return ReputationLevelHonored
	case score >= 501:
		return ReputationLevelFriendly
	case score >= -500:
		return ReputationLevelNeutral
	case score >= -2500:
		return ReputationLevelUnfriendly
	case score >= -5000:
		return ReputationLevelHostile
	case score >= -7500:
		return ReputationLevelHated
	default:
		return ReputationLevelDespised
	}
}

// calculateOverallRank determines overall reputation rank based on total score
func (rs *ReputationSystem) calculateOverallRank(totalScore int64, factionCount int) ReputationRank {
	// Calculate average reputation per faction
	if factionCount == 0 {
		return ReputationRankNeutral
	}

	avgScore := float64(totalScore) / float64(factionCount)

	// Determine rank based on average score and total score
	switch {
	case avgScore >= 6000 && totalScore >= 50000:
		return ReputationRankLegendary
	case avgScore >= 4000 && totalScore >= 25000:
		return ReputationRankRenowned
	case avgScore >= 2000 && totalScore >= 10000:
		return ReputationRankRespected
	case avgScore >= 500 && totalScore >= 2500:
		return ReputationRankKnown
	case avgScore >= -500 && totalScore >= -2500:
		return ReputationRankNeutral
	case avgScore >= -2000 && totalScore >= -10000:
		return ReputationRankUnknown
	case avgScore >= -4000 && totalScore >= -25000:
		return ReputationRankDisliked
	case avgScore >= -6000 && totalScore >= -50000:
		return ReputationRankNotorious
	default:
		return ReputationRankInfamous
	}
}

// applyFactionInfluenceUnsafe applies reputation changes to allied/enemy factions (must be called with lock held)
func (rs *ReputationSystem) applyFactionInfluenceUnsafe(playerID, factionID string, change int64) {
	factionRep := rs.getFactionReputation(factionID)

	// Apply to allied factions (positive influence)
	for _, alliedID := range factionRep.AlliedFactions {
		if alliedStanding, exists := rs.PlayerReputations[playerID].FactionStandings[alliedID]; exists {
			influenceChange := int64(float64(change) * 0.25) // 25% of original change
			if influenceChange != 0 {
				alliedStanding.ReputationScore += influenceChange
				alliedStanding.ReputationLevel = rs.calculateReputationLevel(alliedStanding.ReputationScore)

				// Record influence event
				event := &ReputationEvent{
					ID:            fmt.Sprintf("influence_%d_%s_%s", time.Now().UnixNano(), playerID, alliedID),
					PlayerID:      playerID,
					FactionID:     alliedID,
					Change:        influenceChange,
					Reason:        fmt.Sprintf("allied faction influence from %s", factionID),
					ActionType:    "influence",
					Timestamp:     time.Now(),
					PreviousScore: alliedStanding.ReputationScore - influenceChange,
					NewScore:      alliedStanding.ReputationScore,
					Properties:    make(map[string]interface{}),
				}
				rs.ReputationHistory = append(rs.ReputationHistory, event)
			}
		}
	}

	// Apply to enemy factions (negative influence)
	for _, enemyID := range factionRep.EnemyFactions {
		if enemyStanding, exists := rs.PlayerReputations[playerID].FactionStandings[enemyID]; exists {
			influenceChange := int64(float64(change) * -0.15) // 15% inverse of original change
			if influenceChange != 0 {
				enemyStanding.ReputationScore += influenceChange
				enemyStanding.ReputationLevel = rs.calculateReputationLevel(enemyStanding.ReputationScore)

				// Record influence event
				event := &ReputationEvent{
					ID:            fmt.Sprintf("influence_%d_%s_%s", time.Now().UnixNano(), playerID, enemyID),
					PlayerID:      playerID,
					FactionID:     enemyID,
					Change:        influenceChange,
					Reason:        fmt.Sprintf("enemy faction influence from %s", factionID),
					ActionType:    "influence",
					Timestamp:     time.Now(),
					PreviousScore: enemyStanding.ReputationScore - influenceChange,
					NewScore:      enemyStanding.ReputationScore,
					Properties:    make(map[string]interface{}),
				}
				rs.ReputationHistory = append(rs.ReputationHistory, event)
			}
		}
	}
}

// updateReputationEffectsUnsafe updates active effects based on current reputation (must be called with lock held)
func (rs *ReputationSystem) updateReputationEffectsUnsafe(playerID, factionID string) {
	standing, exists := rs.PlayerReputations[playerID].FactionStandings[factionID]
	if !exists {
		return
	}

	// Define effect key
	effectKey := fmt.Sprintf("%s_%s", playerID, factionID)

	// Remove existing effects for this player-faction combination
	for key := range rs.ReputationEffects {
		if key == effectKey {
			delete(rs.ReputationEffects, key)
		}
	}

	// Create new effects based on current reputation level
	now := time.Now()

	// Price effects
	if standing.ReputationLevel == ReputationLevelFriendly || standing.ReputationLevel == ReputationLevelHonored ||
		standing.ReputationLevel == ReputationLevelExalted || standing.ReputationLevel == ReputationLevelRevered {

		magnitude, _ := rs.calculateEffectUnsafe(playerID, factionID, ReputationEffectPriceDiscount)
		effect := &ReputationEffect{
			ID:          fmt.Sprintf("price_discount_%s", effectKey),
			PlayerID:    playerID,
			FactionID:   factionID,
			EffectType:  ReputationEffectPriceDiscount,
			Magnitude:   magnitude,
			Description: fmt.Sprintf("%.0f%% discount with %s faction", math.Abs(magnitude)*100, factionID),
			IsActive:    true,
			StartTime:   now,
			Properties:  make(map[string]interface{}),
		}
		rs.ReputationEffects[effect.ID] = effect
	}

	if standing.ReputationLevel == ReputationLevelUnfriendly || standing.ReputationLevel == ReputationLevelHostile ||
		standing.ReputationLevel == ReputationLevelHated || standing.ReputationLevel == ReputationLevelDespised {

		magnitude, _ := rs.calculateEffectUnsafe(playerID, factionID, ReputationEffectPricePenalty)
		effect := &ReputationEffect{
			ID:          fmt.Sprintf("price_penalty_%s", effectKey),
			PlayerID:    playerID,
			FactionID:   factionID,
			EffectType:  ReputationEffectPricePenalty,
			Magnitude:   math.Abs(magnitude),
			Description: fmt.Sprintf("%.0f%% price increase with %s faction", math.Abs(magnitude)*100, factionID),
			IsActive:    true,
			StartTime:   now,
			Properties:  make(map[string]interface{}),
		}
		rs.ReputationEffects[effect.ID] = effect
	}

	// Quest access effects
	if standing.ReputationScore >= rs.getFactionReputation(factionID).UnlockThreshold {
		effect := &ReputationEffect{
			ID:          fmt.Sprintf("quest_access_%s", effectKey),
			PlayerID:    playerID,
			FactionID:   factionID,
			EffectType:  ReputationEffectQuestAccess,
			Magnitude:   1.0,
			Description: fmt.Sprintf("Access to special %s faction quests", factionID),
			IsActive:    true,
			StartTime:   now,
			Properties:  make(map[string]interface{}),
		}
		rs.ReputationEffects[effect.ID] = effect
	}
}

// applyFactionInfluence applies reputation changes to allied/enemy factions (public wrapper)
func (rs *ReputationSystem) applyFactionInfluence(playerID, factionID string, change int64) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.applyFactionInfluenceUnsafe(playerID, factionID, change)
}

// updateReputationEffects updates active effects based on current reputation (public wrapper)
func (rs *ReputationSystem) updateReputationEffects(playerID, factionID string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.updateReputationEffectsUnsafe(playerID, factionID)
}
