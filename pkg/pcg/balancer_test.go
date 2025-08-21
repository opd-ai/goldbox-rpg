package pcg

import (
	"context"
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContentBalancer(t *testing.T) {
	tests := []struct {
		name   string
		logger *logrus.Logger
	}{
		{
			name:   "with logger",
			logger: logrus.New(),
		},
		{
			name:   "with nil logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balancer := NewContentBalancer(tt.logger)

			assert.NotNil(t, balancer)
			assert.NotNil(t, balancer.logger)
			assert.Equal(t, "1.0.0", balancer.GetVersion())
			assert.NotNil(t, balancer.metrics)
			assert.NotNil(t, balancer.powerCurves)
			assert.NotNil(t, balancer.scalingRules)
			assert.NotNil(t, balancer.resourceLimits)
		})
	}
}

func TestContentBalancer_BalanceContent(t *testing.T) {
	balancer := NewContentBalancer(logrus.New())
	ctx := context.Background()

	tests := []struct {
		name        string
		request     BalanceRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid quest balance request",
			request: BalanceRequest{
				ContentType:  ContentTypeQuests,
				PlayerLevel:  5,
				Difficulty:   3,
				ContentValue: createTestQuest(),
				Context:      map[string]interface{}{"test": true},
			},
			expectError: false,
		},
		{
			name: "valid character balance request",
			request: BalanceRequest{
				ContentType:  ContentTypeCharacters,
				PlayerLevel:  10,
				Difficulty:   5,
				ContentValue: createTestCharacter(),
				Context:      map[string]interface{}{"test": true},
			},
			expectError: false,
		},
		{
			name: "valid dungeon balance request",
			request: BalanceRequest{
				ContentType:  ContentTypeDungeon,
				PlayerLevel:  7,
				Difficulty:   4,
				ContentValue: createTestDungeon(),
				Context:      map[string]interface{}{"test": true},
			},
			expectError: false,
		},
		{
			name: "valid item balance request",
			request: BalanceRequest{
				ContentType:  ContentTypeItems,
				PlayerLevel:  3,
				Difficulty:   2,
				ContentValue: createTestItem(),
				Context:      map[string]interface{}{"test": true},
			},
			expectError: false,
		},
		{
			name: "invalid player level",
			request: BalanceRequest{
				ContentType:  ContentTypeQuests,
				PlayerLevel:  0,
				Difficulty:   3,
				ContentValue: createTestQuest(),
			},
			expectError: true,
			errorMsg:    "player level must be positive",
		},
		{
			name: "negative difficulty",
			request: BalanceRequest{
				ContentType:  ContentTypeQuests,
				PlayerLevel:  5,
				Difficulty:   -1,
				ContentValue: createTestQuest(),
			},
			expectError: true,
			errorMsg:    "difficulty must be non-negative",
		},
		{
			name: "nil content value",
			request: BalanceRequest{
				ContentType:  ContentTypeQuests,
				PlayerLevel:  5,
				Difficulty:   3,
				ContentValue: nil,
			},
			expectError: true,
			errorMsg:    "content value cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := balancer.BalanceContent(ctx, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.NotNil(t, result.BalancedContent)
				assert.Greater(t, result.AppliedScaling, 0.0)
				assert.GreaterOrEqual(t, result.DifficultyScore, 0.0)
				assert.GreaterOrEqual(t, result.RewardScore, 0.0)
				assert.NotNil(t, result.ResourceCost)
				assert.NotNil(t, result.Metadata)
				assert.GreaterOrEqual(t, result.BalanceQuality, 0.0)
				assert.LessOrEqual(t, result.BalanceQuality, 1.0)
			}
		})
	}
}

func TestContentBalancer_calculatePowerScaling(t *testing.T) {
	balancer := NewContentBalancer(logrus.New())

	curve := PowerCurve{
		BaseValue:       1.0,
		ScalingFactor:   0.1,
		ExponentFactor:  1.2,
		CapValue:        10.0,
		VarianceRange:   0.1,
		BreakpointLevel: 10,
	}

	tests := []struct {
		name        string
		playerLevel int
		difficulty  int
		curve       PowerCurve
		expectMin   float64
		expectMax   float64
	}{
		{
			name:        "level 1 easy",
			playerLevel: 1,
			difficulty:  1,
			curve:       curve,
			expectMin:   1.0,
			expectMax:   5.0,
		},
		{
			name:        "level 5 medium",
			playerLevel: 5,
			difficulty:  5,
			curve:       curve,
			expectMin:   2.0,
			expectMax:   15.0,
		},
		{
			name:        "level 15 hard with breakpoint",
			playerLevel: 15,
			difficulty:  8,
			curve:       curve,
			expectMin:   5.0,
			expectMax:   25.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scaling := balancer.calculatePowerScaling(tt.playerLevel, tt.difficulty, tt.curve)

			assert.GreaterOrEqual(t, scaling, tt.expectMin)
			assert.LessOrEqual(t, scaling, tt.expectMax)
			assert.GreaterOrEqual(t, scaling, curve.BaseValue)

			// If cap is set, scaling should not exceed it
			if curve.CapValue > 0 {
				assert.LessOrEqual(t, scaling, curve.CapValue)
			}
		})
	}
}

func TestContentBalancer_balanceQuest(t *testing.T) {
	balancer := NewContentBalancer(logrus.New())
	rule := ScalingRule{
		DifficultyFactor: 1.5,
		RewardMultiplier: 2.0,
		ResourceCost:     1.0,
	}

	originalQuest := createTestQuest()
	originalExpReward := 0
	originalGoldReward := 0

	// Find original rewards
	for _, reward := range originalQuest.Rewards {
		if reward.Type == "exp" {
			originalExpReward = reward.Value
		}
		if reward.Type == "gold" {
			originalGoldReward = reward.Value
		}
	}

	scaling := 2.0
	result, err := balancer.balanceQuest(originalQuest, rule, scaling)

	require.NoError(t, err)
	require.NotNil(t, result)

	balancedQuest, ok := result.(*game.Quest)
	require.True(t, ok)

	// Check that original quest is not modified
	assert.Equal(t, originalQuest.ID, balancedQuest.ID)

	// Check scaled rewards
	for _, reward := range balancedQuest.Rewards {
		switch reward.Type {
		case "exp":
			expectedExp := int(float64(originalExpReward) * scaling * rule.RewardMultiplier)
			assert.Equal(t, expectedExp, reward.Value)
		case "gold":
			expectedGold := int(float64(originalGoldReward) * scaling * rule.RewardMultiplier)
			assert.Equal(t, expectedGold, reward.Value)
		}
	}
}

func TestContentBalancer_balanceCharacter(t *testing.T) {
	balancer := NewContentBalancer(logrus.New())
	rule := ScalingRule{
		DifficultyFactor: 1.5,
		RewardMultiplier: 1.0,
		ResourceCost:     1.0,
	}

	originalChar := createTestCharacter()
	originalHP := originalChar.MaxHP
	originalAC := originalChar.ArmorClass
	originalTHAC0 := originalChar.THAC0

	scaling := 2.0
	result, err := balancer.balanceCharacter(originalChar, rule, scaling)

	require.NoError(t, err)
	require.NotNil(t, result)

	balancedChar, ok := result.(*game.Character)
	require.True(t, ok)

	// Check HP scaling
	expectedHP := int(float64(originalHP) * scaling * rule.DifficultyFactor)
	assert.Equal(t, expectedHP, balancedChar.MaxHP)
	assert.Equal(t, balancedChar.MaxHP, balancedChar.HP)

	// Check AC improvement (lower is better)
	if scaling > 1.0 {
		assert.LessOrEqual(t, balancedChar.ArmorClass, originalAC)
	}

	// Check THAC0 improvement (lower is better)
	if scaling > 1.0 {
		assert.LessOrEqual(t, balancedChar.THAC0, originalTHAC0)
		assert.GreaterOrEqual(t, balancedChar.THAC0, 1) // Minimum THAC0 is 1
	}
}

func TestContentBalancer_balanceItem(t *testing.T) {
	balancer := NewContentBalancer(logrus.New())
	rule := ScalingRule{
		DifficultyFactor: 1.0,
		RewardMultiplier: 1.5,
		ResourceCost:     1.0,
	}

	originalItem := createTestItem()
	originalValue := originalItem.Value

	scaling := 3.0
	result, err := balancer.balanceItem(originalItem, rule, scaling)

	require.NoError(t, err)
	require.NotNil(t, result)

	balancedItem, ok := result.(*game.Item)
	require.True(t, ok)

	// Check value scaling
	if originalValue > 0 {
		expectedValue := int(float64(originalValue) * scaling * rule.RewardMultiplier)
		assert.Equal(t, expectedValue, balancedItem.Value)
	}

	// Check that other properties are preserved
	assert.Equal(t, originalItem.ID, balancedItem.ID)
	assert.Equal(t, originalItem.Name, balancedItem.Name)
}

func TestContentBalancer_validateResourceAvailability(t *testing.T) {
	balancer := NewContentBalancer(logrus.New())

	// Set up initial resource state
	balancer.metrics.ResourceUsageMetrics["generation_budget"] = ResourceMetrics{
		CurrentReserve: 100.0,
	}

	tests := []struct {
		name         string
		resourceCost map[string]float64
		expectError  bool
		errorMsg     string
	}{
		{
			name: "valid resource usage",
			resourceCost: map[string]float64{
				"generation_budget": 50.0,
			},
			expectError: false,
		},
		{
			name: "exceeds absolute maximum",
			resourceCost: map[string]float64{
				"generation_budget": 950.0, // Would exceed 1000.0 absolute max
			},
			expectError: true,
			errorMsg:    "would exceed absolute maximum",
		},
		{
			name: "below critical reserve",
			resourceCost: map[string]float64{
				"generation_budget": 95.0, // Would leave 5.0, below 10.0 critical reserve
			},
			expectError: true,
			errorMsg:    "would fall below critical reserve",
		},
		{
			name: "unknown resource type",
			resourceCost: map[string]float64{
				"unknown_resource": 100.0,
			},
			expectError: false, // Should skip validation for unknown resources
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := balancer.validateResourceAvailability(tt.resourceCost)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContentBalancer_GetMetrics(t *testing.T) {
	balancer := NewContentBalancer(logrus.New())

	// Perform some balance operations to generate metrics
	ctx := context.Background()
	request := BalanceRequest{
		ContentType:  ContentTypeQuests,
		PlayerLevel:  5,
		Difficulty:   3,
		ContentValue: createTestQuest(),
	}

	_, err := balancer.BalanceContent(ctx, request)
	require.NoError(t, err)

	metrics := balancer.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Greater(t, metrics.TotalBalanceChecks, int64(0))
	assert.Greater(t, metrics.SuccessfulBalances, int64(0))
	assert.NotEmpty(t, metrics.ContentTypeMetrics)
	assert.Contains(t, metrics.ContentTypeMetrics, ContentTypeQuests)
}

func TestContentBalancer_assessBalanceQuality(t *testing.T) {
	balancer := NewContentBalancer(logrus.New())

	request := BalanceRequest{
		ContentType:  ContentTypeQuests,
		PlayerLevel:  5,
		Difficulty:   3,
		ContentValue: createTestQuest(),
	}

	tests := []struct {
		name         string
		scaling      float64
		resourceCost map[string]float64
		expectRange  [2]float64 // [min, max]
	}{
		{
			name:    "good scaling and moderate cost",
			scaling: 1.0,
			resourceCost: map[string]float64{
				"generation_budget": 7.5, // Expected cost for level 5 * 1.5
			},
			expectRange: [2]float64{0.5, 1.0},
		},
		{
			name:    "extreme scaling",
			scaling: 10.0, // Way above tolerance
			resourceCost: map[string]float64{
				"generation_budget": 7.5,
			},
			expectRange: [2]float64{0.0, 0.8}, // Reduced quality
		},
		{
			name:    "excessive resource cost",
			scaling: 1.0,
			resourceCost: map[string]float64{
				"generation_budget": 50.0, // Much higher than expected
			},
			expectRange: [2]float64{0.0, 0.9}, // Reduced quality
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quality := balancer.assessBalanceQuality(request, tt.scaling, tt.resourceCost)

			assert.GreaterOrEqual(t, quality, tt.expectRange[0])
			assert.LessOrEqual(t, quality, tt.expectRange[1])
			assert.GreaterOrEqual(t, quality, 0.0)
			assert.LessOrEqual(t, quality, 1.0)
		})
	}
}

func TestContentBalancer_collectBalanceWarnings(t *testing.T) {
	balancer := NewContentBalancer(logrus.New())

	tests := []struct {
		name           string
		request        BalanceRequest
		scaling        float64
		balanceQuality float64
		expectWarnings int
		checkMessages  []string
	}{
		{
			name: "no warnings",
			request: BalanceRequest{
				PlayerLevel: 5,
				Difficulty:  3,
			},
			scaling:        1.0,
			balanceQuality: 0.8,
			expectWarnings: 0,
		},
		{
			name: "scaling too high",
			request: BalanceRequest{
				PlayerLevel: 5,
				Difficulty:  3,
			},
			scaling:        2.0, // Above default tolerance max of 1.3
			balanceQuality: 0.8,
			expectWarnings: 1,
			checkMessages:  []string{"exceeds maximum tolerance"},
		},
		{
			name: "scaling too low",
			request: BalanceRequest{
				PlayerLevel: 5,
				Difficulty:  3,
			},
			scaling:        0.5, // Below default tolerance min of 0.7
			balanceQuality: 0.8,
			expectWarnings: 1,
			checkMessages:  []string{"below minimum tolerance"},
		},
		{
			name: "poor balance quality",
			request: BalanceRequest{
				PlayerLevel: 5,
				Difficulty:  3,
			},
			scaling:        1.0,
			balanceQuality: 0.3, // Below default critical threshold of 0.5
			expectWarnings: 1,
			checkMessages:  []string{"below critical threshold"},
		},
		{
			name: "player level too high",
			request: BalanceRequest{
				PlayerLevel: 25, // Above default max of 20
				Difficulty:  3,
			},
			scaling:        1.0,
			balanceQuality: 0.8,
			expectWarnings: 1,
			checkMessages:  []string{"exceeds maximum configured level"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := balancer.collectBalanceWarnings(tt.request, tt.scaling, tt.balanceQuality)

			assert.Len(t, warnings, tt.expectWarnings)

			for _, expectedMsg := range tt.checkMessages {
				found := false
				for _, warning := range warnings {
					if assert.Contains(t, warning, expectedMsg) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected warning message not found: %s", expectedMsg)
			}
		})
	}
}

// Helper functions to create test data

func createTestQuest() *game.Quest {
	return &game.Quest{
		ID:          "test-quest-1",
		Title:       "Test Quest",
		Description: "A test quest for validation",
		Status:      game.QuestNotStarted,
		Objectives: []game.QuestObjective{
			{
				Description: "Kill 5 goblins",
				Required:    5,
				Progress:    0,
				Completed:   false,
			},
		},
		Rewards: []game.QuestReward{
			{
				Type:  "exp",
				Value: 100,
			},
			{
				Type:  "gold",
				Value: 50,
			},
		},
	}
}

func createTestCharacter() *game.Character {
	return &game.Character{
		ID:           "test-char-1",
		Name:         "Test Character",
		Level:        5,
		HP:           50,
		MaxHP:        50,
		ArmorClass:   10,
		THAC0:        15,
		Strength:     15,
		Dexterity:    14,
		Constitution: 13,
		Intelligence: 12,
		Wisdom:       11,
		Charisma:     10,
	}
}

func createTestDungeon() *DungeonComplex {
	return &DungeonComplex{
		ID:   "test-dungeon-1",
		Name: "Test Dungeon",
		Difficulty: DifficultyProgression{
			BaseDifficulty: 3,
			ScalingFactor:  1.2,
			MaxDifficulty:  10,
		},
		Levels: make(map[int]*DungeonLevel),
		Theme:  ThemeClassic,
	}
}

func createTestItem() *game.Item {
	return &game.Item{
		ID:     "test-item-1",
		Name:   "Test Sword",
		Type:   "weapon",
		Value:  100,
		Weight: 3,
		Damage: "1d8",
	}
}
