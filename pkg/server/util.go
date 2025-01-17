package server

import (
	"goldbox-rpg/pkg/game"
	"regexp"
	"sort"
	"strconv"

	"golang.org/x/exp/rand"
)

// rollInitiative determines the combat turn order for a list of participants by rolling initiative.
//
// Parameters:
//   - participants: A slice of entity IDs (strings) representing the combatants
//
// Returns:
//   - A slice of entity IDs sorted by initiative roll (highest to lowest)
//
// The initiative roll is calculated as:
//   - For Characters: d20 + DEX modifier (Dexterity-10)/2
//   - For other entities: d20 only
//
// Note: Characters must exist in s.state.WorldState.Objects to have their DEX bonus applied.
// Non-existent entities are skipped in the result.
//
// Related:
//   - game.Character struct - for character stats
//   - s.state.WorldState.Objects - entity storage
func (s *RPCServer) rollInitiative(participants []string) []string {
	type initiativeRoll struct {
		entityID string
		roll     int
	}

	rolls := make([]initiativeRoll, len(participants))
	for i, id := range participants {
		if obj, exists := s.state.WorldState.Objects[id]; exists {
			if char, ok := obj.(*game.Character); ok {
				rolls[i] = initiativeRoll{
					entityID: id,
					roll:     rand.Intn(20) + 1 + (char.Dexterity-10)/2,
				}
			} else {
				rolls[i] = initiativeRoll{
					entityID: id,
					roll:     rand.Intn(20) + 1,
				}
			}
		}
	}

	sort.Slice(rolls, func(i, j int) bool {
		return rolls[i].roll > rolls[j].roll
	})

	result := make([]string, len(rolls))
	for i, roll := range rolls {
		result[i] = roll.entityID
	}

	return result
}

// getVisibleObjects returns all game objects that are within the player's visible range.
// The visibility is determined by the isPositionVisible method which checks if the object's
// position is within line of sight and range of the player.
//
// Parameters:
//   - player: *game.Player - The player whose visibility range is being checked
//
// Returns:
//   - []game.GameObject - Slice containing all visible game objects from the world state
//
// Related:
//   - isPositionVisible() - Used to check if a position is visible from player's position
//   - game.GameObject - Interface implemented by all game objects
//   - game.Player - Player entity struct
func (s *RPCServer) getVisibleObjects(player *game.Player) []game.GameObject {
	playerPos := player.GetPosition()
	visibleObjects := make([]game.GameObject, 0)

	for _, obj := range s.state.WorldState.Objects {
		objPos := obj.GetPosition()
		if s.isPositionVisible(playerPos, objPos) {
			visibleObjects = append(visibleObjects, obj)
		}
	}

	return visibleObjects
}

// getActiveEffects retrieves all active effects currently applied to a player
//
// Parameters:
//   - player *game.Player: The player object to check for effects. Must not be nil.
//
// Returns:
//   - []*game.Effect: Slice of active effects on the player. Returns nil if player
//     does not implement game.EffectHolder interface.
//
// Related types:
//   - game.Effect
//   - game.EffectHolder
//   - game.Player
//
// Note: Uses type assertion to check if player implements EffectHolder interface.
func (s *RPCServer) getActiveEffects(player *game.Player) []*game.Effect {
	if holder, ok := interface{}(player).(game.EffectHolder); ok {
		return holder.GetEffects()
	}
	return nil
}

// getCombatStateIfActive retrieves the current combat state for an active combat session.
// If there is no active combat, it returns nil.
//
// Parameters:
//   - player: *game.Player - The player for whom to get the combat state
//
// Returns:
//   - *CombatState - Contains combat information including:
//   - Active combatants in initiative order
//   - Current round count
//   - Combat zone position
//   - Active status effects
//     Returns nil if no combat is active
//
// Related:
//   - TurnManager.IsInCombat
//   - CombatState struct
func (s *RPCServer) getCombatStateIfActive(player *game.Player) *CombatState {
	if !s.state.TurnManager.IsInCombat {
		return nil
	}

	return &CombatState{
		ActiveCombatants: s.state.TurnManager.Initiative,
		RoundCount:       s.state.TurnManager.CurrentRound,
		CombatZone:       player.GetPosition(),
		StatusEffects:    s.getCombatEffects(),
	}
}

// getCombatEffects returns a map of active effects for all objects in the current combat initiative order.
//
// The function iterates through all objects in the TurnManager's initiative order and collects
// any active effects on objects that implement the EffectHolder interface.
//
// Returns:
//   - map[string][]game.Effect: A map where keys are object IDs and values are slices of active effects
//
// Related types:
//   - game.Effect: The effect type being collected
//   - game.EffectHolder: Interface for objects that can have effects
//
// Note: Objects that don't exist in WorldState or don't implement EffectHolder are skipped.
// Only objects with active effects will have entries in the returned map.
func (s *RPCServer) getCombatEffects() map[string][]game.Effect {
	effects := make(map[string][]game.Effect)

	for _, id := range s.state.TurnManager.Initiative {
		if obj, exists := s.state.WorldState.Objects[id]; exists {
			if holder, ok := obj.(game.EffectHolder); ok {
				activeEffects := holder.GetEffects()
				if len(activeEffects) > 0 {
					effects[id] = make([]game.Effect, len(activeEffects))
					for i, effect := range activeEffects {
						effects[id][i] = *effect
					}
				}
			}
		}
	}

	return effects
}

// isPositionVisible checks if a target position is visible from a given source position.
// It determines visibility based on Manhattan distance and level matching.
//
// Parameters:
//   - from: The source Position containing X,Y coordinates and Level
//   - to: The target Position to check visibility for
//
// Returns:
//   - bool: true if target position is visible (within 10 unit distance and on same level),
//     false otherwise
//
// Notes:
//   - Uses square of Euclidean distance (dx²+dy²) <= 100 for performance
//   - Requires positions to be on the same level
//   - Distance check uses a radius of 10 units (square root of 100)
func (s *RPCServer) isPositionVisible(from, to game.Position) bool {
	dx := from.X - to.X
	dy := from.Y - to.Y
	distanceSquared := dx*dx + dy*dy

	return distanceSquared <= 100 && from.Level == to.Level
}

// processEndTurnEffects processes any effects that should trigger at the end of a turn for a given game object.
// It checks if the object implements the EffectHolder interface and if so, iterates through its effects,
// processing any that should tick based on the current game time.
//
// Parameters:
//   - character: The game object to process end-turn effects for. Must implement game.GameObject interface.
//
// The function handles the following cases:
//   - If character does not implement EffectHolder, no effects are processed
//   - Each effect is checked against current time to determine if it should tick
//
// Related types:
//   - game.GameObject
//   - game.EffectHolder
//   - game.Effect
func (s *RPCServer) processEndTurnEffects(character game.GameObject) {
	if holder, ok := character.(game.EffectHolder); ok {
		for _, effect := range holder.GetEffects() {
			if effect.ShouldTick(s.state.TimeManager.CurrentTime.RealTime) {
				s.state.processEffectTick(effect)
			}
		}
	}
}

// processEndRound handles end-of-round processing for the game state:
// 1. Increments the current round counter
// 2. Processes any delayed/queued actions
// 3. Checks if combat has ended
//
// Related:
// - TurnManager.CurrentRound
// - processDelayedActions()
// - checkCombatEnd()
func (s *RPCServer) processEndRound() {
	s.state.TurnManager.CurrentRound++
	s.processDelayedActions()
	s.checkCombatEnd()
}

// isTimeToExecute checks if a given game time has been reached based on tick counts
//
// Parameters:
//   - current: The current game time
//   - trigger: The target game time to compare against
//
// Returns:
//
//	bool: true if current game ticks is greater than or equal to trigger ticks,
//	false otherwise
//
// Related:
//   - game.GameTime struct
func isTimeToExecute(current, trigger game.GameTime) bool {
	return current.GameTicks >= trigger.GameTicks
}

// findSpell searches for a spell in the provided slice of spells by ID.
// Parameters:
//   - spells: Slice of game.Spell objects to search through
//   - spellID: String ID of the spell to find
//
// Returns:
//   - *game.Spell: Pointer to the found spell, or nil if not found
//
// Related:
//   - game.Spell struct
func findSpell(spells []game.Spell, spellID string) *game.Spell {
	for i := range spells {
		if spells[i].ID == spellID {
			return &spells[i]
		}
	}
	return nil
}

// findInventoryItem searches for an item in the inventory by its ID and returns a pointer to it if found.
//
// Parameters:
//   - inventory: []game.Item - slice of game items to search through
//   - itemID: string - unique identifier of the item to find
//
// Returns:
//   - *game.Item - pointer to the found item, or nil if not found
//
// Related:
//   - game.Item type
//
// Note: Returns nil if the item is not found in the inventory
func findInventoryItem(inventory []game.Item, itemID string) *game.Item {
	for i := range inventory {
		if inventory[i].ID == itemID {
			return &inventory[i]
		}
	}
	return nil
}

// parseDamageString takes a damage string in dice notation format (e.g. "2d6+3") and returns the average damage value.
//
// The function accepts the following formats:
//   - Plain number (e.g. "5")
//   - Dice notation "XdY+Z" where:
//     X = number of dice (optional, defaults to 1)
//     Y = number of sides on each die
//     Z = fixed modifier to add (optional)
//
// Parameters:
//
//	damage string - The damage string to parse in dice notation format
//
// Returns:
//
//	int - The calculated average damage:
//	- For plain numbers, returns the number as-is
//	- For dice notation, returns average roll value of dice + modifier
//	- Returns 0 for invalid input formats
//
// Examples:
//
//	parseDamageString("5")    // Returns 5
//	parseDamageString("2d6")  // Returns 7 (avg of 2 six-sided dice)
//	parseDamageString("d8+2") // Returns 6.5 rounded to 6 (avg of 1d8 + 2)
//	parseDamageString("foo")  // Returns 0 (invalid format)
func parseDamageString(damage string) int {
	// Regular expression to match dice notation: XdY+Z
	re := regexp.MustCompile(`^(\d+)?d(\d+)(?:\+(\d+))?$`)

	// If it's just a number, return it
	if num, err := strconv.Atoi(damage); err == nil {
		return num
	}

	matches := re.FindStringSubmatch(damage)
	if matches == nil {
		return 0 // Invalid format
	}

	// Parse components
	numDice := 1
	if matches[1] != "" {
		numDice, _ = strconv.Atoi(matches[1])
	}

	dieSize, _ := strconv.Atoi(matches[2])

	modifier := 0
	if matches[3] != "" {
		modifier, _ = strconv.Atoi(matches[3])
	}

	// Calculate average damage
	// Average roll on a die is (1 + size) / 2
	averageDamage := int(float64(numDice) * (float64(dieSize) + 1) / 2)
	return averageDamage + modifier
}

// min returns the smaller of two integers.
// Parameters:
//   - a: first integer to compare
//   - b: second integer to compare
//
// Returns:
//
//	The smaller of a and b
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
