package pcg

import (
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestContentBalancer_balanceTerrain tests the terrain balancing function
func TestContentBalancer_balanceTerrain(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Quiet logging for tests

	cb := NewContentBalancer(logger)

	tests := []struct {
		name    string
		content interface{}
		rule    ScalingRule
		scaling float64
	}{
		{
			name:    "basic terrain balancing",
			content: map[string]interface{}{"terrain": "forest"},
			rule: ScalingRule{
				ContentType:      ContentTypeTerrain,
				ScalingType:      ScalingLinear,
				DifficultyFactor: 1.0,
			},
			scaling: 1.5,
		},
		{
			name:    "terrain with higher scaling",
			content: map[string]interface{}{"terrain": "dungeon"},
			rule: ScalingRule{
				ContentType:      ContentTypeTerrain,
				ScalingType:      ScalingExponential,
				DifficultyFactor: 2.0,
			},
			scaling: 2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cb.balanceTerrain(tt.content, tt.rule, tt.scaling)
			assert.NoError(t, err)
			assert.Equal(t, tt.content, result) // Terrain is returned unchanged
		})
	}
}

// TestContentBalancer_balanceGenericContent tests generic content balancing
func TestContentBalancer_balanceGenericContent(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cb := NewContentBalancer(logger)

	tests := []struct {
		name    string
		content interface{}
		rule    ScalingRule
		scaling float64
	}{
		{
			name:    "generic content balancing",
			content: map[string]interface{}{"type": "unknown"},
			rule: ScalingRule{
				ContentType:      ContentType("custom"),
				ScalingType:      ScalingLinear,
				DifficultyFactor: 1.0,
			},
			scaling: 1.0,
		},
		{
			name:    "string content",
			content: "some_content",
			rule:    ScalingRule{},
			scaling: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cb.balanceGenericContent(tt.content, tt.rule, tt.scaling)
			assert.NoError(t, err)
			assert.Equal(t, tt.content, result) // Generic content is returned unchanged
		})
	}
}

// TestContentBalancer_getDefaultScalingRule tests default scaling rule generation
func TestContentBalancer_getDefaultScalingRule(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cb := NewContentBalancer(logger)

	tests := []struct {
		name        string
		contentType ContentType
	}{
		{name: "quests type", contentType: ContentTypeQuests},
		{name: "items type", contentType: ContentTypeItems},
		{name: "terrain type", contentType: ContentTypeTerrain},
		{name: "npcs type", contentType: ContentTypeNPCs},
		{name: "custom type", contentType: ContentType("custom")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := cb.getDefaultScalingRule(tt.contentType)

			assert.Equal(t, tt.contentType, rule.ContentType)
			assert.Equal(t, ScalingLinear, rule.ScalingType)
			assert.Equal(t, 1.0, rule.DifficultyFactor)
			assert.Equal(t, 1.0, rule.RewardMultiplier)
			assert.Equal(t, 1.0, rule.ResourceCost)
			assert.Equal(t, 1, rule.MinLevel)
			assert.Equal(t, 20, rule.MaxLevel)
			assert.NotNil(t, rule.CustomParameters)
		})
	}
}

// TestCharacterGenerator_ageRangeToDescription tests age range description
func TestCharacterGenerator_ageRangeToDescription(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	gen := NewNPCGenerator(logger)

	tests := []struct {
		name     string
		ageRange AgeRange
	}{
		{name: "child", ageRange: AgeRangeChild},
		{name: "adolescent", ageRange: AgeRangeAdolescent},
		{name: "young adult", ageRange: AgeRangeYoungAdult},
		{name: "adult", ageRange: AgeRangeAdult},
		{name: "middle aged", ageRange: AgeRangeMiddleAged},
		{name: "elderly", ageRange: AgeRangeElderly},
		{name: "ancient", ageRange: AgeRangeAncient},
		{name: "unknown", ageRange: AgeRange("unknown")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.ageRangeToDescription(tt.ageRange)
			assert.NotEmpty(t, result)
		})
	}
}

// TestCharacterGenerator_elevatedSocialClass tests social class elevation
func TestCharacterGenerator_elevatedSocialClass(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	gen := NewNPCGenerator(logger)

	tests := []struct {
		name        string
		socialClass SocialClass
	}{
		{name: "slave", socialClass: SocialClassSlave},
		{name: "serf", socialClass: SocialClassSerf},
		{name: "peasant", socialClass: SocialClassPeasant},
		{name: "crafter", socialClass: SocialClassCrafter},
		{name: "merchant", socialClass: SocialClassMerchant},
		{name: "gentry", socialClass: SocialClassGentry},
		{name: "noble", socialClass: SocialClassNoble},
		{name: "royalty", socialClass: SocialClassRoyalty},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.elevatedSocialClass(tt.socialClass)
			// Should return same class or one level higher (capped at Royalty)
			assert.NotEmpty(t, result)
		})
	}
}

// TestFactionGenerator_selectConflictType tests conflict type selection
func TestFactionGenerator_selectConflictType(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	gen := NewFactionGenerator(logger)

	// Call multiple times to test randomness
	for i := 0; i < 10; i++ {
		result := gen.selectConflictType()
		assert.NotEmpty(t, result)
	}
}

// TestFactionGenerator_generateConflictName tests conflict name generation
func TestFactionGenerator_generateConflictName(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	gen := NewFactionGenerator(logger)

	// Call multiple times to get various names
	for i := 0; i < 10; i++ {
		result := gen.generateConflictName()
		assert.NotEmpty(t, result)
	}
}

// TestFactionGenerator_generateConflictCause tests conflict cause generation
func TestFactionGenerator_generateConflictCause(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	gen := NewFactionGenerator(logger)

	// Call multiple times to get various causes
	for i := 0; i < 10; i++ {
		result := gen.generateConflictCause()
		assert.NotEmpty(t, result)
	}
}

// TestCharacterGenerator_establishGroupRelationships tests group relationship establishment
func TestCharacterGenerator_establishGroupRelationships(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	gen := NewNPCGenerator(logger)

	// Create test NPCs - NPC embeds Character, so we need to set Character fields
	npcs := []*game.NPC{
		{
			Character: game.Character{
				ID:   "npc1",
				Name: "NPC One",
			},
		},
		{
			Character: game.Character{
				ID:   "npc2",
				Name: "NPC Two",
			},
		},
		{
			Character: game.Character{
				ID:   "npc3",
				Name: "NPC Three",
			},
		},
	}

	// Test with different group types
	groupTypes := []NPCGroupType{
		NPCGroupFamily,
		NPCGroupGuards,
		NPCGroupMerchants,
		NPCGroupCultists,
	}

	for _, groupType := range groupTypes {
		// Should not panic
		gen.establishGroupRelationships(npcs, groupType)
	}
}
