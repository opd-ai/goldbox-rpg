// Package game provides morale system for tactical combat with NPCs.
package game

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// MoraleState represents the morale level of an NPC.
type MoraleState int

const (
	// MoraleSteadfast represents high morale - NPC fights normally.
	MoraleSteadfast MoraleState = iota
	// MoraleShaken represents reduced morale - NPC may fight defensively.
	MoraleShaken
	// MoraleBroken represents broken morale - NPC will attempt to flee.
	MoraleBroken
	// MoralePanicked represents complete panic - NPC flees at full speed.
	MoralePanicked
)

// MoraleThreshold defines the morale point thresholds for state transitions.
const (
	MoraleThresholdSteadfast = 70 // Above this is Steadfast
	MoraleThresholdShaken    = 40 // Above this is Shaken
	MoraleThresholdBroken    = 20 // Above this is Broken, below is Panicked
)

// MoraleModifier represents events that affect morale.
type MoraleModifier int

const (
	// MoraleModAllyDeath reduces morale when an ally dies.
	MoraleModAllyDeath MoraleModifier = -15
	// MoraleModAllyFlee reduces morale when an ally flees.
	MoraleModAllyFlee MoraleModifier = -10
	// MoraleModDamageTaken reduces morale based on damage severity.
	MoraleModDamageTaken MoraleModifier = -5
	// MoraleModLeaderPresent boosts morale when a leader is nearby.
	MoraleModLeaderPresent MoraleModifier = 10
	// MoraleModVictory boosts morale when defeating an enemy.
	MoraleModVictory MoraleModifier = 5
	// MoraleModCriticalHit reduces morale when critically hit.
	MoraleModCriticalHit MoraleModifier = -10
	// MoraleModSurrounded reduces morale when surrounded by enemies.
	MoraleModSurrounded MoraleModifier = -5
	// MoraleModHealReceived boosts morale when healed.
	MoraleModHealReceived MoraleModifier = 5
)

// MoraleSystem manages morale for all NPCs in combat.
type MoraleSystem struct {
	mu          sync.RWMutex
	moraleScore map[string]int    // Entity ID to current morale (0-100)
	factions    map[string]string // Entity ID to faction
	leaders     map[string]bool   // Entity IDs that are leaders
}

// NewMoraleSystem creates a new morale management system.
func NewMoraleSystem() *MoraleSystem {
	return &MoraleSystem{
		moraleScore: make(map[string]int),
		factions:    make(map[string]string),
		leaders:     make(map[string]bool),
	}
}

// RegisterNPC adds an NPC to the morale system with initial morale.
//
// Parameters:
//   - npcID: Unique identifier for the NPC
//   - faction: The NPC's faction for group morale effects
//   - isLeader: Whether this NPC is a leader (affects nearby allies' morale)
//   - initialMorale: Starting morale score (0-100)
func (ms *MoraleSystem) RegisterNPC(npcID, faction string, isLeader bool, initialMorale int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	morale := clampMorale(initialMorale)
	ms.moraleScore[npcID] = morale
	ms.factions[npcID] = faction
	ms.leaders[npcID] = isLeader

	logrus.WithFields(logrus.Fields{
		"function": "RegisterNPC",
		"npcID":    npcID,
		"faction":  faction,
		"isLeader": isLeader,
		"morale":   morale,
	}).Debug("NPC registered for morale tracking")
}

// UnregisterNPC removes an NPC from the morale system.
func (ms *MoraleSystem) UnregisterNPC(npcID string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	delete(ms.moraleScore, npcID)
	delete(ms.factions, npcID)
	delete(ms.leaders, npcID)

	logrus.WithFields(logrus.Fields{
		"function": "UnregisterNPC",
		"npcID":    npcID,
	}).Debug("NPC unregistered from morale tracking")
}

// ApplyMoraleModifier adjusts an NPC's morale based on a combat event.
//
// Parameters:
//   - npcID: The NPC whose morale is affected
//   - modifier: The morale change amount
//   - wisdomMod: Wisdom modifier for resistance (reduces negative effects)
//
// Returns:
//   - newMorale: The resulting morale score
//   - stateChanged: Whether the morale state changed
func (ms *MoraleSystem) ApplyMoraleModifier(npcID string, modifier MoraleModifier, wisdomMod int) (int, bool) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	currentMorale, exists := ms.moraleScore[npcID]
	if !exists {
		return 0, false
	}

	oldState := getMoraleState(currentMorale)

	// Wisdom provides resistance to negative morale effects
	change := int(modifier)
	if change < 0 && wisdomMod > 0 {
		// Reduce negative effect by wisdom modifier (minimum 1 point reduction)
		reduction := change * wisdomMod / 10
		if reduction > -1 {
			reduction = -1
		}
		change -= reduction // Makes negative change less severe
	}

	newMorale := clampMorale(currentMorale + change)
	ms.moraleScore[npcID] = newMorale

	newState := getMoraleState(newMorale)
	stateChanged := oldState != newState

	logrus.WithFields(logrus.Fields{
		"function":     "ApplyMoraleModifier",
		"npcID":        npcID,
		"oldMorale":    currentMorale,
		"newMorale":    newMorale,
		"modifier":     modifier,
		"wisdomMod":    wisdomMod,
		"stateChanged": stateChanged,
	}).Debug("morale modifier applied")

	return newMorale, stateChanged
}

// GetMoraleState returns the current morale state for an NPC.
func (ms *MoraleSystem) GetMoraleState(npcID string) MoraleState {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	morale, exists := ms.moraleScore[npcID]
	if !exists {
		return MoraleSteadfast // Default for untracked entities
	}

	return getMoraleState(morale)
}

// GetMoraleScore returns the raw morale score for an NPC.
func (ms *MoraleSystem) GetMoraleScore(npcID string) int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.moraleScore[npcID]
}

// OnAllyDeath processes morale effects when an ally dies in combat.
// All faction members lose morale based on proximity and relationship.
func (ms *MoraleSystem) OnAllyDeath(deadNPCID string, factionAllies []string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	deadFaction := ms.factions[deadNPCID]
	wasLeader := ms.leaders[deadNPCID]

	for _, allyID := range factionAllies {
		if allyID == deadNPCID {
			continue
		}

		if ms.factions[allyID] != deadFaction {
			continue
		}

		modifier := MoraleModAllyDeath
		if wasLeader {
			modifier *= 2 // Double penalty for leader death
		}

		currentMorale := ms.moraleScore[allyID]
		newMorale := clampMorale(currentMorale + int(modifier))
		ms.moraleScore[allyID] = newMorale

		logrus.WithFields(logrus.Fields{
			"function":  "OnAllyDeath",
			"allyID":    allyID,
			"deadID":    deadNPCID,
			"wasLeader": wasLeader,
			"newMorale": newMorale,
		}).Debug("ally death morale effect applied")
	}
}

// OnEnemyDefeated processes morale boost when defeating an enemy.
func (ms *MoraleSystem) OnEnemyDefeated(victorID string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	currentMorale, exists := ms.moraleScore[victorID]
	if !exists {
		return
	}

	newMorale := clampMorale(currentMorale + int(MoraleModVictory))
	ms.moraleScore[victorID] = newMorale

	logrus.WithFields(logrus.Fields{
		"function":  "OnEnemyDefeated",
		"victorID":  victorID,
		"newMorale": newMorale,
	}).Debug("victory morale boost applied")
}

// CheckLeaderBonus applies morale bonus if a leader is nearby.
//
// Parameters:
//   - npcID: The NPC to check for leader bonus
//   - nearbyEntities: Entity IDs of nearby allies
//
// Returns:
//   - bool: Whether a leader bonus was applied
func (ms *MoraleSystem) CheckLeaderBonus(npcID string, nearbyEntities []string) bool {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	npcFaction := ms.factions[npcID]

	for _, entityID := range nearbyEntities {
		if !ms.leaders[entityID] {
			continue
		}

		if ms.factions[entityID] != npcFaction {
			continue
		}

		// Apply leader bonus
		currentMorale := ms.moraleScore[npcID]
		newMorale := clampMorale(currentMorale + int(MoraleModLeaderPresent))
		ms.moraleScore[npcID] = newMorale

		logrus.WithFields(logrus.Fields{
			"function":  "CheckLeaderBonus",
			"npcID":     npcID,
			"leaderID":  entityID,
			"newMorale": newMorale,
		}).Debug("leader presence bonus applied")

		return true
	}

	return false
}

// ShouldFlee determines if an NPC should attempt to flee based on morale.
//
// Parameters:
//   - npcID: The NPC to check
//   - charismaMod: Charisma modifier of the NPC's leader (if any) for rally chance
//
// Returns:
//   - bool: True if the NPC should flee
func (ms *MoraleSystem) ShouldFlee(npcID string, charismaMod int) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	state := ms.GetMoraleState(npcID)

	switch state {
	case MoralePanicked:
		return true // Always flee when panicked
	case MoraleBroken:
		// 75% chance to flee, charisma can reduce
		fleeChance := 75 - (charismaMod * 5)
		roll, err := GlobalDiceRoller.Roll("1d100")
		if err != nil {
			return true // Default to flee on error
		}
		return roll.Final <= fleeChance
	case MoraleShaken:
		// 25% chance to flee, charisma can reduce
		fleeChance := 25 - (charismaMod * 5)
		if fleeChance < 0 {
			fleeChance = 0
		}
		roll, err := GlobalDiceRoller.Roll("1d100")
		if err != nil {
			return false // Default to not flee on error
		}
		return roll.Final <= fleeChance
	default:
		return false
	}
}

// ResetMorale restores an NPC's morale to a default value.
// Used when combat ends or for special abilities.
func (ms *MoraleSystem) ResetMorale(npcID string, newMorale int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.moraleScore[npcID] = clampMorale(newMorale)

	logrus.WithFields(logrus.Fields{
		"function":  "ResetMorale",
		"npcID":     npcID,
		"newMorale": newMorale,
	}).Debug("morale reset")
}

// GetFactionMorale returns the average morale of all NPCs in a faction.
func (ms *MoraleSystem) GetFactionMorale(faction string) int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	total := 0
	count := 0

	for npcID, npcFaction := range ms.factions {
		if npcFaction == faction {
			total += ms.moraleScore[npcID]
			count++
		}
	}

	if count == 0 {
		return 100 // No NPCs means no morale issues
	}

	return total / count
}

// IsLeader returns whether an NPC is a leader.
func (ms *MoraleSystem) IsLeader(npcID string) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.leaders[npcID]
}

// getMoraleState converts a morale score to a morale state.
func getMoraleState(morale int) MoraleState {
	switch {
	case morale >= MoraleThresholdSteadfast:
		return MoraleSteadfast
	case morale >= MoraleThresholdShaken:
		return MoraleShaken
	case morale >= MoraleThresholdBroken:
		return MoraleBroken
	default:
		return MoralePanicked
	}
}

// clampMorale ensures morale stays within valid range (0-100).
func clampMorale(morale int) int {
	if morale < 0 {
		return 0
	}
	if morale > 100 {
		return 100
	}
	return morale
}

// MoraleStateString returns a human-readable string for a morale state.
func MoraleStateString(state MoraleState) string {
	switch state {
	case MoraleSteadfast:
		return "Steadfast"
	case MoraleShaken:
		return "Shaken"
	case MoraleBroken:
		return "Broken"
	case MoralePanicked:
		return "Panicked"
	default:
		return "Unknown"
	}
}
