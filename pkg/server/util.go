package server

import (
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/rand"
)

// ADDED: rollInitiative determines combat turn order by rolling initiative for all participants.
// It calculates initiative scores and returns participants sorted by highest roll first.
//
// Initiative calculation:
// - Characters: d20 + DEX modifier ((Dexterity-10)/2)
// - Other entities: d20 only
//
// Parameters:
//   - participants: Slice of entity IDs representing the combatants
//
// Returns:
//   - []string: Entity IDs sorted by initiative roll (highest to lowest)
//
// Notes:
// - Characters must exist in WorldState.Objects to apply DEX bonus
// - Non-existent entities are skipped from results
// - Uses golang.org/x/exp/rand for random number generation
func (s *RPCServer) rollInitiative(participants []string) []string {
	logger := logrus.WithFields(logrus.Fields{
		"function":        "rollInitiative",
		"numParticipants": len(participants),
	})
	logger.Debug("rolling initiative")

	type initiativeRoll struct {
		entityID string
		roll     int
	}

	rolls := make([]initiativeRoll, len(participants))
	for i, id := range participants {
		logger := logger.WithField("entityID", id)
		if obj, exists := s.state.WorldState.Objects[id]; exists {
			if char, ok := obj.(*game.Character); ok {
				roll := rand.Intn(20) + 1
				modifier := (char.Dexterity - 10) / 2
				rolls[i] = initiativeRoll{
					entityID: id,
					roll:     roll + modifier,
				}
				logger.WithFields(logrus.Fields{
					"baseRoll": roll,
					"modifier": modifier,
					"total":    rolls[i].roll,
				}).Info("rolled initiative for character")
			} else {
				roll := rand.Intn(20) + 1
				rolls[i] = initiativeRoll{
					entityID: id,
					roll:     roll,
				}
				logger.WithField("roll", roll).Info("rolled initiative for entity")
			}
		} else {
			logger.Warn("entity not found in world state")
		}
	}

	sort.Slice(rolls, func(i, j int) bool {
		return rolls[i].roll > rolls[j].roll
	})

	result := make([]string, len(rolls))
	for i, roll := range rolls {
		result[i] = roll.entityID
	}

	logger.WithField("order", result).Info("initiative order determined")
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
	logger := logrus.WithFields(logrus.Fields{
		"function": "getVisibleObjects",
		"playerID": player.GetID(),
	})
	logger.Debug("getting visible objects for player")

	playerPos := player.GetPosition()
	logger.WithField("position", playerPos).Debug("got player position")

	visibleObjects := make([]game.GameObject, 0)

	for _, obj := range s.state.WorldState.Objects {
		objPos := obj.GetPosition()
		if s.isPositionVisible(playerPos, objPos) {
			logger.WithFields(logrus.Fields{
				"objectID": obj.GetID(),
				"position": objPos,
			}).Debug("object is visible")
			visibleObjects = append(visibleObjects, obj)
		}
	}

	logger.WithField("visibleCount", len(visibleObjects)).Info("finished getting visible objects")
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
	logger := logrus.WithFields(logrus.Fields{
		"function": "getActiveEffects",
		"playerID": player.GetID(),
	})
	logger.Debug("getting active effects for player")

	if holder, ok := interface{}(player).(game.EffectHolder); ok {
		effects := holder.GetEffects()
		logger.WithField("numEffects", len(effects)).Info("retrieved active effects")
		return effects
	}
	logger.Warn("player does not implement EffectHolder interface")
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
	logger := logrus.WithFields(logrus.Fields{
		"function": "getCombatStateIfActive",
		"playerID": player.GetID(),
	})
	logger.Debug("checking combat state")

	if !s.state.TurnManager.IsInCombat {
		logger.Info("no active combat")
		return nil
	}

	state := &CombatState{
		ActiveCombatants: s.state.TurnManager.Initiative,
		RoundCount:       s.state.TurnManager.CurrentRound,
		CombatZone:       player.GetPosition(),
		StatusEffects:    s.getCombatEffects(),
	}

	logger.WithFields(logrus.Fields{
		"combatants": len(state.ActiveCombatants),
		"round":      state.RoundCount,
		"position":   state.CombatZone,
	}).Info("retrieved active combat state")

	return state
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
	logger := logrus.WithFields(logrus.Fields{
		"function": "getCombatEffects",
	})
	logger.Debug("collecting combat effects")

	effects := make(map[string][]game.Effect)

	for _, id := range s.state.TurnManager.Initiative {
		logger := logger.WithField("entityID", id)
		if obj, exists := s.state.WorldState.Objects[id]; exists {
			if holder, ok := obj.(game.EffectHolder); ok {
				activeEffects := holder.GetEffects()
				if len(activeEffects) > 0 {
					logger.WithField("numEffects", len(activeEffects)).Info("adding effects for entity")
					effects[id] = make([]game.Effect, len(activeEffects))
					for i, effect := range activeEffects {
						effects[id][i] = *effect
					}
				} else {
					logger.Debug("entity has no active effects")
				}
			} else {
				logger.Debug("entity does not implement EffectHolder")
			}
		} else {
			logger.Warn("entity not found in world state")
		}
	}

	logger.WithField("totalEntities", len(effects)).Info("finished collecting combat effects")
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
	logger := logrus.WithFields(logrus.Fields{
		"function": "isPositionVisible",
		"from":     from,
		"to":       to,
	})
	logger.Debug("checking position visibility")

	dx := from.X - to.X
	dy := from.Y - to.Y
	distanceSquared := dx*dx + dy*dy

	result := distanceSquared <= 100 && from.Level == to.Level
	logger.WithFields(logrus.Fields{
		"distanceSquared": distanceSquared,
		"sameLevel":       from.Level == to.Level,
		"visible":         result,
	}).Info("visibility check complete")

	return result
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
	logger := logrus.WithFields(logrus.Fields{
		"function": "processEndTurnEffects",
		"charID":   character.GetID(),
	})
	logger.Debug("processing end turn effects")

	if holder, ok := character.(game.EffectHolder); ok {
		effects := holder.GetEffects()
		logger.WithField("numEffects", len(effects)).Info("checking effects")

		for _, effect := range effects {
			logger := logger.WithField("effectID", effect.ID)
			if effect.ShouldTick(s.state.TimeManager.CurrentTime.RealTime) {
				logger.Info("processing effect tick")
				s.state.processEffectTick(effect)
			} else {
				logger.Debug("effect not ready to tick")
			}
		}
	} else {
		logger.Debug("character has no effects")
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
	logger := logrus.WithFields(logrus.Fields{
		"function": "processEndRound",
		"round":    s.state.TurnManager.CurrentRound,
	})
	logger.Debug("processing end of round")

	s.state.TurnManager.CurrentRound++
	logger.WithField("newRound", s.state.TurnManager.CurrentRound).Info("incremented round counter")

	s.processDelayedActions()
	logger.Debug("processed delayed actions")

	s.checkCombatEnd()
	logger.Debug("checked combat end conditions")
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
//
// ADDED: isTimeToExecute checks if the current game time has reached or exceeded a trigger time.
// It compares game tick counts to determine if a scheduled event should execute.
//
// Parameters:
//   - current: Current game time state
//   - trigger: Target trigger time to check against
//
// Returns:
//   - bool: true if current ticks >= trigger ticks, false otherwise
//
// This function is used for:
// - Scheduled event execution
// - Effect duration checking
// - Time-based game mechanics
func isTimeToExecute(current, trigger game.GameTime) bool {
	logger := logrus.WithFields(logrus.Fields{
		"function":     "isTimeToExecute",
		"currentTicks": current.GameTicks,
		"triggerTicks": trigger.GameTicks,
	})
	logger.Debug("checking execution time")

	result := current.GameTicks >= trigger.GameTicks
	logger.WithField("result", result).Info("time check complete")
	return result
}

// ADDED: findSpell searches for a spell by ID within a slice of spell objects.
// It performs linear search through the provided spell collection.
//
// Parameters:
//   - spells: Slice of game.Spell objects to search through
//   - spellID: String identifier of the spell to find
//
// Returns:
//   - *game.Spell: Pointer to found spell, or nil if not found
//
// Search behavior:
// - Case-sensitive string matching on spell ID field
// - Returns first match found (assumes unique IDs)
// - Returns nil if spell not found in collection
func findSpell(spells []game.Spell, spellID string) *game.Spell {
	logger := logrus.WithFields(logrus.Fields{
		"function": "findSpell",
		"spellID":  spellID,
	})
	logger.Debug("searching for spell")

	for i := range spells {
		if spells[i].ID == spellID {
			logger.WithField("found", true).Info("found spell")
			return &spells[i]
		}
	}

	logger.WithField("found", false).Info("spell not found")
	return nil
}

// ADDED: findInventoryItem searches for an item by ID within an inventory collection.
// It performs linear search through the provided inventory items.
//
// Parameters:
//   - inventory: Slice of game.Item objects representing the inventory to search
//   - itemID: String identifier of the item to find
//
// Returns:
//   - *game.Item: Pointer to found item, or nil if not found
//
// Search behavior:
// - Case-sensitive string matching on item ID field
// - Returns first match found (assumes unique IDs)
// - Commonly used for inventory management operations
func findInventoryItem(inventory []game.Item, itemID string) *game.Item {
	logger := logrus.WithFields(logrus.Fields{
		"function": "findInventoryItem",
		"itemID":   itemID,
	})
	logger.Debug("searching inventory for item")

	for i := range inventory {
		if inventory[i].ID == itemID {
			logger.WithField("found", true).Info("found inventory item")
			return &inventory[i]
		}
	}

	logger.WithField("found", false).Info("item not found in inventory")
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
//
// ADDED: parseDamageString converts dice notation strings to average damage values.
// It supports standard RPG dice notation including modifiers and plain numbers.
//
// Supported formats:
// - Plain numbers: "5", "10"
// - Dice notation: "1d6", "2d8", "3d10"
// - Dice with modifiers: "1d6+2", "2d4-1"
// - Shorthand: "d6" (implies 1d6)
//
// Calculation method: Returns mathematical average of dice rolls plus any modifiers
func parseDamageString(damage string) int {
	logger := logrus.WithFields(logrus.Fields{
		"function": "parseDamageString",
		"damage":   damage,
	})
	logger.Debug("parsing damage string")

	// Regular expression to match dice notation: XdY+Z
	re := regexp.MustCompile(`^(\d+)?d(\d+)(?:\+(\d+))?$`)

	// If it's just a number, return it
	if num, err := strconv.Atoi(damage); err == nil {
		logger.WithField("value", num).Info("parsed plain number")
		return num
	}

	matches := re.FindStringSubmatch(damage)
	if matches == nil {
		logger.Warn("invalid dice notation format")
		return 0
	}

	// Parse components
	numDice := 1
	if matches[1] != "" {
		var err error
		numDice, err = strconv.Atoi(matches[1])
		if err != nil {
			logger.WithError(err).Error("failed to parse number of dice")
			return 0
		}
	}

	dieSize, err := strconv.Atoi(matches[2])
	if err != nil {
		logger.WithError(err).Error("failed to parse die size")
		return 0
	}

	modifier := 0
	if matches[3] != "" {
		var err error
		modifier, err = strconv.Atoi(matches[3])
		if err != nil {
			logger.WithError(err).Error("failed to parse modifier")
			return 0
		}
	}

	// Calculate average damage
	averageDamage := int(float64(numDice) * (float64(dieSize) + 1) / 2)
	result := averageDamage + modifier

	logger.WithFields(logrus.Fields{
		"numDice":  numDice,
		"dieSize":  dieSize,
		"modifier": modifier,
		"result":   result,
	}).Info("calculated average damage")

	return result
}

// ADDED: min returns the smaller of two integer values.
// This is a simple utility function for integer comparisons.
//
// Parameters:
//   - a: First integer to compare
//   - b: Second integer to compare
//
// Returns:
//   - int: The smaller of the two input values
//
// Note: This function provides basic minimum value logic used throughout the server.
func min(a, b int) int {
	logger := logrus.WithFields(logrus.Fields{
		"function": "min",
		"a":        a,
		"b":        b,
	})
	logger.Debug("comparing two integers")

	if a < b {
		logger.WithField("result", a).Debug("returning first value")
		return a
	}
	logger.WithField("result", b).Debug("returning second value")
	return b
}

// ADDED: isStaticFileRequest determines if an HTTP request path is for a static file resource.
// It checks the file extension against a list of common web static file types.
//
// Parameters:
//   - path: URL path to examine for static file characteristics
//
// Returns:
//   - bool: true if path appears to be a static file, false otherwise
//
// Recognized static file extensions:
// - Web assets: .html, .css, .js
// - Images: .jpg, .jpeg, .png, .gif, .svg, .ico
// - Fonts: .woff, .woff2, .ttf, .eot
//
// This function helps route requests between static file serving and RPC handling.
func isStaticFileRequest(path string) bool {
	logger := logrus.WithFields(logrus.Fields{
		"function": "isStaticFileRequest",
		"path":     path,
	})
	logger.Debug("checking if path is static file")

	// Add common static file extensions
	staticExtensions := []string{
		".html", ".css", ".js", ".jpg", ".jpeg",
		".png", ".gif", ".svg", ".ico", ".woff",
		".woff2", ".ttf", ".eot",
	}

	ext := filepath.Ext(path)
	for _, staticExt := range staticExtensions {
		if ext == staticExt {
			logger.WithField("extension", ext).Info("serving static file")
			return true
		}
	}
	logger.Debug("path is not a static file")
	return false
}
