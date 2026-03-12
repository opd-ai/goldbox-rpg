package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoverBonus(t *testing.T) {
	tests := []struct {
		name     string
		cover    CoverType
		expected int
	}{
		{"no cover", CoverNone, 0},
		{"half cover", CoverHalf, 2},
		{"three quarters cover", CoverThreeQuarters, 5},
		{"full cover", CoverFull, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bonus := CoverBonus(tt.cover)
			assert.Equal(t, tt.expected, bonus)
		})
	}
}

func TestNewCombatModifiers(t *testing.T) {
	world := CreateDefaultWorld()
	cm := NewCombatModifiers(world)

	assert.NotNil(t, cm)
	assert.Equal(t, world, cm.world)
}

func TestCombatModifiers_CalculateCover(t *testing.T) {
	world := CreateDefaultWorld()
	cm := NewCombatModifiers(world)

	tests := []struct {
		name     string
		attacker Position
		defender Position
		expected CoverType
	}{
		{
			name:     "adjacent positions no cover",
			attacker: Position{X: 1, Y: 1},
			defender: Position{X: 2, Y: 1},
			expected: CoverNone,
		},
		{
			name:     "same position",
			attacker: Position{X: 1, Y: 1},
			defender: Position{X: 1, Y: 1},
			expected: CoverNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cover := cm.CalculateCover(tt.attacker, tt.defender)
			assert.Equal(t, tt.expected, cover)
		})
	}
}

func TestCombatModifiers_CalculateCover_NilWorld(t *testing.T) {
	cm := NewCombatModifiers(nil)
	cover := cm.CalculateCover(Position{X: 0, Y: 0}, Position{X: 5, Y: 5})
	assert.Equal(t, CoverNone, cover)
}

func TestCombatModifiers_isAdjacent(t *testing.T) {
	cm := &CombatModifiers{}

	tests := []struct {
		name     string
		a        Position
		b        Position
		expected bool
	}{
		{
			name:     "horizontally adjacent",
			a:        Position{X: 1, Y: 1},
			b:        Position{X: 2, Y: 1},
			expected: true,
		},
		{
			name:     "vertically adjacent",
			a:        Position{X: 1, Y: 1},
			b:        Position{X: 1, Y: 2},
			expected: true,
		},
		{
			name:     "diagonally adjacent",
			a:        Position{X: 1, Y: 1},
			b:        Position{X: 2, Y: 2},
			expected: true,
		},
		{
			name:     "same position",
			a:        Position{X: 1, Y: 1},
			b:        Position{X: 1, Y: 1},
			expected: false,
		},
		{
			name:     "too far horizontally",
			a:        Position{X: 1, Y: 1},
			b:        Position{X: 3, Y: 1},
			expected: false,
		},
		{
			name:     "too far diagonally",
			a:        Position{X: 1, Y: 1},
			b:        Position{X: 3, Y: 3},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cm.isAdjacent(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCombatModifiers_isOpposite(t *testing.T) {
	cm := &CombatModifiers{}

	tests := []struct {
		name     string
		attacker Position
		defender Position
		ally     Position
		expected bool
	}{
		{
			name:     "opposite sides horizontally",
			attacker: Position{X: 0, Y: 1},
			defender: Position{X: 1, Y: 1},
			ally:     Position{X: 2, Y: 1},
			expected: true,
		},
		{
			name:     "opposite sides vertically",
			attacker: Position{X: 1, Y: 0},
			defender: Position{X: 1, Y: 1},
			ally:     Position{X: 1, Y: 2},
			expected: true,
		},
		{
			name:     "opposite sides diagonally",
			attacker: Position{X: 0, Y: 0},
			defender: Position{X: 1, Y: 1},
			ally:     Position{X: 2, Y: 2},
			expected: true,
		},
		{
			name:     "same side",
			attacker: Position{X: 0, Y: 1},
			defender: Position{X: 1, Y: 1},
			ally:     Position{X: 0, Y: 2},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cm.isOpposite(tt.attacker, tt.defender, tt.ally)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCombatModifiers_CalculateFlanking(t *testing.T) {
	world := CreateDefaultWorld()
	cm := NewCombatModifiers(world)

	tests := []struct {
		name        string
		attacker    Position
		defender    Position
		allies      []Position
		expectFlank bool
		expectBonus int
	}{
		{
			name:        "no allies",
			attacker:    Position{X: 1, Y: 1},
			defender:    Position{X: 2, Y: 1},
			allies:      []Position{},
			expectFlank: false,
			expectBonus: 0,
		},
		{
			name:        "ally on opposite side",
			attacker:    Position{X: 1, Y: 1},
			defender:    Position{X: 2, Y: 1},
			allies:      []Position{{X: 3, Y: 1}},
			expectFlank: true,
			expectBonus: 2,
		},
		{
			name:        "ally not adjacent to defender",
			attacker:    Position{X: 1, Y: 1},
			defender:    Position{X: 2, Y: 1},
			allies:      []Position{{X: 5, Y: 1}},
			expectFlank: false,
			expectBonus: 0,
		},
		{
			name:        "attacker not adjacent to defender",
			attacker:    Position{X: 0, Y: 1},
			defender:    Position{X: 3, Y: 1},
			allies:      []Position{{X: 4, Y: 1}},
			expectFlank: false,
			expectBonus: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isFlank, bonus := cm.CalculateFlanking(tt.attacker, tt.defender, tt.allies)
			assert.Equal(t, tt.expectFlank, isFlank)
			assert.Equal(t, tt.expectBonus, bonus)
		})
	}
}

func TestCombatModifiers_GetCombatModifiers(t *testing.T) {
	world := CreateDefaultWorld()
	cm := NewCombatModifiers(world)

	// Test basic modifiers
	attackBonus, defenseBonus := cm.GetCombatModifiers(
		Position{X: 1, Y: 1},
		Position{X: 2, Y: 1},
		[]Position{{X: 3, Y: 1}}, // Ally for flanking
	)

	// Should have flanking bonus
	assert.Equal(t, 2, attackBonus)
	// No cover between adjacent tiles
	assert.Equal(t, 0, defenseBonus)
}

func TestCombatModifiers_GetAdjacentPositions(t *testing.T) {
	cm := &CombatModifiers{}

	positions := cm.GetAdjacentPositions(Position{X: 5, Y: 5, Level: 0})

	assert.Equal(t, 8, len(positions))

	// Verify all adjacent positions are returned
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

func TestCombatModifiers_getLinePoints(t *testing.T) {
	cm := &CombatModifiers{}

	points := cm.getLinePoints(Position{X: 0, Y: 0}, Position{X: 3, Y: 0})

	assert.Equal(t, 4, len(points))
	assert.Equal(t, Position{X: 0, Y: 0}, points[0])
	assert.Equal(t, Position{X: 3, Y: 0}, points[len(points)-1])
}

func TestAbsInt(t *testing.T) {
	assert.Equal(t, 5, absInt(5))
	assert.Equal(t, 5, absInt(-5))
	assert.Equal(t, 0, absInt(0))
}

func TestDistanceBetween(t *testing.T) {
	tests := []struct {
		name     string
		a        Position
		b        Position
		expected float64
	}{
		{
			name:     "same position",
			a:        Position{X: 0, Y: 0},
			b:        Position{X: 0, Y: 0},
			expected: 0,
		},
		{
			name:     "horizontal distance",
			a:        Position{X: 0, Y: 0},
			b:        Position{X: 3, Y: 0},
			expected: 3,
		},
		{
			name:     "vertical distance",
			a:        Position{X: 0, Y: 0},
			b:        Position{X: 0, Y: 4},
			expected: 4,
		},
		{
			name:     "diagonal distance",
			a:        Position{X: 0, Y: 0},
			b:        Position{X: 3, Y: 4},
			expected: 5, // 3-4-5 triangle
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := DistanceBetween(tt.a, tt.b)
			assert.InDelta(t, tt.expected, dist, 0.001)
		})
	}
}

func TestCombatModifiers_HighGroundBonus(t *testing.T) {
	cm := &CombatModifiers{}

	// Currently returns 0 as elevation not yet implemented
	bonus := cm.HighGroundBonus(Position{X: 0, Y: 0}, Position{X: 1, Y: 1})
	assert.Equal(t, 0, bonus)
}
