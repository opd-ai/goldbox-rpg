package server

import (
	"testing"

	"goldbox-rpg/pkg/game"

	"github.com/stretchr/testify/assert"
)

// TestRollDice tests the dice rolling function
func TestRollDice(t *testing.T) {
	tests := []struct {
		name     string
		numDice  int
		dieSize  int
		minValue int
		maxValue int
	}{
		{
			name:     "1d6",
			numDice:  1,
			dieSize:  6,
			minValue: 1,
			maxValue: 6,
		},
		{
			name:     "2d10",
			numDice:  2,
			dieSize:  10,
			minValue: 2,
			maxValue: 20,
		},
		{
			name:     "3d4",
			numDice:  3,
			dieSize:  4,
			minValue: 3,
			maxValue: 12,
		},
		{
			name:     "zero dice returns 0",
			numDice:  0,
			dieSize:  6,
			minValue: 0,
			maxValue: 0,
		},
		{
			name:     "negative dice returns 0",
			numDice:  -1,
			dieSize:  6,
			minValue: 0,
			maxValue: 0,
		},
		{
			name:     "zero die size returns 0",
			numDice:  2,
			dieSize:  0,
			minValue: 0,
			maxValue: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rollDice(tt.numDice, tt.dieSize)
			assert.GreaterOrEqual(t, result, tt.minValue)
			assert.LessOrEqual(t, result, tt.maxValue)
		})
	}
}

// TestCalculateDamage tests spell damage calculation
func TestCalculateDamage(t *testing.T) {
	tests := []struct {
		name       string
		spell      *game.Spell
		spellPower int
		minDamage  int
	}{
		{
			name:       "fireball damage",
			spell:      &game.Spell{ID: "fireball", Level: 3},
			spellPower: 5,
			minDamage:  13, // 8d6 (min 8) + 5 power
		},
		{
			name:       "lightning bolt damage",
			spell:      &game.Spell{ID: "lightning_bolt", Level: 3},
			spellPower: 3,
			minDamage:  9, // 6d8 (min 6) + 3 power
		},
		{
			name:       "magic missile damage",
			spell:      &game.Spell{ID: "magic_missile", Level: 1},
			spellPower: 2,
			minDamage:  8, // 3d4 (min 3) + 3 + 2 power
		},
		{
			name:       "generic spell damage",
			spell:      &game.Spell{ID: "unknown_spell", Level: 2},
			spellPower: 4,
			minDamage:  6, // 2d6 (min 2) + 4 power
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			damage := calculateDamage(tt.spell, tt.spellPower)
			assert.GreaterOrEqual(t, damage, tt.minDamage)
		})
	}
}

// TestCalculateHealing tests spell healing calculation
func TestCalculateHealing(t *testing.T) {
	tests := []struct {
		name       string
		spell      *game.Spell
		spellPower int
		minHealing int
	}{
		{
			name:       "heal spell",
			spell:      &game.Spell{ID: "heal", Level: 4},
			spellPower: 5,
			minHealing: 9, // 4d8 (min 4) + 5 power
		},
		{
			name:       "cure wounds spell",
			spell:      &game.Spell{ID: "cure_wounds", Level: 2},
			spellPower: 3,
			minHealing: 5, // 2d8 (min 2) + 3 power
		},
		{
			name:       "healing word spell",
			spell:      &game.Spell{ID: "healing_word", Level: 1},
			spellPower: 2,
			minHealing: 3, // 1d4 (min 1) + 2 power
		},
		{
			name:       "generic healing spell",
			spell:      &game.Spell{ID: "unknown_heal", Level: 3},
			spellPower: 4,
			minHealing: 7, // 3d4 (min 3) + 4 power
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			healing := calculateHealing(tt.spell, tt.spellPower)
			assert.GreaterOrEqual(t, healing, tt.minHealing)
		})
	}
}

// TestCalculateSpellPower tests spell power calculation
func TestCalculateSpellPower(t *testing.T) {
	tests := []struct {
		name     string
		caster   *game.Player
		spell    *game.Spell
		minPower int
	}{
		{
			name: "low level caster",
			caster: &game.Player{
				Character: game.Character{
					Level:        1,
					Intelligence: 10,
				},
			},
			spell:    &game.Spell{Level: 1},
			minPower: 5, // 1*5 + 0 + 0 = 5
		},
		{
			name: "high intelligence bonus",
			caster: &game.Player{
				Character: game.Character{
					Level:        5,
					Intelligence: 18,
				},
			},
			spell:    &game.Spell{Level: 2},
			minPower: 14, // 2*5 + 4 + 2 = 16
		},
		{
			name: "low intelligence - modifier clamped to 0",
			caster: &game.Player{
				Character: game.Character{
					Level:        3,
					Intelligence: 8,
				},
			},
			spell:    &game.Spell{Level: 2},
			minPower: 10, // 2*5 + 0 (clamped) + 1 = 11, but actual formula: level*5 + (int-10)/2 clamped + level/2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			power := calculateSpellPower(tt.caster, tt.spell)
			assert.GreaterOrEqual(t, power, tt.minPower)
		})
	}
}

// TestFindSpell_Additional tests additional spell lookup scenarios
func TestFindSpell_Additional(t *testing.T) {
	spells := []game.Spell{
		{ID: "fireball", Name: "Fireball", Level: 3},
		{ID: "heal", Name: "Heal", Level: 2},
		{ID: "magic_missile", Name: "Magic Missile", Level: 1},
	}

	tests := []struct {
		name     string
		spellID  string
		found    bool
		expected string
	}{
		{
			name:     "find existing spell",
			spellID:  "fireball",
			found:    true,
			expected: "Fireball",
		},
		{
			name:     "find another spell",
			spellID:  "heal",
			found:    true,
			expected: "Heal",
		},
		{
			name:    "spell not found",
			spellID: "nonexistent",
			found:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findSpell(spells, tt.spellID)
			if tt.found {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected, result.Name)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

// TestMinFunction_Coverage tests the min helper function with additional cases
func TestMinFunction_Coverage(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{name: "a smaller", a: 1, b: 5, expected: 1},
		{name: "b smaller", a: 10, b: 3, expected: 3},
		{name: "equal", a: 7, b: 7, expected: 7},
		{name: "negative numbers", a: -5, b: -10, expected: -10},
		{name: "zero and positive", a: 0, b: 5, expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
