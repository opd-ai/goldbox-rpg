// Package game provides inter-faction diplomacy mechanics.
package game

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// DiplomaticState represents the diplomatic relationship between factions.
type DiplomaticState string

const (
	DiplomaticStateWar      DiplomaticState = "war"
	DiplomaticStateHostile  DiplomaticState = "hostile"
	DiplomaticStateTense    DiplomaticState = "tense"
	DiplomaticStateNeutral  DiplomaticState = "neutral"
	DiplomaticStateFriendly DiplomaticState = "friendly"
	DiplomaticStateAllied   DiplomaticState = "allied"
	DiplomaticStateVassal   DiplomaticState = "vassal"
	DiplomaticStateSuzerain DiplomaticState = "suzerain"
)

// DiplomaticAction represents actions that affect faction relations.
type DiplomaticAction string

const (
	ActionDeclareWar      DiplomaticAction = "declare_war"
	ActionOfferPeace      DiplomaticAction = "offer_peace"
	ActionAcceptPeace     DiplomaticAction = "accept_peace"
	ActionProposeAlliance DiplomaticAction = "propose_alliance"
	ActionAcceptAlliance  DiplomaticAction = "accept_alliance"
	ActionBreakAlliance   DiplomaticAction = "break_alliance"
	ActionTradeAgreement  DiplomaticAction = "trade_agreement"
	ActionSendGift        DiplomaticAction = "send_gift"
	ActionInsult          DiplomaticAction = "insult"
	ActionDemandTribute   DiplomaticAction = "demand_tribute"
	ActionPayTribute      DiplomaticAction = "pay_tribute"
)

// FactionRelation represents the diplomatic relationship between two factions.
type FactionRelation struct {
	mu               sync.RWMutex       `yaml:"-"`
	ID               string             `yaml:"relation_id"`
	Faction1ID       string             `yaml:"faction1_id"`
	Faction2ID       string             `yaml:"faction2_id"`
	State            DiplomaticState    `yaml:"state"`
	Opinion          int                `yaml:"opinion"` // -100 to +100
	Trust            int                `yaml:"trust"`   // -100 to +100
	TradeTreaty      bool               `yaml:"trade_treaty"`
	MilitaryAccess   bool               `yaml:"military_access"`
	DefensivePact    bool               `yaml:"defensive_pact"`
	TributeDirection string             `yaml:"tribute_direction"` // Faction ID paying tribute
	TributeAmount    int                `yaml:"tribute_amount"`
	History          []*DiplomaticEvent `yaml:"history"`
	LastInteraction  time.Time          `yaml:"last_interaction"`
}

// DiplomaticEvent records a historical diplomatic action.
type DiplomaticEvent struct {
	ID          string           `yaml:"event_id"`
	Timestamp   time.Time        `yaml:"timestamp"`
	Action      DiplomaticAction `yaml:"action"`
	InitiatorID string           `yaml:"initiator_id"`
	TargetID    string           `yaml:"target_id"`
	OldState    DiplomaticState  `yaml:"old_state"`
	NewState    DiplomaticState  `yaml:"new_state"`
	Details     string           `yaml:"details"`
}

// DiplomacyManager handles all inter-faction diplomatic relations.
type DiplomacyManager struct {
	mu        sync.RWMutex
	relations map[string]*FactionRelation // RelationKey -> Relation
	logger    *logrus.Logger
}

// NewDiplomacyManager creates a new diplomacy management system.
func NewDiplomacyManager(logger *logrus.Logger) *DiplomacyManager {
	if logger == nil {
		logger = logrus.New()
	}
	return &DiplomacyManager{
		relations: make(map[string]*FactionRelation),
		logger:    logger,
	}
}

// Diplomacy errors.
var (
	ErrRelationNotFound   = errors.New("diplomatic relation not found")
	ErrInvalidAction      = errors.New("invalid diplomatic action")
	ErrAlreadyAtWar       = errors.New("factions are already at war")
	ErrNotAtWar           = errors.New("factions are not at war")
	ErrAlreadyAllied      = errors.New("factions are already allied")
	ErrNotAllied          = errors.New("factions are not allied")
	ErrCannotSelfRelation = errors.New("cannot create relation with self")
	ErrPendingProposal    = errors.New("there is already a pending proposal")
)

// clampValue clamps an integer to the range [-100, 100].
func clampValue(value int) int {
	if value > 100 {
		return 100
	}
	if value < -100 {
		return -100
	}
	return value
}

// relationKey creates a consistent key for two faction IDs.
func relationKey(faction1ID, faction2ID string) string {
	if faction1ID < faction2ID {
		return faction1ID + ":" + faction2ID
	}
	return faction2ID + ":" + faction1ID
}

// recordDiplomaticEvent appends a new diplomatic event to the relation's history.
func (rel *FactionRelation) recordDiplomaticEvent(action DiplomaticAction, initiatorID, targetID string, oldState, newState DiplomaticState, details string) {
	rel.History = append(rel.History, &DiplomaticEvent{
		ID:          uuid.New().String(),
		Timestamp:   time.Now(),
		Action:      action,
		InitiatorID: initiatorID,
		TargetID:    targetID,
		OldState:    oldState,
		NewState:    newState,
		Details:     details,
	})
}

// InitializeRelation creates a new diplomatic relation between factions.
func (dm *DiplomacyManager) InitializeRelation(faction1ID, faction2ID string) (*FactionRelation, error) {
	if faction1ID == faction2ID {
		return nil, ErrCannotSelfRelation
	}

	dm.mu.Lock()
	defer dm.mu.Unlock()

	key := relationKey(faction1ID, faction2ID)
	if rel, exists := dm.relations[key]; exists {
		return rel, nil
	}

	relation := &FactionRelation{
		ID:              uuid.New().String(),
		Faction1ID:      faction1ID,
		Faction2ID:      faction2ID,
		State:           DiplomaticStateNeutral,
		Opinion:         0,
		Trust:           0,
		TradeTreaty:     false,
		MilitaryAccess:  false,
		DefensivePact:   false,
		History:         make([]*DiplomaticEvent, 0),
		LastInteraction: time.Now(),
	}

	dm.relations[key] = relation

	dm.logger.WithFields(logrus.Fields{
		"faction1": faction1ID,
		"faction2": faction2ID,
	}).Info("diplomatic relation initialized")

	return relation, nil
}

// GetRelation retrieves the relation between two factions.
func (dm *DiplomacyManager) GetRelation(faction1ID, faction2ID string) (*FactionRelation, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	key := relationKey(faction1ID, faction2ID)
	rel, exists := dm.relations[key]
	if !exists {
		return nil, ErrRelationNotFound
	}
	return rel, nil
}

// DeclareWar initiates war between two factions.
func (dm *DiplomacyManager) DeclareWar(aggressorID, targetID, reason string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	key := relationKey(aggressorID, targetID)
	rel, exists := dm.relations[key]
	if !exists {
		return ErrRelationNotFound
	}

	rel.mu.Lock()
	defer rel.mu.Unlock()

	if rel.State == DiplomaticStateWar {
		return ErrAlreadyAtWar
	}

	oldState := rel.State
	rel.State = DiplomaticStateWar
	rel.Opinion = -100
	rel.Trust = -100
	rel.TradeTreaty = false
	rel.MilitaryAccess = false
	rel.DefensivePact = false
	rel.LastInteraction = time.Now()

	rel.recordDiplomaticEvent(ActionDeclareWar, aggressorID, targetID, oldState, DiplomaticStateWar, reason)

	dm.logger.WithFields(logrus.Fields{
		"aggressor": aggressorID,
		"target":    targetID,
		"reason":    reason,
	}).Info("war declared")

	return nil
}

// OfferPeace sends a peace offer from one faction to another.
func (dm *DiplomacyManager) OfferPeace(initiatorID, targetID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	key := relationKey(initiatorID, targetID)
	rel, exists := dm.relations[key]
	if !exists {
		return ErrRelationNotFound
	}

	rel.mu.Lock()
	defer rel.mu.Unlock()

	if rel.State != DiplomaticStateWar {
		return ErrNotAtWar
	}

	rel.LastInteraction = time.Now()

	rel.recordDiplomaticEvent(ActionOfferPeace, initiatorID, targetID, rel.State, rel.State, "peace proposal pending")

	dm.logger.WithFields(logrus.Fields{
		"initiator": initiatorID,
		"target":    targetID,
	}).Info("peace offer sent")

	return nil
}

// AcceptPeace ends a war and establishes hostile relations.
func (dm *DiplomacyManager) AcceptPeace(initiatorID, targetID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	key := relationKey(initiatorID, targetID)
	rel, exists := dm.relations[key]
	if !exists {
		return ErrRelationNotFound
	}

	rel.mu.Lock()
	defer rel.mu.Unlock()

	if rel.State != DiplomaticStateWar {
		return ErrNotAtWar
	}

	oldState := rel.State
	rel.State = DiplomaticStateHostile
	rel.Opinion = -50 // Still hostile after war
	rel.Trust = -50
	rel.LastInteraction = time.Now()

	rel.recordDiplomaticEvent(ActionAcceptPeace, initiatorID, targetID, oldState, DiplomaticStateHostile, "peace treaty signed")

	dm.logger.WithFields(logrus.Fields{
		"faction1": initiatorID,
		"faction2": targetID,
	}).Info("peace accepted")

	return nil
}

// ProposeAlliance sends an alliance proposal.
func (dm *DiplomacyManager) ProposeAlliance(initiatorID, targetID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	key := relationKey(initiatorID, targetID)
	rel, exists := dm.relations[key]
	if !exists {
		return ErrRelationNotFound
	}

	rel.mu.Lock()
	defer rel.mu.Unlock()

	if rel.State == DiplomaticStateWar || rel.State == DiplomaticStateHostile {
		return ErrInvalidAction
	}

	if rel.State == DiplomaticStateAllied {
		return ErrAlreadyAllied
	}

	rel.LastInteraction = time.Now()

	rel.recordDiplomaticEvent(ActionProposeAlliance, initiatorID, targetID, rel.State, rel.State, "alliance proposal pending")

	dm.logger.WithFields(logrus.Fields{
		"initiator": initiatorID,
		"target":    targetID,
	}).Info("alliance proposed")

	return nil
}

// AcceptAlliance forms an alliance between factions.
func (dm *DiplomacyManager) AcceptAlliance(initiatorID, targetID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	key := relationKey(initiatorID, targetID)
	rel, exists := dm.relations[key]
	if !exists {
		return ErrRelationNotFound
	}

	rel.mu.Lock()
	defer rel.mu.Unlock()

	if rel.State == DiplomaticStateAllied {
		return ErrAlreadyAllied
	}

	oldState := rel.State
	rel.State = DiplomaticStateAllied
	rel.Opinion = 75
	rel.Trust = 50
	rel.DefensivePact = true
	rel.MilitaryAccess = true
	rel.LastInteraction = time.Now()

	rel.recordDiplomaticEvent(ActionAcceptAlliance, initiatorID, targetID, oldState, DiplomaticStateAllied, "alliance formed")

	dm.logger.WithFields(logrus.Fields{
		"faction1": initiatorID,
		"faction2": targetID,
	}).Info("alliance formed")

	return nil
}

// BreakAlliance dissolves an existing alliance.
func (dm *DiplomacyManager) BreakAlliance(initiatorID, targetID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	key := relationKey(initiatorID, targetID)
	rel, exists := dm.relations[key]
	if !exists {
		return ErrRelationNotFound
	}

	rel.mu.Lock()
	defer rel.mu.Unlock()

	if rel.State != DiplomaticStateAllied {
		return ErrNotAllied
	}

	oldState := rel.State
	rel.State = DiplomaticStateTense
	rel.Opinion = -25
	rel.Trust = -50 // Trust damaged by breaking alliance
	rel.DefensivePact = false
	rel.MilitaryAccess = false
	rel.LastInteraction = time.Now()

	rel.recordDiplomaticEvent(ActionBreakAlliance, initiatorID, targetID, oldState, DiplomaticStateTense, "alliance broken")

	dm.logger.WithFields(logrus.Fields{
		"initiator": initiatorID,
		"target":    targetID,
	}).Info("alliance broken")

	return nil
}

// SignTradeAgreement establishes a trade treaty.
func (dm *DiplomacyManager) SignTradeAgreement(faction1ID, faction2ID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	key := relationKey(faction1ID, faction2ID)
	rel, exists := dm.relations[key]
	if !exists {
		return ErrRelationNotFound
	}

	rel.mu.Lock()
	defer rel.mu.Unlock()

	if rel.State == DiplomaticStateWar {
		return ErrInvalidAction
	}

	rel.TradeTreaty = true
	rel.Opinion = clampValue(rel.Opinion + 10)
	rel.LastInteraction = time.Now()

	rel.recordDiplomaticEvent(ActionTradeAgreement, faction1ID, faction2ID, rel.State, rel.State, "trade agreement signed")

	dm.logger.WithFields(logrus.Fields{
		"faction1": faction1ID,
		"faction2": faction2ID,
	}).Info("trade agreement signed")

	return nil
}

// SendGift improves relations by sending a gift.
func (dm *DiplomacyManager) SendGift(senderID, receiverID string, value int) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	key := relationKey(senderID, receiverID)
	rel, exists := dm.relations[key]
	if !exists {
		return ErrRelationNotFound
	}

	rel.mu.Lock()
	defer rel.mu.Unlock()

	// Gift improves opinion based on value
	opinionGain := value / 100
	if opinionGain < 1 {
		opinionGain = 1
	}
	if opinionGain > 25 {
		opinionGain = 25
	}

	rel.Opinion = clampValue(rel.Opinion + opinionGain)
	rel.Trust = clampValue(rel.Trust + opinionGain/2)

	rel.LastInteraction = time.Now()
	dm.updateState(rel)

	rel.recordDiplomaticEvent(ActionSendGift, senderID, receiverID, rel.State, rel.State, "diplomatic gift")

	return nil
}

// ModifyOpinion directly changes opinion (from events, quests, etc.).
func (dm *DiplomacyManager) ModifyOpinion(faction1ID, faction2ID string, change int) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	key := relationKey(faction1ID, faction2ID)
	rel, exists := dm.relations[key]
	if !exists {
		return ErrRelationNotFound
	}

	rel.mu.Lock()
	defer rel.mu.Unlock()

	rel.Opinion = clampValue(rel.Opinion + change)
	dm.updateState(rel)
	return nil
}

// ModifyTrust directly changes trust level.
func (dm *DiplomacyManager) ModifyTrust(faction1ID, faction2ID string, change int) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	key := relationKey(faction1ID, faction2ID)
	rel, exists := dm.relations[key]
	if !exists {
		return ErrRelationNotFound
	}

	rel.mu.Lock()
	defer rel.mu.Unlock()

	rel.Trust = clampValue(rel.Trust + change)
	return nil
}

// updateState adjusts diplomatic state based on opinion.
// Caller must hold rel.mu lock.
func (dm *DiplomacyManager) updateState(rel *FactionRelation) {
	// Don't auto-change war or alliance states
	if rel.State == DiplomaticStateWar || rel.State == DiplomaticStateAllied {
		return
	}

	oldState := rel.State

	switch {
	case rel.Opinion <= -75:
		rel.State = DiplomaticStateHostile
	case rel.Opinion <= -25:
		rel.State = DiplomaticStateTense
	case rel.Opinion <= 25:
		rel.State = DiplomaticStateNeutral
	case rel.Opinion <= 75:
		rel.State = DiplomaticStateFriendly
	default:
		rel.State = DiplomaticStateFriendly // Alliance requires explicit action
	}

	if rel.State != oldState {
		dm.logger.WithFields(logrus.Fields{
			"faction1":  rel.Faction1ID,
			"faction2":  rel.Faction2ID,
			"old_state": oldState,
			"new_state": rel.State,
		}).Debug("diplomatic state changed")
	}
}

// GetAllRelations returns all diplomatic relations.
func (dm *DiplomacyManager) GetAllRelations() []*FactionRelation {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	relations := make([]*FactionRelation, 0, len(dm.relations))
	for _, rel := range dm.relations {
		relations = append(relations, rel)
	}
	return relations
}

// GetFactionRelations returns all relations for a specific faction.
func (dm *DiplomacyManager) GetFactionRelations(factionID string) []*FactionRelation {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	relations := make([]*FactionRelation, 0)
	for _, rel := range dm.relations {
		if rel.Faction1ID == factionID || rel.Faction2ID == factionID {
			relations = append(relations, rel)
		}
	}
	return relations
}

// AreAllied checks if two factions are allied.
func (dm *DiplomacyManager) AreAllied(faction1ID, faction2ID string) bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	key := relationKey(faction1ID, faction2ID)
	rel, exists := dm.relations[key]
	if !exists {
		return false
	}

	rel.mu.RLock()
	defer rel.mu.RUnlock()
	return rel.State == DiplomaticStateAllied
}

// AreAtWar checks if two factions are at war.
func (dm *DiplomacyManager) AreAtWar(faction1ID, faction2ID string) bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	key := relationKey(faction1ID, faction2ID)
	rel, exists := dm.relations[key]
	if !exists {
		return false
	}

	rel.mu.RLock()
	defer rel.mu.RUnlock()
	return rel.State == DiplomaticStateWar
}

// GetOpinion returns the opinion between two factions.
func (dm *DiplomacyManager) GetOpinion(faction1ID, faction2ID string) (int, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	key := relationKey(faction1ID, faction2ID)
	rel, exists := dm.relations[key]
	if !exists {
		return 0, ErrRelationNotFound
	}

	rel.mu.RLock()
	defer rel.mu.RUnlock()
	return rel.Opinion, nil
}

// DecayRelations applies natural decay to relations over time.
func (dm *DiplomacyManager) DecayRelations(decayRate float64) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for _, rel := range dm.relations {
		rel.mu.Lock()

		// Trust decays toward 0
		if rel.Trust > 0 {
			rel.Trust -= int(float64(rel.Trust) * decayRate)
		} else if rel.Trust < 0 {
			rel.Trust += int(float64(-rel.Trust) * decayRate)
		}

		// Opinion decays toward 0 more slowly
		if rel.Opinion > 0 {
			rel.Opinion -= int(float64(rel.Opinion) * decayRate * 0.5)
		} else if rel.Opinion < 0 {
			rel.Opinion += int(float64(-rel.Opinion) * decayRate * 0.5)
		}

		dm.updateState(rel)
		rel.mu.Unlock()
	}
}
