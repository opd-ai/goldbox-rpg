package pcg

import (
	"time"

	"goldbox-rpg/pkg/game"
)

// FactionType represents different categories of factions
type FactionType string

const (
	FactionTypeMilitary  FactionType = "military"
	FactionTypeEconomic  FactionType = "economic"
	FactionTypeReligious FactionType = "religious"
	FactionTypeCriminal  FactionType = "criminal"
	FactionTypeScholarly FactionType = "scholarly"
	FactionTypePolitical FactionType = "political"
	FactionTypeMercenary FactionType = "mercenary"
	FactionTypeMagical   FactionType = "magical"
)

// RelationshipStatus represents the diplomatic status between factions
type RelationshipStatus string

const (
	RelationStatusAllied   RelationshipStatus = "allied"
	RelationStatusFriendly RelationshipStatus = "friendly"
	RelationStatusNeutral  RelationshipStatus = "neutral"
	RelationStatusTense    RelationshipStatus = "tense"
	RelationStatusHostile  RelationshipStatus = "hostile"
	RelationStatusWar      RelationshipStatus = "war"
)

// TerritoryType represents different types of faction territories
type TerritoryType string

const (
	TerritoryTypeCapital     TerritoryType = "capital"
	TerritoryTypeCity        TerritoryType = "city"
	TerritoryTypeOutpost     TerritoryType = "outpost"
	TerritoryTypeFortress    TerritoryType = "fortress"
	TerritoryTypeTradingPost TerritoryType = "trading_post"
	TerritoryTypeResource    TerritoryType = "resource"
)

// ConflictType represents different types of conflicts between factions
type ConflictType string

const (
	ConflictTypeTrade      ConflictType = "trade"
	ConflictTypeTerritory  ConflictType = "territory"
	ConflictTypeReligious  ConflictType = "religious"
	ConflictTypeResource   ConflictType = "resource"
	ConflictTypeSuccession ConflictType = "succession"
	ConflictTypeRevenge    ConflictType = "revenge"
)

// Faction represents a political/social organization
type Faction struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       FactionType            `json:"type"`
	Government GovernmentType         `json:"government"`
	Ideology   string                 `json:"ideology"`
	Power      int                    `json:"power"`
	Wealth     int                    `json:"wealth"`
	Military   int                    `json:"military"`
	Influence  int                    `json:"influence"`
	Stability  float64                `json:"stability"`
	Goals      []string               `json:"goals"`
	Resources  []ResourceType         `json:"resources"`
	Leaders    []*FactionLeader       `json:"leaders"`
	Properties map[string]interface{} `json:"properties"`
}

// FactionLeader represents a leader within a faction
type FactionLeader struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Title      string                 `json:"title"`
	Age        int                    `json:"age"`
	Traits     []string               `json:"traits"`
	Loyalty    float64                `json:"loyalty"`
	Competence float64                `json:"competence"`
	Influence  float64                `json:"influence"`
	Properties map[string]interface{} `json:"properties"`
}

// FactionRelationship represents diplomatic relations between two factions
type FactionRelationship struct {
	ID          string                 `json:"id"`
	Faction1ID  string                 `json:"faction1_id"`
	Faction2ID  string                 `json:"faction2_id"`
	Status      RelationshipStatus     `json:"status"`
	Opinion     float64                `json:"opinion"`     // -1.0 to 1.0
	TrustLevel  float64                `json:"trust_level"` // 0.0 to 1.0
	TradeLevel  float64                `json:"trade_level"` // 0.0 to 1.0
	Hostility   float64                `json:"hostility"`   // 0.0 to 1.0
	History     []string               `json:"history"`
	LastChanged time.Time              `json:"last_changed"`
	Properties  map[string]interface{} `json:"properties"`
}

// Territory represents a geographic area controlled by a faction
type Territory struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         TerritoryType          `json:"type"`
	ControllerID string                 `json:"controller_id"`
	Position     game.Position          `json:"position"`
	Size         int                    `json:"size"`
	Population   int                    `json:"population"`
	Defenses     int                    `json:"defenses"`
	Resources    []ResourceType         `json:"resources"`
	Strategic    bool                   `json:"strategic"`
	Properties   map[string]interface{} `json:"properties"`
}

// TradeDeal represents economic agreements between factions
type TradeDeal struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Faction1ID string                 `json:"faction1_id"`
	Faction2ID string                 `json:"faction2_id"`
	Resource1  ResourceType           `json:"resource1"`
	Resource2  ResourceType           `json:"resource2"`
	Volume1    int                    `json:"volume1"`
	Volume2    int                    `json:"volume2"`
	Duration   int                    `json:"duration"` // Duration in days
	Profit1    int                    `json:"profit1"`
	Profit2    int                    `json:"profit2"`
	Active     bool                   `json:"active"`
	Properties map[string]interface{} `json:"properties"`
}

// Conflict represents ongoing conflicts between factions
type Conflict struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       ConflictType           `json:"type"`
	Factions   []string               `json:"factions"`
	Cause      string                 `json:"cause"`
	Intensity  float64                `json:"intensity"` // 0.0 to 1.0
	Duration   int                    `json:"duration"`  // Duration in days
	Territory  string                 `json:"territory"` // Disputed territory ID
	Resolution string                 `json:"resolution"`
	Active     bool                   `json:"active"`
	Properties map[string]interface{} `json:"properties"`
}

// GeneratedFactionSystem represents a complete faction system
type GeneratedFactionSystem struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Factions      []*Faction             `json:"factions"`
	Relationships []*FactionRelationship `json:"relationships"`
	Territories   []*Territory           `json:"territories"`
	TradeDeals    []*TradeDeal           `json:"trade_deals"`
	Conflicts     []*Conflict            `json:"conflicts"`
	Metadata      map[string]interface{} `json:"metadata"`
	Generated     time.Time              `json:"generated"`
}

// FactionParams provides faction-specific generation parameters
type FactionParams struct {
	GenerationParams   `yaml:",inline"`
	FactionCount       int     `yaml:"faction_count"`       // Number of factions to generate
	MinPower           int     `yaml:"min_power"`           // Minimum faction power level
	MaxPower           int     `yaml:"max_power"`           // Maximum faction power level
	ConflictLevel      float64 `yaml:"conflict_level"`      // Overall conflict intensity (0.0-1.0)
	EconomicFocus      float64 `yaml:"economic_focus"`      // Economic activity emphasis (0.0-1.0)
	MilitaryFocus      float64 `yaml:"military_focus"`      // Military emphasis (0.0-1.0)
	CulturalFocus      float64 `yaml:"cultural_focus"`      // Cultural/religious emphasis (0.0-1.0)
	TerritoryCount     int     `yaml:"territory_count"`     // Number of territories per faction
	TradeVolume        float64 `yaml:"trade_volume"`        // Overall trade activity (0.0-1.0)
	PoliticalStability float64 `yaml:"political_stability"` // Overall political stability (0.0-1.0)
}
