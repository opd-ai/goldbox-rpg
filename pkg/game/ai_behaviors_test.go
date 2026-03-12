package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a test NPC for behavior tests
func createTestNPCForBehavior(id string, x, y int) *NPC {
	return &NPC{
		Character: Character{
			ID:       id,
			Name:     "Test NPC",
			HP:       100,
			MaxHP:    100,
			Position: Position{X: x, Y: y, Level: 0},
		},
	}
}

// Helper to create a test character for targeting
func createTestCharacterForBehavior(name string, x, y, hp int) *Character {
	return &Character{
		Name:     name,
		HP:       hp,
		MaxHP:    100,
		Position: Position{X: x, Y: y, Level: 0},
	}
}

// Helper to create basic test context
func createTestBehaviorContext() *BehaviorContext {
	return &BehaviorContext{
		Enemies:   make([]*Character, 0),
		Allies:    make([]*Character, 0),
		DeltaTime: 0.016, // ~60 FPS
	}
}

// TestSequenceNode tests the sequence node behavior
func TestSequenceNode(t *testing.T) {
	tests := []struct {
		name           string
		childResults   []BehaviorStatus
		expectedResult BehaviorStatus
	}{
		{
			name:           "all_success",
			childResults:   []BehaviorStatus{StatusSuccess, StatusSuccess, StatusSuccess},
			expectedResult: StatusSuccess,
		},
		{
			name:           "first_fails",
			childResults:   []BehaviorStatus{StatusFailure, StatusSuccess, StatusSuccess},
			expectedResult: StatusFailure,
		},
		{
			name:           "middle_fails",
			childResults:   []BehaviorStatus{StatusSuccess, StatusFailure, StatusSuccess},
			expectedResult: StatusFailure,
		},
		{
			name:           "last_fails",
			childResults:   []BehaviorStatus{StatusSuccess, StatusSuccess, StatusFailure},
			expectedResult: StatusFailure,
		},
		{
			name:           "running_stops_sequence",
			childResults:   []BehaviorStatus{StatusSuccess, StatusRunning, StatusSuccess},
			expectedResult: StatusRunning,
		},
		{
			name:           "empty_sequence",
			childResults:   []BehaviorStatus{},
			expectedResult: StatusSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock children that return predetermined results
			children := make([]BehaviorNode, len(tt.childResults))
			for i, result := range tt.childResults {
				children[i] = &mockBehaviorNode{result: result}
			}

			seq := NewSequenceNode("test_sequence", children...)
			npc := createTestNPCForBehavior("test_npc", 5, 5)
			ctx := createTestBehaviorContext()

			result := seq.Tick(npc, ctx)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

// TestSelectorNode tests the selector node behavior
func TestSelectorNode(t *testing.T) {
	tests := []struct {
		name           string
		childResults   []BehaviorStatus
		expectedResult BehaviorStatus
	}{
		{
			name:           "first_succeeds",
			childResults:   []BehaviorStatus{StatusSuccess, StatusFailure, StatusFailure},
			expectedResult: StatusSuccess,
		},
		{
			name:           "second_succeeds",
			childResults:   []BehaviorStatus{StatusFailure, StatusSuccess, StatusFailure},
			expectedResult: StatusSuccess,
		},
		{
			name:           "last_succeeds",
			childResults:   []BehaviorStatus{StatusFailure, StatusFailure, StatusSuccess},
			expectedResult: StatusSuccess,
		},
		{
			name:           "all_fail",
			childResults:   []BehaviorStatus{StatusFailure, StatusFailure, StatusFailure},
			expectedResult: StatusFailure,
		},
		{
			name:           "running_stops_selector",
			childResults:   []BehaviorStatus{StatusFailure, StatusRunning, StatusSuccess},
			expectedResult: StatusRunning,
		},
		{
			name:           "empty_selector",
			childResults:   []BehaviorStatus{},
			expectedResult: StatusFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := make([]BehaviorNode, len(tt.childResults))
			for i, result := range tt.childResults {
				children[i] = &mockBehaviorNode{result: result}
			}

			sel := NewSelectorNode("test_selector", children...)
			npc := createTestNPCForBehavior("test_npc", 5, 5)
			ctx := createTestBehaviorContext()

			result := sel.Tick(npc, ctx)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

// TestInverterNode tests the inverter decorator
func TestInverterNode(t *testing.T) {
	tests := []struct {
		name           string
		childResult    BehaviorStatus
		expectedResult BehaviorStatus
	}{
		{
			name:           "inverts_success",
			childResult:    StatusSuccess,
			expectedResult: StatusFailure,
		},
		{
			name:           "inverts_failure",
			childResult:    StatusFailure,
			expectedResult: StatusSuccess,
		},
		{
			name:           "passes_running",
			childResult:    StatusRunning,
			expectedResult: StatusRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := &mockBehaviorNode{result: tt.childResult}
			inv := NewInverterNode(child)
			npc := createTestNPCForBehavior("test_npc", 5, 5)
			ctx := createTestBehaviorContext()

			result := inv.Tick(npc, ctx)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

// TestRepeatNode tests the repeat decorator
func TestRepeatNode(t *testing.T) {
	tests := []struct {
		name           string
		repeatCount    int
		childResult    BehaviorStatus
		tickCount      int
		expectedResult BehaviorStatus
	}{
		{
			name:           "returns_running_while_repeating",
			repeatCount:    3,
			childResult:    StatusSuccess,
			tickCount:      1,
			expectedResult: StatusRunning,
		},
		{
			name:           "returns_success_after_all_repeats",
			repeatCount:    3,
			childResult:    StatusSuccess,
			tickCount:      4, // Need to tick 4 times: 3 iterations + 1 final
			expectedResult: StatusSuccess,
		},
		{
			name:           "fails_on_child_failure",
			repeatCount:    3,
			childResult:    StatusFailure,
			tickCount:      1,
			expectedResult: StatusFailure,
		},
		{
			name:           "zero_repeats_success_immediately",
			repeatCount:    0,
			childResult:    StatusSuccess,
			tickCount:      1,
			expectedResult: StatusSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := &mockBehaviorNode{result: tt.childResult}
			rep := NewRepeatNode(child, tt.repeatCount)
			npc := createTestNPCForBehavior("test_npc", 5, 5)
			ctx := createTestBehaviorContext()

			var result BehaviorStatus
			for i := 0; i < tt.tickCount; i++ {
				result = rep.Tick(npc, ctx)
			}
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

// TestConditionNode tests condition evaluation
func TestConditionNode(t *testing.T) {
	tests := []struct {
		name           string
		condition      func(npc *NPC, ctx *BehaviorContext) bool
		expectedResult BehaviorStatus
	}{
		{
			name: "condition_true",
			condition: func(npc *NPC, ctx *BehaviorContext) bool {
				return true
			},
			expectedResult: StatusSuccess,
		},
		{
			name: "condition_false",
			condition: func(npc *NPC, ctx *BehaviorContext) bool {
				return false
			},
			expectedResult: StatusFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := NewConditionNode("test_condition", tt.condition)
			npc := createTestNPCForBehavior("test_npc", 5, 5)
			ctx := createTestBehaviorContext()

			result := cond.Tick(npc, ctx)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

// TestActionNode tests action execution
func TestActionNode(t *testing.T) {
	tests := []struct {
		name           string
		actionResult   BehaviorStatus
		expectedResult BehaviorStatus
	}{
		{
			name:           "action_success",
			actionResult:   StatusSuccess,
			expectedResult: StatusSuccess,
		},
		{
			name:           "action_failure",
			actionResult:   StatusFailure,
			expectedResult: StatusFailure,
		},
		{
			name:           "action_running",
			actionResult:   StatusRunning,
			expectedResult: StatusRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewActionNode("test_action", func(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
				return tt.actionResult
			})
			npc := createTestNPCForBehavior("test_npc", 5, 5)
			ctx := createTestBehaviorContext()

			result := action.Tick(npc, ctx)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

// TestHealthBelowThreshold tests the health condition factory
func TestHealthBelowThreshold(t *testing.T) {
	tests := []struct {
		name      string
		hp        int
		maxHP     int
		threshold float64
		expected  bool
	}{
		{
			name:      "below_threshold",
			hp:        20,
			maxHP:     100,
			threshold: 0.25,
			expected:  true,
		},
		{
			name:      "above_threshold",
			hp:        50,
			maxHP:     100,
			threshold: 0.25,
			expected:  false,
		},
		{
			name:      "at_threshold",
			hp:        25,
			maxHP:     100,
			threshold: 0.25,
			expected:  false,
		},
		{
			name:      "zero_maxhp",
			hp:        0,
			maxHP:     0,
			threshold: 0.5,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := createTestNPCForBehavior("test_npc", 5, 5)
			npc.Character.HP = tt.hp
			npc.Character.MaxHP = tt.maxHP
			ctx := createTestBehaviorContext()

			condFn := HealthBelowThreshold(tt.threshold)
			result := condFn(npc, ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDistanceToTargetBelow tests the distance condition factory
func TestDistanceToTargetBelow(t *testing.T) {
	tests := []struct {
		name     string
		npcX     int
		npcY     int
		targetX  int
		targetY  int
		maxDist  float64
		expected bool
	}{
		{
			name:     "within_range",
			npcX:     0,
			npcY:     0,
			targetX:  2,
			targetY:  2,
			maxDist:  5.0,
			expected: true,
		},
		{
			name:     "outside_range",
			npcX:     0,
			npcY:     0,
			targetX:  10,
			targetY:  10,
			maxDist:  5.0,
			expected: false,
		},
		{
			name:     "no_target",
			npcX:     0,
			npcY:     0,
			targetX:  0,
			targetY:  0,
			maxDist:  5.0,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := createTestNPCForBehavior("test_npc", tt.npcX, tt.npcY)
			ctx := createTestBehaviorContext()

			if tt.name != "no_target" {
				ctx.Target = createTestCharacterForBehavior("target", tt.targetX, tt.targetY, 100)
			}

			condFn := DistanceToTargetBelow(tt.maxDist)
			result := condFn(npc, ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHasEnemiesNearby tests the enemy proximity condition
func TestHasEnemiesNearby(t *testing.T) {
	tests := []struct {
		name     string
		npcX     int
		npcY     int
		enemies  []struct{ x, y, hp int }
		radius   float64
		expected bool
	}{
		{
			name: "enemy_within_radius",
			npcX: 5,
			npcY: 5,
			enemies: []struct{ x, y, hp int }{
				{x: 7, y: 7, hp: 100},
			},
			radius:   5.0,
			expected: true,
		},
		{
			name: "enemy_outside_radius",
			npcX: 5,
			npcY: 5,
			enemies: []struct{ x, y, hp int }{
				{x: 20, y: 20, hp: 100},
			},
			radius:   5.0,
			expected: false,
		},
		{
			name: "dead_enemy_ignored",
			npcX: 5,
			npcY: 5,
			enemies: []struct{ x, y, hp int }{
				{x: 6, y: 6, hp: 0},
			},
			radius:   5.0,
			expected: false,
		},
		{
			name:     "no_enemies",
			npcX:     5,
			npcY:     5,
			enemies:  []struct{ x, y, hp int }{},
			radius:   5.0,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := createTestNPCForBehavior("test_npc", tt.npcX, tt.npcY)
			ctx := createTestBehaviorContext()

			for i, e := range tt.enemies {
				enemy := createTestCharacterForBehavior("enemy"+string(rune('A'+i)), e.x, e.y, e.hp)
				ctx.Enemies = append(ctx.Enemies, enemy)
			}

			condFn := HasEnemiesNearby(tt.radius)
			result := condFn(npc, ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsTargetAlive tests the target alive condition
func TestIsTargetAlive(t *testing.T) {
	tests := []struct {
		name      string
		hasTarget bool
		targetHP  int
		expected  bool
	}{
		{
			name:      "alive_target",
			hasTarget: true,
			targetHP:  50,
			expected:  true,
		},
		{
			name:      "dead_target",
			hasTarget: true,
			targetHP:  0,
			expected:  false,
		},
		{
			name:      "no_target",
			hasTarget: false,
			targetHP:  0,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := createTestNPCForBehavior("test_npc", 5, 5)
			ctx := createTestBehaviorContext()

			if tt.hasTarget {
				ctx.Target = createTestCharacterForBehavior("target", 10, 10, tt.targetHP)
			}

			condFn := IsTargetAlive()
			result := condFn(npc, ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIdleAction tests the idle action
func TestIdleAction(t *testing.T) {
	npc := createTestNPCForBehavior("test_npc", 5, 5)
	ctx := createTestBehaviorContext()

	idleFn := Idle()
	result := idleFn(npc, ctx)
	assert.Equal(t, StatusSuccess, result)
}

// TestSelectNearestEnemy tests enemy selection
func TestSelectNearestEnemy(t *testing.T) {
	tests := []struct {
		name            string
		npcX            int
		npcY            int
		enemies         []struct{ x, y, hp int }
		expectedTargetX int
		expectedTargetY int
		expectedStatus  BehaviorStatus
	}{
		{
			name: "selects_nearest",
			npcX: 0,
			npcY: 0,
			enemies: []struct{ x, y, hp int }{
				{x: 10, y: 10, hp: 100}, // farther
				{x: 2, y: 2, hp: 100},   // nearest
				{x: 5, y: 5, hp: 100},   // middle
			},
			expectedTargetX: 2,
			expectedTargetY: 2,
			expectedStatus:  StatusSuccess,
		},
		{
			name: "ignores_dead",
			npcX: 0,
			npcY: 0,
			enemies: []struct{ x, y, hp int }{
				{x: 1, y: 1, hp: 0},   // dead, closest
				{x: 5, y: 5, hp: 100}, // alive
			},
			expectedTargetX: 5,
			expectedTargetY: 5,
			expectedStatus:  StatusSuccess,
		},
		{
			name:            "no_enemies",
			npcX:            0,
			npcY:            0,
			enemies:         []struct{ x, y, hp int }{},
			expectedTargetX: 0,
			expectedTargetY: 0,
			expectedStatus:  StatusFailure,
		},
		{
			name: "all_dead",
			npcX: 0,
			npcY: 0,
			enemies: []struct{ x, y, hp int }{
				{x: 1, y: 1, hp: 0},
				{x: 2, y: 2, hp: 0},
			},
			expectedTargetX: 0,
			expectedTargetY: 0,
			expectedStatus:  StatusFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := createTestNPCForBehavior("test_npc", tt.npcX, tt.npcY)
			ctx := createTestBehaviorContext()

			for i, e := range tt.enemies {
				enemy := createTestCharacterForBehavior("enemy"+string(rune('A'+i)), e.x, e.y, e.hp)
				ctx.Enemies = append(ctx.Enemies, enemy)
			}

			selectFn := SelectNearestEnemy()
			result := selectFn(npc, ctx)
			assert.Equal(t, tt.expectedStatus, result)

			if tt.expectedStatus == StatusSuccess {
				require.NotNil(t, ctx.Target)
				assert.Equal(t, tt.expectedTargetX, ctx.Target.Position.X)
				assert.Equal(t, tt.expectedTargetY, ctx.Target.Position.Y)
			}
		})
	}
}

// TestStandardBehaviorTrees tests the predefined behavior trees
func TestStandardBehaviorTrees(t *testing.T) {
	trees := StandardBehaviorTrees{}

	t.Run("aggressive_tree_creation", func(t *testing.T) {
		tree := trees.AggressiveTree()
		assert.NotNil(t, tree)
	})

	t.Run("guard_tree_creation", func(t *testing.T) {
		tree := trees.GuardTree()
		assert.NotNil(t, tree)
	})

	t.Run("patrol_tree_creation", func(t *testing.T) {
		waypoints := []Position{
			{X: 0, Y: 0, Level: 0},
			{X: 10, Y: 0, Level: 0},
			{X: 10, Y: 10, Level: 0},
		}
		tree := trees.PatrolTree(waypoints)
		assert.NotNil(t, tree)
	})

	t.Run("coward_tree_creation", func(t *testing.T) {
		tree := trees.CowardTree()
		assert.NotNil(t, tree)
	})
}

// TestBehaviorTreeBuilder tests the fluent builder API
func TestBehaviorTreeBuilder(t *testing.T) {
	t.Run("build_sequence_tree", func(t *testing.T) {
		builder := NewBehaviorTreeBuilder()
		tree := builder.
			Sequence("main_sequence",
				NewConditionNode("check", func(npc *NPC, ctx *BehaviorContext) bool { return true }),
				NewActionNode("act", func(npc *NPC, ctx *BehaviorContext) BehaviorStatus { return StatusSuccess }),
			).
			Build()

		assert.NotNil(t, tree)
		npc := createTestNPCForBehavior("test_npc", 5, 5)
		ctx := createTestBehaviorContext()
		result := tree.Tick(npc, ctx)
		assert.Equal(t, StatusSuccess, result)
	})

	t.Run("build_selector_tree", func(t *testing.T) {
		builder := NewBehaviorTreeBuilder()
		tree := builder.
			Selector("main_selector",
				NewConditionNode("check_fail", func(npc *NPC, ctx *BehaviorContext) bool { return false }),
				NewActionNode("fallback", func(npc *NPC, ctx *BehaviorContext) BehaviorStatus { return StatusSuccess }),
			).
			Build()

		assert.NotNil(t, tree)
		npc := createTestNPCForBehavior("test_npc", 5, 5)
		ctx := createTestBehaviorContext()
		result := tree.Tick(npc, ctx)
		assert.Equal(t, StatusSuccess, result)
	})
}

// TestBehaviorStatus tests status string representation
func TestBehaviorStatus(t *testing.T) {
	tests := []struct {
		status   BehaviorStatus
		expected string
	}{
		{StatusSuccess, "Success"},
		{StatusFailure, "Failure"},
		{StatusRunning, "Running"},
		{BehaviorStatus(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

// Mock types for testing

// mockBehaviorNode always returns a predetermined result
type mockBehaviorNode struct {
	result BehaviorStatus
}

func (m *mockBehaviorNode) Tick(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
	return m.result
}

// mockBehaviorNodeCounter returns results in sequence
type mockBehaviorNodeCounter struct {
	results []BehaviorStatus
	index   int
}

func (m *mockBehaviorNodeCounter) Tick(npc *NPC, ctx *BehaviorContext) BehaviorStatus {
	if m.index >= len(m.results) {
		return StatusSuccess
	}
	result := m.results[m.index]
	m.index++
	return result
}

func TestDistanceToTargetAbove(t *testing.T) {
	npc := &NPC{
		Character: Character{Position: Position{X: 0, Y: 0}},
	}
	target := &Character{Position: Position{X: 3, Y: 4}}

	tests := []struct {
		name     string
		minDist  float64
		target   *Character
		expected bool
	}{
		{"above threshold", 4.0, target, true},  // distance is 5
		{"at threshold", 5.0, target, false},    // distance is exactly 5
		{"below threshold", 6.0, target, false}, // distance is 5
		{"nil target", 5.0, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &BehaviorContext{Target: tt.target}
			cond := DistanceToTargetAbove(tt.minDist)
			got := cond(npc, ctx)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestAllyCountNearby(t *testing.T) {
	npc := &NPC{
		Character: Character{Position: Position{X: 0, Y: 0}},
	}

	tests := []struct {
		name     string
		minCount int
		range_   float64
		allies   []*Character
		expected bool
	}{
		{
			"enough allies nearby", 2, 5.0,
			[]*Character{
				{Position: Position{X: 1, Y: 1}},
				{Position: Position{X: 2, Y: 2}},
			},
			true,
		},
		{
			"not enough allies", 3, 5.0,
			[]*Character{
				{Position: Position{X: 1, Y: 1}},
			},
			false,
		},
		{
			"allies out of range", 1, 2.0,
			[]*Character{
				{Position: Position{X: 10, Y: 10}},
			},
			false,
		},
		{"no allies", 1, 5.0, []*Character{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &BehaviorContext{Allies: tt.allies}
			cond := AllyCountNearby(tt.minCount, tt.range_)
			got := cond(npc, ctx)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestMoveToPosition(t *testing.T) {
	npc := &NPC{
		Character: Character{Position: Position{X: 0, Y: 0}},
	}

	action := MoveToPosition()

	// Test failure with nil TargetPos
	ctx := &BehaviorContext{TargetPos: nil}
	assert.Equal(t, StatusFailure, action(npc, ctx))

	// Test failure with nil PathFinder
	pos := Position{X: 5, Y: 5}
	ctx = &BehaviorContext{TargetPos: &pos, PathFinder: nil}
	assert.Equal(t, StatusFailure, action(npc, ctx))
}
