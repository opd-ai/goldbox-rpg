package game

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCharacter_CloneBasic tests the Clone method for creating deep copies of characters
func TestCharacter_CloneBasic(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Character
		validate func(t *testing.T, original, clone *Character)
	}{
		{
			name: "basic character clone",
			setup: func() *Character {
				char := &Character{
					ID:              "char-001",
					Name:            "Test Warrior",
					Description:     "A brave warrior",
					Strength:        18,
					Dexterity:       14,
					Constitution:    16,
					Intelligence:    10,
					Wisdom:          12,
					Charisma:        13,
					HP:              50,
					MaxHP:           50,
					ArmorClass:      15,
					THAC0:           10,
					ActionPoints:    4,
					MaxActionPoints: 4,
					Level:           5,
					Experience:      10000,
					Gold:            500,
					Equipment:       make(map[EquipmentSlot]Item),
					Inventory:       []Item{},
					active:          true,
					tags:            []string{"warrior", "human"},
				}
				char.Position = Position{X: 10, Y: 20, Level: 1}
				return char
			},
			validate: func(t *testing.T, original, clone *Character) {
				assert.Equal(t, original.ID, clone.ID)
				assert.Equal(t, original.Name, clone.Name)
				assert.Equal(t, original.Description, clone.Description)
				assert.Equal(t, original.Strength, clone.Strength)
				assert.Equal(t, original.Dexterity, clone.Dexterity)
				assert.Equal(t, original.Constitution, clone.Constitution)
				assert.Equal(t, original.Intelligence, clone.Intelligence)
				assert.Equal(t, original.Wisdom, clone.Wisdom)
				assert.Equal(t, original.Charisma, clone.Charisma)
				assert.Equal(t, original.HP, clone.HP)
				assert.Equal(t, original.MaxHP, clone.MaxHP)
				assert.Equal(t, original.ArmorClass, clone.ArmorClass)
				assert.Equal(t, original.THAC0, clone.THAC0)
				assert.Equal(t, original.ActionPoints, clone.ActionPoints)
				assert.Equal(t, original.MaxActionPoints, clone.MaxActionPoints)
				assert.Equal(t, original.Level, clone.Level)
				assert.Equal(t, original.Experience, clone.Experience)
				assert.Equal(t, original.Gold, clone.Gold)
				assert.Equal(t, original.Position, clone.Position)
				assert.Equal(t, original.active, clone.active)
				assert.Equal(t, original.tags, clone.tags)
			},
		},
		{
			name: "clone with equipment",
			setup: func() *Character {
				char := &Character{
					ID:        "char-002",
					Name:      "Test Mage",
					Equipment: make(map[EquipmentSlot]Item),
					Inventory: []Item{},
				}
				char.Equipment[SlotWeaponMain] = Item{
					ID:     "sword-001",
					Name:   "Iron Sword",
					Type:   "weapon",
					Weight: 5,
				}
				char.Equipment[SlotHead] = Item{
					ID:     "helm-001",
					Name:   "Iron Helm",
					Type:   "armor",
					Weight: 3,
				}
				return char
			},
			validate: func(t *testing.T, original, clone *Character) {
				assert.Equal(t, len(original.Equipment), len(clone.Equipment))
				for slot, item := range original.Equipment {
					cloneItem, exists := clone.Equipment[slot]
					assert.True(t, exists)
					assert.Equal(t, item.ID, cloneItem.ID)
					assert.Equal(t, item.Name, cloneItem.Name)
				}

				// Verify deep copy - modifying clone doesn't affect original
				clone.Equipment[SlotWeaponMain] = Item{ID: "different-sword"}
				assert.NotEqual(t, original.Equipment[SlotWeaponMain].ID, clone.Equipment[SlotWeaponMain].ID)
			},
		},
		{
			name: "clone with inventory",
			setup: func() *Character {
				char := &Character{
					ID:        "char-003",
					Name:      "Test Rogue",
					Equipment: make(map[EquipmentSlot]Item),
					Inventory: []Item{
						{ID: "potion-001", Name: "Health Potion", Type: "potion", Weight: 1},
						{ID: "potion-002", Name: "Mana Potion", Type: "potion", Weight: 1},
						{ID: "scroll-001", Name: "Magic Scroll", Type: "scroll", Weight: 0},
					},
				}
				return char
			},
			validate: func(t *testing.T, original, clone *Character) {
				assert.Equal(t, len(original.Inventory), len(clone.Inventory))
				for i, item := range original.Inventory {
					assert.Equal(t, item.ID, clone.Inventory[i].ID)
					assert.Equal(t, item.Name, clone.Inventory[i].Name)
				}

				// Verify deep copy - modifying clone doesn't affect original
				clone.Inventory[0] = Item{ID: "different-potion"}
				assert.NotEqual(t, original.Inventory[0].ID, clone.Inventory[0].ID)
			},
		},
		{
			name: "clone independence - equipment maps",
			setup: func() *Character {
				char := &Character{
					ID:        "char-004",
					Name:      "Test Cleric",
					Equipment: make(map[EquipmentSlot]Item),
					Inventory: []Item{},
				}
				char.Equipment[SlotWeaponMain] = Item{ID: "mace-001", Name: "Iron Mace", Type: "weapon"}
				return char
			},
			validate: func(t *testing.T, original, clone *Character) {
				// Add item to clone's equipment
				clone.Equipment[SlotWeaponOff] = Item{ID: "shield-001", Name: "Wooden Shield", Type: "shield"}

				// Original should not have the new item
				_, exists := original.Equipment[SlotWeaponOff]
				assert.False(t, exists)

				// Clone should have both items
				assert.Len(t, clone.Equipment, 2)
				assert.Len(t, original.Equipment, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := tt.setup()
			clone := original.Clone()

			require.NotNil(t, clone)
			tt.validate(t, original, clone)

			// Verify different mutex instances
			assert.NotSame(t, &original.mu, &clone.mu)
		})
	}
}

// TestCharacter_ConcurrentAccess tests thread-safety of character operations
func TestCharacter_ConcurrentAccess(t *testing.T) {
	char := &Character{
		ID:              "char-concurrent",
		Name:            "Concurrent Test",
		HP:              100,
		MaxHP:           100,
		ActionPoints:    10,
		MaxActionPoints: 10,
		Equipment:       make(map[EquipmentSlot]Item),
		Inventory:       []Item{},
		Position:        Position{X: 0, Y: 0, Level: 0},
	}

	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3) // 3 types of operations

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				_ = char.GetHealth()
				_ = char.GetActionPoints()
				_ = char.GetPosition()
				_ = char.GetID()
				_ = char.GetName()
			}
		}()
	}

	// Concurrent writes to position
	for i := 0; i < numGoroutines; i++ {
		go func(n int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				_ = char.SetPosition(Position{X: n % 50, Y: j % 50, Level: 0})
			}
		}(i)
	}

	// Concurrent writes to health and action points
	for i := 0; i < numGoroutines; i++ {
		go func(n int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				char.SetHealth(50 + (n+j)%50)
				char.SetActionPoints(5 + (n+j)%5)
			}
		}(i)
	}

	wg.Wait()

	// Verify character is still in valid state
	assert.GreaterOrEqual(t, char.GetHealth(), 0)
	assert.LessOrEqual(t, char.GetHealth(), char.MaxHP)
	assert.GreaterOrEqual(t, char.GetActionPoints(), 0)
}

// TestCharacter_CloneConcurrent tests that Clone is thread-safe
func TestCharacter_CloneConcurrent(t *testing.T) {
	original := &Character{
		ID:              "char-concurrent-clone",
		Name:            "Concurrent Clone Test",
		HP:              100,
		MaxHP:           100,
		Strength:        15,
		Dexterity:       14,
		Constitution:    13,
		Equipment:       make(map[EquipmentSlot]Item),
		Inventory:       []Item{{ID: "item-001", Name: "Test Item", Type: "misc"}},
		Position:        Position{X: 10, Y: 10, Level: 1},
	}

	const numGoroutines = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Concurrent clones
	clones := make([]*Character, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			clones[index] = original.Clone()
		}(i)
	}

	// Concurrent reads from original while cloning
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = original.GetID()
			_ = original.GetName()
			_ = original.GetHealth()
		}()
	}

	wg.Wait()

	// Verify all clones are valid
	for i, clone := range clones {
		require.NotNil(t, clone, "Clone %d should not be nil", i)
		assert.Equal(t, original.ID, clone.ID)
		assert.Equal(t, original.Name, clone.Name)
	}
}

// TestCharacter_ToJSONAndFromJSON tests JSON serialization
func TestCharacter_ToJSONAndFromJSON(t *testing.T) {
	original := &Character{
		ID:              "char-json-001",
		Name:            "JSON Test",
		Description:     "Testing JSON serialization",
		Strength:        15,
		Dexterity:       14,
		Constitution:    13,
		Intelligence:    12,
		Wisdom:          11,
		Charisma:        10,
		HP:              50,
		MaxHP:           60,
		ArmorClass:      14,
		THAC0:           15,
		ActionPoints:    5,
		MaxActionPoints: 6,
		Level:           3,
		Experience:      5000,
		Gold:            250,
		Equipment:       make(map[EquipmentSlot]Item),
		Inventory:       []Item{{ID: "item-001", Name: "Test Item", Type: "misc"}},
		active:          true,
		tags:            []string{"test", "json"},
	}
	original.Position = Position{X: 5, Y: 10, Level: 2}

	// Serialize to JSON
	data, err := original.ToJSON()
	require.NoError(t, err)
	require.NotNil(t, data)

	// Deserialize from JSON
	restored := &Character{
		Equipment: make(map[EquipmentSlot]Item),
		Inventory: []Item{},
	}
	err = restored.FromJSON(data)
	require.NoError(t, err)

	// Verify critical fields
	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.Name, restored.Name)
	assert.Equal(t, original.Description, restored.Description)
	assert.Equal(t, original.Strength, restored.Strength)
	assert.Equal(t, original.Dexterity, restored.Dexterity)
	assert.Equal(t, original.Constitution, restored.Constitution)
	assert.Equal(t, original.Intelligence, restored.Intelligence)
	assert.Equal(t, original.Wisdom, restored.Wisdom)
	assert.Equal(t, original.Charisma, restored.Charisma)
	assert.Equal(t, original.HP, restored.HP)
	assert.Equal(t, original.MaxHP, restored.MaxHP)
	assert.Equal(t, original.ArmorClass, restored.ArmorClass)
	assert.Equal(t, original.THAC0, restored.THAC0)
	assert.Equal(t, original.Level, restored.Level)
	assert.Equal(t, original.Experience, restored.Experience)
	assert.Equal(t, original.Gold, restored.Gold)
	assert.Equal(t, original.Position, restored.Position)
}

// TestCharacter_FromJSONInvalidData tests error handling in JSON deserialization
func TestCharacter_FromJSONInvalidData(t *testing.T) {
	char := &Character{
		Equipment: make(map[EquipmentSlot]Item),
		Inventory: []Item{},
	}

	// Test with invalid JSON
	err := char.FromJSON([]byte("invalid json"))
	assert.Error(t, err)

	// Test with empty data
	err = char.FromJSON([]byte{})
	assert.Error(t, err)
}
