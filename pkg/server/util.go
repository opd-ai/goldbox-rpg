package server

import (
	"fmt"

	"goldbox-rpg/pkg/game"
)

// Additional EventType constants
const (
	EventCombatStart game.EventType = 100 + iota
	EventCombatEnd
	EventTurnStart
	EventTurnEnd
	EventMovement
)

// Add methods to TurnManager
func (tm *TurnManager) IsCurrentTurn(entityID string) bool {
	if !tm.IsInCombat || tm.CurrentIndex >= len(tm.Initiative) {
		return false
	}
	return tm.Initiative[tm.CurrentIndex] == entityID
}

func (tm *TurnManager) StartCombat(initiative []string) {
	tm.IsInCombat = true
	tm.Initiative = initiative
	tm.CurrentIndex = 0
	tm.CurrentRound = 1
}

func (tm *TurnManager) AdvanceTurn() string {
	if !tm.IsInCombat {
		return ""
	}

	tm.CurrentIndex = (tm.CurrentIndex + 1) % len(tm.Initiative)
	if tm.CurrentIndex == 0 {
		tm.CurrentRound++
	}

	return tm.Initiative[tm.CurrentIndex]
}

// Add helper methods to RPCServer
func (s *RPCServer) processSpellCast(caster *game.Player, spell *game.Spell, targetID string, pos game.Position) (interface{}, error) {
	// Validate spell requirements
	if err := s.validateSpellCast(caster, spell); err != nil {
		return nil, err
	}

	// Process spell effects based on type
	switch spell.School {
	case game.SchoolEvocation:
		return s.processEvocationSpell(spell, caster, targetID)
	case game.SchoolEnchantment:
		return s.processEnchantmentSpell(spell, caster, targetID)
	case game.SchoolIllusion:
		return s.processIllusionSpell(spell, caster, pos)
	default:
		return s.processGenericSpell(spell, caster, targetID)
	}
}

func (s *RPCServer) validateSpellCast(caster *game.Player, spell *game.Spell) error {
	// Check level requirements
	if caster.Level < spell.Level {
		return fmt.Errorf("insufficient level to cast spell")
	}

	// Check components
	for _, component := range spell.Components {
		if !s.hasSpellComponent(caster, component) {
			return fmt.Errorf("missing required spell component: %v", component)
		}
	}

	return nil
}

func (s *RPCServer) getVisibleObjects(player *game.Player) []game.GameObject {
	playerPos := player.GetPosition()
	visibleObjects := make([]game.GameObject, 0)

	// Get objects in visible range
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
		CombatZone:       player.GetPosition(), // Center on player
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
	// Implement line of sight checking
	// This is a simple distance check - replace with proper LoS algorithm
	dx := from.X - to.X
	dy := from.Y - to.Y
	distanceSquared := dx*dx + dy*dy

	// Arbitrary visibility radius of 10 tiles
	return distanceSquared <= 100 && from.Level == to.Level
}

func (s *RPCServer) hasSpellComponent(caster *game.Player, component game.SpellComponent) bool {
	// For verbal/somatic components, check if character is able to speak/move
	if component == game.ComponentVerbal || component == game.ComponentSomatic {
		return !s.isCharacterImpaired(caster)
	}

	// For material components, check inventory
	if component == game.ComponentMaterial {
		// Implementation depends on how material components are tracked
		return true // Simplified for now
	}

	return false
}

func (s *RPCServer) isCharacterImpaired(character *game.Player) bool {
	if holder, ok := interface{}(character).(game.EffectHolder); ok {
		for _, effect := range holder.GetEffects() {
			if effect.Type == game.EffectStun || effect.Type == game.EffectRoot {
				return true
			}
		}
	}
	return false
}

// Spell processing methods
func (s *RPCServer) processEvocationSpell(spell *game.Spell, caster *game.Player, targetID string) (interface{}, error) {
	// Implement damage/healing spells
	return map[string]interface{}{
		"success":  true,
		"spell_id": spell.ID,
	}, nil
}

func (s *RPCServer) processEnchantmentSpell(spell *game.Spell, caster *game.Player, targetID string) (interface{}, error) {
	// Implement buff/debuff spells
	return map[string]interface{}{
		"success":  true,
		"spell_id": spell.ID,
	}, nil
}

func (s *RPCServer) processIllusionSpell(spell *game.Spell, caster *game.Player, pos game.Position) (interface{}, error) {
	// Implement area effect spells
	return map[string]interface{}{
		"success":  true,
		"spell_id": spell.ID,
	}, nil
}

func (s *RPCServer) processGenericSpell(spell *game.Spell, caster *game.Player, targetID string) (interface{}, error) {
	// Default spell processing
	return map[string]interface{}{
		"success":  true,
		"spell_id": spell.ID,
	}, nil
}
