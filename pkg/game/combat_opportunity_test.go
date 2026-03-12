package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOpportunityAttackManager(t *testing.T) {
	world := CreateDefaultWorld()
	oam := NewOpportunityAttackManager(world)

	assert.NotNil(t, oam)
	assert.Equal(t, world, oam.world)
	assert.NotNil(t, oam.reactionUsed)
	assert.NotNil(t, oam.threatenedBy)
	assert.NotNil(t, oam.entityPositions)
}

func TestOpportunityAttackManager_RegisterEntity(t *testing.T) {
	world := CreateDefaultWorld()
	oam := NewOpportunityAttackManager(world)

	pos := Position{X: 5, Y: 5, Level: 0}
	oam.RegisterEntity("entity1", pos)

	assert.Equal(t, pos, oam.entityPositions["entity1"])

	// Check that adjacent squares are threatened
	threatKey := positionKey(Position{X: 4, Y: 5})
	assert.Contains(t, oam.threatenedBy[threatKey], "entity1")
}

func TestOpportunityAttackManager_UnregisterEntity(t *testing.T) {
	world := CreateDefaultWorld()
	oam := NewOpportunityAttackManager(world)

	pos := Position{X: 5, Y: 5, Level: 0}
	oam.RegisterEntity("entity1", pos)
	oam.UnregisterEntity("entity1")

	_, exists := oam.entityPositions["entity1"]
	assert.False(t, exists)
}

func TestOpportunityAttackManager_CheckMovement_NoAttacks(t *testing.T) {
	world := CreateDefaultWorld()
	oam := NewOpportunityAttackManager(world)

	// Register mover only - no one to make opportunity attacks
	oam.RegisterEntity("mover", Position{X: 5, Y: 5})

	attacks := oam.CheckMovement("mover", Position{X: 5, Y: 5}, Position{X: 6, Y: 5}, false)
	assert.Empty(t, attacks)
}

func TestOpportunityAttackManager_CheckMovement_WithDisengage(t *testing.T) {
	world := CreateDefaultWorld()
	oam := NewOpportunityAttackManager(world)

	// Register attacker and mover
	oam.RegisterEntity("attacker", Position{X: 5, Y: 5})
	oam.RegisterEntity("mover", Position{X: 5, Y: 6})

	// Move away with disengage - no opportunity attack
	attacks := oam.CheckMovement("mover", Position{X: 5, Y: 6}, Position{X: 5, Y: 8}, true)
	assert.Empty(t, attacks)
}

func TestOpportunityAttackManager_CheckMovement_TriggerAttack(t *testing.T) {
	world := CreateDefaultWorld()
	oam := NewOpportunityAttackManager(world)

	// Register attacker adjacent to mover
	oam.RegisterEntity("attacker", Position{X: 5, Y: 5})
	oam.RegisterEntity("mover", Position{X: 5, Y: 6})

	// Move away from attacker without disengage - triggers opportunity attack
	attacks := oam.CheckMovement("mover", Position{X: 5, Y: 6}, Position{X: 5, Y: 8}, false)

	assert.Len(t, attacks, 1)
	assert.Equal(t, "attacker", attacks[0].AttackerID)
	assert.Equal(t, "mover", attacks[0].TargetID)
}

func TestOpportunityAttackManager_CheckMovement_StillAdjacent(t *testing.T) {
	world := CreateDefaultWorld()
	oam := NewOpportunityAttackManager(world)

	// Register attacker adjacent to mover
	oam.RegisterEntity("attacker", Position{X: 5, Y: 5})
	oam.RegisterEntity("mover", Position{X: 5, Y: 6})

	// Move but stay adjacent - no opportunity attack
	attacks := oam.CheckMovement("mover", Position{X: 5, Y: 6}, Position{X: 6, Y: 5}, false)
	assert.Empty(t, attacks)
}

func TestOpportunityAttackManager_UseReaction(t *testing.T) {
	world := CreateDefaultWorld()
	oam := NewOpportunityAttackManager(world)

	oam.RegisterEntity("entity1", Position{X: 5, Y: 5})
	assert.True(t, oam.HasReaction("entity1"))

	oam.UseReaction("entity1")
	assert.False(t, oam.HasReaction("entity1"))
}

func TestOpportunityAttackManager_ResetReactions(t *testing.T) {
	world := CreateDefaultWorld()
	oam := NewOpportunityAttackManager(world)

	oam.RegisterEntity("entity1", Position{X: 5, Y: 5})
	oam.UseReaction("entity1")
	assert.False(t, oam.HasReaction("entity1"))

	oam.ResetReactions()
	assert.True(t, oam.HasReaction("entity1"))
}

func TestOpportunityAttackManager_UpdatePosition(t *testing.T) {
	world := CreateDefaultWorld()
	oam := NewOpportunityAttackManager(world)

	oam.RegisterEntity("entity1", Position{X: 5, Y: 5})

	newPos := Position{X: 7, Y: 7}
	oam.UpdatePosition("entity1", newPos)

	assert.Equal(t, newPos, oam.entityPositions["entity1"])
}

func TestOpportunityAttackManager_GetThreateningEntities(t *testing.T) {
	world := CreateDefaultWorld()
	oam := NewOpportunityAttackManager(world)

	oam.RegisterEntity("entity1", Position{X: 5, Y: 5})
	oam.RegisterEntity("entity2", Position{X: 5, Y: 4})

	// Position {5, 5} is threatened by entity2 (adjacent at {5,4})
	threats := oam.GetThreateningEntities(Position{X: 5, Y: 5})
	assert.Contains(t, threats, "entity2")
}

func TestOpportunityAttackManager_ReactionUsedPreventsAttack(t *testing.T) {
	world := CreateDefaultWorld()
	oam := NewOpportunityAttackManager(world)

	oam.RegisterEntity("attacker", Position{X: 5, Y: 5})
	oam.RegisterEntity("mover", Position{X: 5, Y: 6})

	// Use attacker's reaction
	oam.UseReaction("attacker")

	// Move away - no opportunity attack because reaction used
	attacks := oam.CheckMovement("mover", Position{X: 5, Y: 6}, Position{X: 5, Y: 8}, false)
	assert.Empty(t, attacks)
}

func TestIsAdjacent(t *testing.T) {
	tests := []struct {
		name     string
		a        Position
		b        Position
		expected bool
	}{
		{"horizontally adjacent", Position{X: 1, Y: 1}, Position{X: 2, Y: 1}, true},
		{"vertically adjacent", Position{X: 1, Y: 1}, Position{X: 1, Y: 2}, true},
		{"diagonally adjacent", Position{X: 1, Y: 1}, Position{X: 2, Y: 2}, true},
		{"same position", Position{X: 1, Y: 1}, Position{X: 1, Y: 1}, false},
		{"different levels", Position{X: 1, Y: 1, Level: 0}, Position{X: 2, Y: 1, Level: 1}, false},
		{"too far", Position{X: 1, Y: 1}, Position{X: 4, Y: 1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAdjacent(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAdjacentPositions(t *testing.T) {
	positions := getAdjacentPositions(Position{X: 5, Y: 5, Level: 0})

	assert.Equal(t, 8, len(positions))

	// Verify all 8 adjacent positions are returned
	expected := map[Position]bool{
		{X: 4, Y: 4, Level: 0}: true,
		{X: 5, Y: 4, Level: 0}: true,
		{X: 6, Y: 4, Level: 0}: true,
		{X: 4, Y: 5, Level: 0}: true,
		{X: 6, Y: 5, Level: 0}: true,
		{X: 4, Y: 6, Level: 0}: true,
		{X: 5, Y: 6, Level: 0}: true,
		{X: 6, Y: 6, Level: 0}: true,
	}

	for _, pos := range positions {
		assert.True(t, expected[pos], "unexpected position: %v", pos)
	}
}

func TestPositionKey(t *testing.T) {
	pos1 := Position{X: 1, Y: 2, Level: 3}
	pos2 := Position{X: 1, Y: 2, Level: 3}
	pos3 := Position{X: 4, Y: 5, Level: 6}

	key1 := positionKey(pos1)
	key2 := positionKey(pos2)
	key3 := positionKey(pos3)

	assert.Equal(t, key1, key2, "same positions should have same keys")
	assert.NotEqual(t, key1, key3, "different positions should have different keys")
}

func TestResolveOpportunityAttack(t *testing.T) {
	attacker := &Character{
		ID:         "attacker",
		Strength:   14, // +2 modifier
		HP:         20,
		MaxHP:      20,
		ArmorClass: 10,
	}

	target := &Character{
		ID:         "target",
		Strength:   10,
		HP:         20,
		MaxHP:      20,
		ArmorClass: 10,
	}

	// Run multiple times to test hit/miss scenarios
	hitCount := 0
	for i := 0; i < 100; i++ {
		hit, damage := ResolveOpportunityAttack(attacker, target)
		if hit {
			hitCount++
			assert.Greater(t, damage, 0, "damage should be positive on hit")
		}
	}

	// With +2 strength mod attacking AC 10, should hit frequently
	assert.Greater(t, hitCount, 30, "should hit at least 30% of the time")
	assert.Less(t, hitCount, 90, "should miss sometimes")
}

func TestResolveOpportunityAttack_NilCharacters(t *testing.T) {
	hit, damage := ResolveOpportunityAttack(nil, nil)
	assert.False(t, hit)
	assert.Equal(t, 0, damage)

	attacker := &Character{ID: "attacker", Strength: 10}
	hit, damage = ResolveOpportunityAttack(attacker, nil)
	assert.False(t, hit)
	assert.Equal(t, 0, damage)

	target := &Character{ID: "target", ArmorClass: 10}
	hit, damage = ResolveOpportunityAttack(nil, target)
	assert.False(t, hit)
	assert.Equal(t, 0, damage)
}
