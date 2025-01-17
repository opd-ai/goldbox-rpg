package server

import (
	"goldbox-rpg/pkg/game"
	"regexp"
	"sort"
	"strconv"

	"golang.org/x/exp/rand"
)

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

func (s *RPCServer) getActiveEffects(player *game.Player) []*game.Effect {
	if holder, ok := interface{}(player).(game.EffectHolder); ok {
		return holder.GetEffects()
	}
	return nil
}

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

func (s *RPCServer) isPositionVisible(from, to game.Position) bool {
	dx := from.X - to.X
	dy := from.Y - to.Y
	distanceSquared := dx*dx + dy*dy

	return distanceSquared <= 100 && from.Level == to.Level
}

func (s *RPCServer) processEndTurnEffects(character game.GameObject) {
	if holder, ok := character.(game.EffectHolder); ok {
		for _, effect := range holder.GetEffects() {
			if effect.ShouldTick(s.state.TimeManager.CurrentTime.RealTime) {
				s.state.processEffectTick(effect)
			}
		}
	}
}

func (s *RPCServer) processEndRound() {
	s.state.TurnManager.CurrentRound++
	s.processDelayedActions()
	s.checkCombatEnd()
}

func isTimeToExecute(current, trigger game.GameTime) bool {
	return current.GameTicks >= trigger.GameTicks
}

func findSpell(spells []game.Spell, spellID string) *game.Spell {
	for i := range spells {
		if spells[i].ID == spellID {
			return &spells[i]
		}
	}
	return nil
}

func findInventoryItem(inventory []game.Item, itemID string) *game.Item {
	for i := range inventory {
		if inventory[i].ID == itemID {
			return &inventory[i]
		}
	}
	return nil
}

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
