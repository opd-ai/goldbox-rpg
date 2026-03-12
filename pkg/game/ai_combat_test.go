package game

import (
	"testing"
)

func TestCombatAI_SelectTarget(t *testing.T) {
	tests := []struct {
		name       string
		difficulty AIDifficulty
		npcHP      int
		enemies    []*Character
		wantNil    bool
		checkFunc  func(*testing.T, *Character, []*Character)
	}{
		{
			name:       "no enemies returns nil",
			difficulty: AIDifficultyMedium,
			npcHP:      100,
			enemies:    []*Character{},
			wantNil:    true,
		},
		{
			name:       "single enemy selected",
			difficulty: AIDifficultyMedium,
			npcHP:      100,
			enemies: []*Character{
				{Name: "Enemy1", HP: 50, MaxHP: 100, Position: Position{X: 5, Y: 5}},
			},
			wantNil: false,
			checkFunc: func(t *testing.T, target *Character, enemies []*Character) {
				if target.Name != "Enemy1" {
					t.Errorf("Expected Enemy1, got %s", target.Name)
				}
			},
		},
		{
			name:       "medium difficulty selects nearest",
			difficulty: AIDifficultyMedium,
			npcHP:      100,
			enemies: []*Character{
				{Name: "Far", HP: 100, MaxHP: 100, Position: Position{X: 10, Y: 10}},
				{Name: "Near", HP: 100, MaxHP: 100, Position: Position{X: 2, Y: 2}},
			},
			wantNil: false,
			checkFunc: func(t *testing.T, target *Character, enemies []*Character) {
				if target.Name != "Near" {
					t.Errorf("Expected Near enemy, got %s", target.Name)
				}
			},
		},
		{
			name:       "hard difficulty prioritizes low HP",
			difficulty: AIDifficultyHard,
			npcHP:      100,
			enemies: []*Character{
				{Name: "FullHP", HP: 100, MaxHP: 100, Position: Position{X: 2, Y: 2}},
				{Name: "LowHP", HP: 10, MaxHP: 100, Position: Position{X: 3, Y: 3}},
			},
			wantNil: false,
			checkFunc: func(t *testing.T, target *Character, enemies []*Character) {
				if target.Name != "LowHP" {
					t.Errorf("Expected LowHP target, got %s", target.Name)
				}
			},
		},
		{
			name:       "dead enemies ignored",
			difficulty: AIDifficultyMedium,
			npcHP:      100,
			enemies: []*Character{
				{Name: "Dead", HP: 0, MaxHP: 100, Position: Position{X: 1, Y: 1}},
				{Name: "Alive", HP: 50, MaxHP: 100, Position: Position{X: 2, Y: 2}},
			},
			wantNil: false,
			checkFunc: func(t *testing.T, target *Character, enemies []*Character) {
				if target.Name != "Alive" {
					t.Errorf("Expected Alive enemy, got %s", target.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := createTestWorld(20, 20, createAllWalkableGrid(20, 20))
			ai := NewCombatAI(tt.difficulty, world)

			npc := &NPC{
				Character: Character{
					Name:     "TestNPC",
					HP:       tt.npcHP,
					MaxHP:    100,
					Position: Position{X: 1, Y: 1, Level: 0},
				},
				Behavior: "aggressive",
			}

			target := ai.SelectTarget(npc, tt.enemies, world)

			if tt.wantNil && target != nil {
				t.Errorf("Expected nil target, got %v", target)
			}

			if !tt.wantNil && target == nil {
				t.Error("Expected target, got nil")
			}

			if !tt.wantNil && target != nil && tt.checkFunc != nil {
				tt.checkFunc(t, target, tt.enemies)
			}
		})
	}
}

func TestCombatAI_ShouldRetreat(t *testing.T) {
	tests := []struct {
		name        string
		difficulty  AIDifficulty
		behavior    string
		npcHP       int
		npcMaxHP    int
		threatCount int
		wantRetreat bool
	}{
		{
			name:        "healthy NPC does not retreat",
			difficulty:  AIDifficultyMedium,
			behavior:    "aggressive",
			npcHP:       80,
			npcMaxHP:    100,
			threatCount: 1,
			wantRetreat: false,
		},
		{
			name:        "easy difficulty retreats at 50% HP",
			difficulty:  AIDifficultyEasy,
			behavior:    "guard",
			npcHP:       45,
			npcMaxHP:    100,
			threatCount: 1,
			wantRetreat: true,
		},
		{
			name:        "medium difficulty retreats at 30% HP",
			difficulty:  AIDifficultyMedium,
			behavior:    "guard",
			npcHP:       25,
			npcMaxHP:    100,
			threatCount: 1,
			wantRetreat: true,
		},
		{
			name:        "hard difficulty retreats at 20% HP",
			difficulty:  AIDifficultyHard,
			behavior:    "guard",
			npcHP:       15,
			npcMaxHP:    100,
			threatCount: 1,
			wantRetreat: true,
		},
		{
			name:        "aggressive hard AI never retreats",
			difficulty:  AIDifficultyHard,
			behavior:    "aggressive",
			npcHP:       10,
			npcMaxHP:    100,
			threatCount: 1,
			wantRetreat: false,
		},
		{
			name:        "overwhelming odds trigger retreat",
			difficulty:  AIDifficultyMedium,
			behavior:    "guard",
			npcHP:       60,
			npcMaxHP:    100,
			threatCount: 3,
			wantRetreat: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := createTestWorld(10, 10, createAllWalkableGrid(10, 10))
			ai := NewCombatAI(tt.difficulty, world)

			npc := &NPC{
				Character: Character{
					HP:    tt.npcHP,
					MaxHP: tt.npcMaxHP,
				},
				Behavior: tt.behavior,
			}

			threats := make([]*Character, tt.threatCount)
			for i := 0; i < tt.threatCount; i++ {
				threats[i] = &Character{
					HP:    100,
					MaxHP: 100,
				}
			}

			shouldRetreat := ai.ShouldRetreat(npc, threats)

			if shouldRetreat != tt.wantRetreat {
				t.Errorf("ShouldRetreat() = %v, want %v", shouldRetreat, tt.wantRetreat)
			}
		})
	}
}

func TestCombatAI_ChooseAction(t *testing.T) {
	tests := []struct {
		name         string
		npcHP        int
		npcPos       Position
		enemyPos     Position
		wantAction   ActionType
		checkPosFunc func(*testing.T, Position)
	}{
		{
			name:       "attack when in melee range",
			npcHP:      100,
			npcPos:     Position{X: 1, Y: 1, Level: 0},
			enemyPos:   Position{X: 2, Y: 1, Level: 0},
			wantAction: ActionAttack,
			checkPosFunc: func(t *testing.T, pos Position) {
				if pos.X != 2 || pos.Y != 1 {
					t.Errorf("Expected attack position (2,1), got (%d,%d)", pos.X, pos.Y)
				}
			},
		},
		{
			name:       "move towards distant enemy",
			npcHP:      100,
			npcPos:     Position{X: 1, Y: 1, Level: 0},
			enemyPos:   Position{X: 5, Y: 1, Level: 0},
			wantAction: ActionMove,
			checkPosFunc: func(t *testing.T, pos Position) {
				// Should move closer to enemy (towards X=5)
				if pos.X <= 1 {
					t.Errorf("Expected move towards enemy, got position (%d,%d)", pos.X, pos.Y)
				}
			},
		},
		{
			name:       "retreat when low HP",
			npcHP:      10,
			npcPos:     Position{X: 5, Y: 5, Level: 0},
			enemyPos:   Position{X: 6, Y: 5, Level: 0},
			wantAction: ActionRetreat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := createTestWorld(20, 20, createAllWalkableGrid(20, 20))
			ai := NewCombatAI(AIDifficultyMedium, world)

			npc := &NPC{
				Character: Character{
					HP:       tt.npcHP,
					MaxHP:    100,
					Position: tt.npcPos,
				},
				Behavior: "guard",
			}

			enemies := []*Character{
				{
					HP:       100,
					MaxHP:    100,
					Position: tt.enemyPos,
				},
			}

			action, pos := ai.ChooseAction(npc, enemies, world)

			if action != tt.wantAction {
				t.Errorf("ChooseAction() action = %v, want %v", action, tt.wantAction)
			}

			if tt.checkPosFunc != nil {
				tt.checkPosFunc(t, pos)
			}
		})
	}
}

func TestCombatAI_FindRetreatPosition(t *testing.T) {
	world := createTestWorld(20, 20, createAllWalkableGrid(20, 20))
	ai := NewCombatAI(AIDifficultyMedium, world)

	npc := &NPC{
		Character: Character{
			Position: Position{X: 10, Y: 10, Level: 0},
		},
	}

	threats := []*Character{
		{HP: 100, Position: Position{X: 11, Y: 10, Level: 0}},
		{HP: 100, Position: Position{X: 11, Y: 11, Level: 0}},
	}

	retreatPos := ai.findRetreatPosition(npc, threats, world)

	// Retreat position should be away from threats (west/southwest from threats at east/southeast)
	if retreatPos.X > npc.Position.X {
		t.Errorf("Expected retreat to move away (west) from threats, got position (%d,%d)",
			retreatPos.X, retreatPos.Y)
	}
}

func TestCombatAI_DifficultyBehavior(t *testing.T) {
	world := createTestWorld(20, 20, createAllWalkableGrid(20, 20))

	enemies := []*Character{
		{Name: "Nearby", HP: 100, MaxHP: 100, Position: Position{X: 2, Y: 1}},
		{Name: "Wounded", HP: 20, MaxHP: 100, Position: Position{X: 5, Y: 5}},
	}

	npc := &NPC{
		Character: Character{
			Position: Position{X: 1, Y: 1, Level: 0},
			HP:       100,
			MaxHP:    100,
		},
	}

	// Easy AI often picks random/nearby targets
	easyAI := NewCombatAI(AIDifficultyEasy, world)
	easyTarget := easyAI.SelectTarget(npc, enemies, world)
	if easyTarget == nil {
		t.Error("Easy AI should select a target")
	}

	// Medium AI picks nearest
	mediumAI := NewCombatAI(AIDifficultyMedium, world)
	mediumTarget := mediumAI.SelectTarget(npc, enemies, world)
	if mediumTarget == nil {
		t.Error("Medium AI should select a target")
	}
	if mediumTarget.Name != "Nearby" {
		t.Errorf("Medium AI should select nearby target, got %s", mediumTarget.Name)
	}

	// Hard AI prioritizes wounded targets
	hardAI := NewCombatAI(AIDifficultyHard, world)
	hardTarget := hardAI.SelectTarget(npc, enemies, world)
	if hardTarget == nil {
		t.Error("Hard AI should select a target")
	}
	if hardTarget.Name != "Wounded" {
		t.Errorf("Hard AI should prioritize wounded target, got %s", hardTarget.Name)
	}
}

// Helper function to create an all-walkable grid
func createAllWalkableGrid(width, height int) [][]bool {
	grid := make([][]bool, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			grid[y][x] = true
		}
	}
	return grid
}
