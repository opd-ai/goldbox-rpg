// Package game provides opportunity attack mechanics for tactical combat.
package game

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// OpportunityAttackManager tracks entity positions and triggers opportunity attacks
// when enemies leave threatened squares without using the Disengage action.
type OpportunityAttackManager struct {
	mu              sync.RWMutex
	world           *World
	reactionUsed    map[string]bool     // Tracks which entities used their reaction this round
	threatenedBy    map[string][]string // Maps position key to entity IDs threatening it
	entityPositions map[string]Position // Current positions of tracked entities
}

// NewOpportunityAttackManager creates a new opportunity attack manager.
func NewOpportunityAttackManager(world *World) *OpportunityAttackManager {
	return &OpportunityAttackManager{
		world:           world,
		reactionUsed:    make(map[string]bool),
		threatenedBy:    make(map[string][]string),
		entityPositions: make(map[string]Position),
	}
}

// OpportunityAttack represents an opportunity attack that can be made.
type OpportunityAttack struct {
	AttackerID string   // Entity making the opportunity attack
	TargetID   string   // Entity triggering the attack
	FromPos    Position // Position being left
	ToPos      Position // Position moving to
}

// RegisterEntity adds an entity to be tracked for opportunity attacks.
func (oam *OpportunityAttackManager) RegisterEntity(entityID string, pos Position) {
	oam.mu.Lock()
	defer oam.mu.Unlock()

	oam.entityPositions[entityID] = pos
	oam.updateThreatenedSquares(entityID, pos)

	logrus.WithFields(logrus.Fields{
		"function": "RegisterEntity",
		"entityID": entityID,
		"position": pos,
	}).Debug("entity registered for opportunity attacks")
}

// UnregisterEntity removes an entity from opportunity attack tracking.
func (oam *OpportunityAttackManager) UnregisterEntity(entityID string) {
	oam.mu.Lock()
	defer oam.mu.Unlock()

	delete(oam.entityPositions, entityID)
	oam.removeFromThreatenedSquares(entityID)
	delete(oam.reactionUsed, entityID)

	logrus.WithFields(logrus.Fields{
		"function": "UnregisterEntity",
		"entityID": entityID,
	}).Debug("entity unregistered from opportunity attacks")
}

// CheckMovement evaluates if moving from one position to another triggers opportunity attacks.
// Returns a list of opportunity attacks that can be made against the moving entity.
//
// Parameters:
//   - moverID: ID of the entity that is moving
//   - fromPos: Starting position of the move
//   - toPos: Destination position of the move
//   - isDisengage: True if the mover used the Disengage action
//
// Returns:
//   - []OpportunityAttack: Attacks that can be triggered by this movement
func (oam *OpportunityAttackManager) CheckMovement(moverID string, fromPos, toPos Position, isDisengage bool) []OpportunityAttack {
	oam.mu.RLock()
	defer oam.mu.RUnlock()

	var attacks []OpportunityAttack

	// Disengage prevents opportunity attacks
	if isDisengage {
		logrus.WithFields(logrus.Fields{
			"function": "CheckMovement",
			"moverID":  moverID,
		}).Debug("movement with disengage - no opportunity attacks")
		return attacks
	}

	// Check if mover is leaving any threatened squares
	posKey := positionKey(fromPos)
	threateningEntities := oam.threatenedBy[posKey]

	for _, attackerID := range threateningEntities {
		// Skip self
		if attackerID == moverID {
			continue
		}

		// Skip if attacker already used reaction this round
		if oam.reactionUsed[attackerID] {
			continue
		}

		// Check if destination is still adjacent to attacker
		attackerPos := oam.entityPositions[attackerID]
		if isAdjacent(attackerPos, toPos) {
			// Still adjacent, no opportunity attack
			continue
		}

		// Leaving threatened area without disengage triggers opportunity attack
		attacks = append(attacks, OpportunityAttack{
			AttackerID: attackerID,
			TargetID:   moverID,
			FromPos:    fromPos,
			ToPos:      toPos,
		})

		logrus.WithFields(logrus.Fields{
			"function":   "CheckMovement",
			"attackerID": attackerID,
			"targetID":   moverID,
			"fromPos":    fromPos,
		}).Debug("opportunity attack triggered")
	}

	return attacks
}

// UseReaction marks an entity as having used their reaction for this round.
func (oam *OpportunityAttackManager) UseReaction(entityID string) {
	oam.mu.Lock()
	defer oam.mu.Unlock()

	oam.reactionUsed[entityID] = true

	logrus.WithFields(logrus.Fields{
		"function": "UseReaction",
		"entityID": entityID,
	}).Debug("reaction used")
}

// ResetReactions clears all reaction flags for a new round.
func (oam *OpportunityAttackManager) ResetReactions() {
	oam.mu.Lock()
	defer oam.mu.Unlock()

	oam.reactionUsed = make(map[string]bool)

	logrus.WithFields(logrus.Fields{
		"function": "ResetReactions",
	}).Debug("all reactions reset for new round")
}

// UpdatePosition updates an entity's position and recalculates threatened squares.
func (oam *OpportunityAttackManager) UpdatePosition(entityID string, newPos Position) {
	oam.mu.Lock()
	defer oam.mu.Unlock()

	oam.removeFromThreatenedSquares(entityID)
	oam.entityPositions[entityID] = newPos
	oam.updateThreatenedSquares(entityID, newPos)

	logrus.WithFields(logrus.Fields{
		"function": "UpdatePosition",
		"entityID": entityID,
		"newPos":   newPos,
	}).Debug("entity position updated")
}

// HasReaction returns whether an entity still has their reaction available.
func (oam *OpportunityAttackManager) HasReaction(entityID string) bool {
	oam.mu.RLock()
	defer oam.mu.RUnlock()

	return !oam.reactionUsed[entityID]
}

// GetThreateningEntities returns all entity IDs that threaten a given position.
func (oam *OpportunityAttackManager) GetThreateningEntities(pos Position) []string {
	oam.mu.RLock()
	defer oam.mu.RUnlock()

	posKey := positionKey(pos)
	result := make([]string, len(oam.threatenedBy[posKey]))
	copy(result, oam.threatenedBy[posKey])
	return result
}

// updateThreatenedSquares marks all adjacent squares as threatened by this entity.
func (oam *OpportunityAttackManager) updateThreatenedSquares(entityID string, pos Position) {
	adjacentPositions := getAdjacentPositions(pos)

	for _, adjPos := range adjacentPositions {
		posKey := positionKey(adjPos)
		oam.threatenedBy[posKey] = append(oam.threatenedBy[posKey], entityID)
	}
}

// removeFromThreatenedSquares removes an entity from all threatened square mappings.
func (oam *OpportunityAttackManager) removeFromThreatenedSquares(entityID string) {
	for posKey, entities := range oam.threatenedBy {
		filtered := make([]string, 0, len(entities))
		for _, id := range entities {
			if id != entityID {
				filtered = append(filtered, id)
			}
		}
		if len(filtered) > 0 {
			oam.threatenedBy[posKey] = filtered
		} else {
			delete(oam.threatenedBy, posKey)
		}
	}
}

// positionKey creates a string key for a position to use in maps.
func positionKey(pos Position) string {
	return string(rune(pos.X)) + "," + string(rune(pos.Y)) + "," + string(rune(pos.Level))
}

// isAdjacent checks if two positions are within melee range (adjacent).
func isAdjacent(a, b Position) bool {
	if a.Level != b.Level {
		return false
	}
	dx := a.X - b.X
	dy := a.Y - b.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx <= 1 && dy <= 1 && !(dx == 0 && dy == 0)
}

// getAdjacentPositions returns all 8 positions adjacent to the given position.
func getAdjacentPositions(pos Position) []Position {
	adjacent := make([]Position, 0, 8)
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}
			adjacent = append(adjacent, Position{
				X:     pos.X + dx,
				Y:     pos.Y + dy,
				Level: pos.Level,
			})
		}
	}
	return adjacent
}

// ResolveOpportunityAttack executes an opportunity attack, returning damage dealt.
// This is a simplified attack resolution that doesn't consume action points.
//
// Parameters:
//   - attacker: The character making the opportunity attack
//   - target: The character being attacked
//
// Returns:
//   - hit: Whether the attack hit
//   - damage: Damage dealt if hit
func ResolveOpportunityAttack(attacker, target *Character) (hit bool, damage int) {
	if attacker == nil || target == nil {
		return false, 0
	}

	// Simple d20 attack roll vs AC
	roll, err := GlobalDiceRoller.Roll("1d20")
	if err != nil {
		return false, 0
	}
	attackRoll := roll.Final

	// Get strength modifier for melee
	strMod := (attacker.Strength - 10) / 2

	totalAttack := attackRoll + strMod

	// Compare against target AC
	if totalAttack >= target.ArmorClass {
		hit = true
		// Base damage d6 + strength modifier
		damageRoll, err := GlobalDiceRoller.Roll("1d6")
		if err != nil {
			return true, 1
		}
		damage = damageRoll.Final + strMod
		if damage < 1 {
			damage = 1 // Minimum 1 damage on hit
		}

		logrus.WithFields(logrus.Fields{
			"function":   "ResolveOpportunityAttack",
			"attackerID": attacker.ID,
			"targetID":   target.ID,
			"roll":       attackRoll,
			"total":      totalAttack,
			"damage":     damage,
		}).Info("opportunity attack hit")
	} else {
		logrus.WithFields(logrus.Fields{
			"function":   "ResolveOpportunityAttack",
			"attackerID": attacker.ID,
			"targetID":   target.ID,
			"roll":       attackRoll,
			"total":      totalAttack,
			"targetAC":   target.ArmorClass,
		}).Debug("opportunity attack missed")
	}

	return hit, damage
}
