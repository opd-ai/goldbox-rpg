package pcg

import (
	"time"
)

// ReputationLevel represents different tiers of reputation standing
type ReputationLevel string

const (
	ReputationLevelRevered    ReputationLevel = "revered"    // +7501 to +10000
	ReputationLevelExalted    ReputationLevel = "exalted"    // +5001 to +7500
	ReputationLevelHonored    ReputationLevel = "honored"    // +2501 to +5000
	ReputationLevelFriendly   ReputationLevel = "friendly"   // +501 to +2500
	ReputationLevelNeutral    ReputationLevel = "neutral"    // -500 to +500
	ReputationLevelUnfriendly ReputationLevel = "unfriendly" // -2500 to -501
	ReputationLevelHostile    ReputationLevel = "hostile"    // -5000 to -2501
	ReputationLevelHated      ReputationLevel = "hated"      // -7500 to -5001
	ReputationLevelDespised   ReputationLevel = "despised"   // -10000 to -7501
)

// ReputationRank represents overall reputation standing across all factions
type ReputationRank string

const (
	ReputationRankLegendary ReputationRank = "legendary" // Top 1% of players
	ReputationRankRenowned  ReputationRank = "renowned"  // Top 5% of players
	ReputationRankRespected ReputationRank = "respected" // Top 15% of players
	ReputationRankKnown     ReputationRank = "known"     // Top 35% of players
	ReputationRankNeutral   ReputationRank = "neutral"   // Middle 30% of players
	ReputationRankUnknown   ReputationRank = "unknown"   // Bottom 35% of players
	ReputationRankDisliked  ReputationRank = "disliked"  // Bottom 15% of players
	ReputationRankNotorious ReputationRank = "notorious" // Bottom 5% of players
	ReputationRankInfamous  ReputationRank = "infamous"  // Bottom 1% of players
)

// ReputationActionType categorizes actions that affect reputation
type ReputationActionType string

const (
	ReputationActionQuest     ReputationActionType = "quest"     // Quest completion or failure
	ReputationActionTrade     ReputationActionType = "trade"     // Trading with faction members
	ReputationActionCombat    ReputationActionType = "combat"    // Fighting for or against faction
	ReputationActionDiplomacy ReputationActionType = "diplomacy" // Diplomatic interactions
	ReputationActionTheft     ReputationActionType = "theft"     // Stealing from faction
	ReputationActionMurder    ReputationActionType = "murder"    // Killing faction members
	ReputationActionRescue    ReputationActionType = "rescue"    // Rescuing faction members
	ReputationActionBetrayal  ReputationActionType = "betrayal"  // Betraying faction trust
	ReputationActionCharity   ReputationActionType = "charity"   // Charitable acts toward faction
	ReputationActionSabotage  ReputationActionType = "sabotage"  // Sabotaging faction operations
)

// ReputationEffectType represents different types of reputation-based effects
type ReputationEffectType string

const (
	ReputationEffectPriceDiscount     ReputationEffectType = "price_discount"     // Trading price reduction
	ReputationEffectPricePenalty      ReputationEffectType = "price_penalty"      // Trading price increase
	ReputationEffectQuestAccess       ReputationEffectType = "quest_access"       // Access to special quests
	ReputationEffectQuestReward       ReputationEffectType = "quest_reward"       // Enhanced quest rewards
	ReputationEffectServiceAccess     ReputationEffectType = "service_access"     // Access to special services
	ReputationEffectServicePenalty    ReputationEffectType = "service_penalty"    // Denial of services
	ReputationEffectAttitudeBonus     ReputationEffectType = "attitude_bonus"     // Improved NPC attitudes
	ReputationEffectAttitudePenalty   ReputationEffectType = "attitude_penalty"   // Worsened NPC attitudes
	ReputationEffectCombatAssistance  ReputationEffectType = "combat_assistance"  // NPCs help in combat
	ReputationEffectCombatHostility   ReputationEffectType = "combat_hostility"   // NPCs attack on sight
	ReputationEffectInformationAccess ReputationEffectType = "information_access" // Access to faction secrets
	ReputationEffectTerritoryAccess   ReputationEffectType = "territory_access"   // Access to faction territories
)

// ReputationParams provides reputation-specific generation parameters
type ReputationParams struct {
	GenerationParams    `yaml:",inline"`
	DefaultBaseAttitude int64              `yaml:"default_base_attitude"` // Default starting reputation
	AttitudeVariance    int64              `yaml:"attitude_variance"`     // Random variance in base attitudes
	DecayEnabled        bool               `yaml:"decay_enabled"`         // Whether reputation decays over time
	AlliedInfluence     float64            `yaml:"allied_influence"`      // How much allied factions share reputation
	EnemyInfluence      float64            `yaml:"enemy_influence"`       // How much enemy factions oppose reputation
	MemoryDuration      time.Duration      `yaml:"memory_duration"`       // How long factions remember actions
	EffectMultipliers   map[string]float64 `yaml:"effect_multipliers"`    // Multipliers for different effect types
	CustomThresholds    map[string]int64   `yaml:"custom_thresholds"`     // Custom reputation thresholds
}
