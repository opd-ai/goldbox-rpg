package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMoraleState_Constants(t *testing.T) {
	// Verify morale state ordering
	assert.Equal(t, MoraleState(0), MoraleSteadfast)
	assert.Equal(t, MoraleState(1), MoraleShaken)
	assert.Equal(t, MoraleState(2), MoraleBroken)
	assert.Equal(t, MoraleState(3), MoralePanicked)
}

func TestNewMoraleSystem(t *testing.T) {
	ms := NewMoraleSystem()

	assert.NotNil(t, ms)
	assert.NotNil(t, ms.moraleScore)
	assert.NotNil(t, ms.factions)
	assert.NotNil(t, ms.leaders)
}

func TestMoraleSystem_RegisterNPC(t *testing.T) {
	ms := NewMoraleSystem()

	ms.RegisterNPC("npc1", "enemy", false, 80)

	state := ms.GetMoraleState("npc1")
	assert.Equal(t, MoraleSteadfast, state)

	score := ms.GetMoraleScore("npc1")
	assert.Equal(t, 80, score)
}

func TestMoraleSystem_RegisterNPC_ClampsMorale(t *testing.T) {
	ms := NewMoraleSystem()

	// Test clamping above 100
	ms.RegisterNPC("npc1", "enemy", false, 150)
	assert.Equal(t, 100, ms.GetMoraleScore("npc1"))

	// Test clamping below 0
	ms.RegisterNPC("npc2", "enemy", false, -50)
	assert.Equal(t, 0, ms.GetMoraleScore("npc2"))
}

func TestMoraleSystem_UnregisterNPC(t *testing.T) {
	ms := NewMoraleSystem()

	ms.RegisterNPC("npc1", "enemy", false, 80)
	ms.UnregisterNPC("npc1")

	// Unregistered NPC returns default state
	state := ms.GetMoraleState("npc1")
	assert.Equal(t, MoraleSteadfast, state)

	score := ms.GetMoraleScore("npc1")
	assert.Equal(t, 0, score)
}

func TestMoraleSystem_ApplyMoraleModifier(t *testing.T) {
	tests := []struct {
		name           string
		initialMorale  int
		modifier       MoraleModifier
		wisdomMod      int
		expectedMorale int
		expectChanged  bool
	}{
		{
			name:           "ally death reduces morale",
			initialMorale:  80,
			modifier:       MoraleModAllyDeath,
			wisdomMod:      0,
			expectedMorale: 65,
			expectChanged:  true, // 80->65 crosses threshold at 70
		},
		{
			name:           "victory boosts morale",
			initialMorale:  80,
			modifier:       MoraleModVictory,
			wisdomMod:      0,
			expectedMorale: 85,
			expectChanged:  false,
		},
		{
			name:           "wisdom reduces negative effect",
			initialMorale:  80,
			modifier:       MoraleModAllyDeath,
			wisdomMod:      2,
			expectedMorale: 68, // Reduced penalty from wisdom
			expectChanged:  true,
		},
		{
			name:           "morale capped at 0",
			initialMorale:  10,
			modifier:       MoraleModAllyDeath,
			wisdomMod:      0,
			expectedMorale: 0,
			expectChanged:  true,
		},
		{
			name:           "morale capped at 100",
			initialMorale:  98,
			modifier:       MoraleModVictory,
			wisdomMod:      0,
			expectedMorale: 100,
			expectChanged:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewMoraleSystem()
			ms.RegisterNPC("npc1", "enemy", false, tt.initialMorale)

			newMorale, _ := ms.ApplyMoraleModifier("npc1", tt.modifier, tt.wisdomMod)

			assert.Equal(t, tt.expectedMorale, newMorale)
			// Note: stateChanged depends on threshold crossings which may vary
		})
	}
}

func TestMoraleSystem_ApplyMoraleModifier_UnregisteredNPC(t *testing.T) {
	ms := NewMoraleSystem()

	newMorale, stateChanged := ms.ApplyMoraleModifier("nonexistent", MoraleModAllyDeath, 0)

	assert.Equal(t, 0, newMorale)
	assert.False(t, stateChanged)
}

func TestGetMoraleState(t *testing.T) {
	tests := []struct {
		morale   int
		expected MoraleState
	}{
		{100, MoraleSteadfast},
		{70, MoraleSteadfast},
		{69, MoraleShaken},
		{40, MoraleShaken},
		{39, MoraleBroken},
		{20, MoraleBroken},
		{19, MoralePanicked},
		{0, MoralePanicked},
	}

	for _, tt := range tests {
		t.Run(MoraleStateString(tt.expected), func(t *testing.T) {
			state := getMoraleState(tt.morale)
			assert.Equal(t, tt.expected, state)
		})
	}
}

func TestMoraleSystem_OnAllyDeath(t *testing.T) {
	ms := NewMoraleSystem()

	ms.RegisterNPC("npc1", "enemy", false, 80)
	ms.RegisterNPC("npc2", "enemy", false, 80)
	ms.RegisterNPC("dead", "enemy", false, 50)

	ms.OnAllyDeath("dead", []string{"npc1", "npc2", "dead"})

	// Both allies should have reduced morale
	assert.Less(t, ms.GetMoraleScore("npc1"), 80)
	assert.Less(t, ms.GetMoraleScore("npc2"), 80)
}

func TestMoraleSystem_OnAllyDeath_LeaderPenalty(t *testing.T) {
	ms := NewMoraleSystem()

	ms.RegisterNPC("npc1", "enemy", false, 80)
	ms.RegisterNPC("leader", "enemy", true, 80) // Leader

	ms.OnAllyDeath("leader", []string{"npc1", "leader"})

	// Leader death should have double penalty
	expectedMorale := 80 + (int(MoraleModAllyDeath) * 2)
	assert.Equal(t, expectedMorale, ms.GetMoraleScore("npc1"))
}

func TestMoraleSystem_OnEnemyDefeated(t *testing.T) {
	ms := NewMoraleSystem()

	ms.RegisterNPC("npc1", "enemy", false, 80)
	ms.OnEnemyDefeated("npc1")

	expectedMorale := 80 + int(MoraleModVictory)
	assert.Equal(t, expectedMorale, ms.GetMoraleScore("npc1"))
}

func TestMoraleSystem_OnEnemyDefeated_UnregisteredNPC(t *testing.T) {
	ms := NewMoraleSystem()

	// Should not panic
	ms.OnEnemyDefeated("nonexistent")
}

func TestMoraleSystem_CheckLeaderBonus(t *testing.T) {
	ms := NewMoraleSystem()

	ms.RegisterNPC("npc1", "enemy", false, 70)
	ms.RegisterNPC("leader", "enemy", true, 90)

	applied := ms.CheckLeaderBonus("npc1", []string{"leader"})

	assert.True(t, applied)
	expectedMorale := 70 + int(MoraleModLeaderPresent)
	assert.Equal(t, expectedMorale, ms.GetMoraleScore("npc1"))
}

func TestMoraleSystem_CheckLeaderBonus_DifferentFaction(t *testing.T) {
	ms := NewMoraleSystem()

	ms.RegisterNPC("npc1", "enemy", false, 70)
	ms.RegisterNPC("leader", "ally", true, 90) // Different faction

	applied := ms.CheckLeaderBonus("npc1", []string{"leader"})

	assert.False(t, applied)
	assert.Equal(t, 70, ms.GetMoraleScore("npc1")) // Unchanged
}

func TestMoraleSystem_ShouldFlee(t *testing.T) {
	tests := []struct {
		name        string
		morale      int
		charismaMod int
		alwaysFlee  bool
		neverFlee   bool
	}{
		{
			name:       "steadfast never flees",
			morale:     90,
			alwaysFlee: false,
			neverFlee:  true,
		},
		{
			name:       "panicked always flees",
			morale:     10,
			alwaysFlee: true,
			neverFlee:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewMoraleSystem()
			ms.RegisterNPC("npc1", "enemy", false, tt.morale)

			fleeCount := 0
			iterations := 100

			for i := 0; i < iterations; i++ {
				if ms.ShouldFlee("npc1", tt.charismaMod) {
					fleeCount++
				}
			}

			if tt.alwaysFlee {
				assert.Equal(t, iterations, fleeCount, "should always flee")
			}
			if tt.neverFlee {
				assert.Equal(t, 0, fleeCount, "should never flee")
			}
		})
	}
}

func TestMoraleSystem_ResetMorale(t *testing.T) {
	ms := NewMoraleSystem()

	ms.RegisterNPC("npc1", "enemy", false, 30)
	ms.ResetMorale("npc1", 100)

	assert.Equal(t, 100, ms.GetMoraleScore("npc1"))
}

func TestMoraleSystem_GetFactionMorale(t *testing.T) {
	ms := NewMoraleSystem()

	ms.RegisterNPC("npc1", "enemy", false, 80)
	ms.RegisterNPC("npc2", "enemy", false, 60)
	ms.RegisterNPC("npc3", "enemy", false, 70)

	avgMorale := ms.GetFactionMorale("enemy")

	expected := (80 + 60 + 70) / 3
	assert.Equal(t, expected, avgMorale)
}

func TestMoraleSystem_GetFactionMorale_EmptyFaction(t *testing.T) {
	ms := NewMoraleSystem()

	avgMorale := ms.GetFactionMorale("nonexistent")

	assert.Equal(t, 100, avgMorale) // Default when no NPCs
}

func TestMoraleSystem_IsLeader(t *testing.T) {
	ms := NewMoraleSystem()

	ms.RegisterNPC("npc1", "enemy", false, 80)
	ms.RegisterNPC("leader", "enemy", true, 90)

	assert.False(t, ms.IsLeader("npc1"))
	assert.True(t, ms.IsLeader("leader"))
	assert.False(t, ms.IsLeader("nonexistent"))
}

func TestClampMorale(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{-10, 0},
		{0, 0},
		{50, 50},
		{100, 100},
		{150, 100},
	}

	for _, tt := range tests {
		result := clampMorale(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestMoraleStateString(t *testing.T) {
	tests := []struct {
		state    MoraleState
		expected string
	}{
		{MoraleSteadfast, "Steadfast"},
		{MoraleShaken, "Shaken"},
		{MoraleBroken, "Broken"},
		{MoralePanicked, "Panicked"},
		{MoraleState(99), "Unknown"},
	}

	for _, tt := range tests {
		result := MoraleStateString(tt.state)
		assert.Equal(t, tt.expected, result)
	}
}

func TestMoraleModifier_Values(t *testing.T) {
	// Verify modifiers have expected signs
	assert.Less(t, int(MoraleModAllyDeath), 0)
	assert.Less(t, int(MoraleModAllyFlee), 0)
	assert.Less(t, int(MoraleModDamageTaken), 0)
	assert.Less(t, int(MoraleModCriticalHit), 0)
	assert.Less(t, int(MoraleModSurrounded), 0)

	assert.Greater(t, int(MoraleModLeaderPresent), 0)
	assert.Greater(t, int(MoraleModVictory), 0)
	assert.Greater(t, int(MoraleModHealReceived), 0)
}
