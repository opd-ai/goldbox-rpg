Project Path: server

Source Tree:

```
server
├── doc.md
├── state.go
├── util.go
├── handlers.go
├── new.go
└── server.go

```

`/home/user/go/src/github.com/opd-ai/goldbox-rpg/pkg/server/doc.md`:

```md
# server
--
    import "github.com/opd-ai/goldbox-rpg/pkg/server"


## Usage

#### type CombatState

```go
type CombatState struct {
	ActiveCombatants []string                 `yaml:"combat_active_entities"` // Entities in combat
	RoundCount       int                      `yaml:"combat_round_count"`     // Number of rounds
	CombatZone       game.Position            `yaml:"combat_zone_center"`     // Combat area center
	StatusEffects    map[string][]game.Effect `yaml:"combat_status_effects"`  // Active effects
}
```

CombatState tracks active combat information

#### type DelayedAction

```go
type DelayedAction struct {
	ActorID     string        `yaml:"action_actor_id"`     // Entity performing action
	ActionType  string        `yaml:"action_type"`         // Type of action
	Target      game.Position `yaml:"action_target_pos"`   // Target location
	TriggerTime game.GameTime `yaml:"action_trigger_time"` // When to execute
	Parameters  []string      `yaml:"action_parameters"`   // Additional data
}
```

DelayedAction represents a pending combat action

#### type GameState

```go
type GameState struct {
	WorldState  *game.World               `yaml:"state_world"`    // Current world state
	TurnManager *TurnManager              `yaml:"state_turns"`    // Turn management
	TimeManager *TimeManager              `yaml:"state_time"`     // Time tracking
	Sessions    map[string]*PlayerSession `yaml:"state_sessions"` // Active player sessions
}
```

GameState represents the complete server-side game state

#### type PlayerSession

```go
type PlayerSession struct {
	SessionID  string       `yaml:"session_id"`  // Unique session identifier
	Player     *game.Player `yaml:"player"`      // Associated player
	LastActive time.Time    `yaml:"last_active"` // Last activity timestamp
	Connected  bool         `yaml:"connected"`   // Connection status
}
```

PlayerSession represents an active player connection

#### type ScheduledEvent

```go
type ScheduledEvent struct {
	EventID     string        `yaml:"event_id"`           // Event identifier
	EventType   string        `yaml:"event_type"`         // Type of event
	TriggerTime game.GameTime `yaml:"event_trigger_time"` // When to trigger
	Parameters  []string      `yaml:"event_parameters"`   // Event data
	Repeating   bool          `yaml:"event_is_repeating"` // Whether it repeats
}
```

ScheduledEvent represents a future game event

#### type ScriptContext

```go
type ScriptContext struct {
	ScriptID     string                 `yaml:"script_id"`            // Script identifier
	Variables    map[string]interface{} `yaml:"script_variables"`     // Script state
	LastExecuted time.Time              `yaml:"script_last_executed"` // Last run timestamp
	IsActive     bool                   `yaml:"script_is_active"`     // Execution state
}
```

ScriptContext represents the NPC behavior script state

#### type StateUpdate

```go
type StateUpdate struct {
	UpdateType string                 `yaml:"update_type"`      // Type of update
	EntityID   string                 `yaml:"update_entity_id"` // Affected entity
	ChangeData map[string]interface{} `yaml:"update_data"`      // Update details
	Timestamp  time.Time              `yaml:"update_timestamp"` // When it occurred
}
```

StateUpdate represents a game state change notification

#### type TimeManager

```go
type TimeManager struct {
	CurrentTime     game.GameTime    `yaml:"time_current"`          // Current game time
	TimeScale       float64          `yaml:"time_scale"`            // Time progression rate
	LastTick        time.Time        `yaml:"time_last_tick"`        // Last update time
	ScheduledEvents []ScheduledEvent `yaml:"time_scheduled_events"` // Pending events
}
```

TimeManager handles game time progression and scheduled events

#### type TurnManager

```go
type TurnManager struct {
	CurrentRound   int                 `yaml:"turn_current_round"`    // Active combat round
	Initiative     []string            `yaml:"turn_initiative_order"` // Turn order by entity ID
	CurrentIndex   int                 `yaml:"turn_current_index"`    // Current actor index
	IsInCombat     bool                `yaml:"turn_in_combat"`        // Combat state flag
	CombatGroups   map[string][]string `yaml:"turn_combat_groups"`    // Allied entities
	DelayedActions []DelayedAction     `yaml:"turn_delayed_actions"`  // Pending actions
}
```

TurnManager handles combat turns and initiative ordering

```

`/home/user/go/src/github.com/opd-ai/goldbox-rpg/pkg/server/state.go`:

```go
package server

import (
	"sync"
	"time"

	"goldbox-rpg/pkg/game"
)

// PlayerSession represents an active player connection
type PlayerSession struct {
	SessionID  string       `yaml:"session_id"`  // Unique session identifier
	Player     *game.Player `yaml:"player"`      // Associated player
	LastActive time.Time    `yaml:"last_active"` // Last activity timestamp
	Connected  bool         `yaml:"connected"`   // Connection status
}

// GameState represents the complete server-side game state
type GameState struct {
	WorldState  *game.World               `yaml:"state_world"`    // Current world state
	TurnManager *TurnManager              `yaml:"state_turns"`    // Turn management
	TimeManager *TimeManager              `yaml:"state_time"`     // Time tracking
	Sessions    map[string]*PlayerSession `yaml:"state_sessions"` // Active player sessions
	mu          sync.RWMutex              `yaml:"-"`              // State mutex
	updates     chan StateUpdate          `yaml:"-"`              // Update channel
}

// TurnManager handles combat turns and initiative ordering
type TurnManager struct {
	CurrentRound   int                 `yaml:"turn_current_round"`    // Active combat round
	Initiative     []string            `yaml:"turn_initiative_order"` // Turn order by entity ID
	CurrentIndex   int                 `yaml:"turn_current_index"`    // Current actor index
	IsInCombat     bool                `yaml:"turn_in_combat"`        // Combat state flag
	CombatGroups   map[string][]string `yaml:"turn_combat_groups"`    // Allied entities
	DelayedActions []DelayedAction     `yaml:"turn_delayed_actions"`  // Pending actions
}

// TimeManager handles game time progression and scheduled events
type TimeManager struct {
	CurrentTime     game.GameTime    `yaml:"time_current"`          // Current game time
	TimeScale       float64          `yaml:"time_scale"`            // Time progression rate
	LastTick        time.Time        `yaml:"time_last_tick"`        // Last update time
	ScheduledEvents []ScheduledEvent `yaml:"time_scheduled_events"` // Pending events
}

// CombatState tracks active combat information
type CombatState struct {
	ActiveCombatants []string                 `yaml:"combat_active_entities"` // Entities in combat
	RoundCount       int                      `yaml:"combat_round_count"`     // Number of rounds
	CombatZone       game.Position            `yaml:"combat_zone_center"`     // Combat area center
	StatusEffects    map[string][]game.Effect `yaml:"combat_status_effects"`  // Active effects
}

// ScriptContext represents the NPC behavior script state
type ScriptContext struct {
	ScriptID     string                 `yaml:"script_id"`            // Script identifier
	Variables    map[string]interface{} `yaml:"script_variables"`     // Script state
	LastExecuted time.Time              `yaml:"script_last_executed"` // Last run timestamp
	IsActive     bool                   `yaml:"script_is_active"`     // Execution state
}

// StateUpdate represents a game state change notification
type StateUpdate struct {
	UpdateType string                 `yaml:"update_type"`      // Type of update
	EntityID   string                 `yaml:"update_entity_id"` // Affected entity
	ChangeData map[string]interface{} `yaml:"update_data"`      // Update details
	Timestamp  time.Time              `yaml:"update_timestamp"` // When it occurred
}

// DelayedAction represents a pending combat action
type DelayedAction struct {
	ActorID     string        `yaml:"action_actor_id"`     // Entity performing action
	ActionType  string        `yaml:"action_type"`         // Type of action
	Target      game.Position `yaml:"action_target_pos"`   // Target location
	TriggerTime game.GameTime `yaml:"action_trigger_time"` // When to execute
	Parameters  []string      `yaml:"action_parameters"`   // Additional data
}

// ScheduledEvent represents a future game event
type ScheduledEvent struct {
	EventID     string        `yaml:"event_id"`           // Event identifier
	EventType   string        `yaml:"event_type"`         // Type of event
	TriggerTime game.GameTime `yaml:"event_trigger_time"` // When to trigger
	Parameters  []string      `yaml:"event_parameters"`   // Event data
	Repeating   bool          `yaml:"event_is_repeating"` // Whether it repeats
}

```

`/home/user/go/src/github.com/opd-ai/goldbox-rpg/pkg/server/util.go`:

```go
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

```

`/home/user/go/src/github.com/opd-ai/goldbox-rpg/pkg/server/handlers.go`:

```go
package server

import (
	"encoding/json"
	"fmt"
	"sort"

	"goldbox-rpg/pkg/game"

	"golang.org/x/exp/rand"
)

// Additional RPC methods
func (s *RPCServer) handleCastSpell(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string        `json:"session_id"`
		SpellID   string        `json:"spell_id"`
		TargetID  string        `json:"target_id"`
		Position  game.Position `json:"position,omitempty"` // For area spells
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid spell parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	// Validate spell casting
	player := session.Player
	spell := findSpell(player.KnownSpells, req.SpellID)
	if spell == nil {
		return nil, fmt.Errorf("spell not known")
	}

	// Validate turn if in combat
	if s.state.TurnManager.IsInCombat && !s.state.TurnManager.IsCurrentTurn(player.GetID()) {
		return nil, fmt.Errorf("not your turn")
	}

	// Process spell effects
	result, err := s.processSpellCast(player, spell, req.TargetID, req.Position)
	if err != nil {
		return nil, err
	}

	// Emit spell cast event
	s.eventSys.Emit(game.GameEvent{
		Type:     game.EventSpellCast,
		SourceID: player.GetID(),
		TargetID: req.TargetID,
		Data: map[string]interface{}{
			"spell_id": req.SpellID,
			"position": req.Position,
			"effects":  result,
		},
	})

	return result, nil
}

func (s *RPCServer) handleApplyEffect(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID  string          `json:"session_id"`
		EffectType game.EffectType `json:"effect_type"`
		TargetID   string          `json:"target_id"`
		Magnitude  float64         `json:"magnitude"`
		Duration   game.Duration   `json:"duration"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid effect parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	// Create and apply the effect
	effect := game.NewEffect(req.EffectType, req.Duration, req.Magnitude)
	effect.SourceID = session.Player.GetID()

	target, exists := s.state.WorldState.Objects[req.TargetID]
	if !exists {
		return nil, fmt.Errorf("invalid target")
	}

	effectHolder, ok := target.(game.EffectHolder)
	if !ok {
		return nil, fmt.Errorf("target cannot receive effects")
	}

	if err := effectHolder.AddEffect(effect); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success":   true,
		"effect_id": effect.ID,
	}, nil
}

func (s *RPCServer) handleStartCombat(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID    string   `json:"session_id"`
		Participants []string `json:"participant_ids"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid combat parameters")
	}

	if s.state.TurnManager.IsInCombat {
		return nil, fmt.Errorf("combat already in progress")
	}

	// Initialize combat state
	combatState := &CombatState{
		ActiveCombatants: req.Participants,
		RoundCount:       1,
		StatusEffects:    make(map[string][]game.Effect),
	}

	// Roll initiative for all participants
	initiative := s.rollInitiative(req.Participants)
	s.state.TurnManager.StartCombat(initiative)

	// Emit combat start event
	s.eventSys.Emit(game.GameEvent{
		Type: game.EventCombatStart,
		Data: map[string]interface{}{
			"participants": req.Participants,
			"initiative":   initiative,
		},
	})

	return map[string]interface{}{
		"success":    true,
		"initiative": initiative,
		"first_turn": initiative[0],
	}, nil
}

func (s *RPCServer) handleEndTurn(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid turn parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	if !s.state.TurnManager.IsInCombat {
		return nil, fmt.Errorf("not in combat")
	}

	if !s.state.TurnManager.IsCurrentTurn(session.Player.GetID()) {
		return nil, fmt.Errorf("not your turn")
	}

	// Process end of turn effects
	s.processEndTurnEffects(session.Player)

	// Advance to next turn
	nextTurn := s.state.TurnManager.AdvanceTurn()

	// Check for round end
	if s.state.TurnManager.CurrentIndex == 0 {
		s.processEndRound()
	}

	return map[string]interface{}{
		"success":   true,
		"next_turn": nextTurn,
	}, nil
}

func (s *RPCServer) handleGetGameState(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid state request parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	// Get visible game state for player
	player := session.Player
	visibleObjects := s.getVisibleObjects(player)
	activeEffects := s.getActiveEffects(player)
	combatState := s.getCombatStateIfActive(player)

	return map[string]interface{}{
		"player": map[string]interface{}{
			"position":   player.GetPosition(),
			"stats":      player.GetStats(),
			"effects":    activeEffects,
			"inventory":  player.Inventory,
			"spells":     player.KnownSpells,
			"experience": player.Experience,
		},
		"world": map[string]interface{}{
			"visible_objects": visibleObjects,
			"current_time":    s.state.TimeManager.CurrentTime,
			"combat_state":    combatState,
		},
	}, nil
}

func (s *RPCServer) handleUseItem(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string `json:"session_id"`
		ItemID    string `json:"item_id"`
		TargetID  string `json:"target_id,omitempty"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid item parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	// Validate item ownership and usage
	item := findInventoryItem(session.Player.Inventory, req.ItemID)
	if item == nil {
		return nil, fmt.Errorf("item not found in inventory")
	}

	// Process item usage
	result, err := s.processItemUse(session.Player, item, req.TargetID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Helper functions for RPC methods
func (s *RPCServer) rollInitiative(participants []string) []string {
	type initiativeRoll struct {
		entityID string
		roll     int
	}

	rolls := make([]initiativeRoll, len(participants))
	for i, id := range participants {
		if obj, exists := s.state.WorldState.Objects[id]; exists {
			// Base roll on dexterity if character
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

	// Sort by initiative roll
	sort.Slice(rolls, func(i, j int) bool {
		return rolls[i].roll > rolls[j].roll
	})

	// Extract sorted IDs
	result := make([]string, len(rolls))
	for i, roll := range rolls {
		result[i] = roll.entityID
	}

	return result
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
	s.state.TurnManager.RoundCount++
	s.processDelayedActions()
	s.checkCombatEnd()
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

```

`/home/user/go/src/github.com/opd-ai/goldbox-rpg/pkg/server/new.go`:

```go
package server

import (
	"fmt"
	"time"

	"goldbox-rpg/pkg/game"
)

func NewTimeManager() *TimeManager {
	return &TimeManager{
		CurrentTime: game.GameTime{
			RealTime:  time.Now(),
			GameTicks: 0,
			TimeScale: 1.0,
		},
		TimeScale:       1.0,
		LastTick:        time.Now(),
		ScheduledEvents: make([]ScheduledEvent, 0),
	}
}

// Add these methods to GameState
func (gs *GameState) processEffectTick(effect *game.Effect) error {
	if effect == nil {
		return fmt.Errorf("nil effect")
	}

	switch effect.Type {
	case game.EffectDamageOverTime:
		return gs.processDamageEffect(effect)
	case game.EffectHealOverTime:
		return gs.processHealEffect(effect)
	case game.EffectStatBoost, game.EffectStatPenalty:
		return gs.processStatEffect(effect)
	default:
		return fmt.Errorf("unknown effect type: %s", effect.Type)
	}
}

// Add to RPCServer
func (s *RPCServer) processItemUse(player *game.Player, item *game.Item, targetID string) (interface{}, error) {
	switch item.Type {
	case game.ItemTypeWeapon:
		return s.processWeaponUse(player, item, targetID)
	case game.ItemTypeArmor:
		return s.processArmorUse(player, item)
	default:
		return s.processConsumableUse(player, item, targetID)
	}
}

func (s *RPCServer) processDelayedActions() {
	currentTime := s.state.TimeManager.CurrentTime

	for i := len(s.state.TurnManager.DelayedActions) - 1; i >= 0; i-- {
		action := s.state.TurnManager.DelayedActions[i]
		if isTimeToExecute(currentTime, action.TriggerTime) {
			s.executeDelayedAction(action)
			// Remove executed action
			s.state.TurnManager.DelayedActions = append(
				s.state.TurnManager.DelayedActions[:i],
				s.state.TurnManager.DelayedActions[i+1:]...,
			)
		}
	}
}

func (s *RPCServer) checkCombatEnd() bool {
	if !s.state.TurnManager.IsInCombat {
		return false
	}

	// Check if combat should end
	hostileGroups := s.getHostileGroups()
	if len(hostileGroups) <= 1 {
		s.endCombat()
		return true
	}
	return false
}

// Add helper methods
func (s *RPCServer) processWeaponUse(player *game.Player, weapon *game.Item, targetID string) (interface{}, error) {
	target, exists := s.state.WorldState.Objects[targetID]
	if !exists {
		return nil, fmt.Errorf("invalid target")
	}

	// Calculate damage
	damage := calculateWeaponDamage(weapon, player)

	// Apply damage to target
	if err := s.applyDamage(target, damage); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
		"damage":  damage,
	}, nil
}

func (s *RPCServer) processArmorUse(player *game.Player, armor *game.Item) (interface{}, error) {
	// Equip armor
	slot := determineArmorSlot(armor)
	if err := s.equipItem(player, armor, slot); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
		"slot":    slot,
	}, nil
}

func (s *RPCServer) processConsumableUse(player *game.Player, item *game.Item, targetID string) (interface{}, error) {
	// Apply item effects
	effects, err := s.applyItemEffects(player, item, targetID)
	if err != nil {
		return nil, err
	}

	// Remove consumed item
	s.removeItemFromInventory(player, item)

	return map[string]interface{}{
		"success": true,
		"effects": effects,
	}, nil
}

func (s *RPCServer) getHostileGroups() [][]string {
	groups := make([][]string, 0)
	processed := make(map[string]bool)

	for id := range s.state.TurnManager.CombatGroups {
		if !processed[id] {
			group := s.state.TurnManager.CombatGroups[id]
			groups = append(groups, group)
			for _, memberID := range group {
				processed[memberID] = true
			}
		}
	}

	return groups
}

func (s *RPCServer) endCombat() {
	s.state.TurnManager.IsInCombat = false
	s.state.TurnManager.Initiative = nil
	s.state.TurnManager.CurrentIndex = 0

	// Emit combat end event
	s.eventSys.Emit(game.GameEvent{
		Type: EventCombatEnd,
		Data: map[string]interface{}{
			"rounds_completed": s.state.TurnManager.CurrentRound,
		},
	})
}

func isTimeToExecute(current, trigger game.GameTime) bool {
	return current.GameTicks >= trigger.GameTicks
}

func calculateWeaponDamage(weapon *game.Item, attacker *game.Player) int {
	// Basic damage calculation
	baseDamage := parseDamageString(weapon.Damage)
	strBonus := (attacker.Strength - 10) / 2
	return baseDamage + strBonus
}

func determineArmorSlot(armor *game.Item) game.EquipmentSlot {
	// Determine appropriate slot based on armor type
	switch armor.Type {
	case "helmet":
		return game.SlotHead
	case "chest":
		return game.SlotChest
	case "gloves":
		return game.SlotHands
	case "boots":
		return game.SlotFeet
	default:
		return game.SlotChest
	}
}

```

`/home/user/go/src/github.com/opd-ai/goldbox-rpg/pkg/server/server.go`:

```go
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"goldbox-rpg/pkg/game"
)

// RPCServer handles all game server functionality
type RPCServer struct {
	state      *GameState
	eventSys   *game.EventSystem
	mu         sync.RWMutex
	timekeeper *TimeManager

	// Session management
	sessions map[string]*PlayerSession
}

// RPCMethod represents available RPC methods
type RPCMethod string

const (
	MethodMove         RPCMethod = "move"
	MethodAttack       RPCMethod = "attack"
	MethodCastSpell    RPCMethod = "castSpell"
	MethodUseItem      RPCMethod = "useItem"
	MethodApplyEffect  RPCMethod = "applyEffect"
	MethodStartCombat  RPCMethod = "startCombat"
	MethodEndTurn      RPCMethod = "endTurn"
	MethodGetGameState RPCMethod = "getGameState"
	MethodJoinGame     RPCMethod = "joinGame"
	MethodLeaveGame    RPCMethod = "leaveGame"
)

// NewRPCServer creates a new game server instance
func NewRPCServer() *RPCServer {
	return &RPCServer{
		state: &GameState{
			WorldState:  game.NewWorld(),
			TurnManager: &TurnManager{},
			TimeManager: NewTimeManager(),
			Sessions:    make(map[string]*PlayerSession),
		},
		eventSys:   game.NewEventSystem(),
		sessions:   make(map[string]*PlayerSession),
		timekeeper: NewTimeManager(),
	}
}

// ServeHTTP implements http.Handler
func (s *RPCServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		JsonRPC string          `json:"jsonrpc"`
		Method  RPCMethod       `json:"method"`
		Params  json.RawMessage `json:"params"`
		ID      interface{}     `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, -32700, "Parse error", nil)
		return
	}

	// Handle the RPC method
	result, err := s.handleMethod(req.Method, req.Params)
	if err != nil {
		writeError(w, -32603, err.Error(), nil)
		return
	}

	// Write successful response
	writeResponse(w, result, req.ID)
}

// handleMethod processes individual RPC methods
func (s *RPCServer) handleMethod(method RPCMethod, params json.RawMessage) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch method {
	case MethodMove:
		return s.handleMove(params)
	case MethodAttack:
		return s.handleAttack(params)
	case MethodCastSpell:
		return s.handleCastSpell(params)
	case MethodApplyEffect:
		return s.handleApplyEffect(params)
	case MethodStartCombat:
		return s.handleStartCombat(params)
	case MethodEndTurn:
		return s.handleEndTurn(params)
	case MethodGetGameState:
		return s.handleGetGameState(params)
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// handleMove processes movement requests
func (s *RPCServer) handleMove(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string         `json:"session_id"`
		Direction game.Direction `json:"direction"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid movement parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	player := session.Player
	currentPos := player.GetPosition()
	newPos := calculateNewPosition(currentPos, req.Direction)

	if err := s.state.WorldState.ValidateMove(player, newPos); err != nil {
		return nil, err
	}

	if err := player.SetPosition(newPos); err != nil {
		return nil, err
	}

	// Emit movement event
	s.eventSys.Emit(game.GameEvent{
		Type:     game.EventMovement,
		SourceID: player.GetID(),
		Data: map[string]interface{}{
			"old_position": currentPos,
			"new_position": newPos,
		},
	})

	return map[string]interface{}{
		"success":  true,
		"position": newPos,
	}, nil
}

// handleAttack processes combat actions
func (s *RPCServer) handleAttack(params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string `json:"session_id"`
		TargetID  string `json:"target_id"`
		WeaponID  string `json:"weapon_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid attack parameters")
	}

	session, exists := s.sessions[req.SessionID]
	if !exists {
		return nil, fmt.Errorf("invalid session")
	}

	// Validate combat state
	if !s.state.TurnManager.IsInCombat {
		return nil, fmt.Errorf("not in combat")
	}

	if !s.state.TurnManager.IsCurrentTurn(session.Player.GetID()) {
		return nil, fmt.Errorf("not your turn")
	}

	// Process attack
	result, err := s.processCombatAction(session.Player, req.TargetID, req.WeaponID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Helper functions
func writeResponse(w http.ResponseWriter, result interface{}, id interface{}) {
	response := struct {
		JsonRPC string      `json:"jsonrpc"`
		Result  interface{} `json:"result"`
		ID      interface{} `json:"id"`
	}{
		JsonRPC: "2.0",
		Result:  result,
		ID:      id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func writeError(w http.ResponseWriter, code int, message string, data interface{}) {
	response := struct {
		JsonRPC string `json:"jsonrpc"`
		Error   struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data,omitempty"`
		} `json:"error"`
		ID interface{} `json:"id"`
	}{
		JsonRPC: "2.0",
		Error: struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data,omitempty"`
		}{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

```