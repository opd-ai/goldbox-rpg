package pcg

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReputationSystem(t *testing.T) {
	logger := logrus.New()
	rs := NewReputationSystem(logger)

	assert.NotNil(t, rs)
	assert.NotNil(t, rs.PlayerReputations)
	assert.NotNil(t, rs.FactionReputations)
	assert.NotNil(t, rs.ReputationHistory)
	assert.NotNil(t, rs.ReputationEffects)
	assert.NotNil(t, rs.GlobalModifiers)
	assert.NotNil(t, rs.DecaySettings)
	assert.Equal(t, logger, rs.logger)

	// Check default decay settings
	assert.True(t, rs.DecaySettings.EnableDecay)
	assert.Equal(t, 24*time.Hour, rs.DecaySettings.DecayInterval)
	assert.Equal(t, 0.01, rs.DecaySettings.BaseDecayRate)
}

func TestInitializePlayerReputation(t *testing.T) {
	tests := []struct {
		name          string
		playerID      string
		factionSystem *GeneratedFactionSystem
		expectError   bool
		expectedCount int
	}{
		{
			name:     "valid initialization",
			playerID: "player1",
			factionSystem: &GeneratedFactionSystem{
				Factions: []*Faction{
					{ID: "faction1", Name: "Faction One"},
					{ID: "faction2", Name: "Faction Two"},
				},
			},
			expectError:   false,
			expectedCount: 2,
		},
		{
			name:     "empty faction system",
			playerID: "player2",
			factionSystem: &GeneratedFactionSystem{
				Factions: []*Faction{},
			},
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:     "duplicate initialization",
			playerID: "player1", // Same as first test
			factionSystem: &GeneratedFactionSystem{
				Factions: []*Faction{
					{ID: "faction1", Name: "Faction One"},
				},
			},
			expectError: true,
		},
	}

	rs := NewReputationSystem(logrus.New())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rs.InitializePlayerReputation(tt.playerID, tt.factionSystem)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify player reputation was created
			playerRep, exists := rs.PlayerReputations[tt.playerID]
			require.True(t, exists)
			assert.Equal(t, tt.playerID, playerRep.PlayerID)
			assert.Equal(t, tt.expectedCount, len(playerRep.FactionStandings))
			assert.Equal(t, ReputationRankNeutral, playerRep.ReputationRank)

			// Verify faction standings were created with defaults
			for _, faction := range tt.factionSystem.Factions {
				standing, exists := playerRep.FactionStandings[faction.ID]
				require.True(t, exists)
				assert.Equal(t, faction.ID, standing.FactionID)
				assert.Equal(t, int64(0), standing.ReputationScore) // Default base attitude
				assert.Equal(t, ReputationLevelNeutral, standing.ReputationLevel)
				assert.False(t, standing.IsLocked)
				assert.Equal(t, 0, standing.ActionCount)
			}
		})
	}
}

func TestModifyReputation(t *testing.T) {
	tests := []struct {
		name               string
		change             int64
		reason             string
		actionType         ReputationActionType
		expectError        bool
		expectedLevel      ReputationLevel
		expectedScoreRange [2]int64 // min, max
	}{
		{
			name:               "positive change",
			change:             1000,
			reason:             "completed quest",
			actionType:         ReputationActionQuest,
			expectError:        false,
			expectedLevel:      ReputationLevelFriendly,
			expectedScoreRange: [2]int64{1000, 1000},
		},
		{
			name:               "negative change",
			change:             -500,
			reason:             "betrayed faction",
			actionType:         ReputationActionBetrayal,
			expectError:        false,
			expectedLevel:      ReputationLevelNeutral,
			expectedScoreRange: [2]int64{-500, -500},
		},
		{
			name:               "large positive change",
			change:             5000,
			reason:             "saved faction leader",
			actionType:         ReputationActionRescue,
			expectError:        false,
			expectedLevel:      ReputationLevelHonored, // 5000 is honored, not exalted
			expectedScoreRange: [2]int64{5000, 5000},
		},
		{
			name:               "extreme negative change",
			change:             -15000, // Should be clamped to -10000
			reason:             "genocide",
			actionType:         ReputationActionMurder,
			expectError:        false,
			expectedLevel:      ReputationLevelDespised,
			expectedScoreRange: [2]int64{-10000, -10000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh reputation system and player for each test
			rs := NewReputationSystem(logrus.New())
			playerID := "test_player"
			factionID := "test_faction"

			// Initialize player reputation
			factionSystem := &GeneratedFactionSystem{
				Factions: []*Faction{
					{ID: factionID, Name: "Test Faction"},
				},
			}
			require.NoError(t, rs.InitializePlayerReputation(playerID, factionSystem))

			initialScore := rs.PlayerReputations[playerID].FactionStandings[factionID].ReputationScore

			err := rs.ModifyReputation(playerID, factionID, tt.change, tt.reason, tt.actionType)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Check final reputation
			standing := rs.PlayerReputations[playerID].FactionStandings[factionID]
			assert.Equal(t, tt.expectedLevel, standing.ReputationLevel)
			assert.GreaterOrEqual(t, standing.ReputationScore, tt.expectedScoreRange[0])
			assert.LessOrEqual(t, standing.ReputationScore, tt.expectedScoreRange[1])

			// Check that action count increased
			assert.Greater(t, standing.ActionCount, 0)

			// Check that an event was recorded
			events, err := rs.GetReputationHistory(playerID, factionID, 1)
			require.NoError(t, err)
			require.Len(t, events, 1)

			event := events[0]
			assert.Equal(t, playerID, event.PlayerID)
			assert.Equal(t, factionID, event.FactionID)
			assert.Equal(t, tt.reason, event.Reason)
			assert.Equal(t, tt.actionType, event.ActionType)
			assert.Equal(t, initialScore, event.PreviousScore)
		})
	}
}

func TestGetReputation(t *testing.T) {
	rs := NewReputationSystem(logrus.New())
	playerID := "test_player"
	factionID := "test_faction"

	// Test getting reputation for non-existent player
	_, err := rs.GetReputation("nonexistent", factionID)
	assert.Error(t, err)

	// Initialize player reputation
	factionSystem := &GeneratedFactionSystem{
		Factions: []*Faction{
			{ID: factionID, Name: "Test Faction"},
		},
	}
	require.NoError(t, rs.InitializePlayerReputation(playerID, factionSystem))

	// Test getting valid reputation
	standing, err := rs.GetReputation(playerID, factionID)
	require.NoError(t, err)
	assert.Equal(t, factionID, standing.FactionID)
	assert.Equal(t, int64(0), standing.ReputationScore)
	assert.Equal(t, ReputationLevelNeutral, standing.ReputationLevel)

	// Test getting reputation for non-existent faction
	_, err = rs.GetReputation(playerID, "nonexistent")
	assert.Error(t, err)
}

func TestCalculateReputationLevel(t *testing.T) {
	rs := NewReputationSystem(logrus.New())

	tests := []struct {
		score    int64
		expected ReputationLevel
	}{
		{10000, ReputationLevelRevered},
		{7501, ReputationLevelRevered},
		{7500, ReputationLevelExalted},
		{5001, ReputationLevelExalted},
		{5000, ReputationLevelHonored},
		{2501, ReputationLevelHonored},
		{2500, ReputationLevelFriendly},
		{501, ReputationLevelFriendly},
		{500, ReputationLevelNeutral},
		{0, ReputationLevelNeutral},
		{-500, ReputationLevelNeutral},
		{-501, ReputationLevelUnfriendly},
		{-2500, ReputationLevelUnfriendly},
		{-2501, ReputationLevelHostile},
		{-5000, ReputationLevelHostile},
		{-5001, ReputationLevelHated},
		{-7500, ReputationLevelHated},
		{-7501, ReputationLevelDespised},
		{-10000, ReputationLevelDespised},
	}

	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			level := rs.calculateReputationLevel(tt.score)
			assert.Equal(t, tt.expected, level)
		})
	}
}

func TestCalculateOverallRank(t *testing.T) {
	rs := NewReputationSystem(logrus.New())

	tests := []struct {
		name         string
		totalScore   int64
		factionCount int
		expected     ReputationRank
	}{
		{
			name:         "legendary rank",
			totalScore:   60000,
			factionCount: 10,
			expected:     ReputationRankLegendary,
		},
		{
			name:         "renowned rank",
			totalScore:   30000,
			factionCount: 6,
			expected:     ReputationRankRenowned,
		},
		{
			name:         "neutral rank",
			totalScore:   1000,
			factionCount: 5,
			expected:     ReputationRankNeutral,
		},
		{
			name:         "infamous rank",
			totalScore:   -60000,
			factionCount: 8,
			expected:     ReputationRankInfamous,
		},
		{
			name:         "zero factions",
			totalScore:   10000,
			factionCount: 0,
			expected:     ReputationRankNeutral,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rank := rs.calculateOverallRank(tt.totalScore, tt.factionCount)
			assert.Equal(t, tt.expected, rank)
		})
	}
}

func TestCalculateEffect(t *testing.T) {
	rs := NewReputationSystem(logrus.New())
	playerID := "test_player"
	factionID := "test_faction"

	// Initialize player reputation
	factionSystem := &GeneratedFactionSystem{
		Factions: []*Faction{
			{ID: factionID, Name: "Test Faction"},
		},
	}
	require.NoError(t, rs.InitializePlayerReputation(playerID, factionSystem))

	// Set specific reputation level
	require.NoError(t, rs.ModifyReputation(playerID, factionID, 3000, "test", ReputationActionQuest))

	tests := []struct {
		name       string
		effectType ReputationEffectType
		expectSign float64 // positive, negative, or zero
	}{
		{
			name:       "price discount for honored reputation",
			effectType: ReputationEffectPriceDiscount,
			expectSign: 1, // positive (beneficial)
		},
		{
			name:       "quest reward for honored reputation",
			effectType: ReputationEffectQuestReward,
			expectSign: 1, // positive (beneficial)
		},
		{
			name:       "combat assistance for honored reputation",
			effectType: ReputationEffectCombatAssistance,
			expectSign: 1, // positive (Honored level should provide combat assistance)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effect, err := rs.CalculateEffect(playerID, factionID, tt.effectType)
			require.NoError(t, err)

			if tt.expectSign > 0 {
				assert.Positive(t, effect)
			} else if tt.expectSign < 0 {
				assert.Negative(t, effect)
			} else {
				assert.Equal(t, 0.0, effect)
			}
		})
	}
}

func TestReputationLocking(t *testing.T) {
	rs := NewReputationSystem(logrus.New())
	playerID := "test_player"
	factionID := "test_faction"

	// Initialize player reputation
	factionSystem := &GeneratedFactionSystem{
		Factions: []*Faction{
			{ID: factionID, Name: "Test Faction"},
		},
	}
	require.NoError(t, rs.InitializePlayerReputation(playerID, factionSystem))

	// Lock reputation
	standing := rs.PlayerReputations[playerID].FactionStandings[factionID]
	standing.IsLocked = true
	standing.LockReason = "special quest requirement"

	// Try to modify locked reputation
	err := rs.ModifyReputation(playerID, factionID, 1000, "test", ReputationActionQuest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reputation locked")

	// Verify reputation didn't change
	finalStanding, err := rs.GetReputation(playerID, factionID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), finalStanding.ReputationScore)
}

func TestApplyDecay(t *testing.T) {
	rs := NewReputationSystem(logrus.New())
	playerID := "test_player"
	factionID := "test_faction"

	// Initialize player reputation
	factionSystem := &GeneratedFactionSystem{
		Factions: []*Faction{
			{ID: factionID, Name: "Test Faction"},
		},
	}
	require.NoError(t, rs.InitializePlayerReputation(playerID, factionSystem))

	// Set high reputation
	require.NoError(t, rs.ModifyReputation(playerID, factionID, 5000, "test", ReputationActionQuest))

	// Manually age the last interaction to bypass activity protection
	standing := rs.PlayerReputations[playerID].FactionStandings[factionID]
	standing.LastInteraction = time.Now().Add(-4 * 24 * time.Hour) // 4 days ago

	initialScore := standing.ReputationScore

	// Apply decay
	err := rs.ApplyDecay()
	require.NoError(t, err)

	// Check that reputation decreased
	finalStanding, err := rs.GetReputation(playerID, factionID)
	require.NoError(t, err)
	assert.Less(t, finalStanding.ReputationScore, initialScore)

	// Check that decay event was recorded
	events, err := rs.GetReputationHistory(playerID, factionID, 10)
	require.NoError(t, err)

	var decayEvent *ReputationEvent
	for _, event := range events {
		if event.Reason == "time decay" {
			decayEvent = event
			break
		}
	}

	require.NotNil(t, decayEvent, "decay event should have been recorded")
	assert.Negative(t, decayEvent.Change)
}

func TestFactionInfluence(t *testing.T) {
	rs := NewReputationSystem(logrus.New())
	playerID := "test_player"
	faction1ID := "faction1"
	faction2ID := "faction2"
	faction3ID := "faction3"

	// Initialize player reputation with allied and enemy factions
	factionSystem := &GeneratedFactionSystem{
		Factions: []*Faction{
			{ID: faction1ID, Name: "Faction One"},
			{ID: faction2ID, Name: "Faction Two"},
			{ID: faction3ID, Name: "Faction Three"},
		},
	}
	require.NoError(t, rs.InitializePlayerReputation(playerID, factionSystem))

	// Set up faction relationships
	faction1Rep := rs.getFactionReputation(faction1ID)
	faction1Rep.AlliedFactions = []string{faction2ID}
	faction1Rep.EnemyFactions = []string{faction3ID}

	initialScore2 := rs.PlayerReputations[playerID].FactionStandings[faction2ID].ReputationScore
	initialScore3 := rs.PlayerReputations[playerID].FactionStandings[faction3ID].ReputationScore

	// Modify reputation with faction1
	require.NoError(t, rs.ModifyReputation(playerID, faction1ID, 1000, "test", ReputationActionQuest))

	// Check that allied faction gained reputation
	finalScore2 := rs.PlayerReputations[playerID].FactionStandings[faction2ID].ReputationScore
	assert.Greater(t, finalScore2, initialScore2)

	// Check that enemy faction lost reputation
	finalScore3 := rs.PlayerReputations[playerID].FactionStandings[faction3ID].ReputationScore
	assert.Less(t, finalScore3, initialScore3)
}

func TestGetReputationHistory(t *testing.T) {
	rs := NewReputationSystem(logrus.New())
	playerID := "test_player"
	factionID := "test_faction"

	// Initialize player reputation
	factionSystem := &GeneratedFactionSystem{
		Factions: []*Faction{
			{ID: factionID, Name: "Test Faction"},
		},
	}
	require.NoError(t, rs.InitializePlayerReputation(playerID, factionSystem))

	// Make several reputation changes
	changes := []struct {
		change int64
		reason string
	}{
		{1000, "first quest"},
		{500, "second quest"},
		{-200, "minor offense"},
	}

	for _, c := range changes {
		require.NoError(t, rs.ModifyReputation(playerID, factionID, c.change, c.reason, ReputationActionQuest))
	}

	// Get limited history
	events, err := rs.GetReputationHistory(playerID, factionID, 2)
	require.NoError(t, err)
	assert.Len(t, events, 2)

	// Events should be in reverse chronological order
	assert.Equal(t, "minor offense", events[0].Reason)
	assert.Equal(t, "second quest", events[1].Reason)

	// Get all history
	allEvents, err := rs.GetReputationHistory(playerID, factionID, 10)
	require.NoError(t, err)
	assert.Len(t, allEvents, 3)

	// Get history for all factions
	allFactionsEvents, err := rs.GetReputationHistory(playerID, "", 10)
	require.NoError(t, err)
	assert.Len(t, allFactionsEvents, 3) // Should be same as faction-specific since we only have one faction
}

// Benchmark tests for performance validation

func BenchmarkModifyReputation(b *testing.B) {
	rs := NewReputationSystem(logrus.New())
	playerID := "test_player"
	factionID := "test_faction"

	factionSystem := &GeneratedFactionSystem{
		Factions: []*Faction{
			{ID: factionID, Name: "Test Faction"},
		},
	}
	rs.InitializePlayerReputation(playerID, factionSystem)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		change := int64(i%100 - 50) // Varies between -50 and +49
		rs.ModifyReputation(playerID, factionID, change, "benchmark", ReputationActionQuest)
	}
}

func BenchmarkCalculateEffect(b *testing.B) {
	rs := NewReputationSystem(logrus.New())
	playerID := "test_player"
	factionID := "test_faction"

	factionSystem := &GeneratedFactionSystem{
		Factions: []*Faction{
			{ID: factionID, Name: "Test Faction"},
		},
	}
	rs.InitializePlayerReputation(playerID, factionSystem)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rs.CalculateEffect(playerID, factionID, ReputationEffectPriceDiscount)
	}
}

func BenchmarkApplyDecay(b *testing.B) {
	rs := NewReputationSystem(logrus.New())

	// Create multiple players with multiple factions for realistic scenario
	factionSystem := &GeneratedFactionSystem{
		Factions: []*Faction{
			{ID: "faction1", Name: "Faction One"},
			{ID: "faction2", Name: "Faction Two"},
			{ID: "faction3", Name: "Faction Three"},
		},
	}

	for i := 0; i < 100; i++ {
		playerID := fmt.Sprintf("player_%d", i)
		rs.InitializePlayerReputation(playerID, factionSystem)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rs.ApplyDecay()
	}
}
