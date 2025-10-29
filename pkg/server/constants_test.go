package server

import (
	"testing"
	"time"

	"goldbox-rpg/pkg/game"
)

// TestContextKey tests the contextKey type and constants
func TestContextKey_Type(t *testing.T) {
	// Test that sessionKey is of type contextKey
	var key contextKey = sessionKey
	expected := "session"
	if string(key) != expected {
		t.Errorf("sessionKey = %q, want %q", string(key), expected)
	}
}

// TestSessionConfiguration tests session and server configuration constants
func TestSessionConfiguration_Values(t *testing.T) {
	tests := []struct {
		name     string
		value    time.Duration
		expected time.Duration
	}{
		{
			name:     "sessionCleanupInterval",
			value:    sessionCleanupInterval,
			expected: 5 * time.Minute,
		},
		{
			name:     "sessionTimeout",
			value:    sessionTimeout,
			expected: 30 * time.Minute,
		},
		{
			name:     "MessageSendTimeout",
			value:    MessageSendTimeout,
			expected: 50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.value, tt.expected)
			}
		})
	}
}

// TestMessageConfiguration tests message channel configuration constants
func TestMessageConfiguration_BufferSize(t *testing.T) {
	expected := 500
	if MessageChanBufferSize != expected {
		t.Errorf("MessageChanBufferSize = %d, want %d", MessageChanBufferSize, expected)
	}
}

// TestRPCMethod_BasicMethods tests basic RPC method constants
func TestRPCMethod_BasicMethods(t *testing.T) {
	tests := []struct {
		name     string
		method   RPCMethod
		expected string
	}{
		{"MethodMove", MethodMove, "move"},
		{"MethodAttack", MethodAttack, "attack"},
		{"MethodCastSpell", MethodCastSpell, "castSpell"},
		{"MethodUseItem", MethodUseItem, "useItem"},
		{"MethodApplyEffect", MethodApplyEffect, "applyEffect"},
		{"MethodStartCombat", MethodStartCombat, "startCombat"},
		{"MethodEndTurn", MethodEndTurn, "endTurn"},
		{"MethodGetGameState", MethodGetGameState, "getGameState"},
		{"MethodJoinGame", MethodJoinGame, "joinGame"},
		{"MethodLeaveGame", MethodLeaveGame, "leaveGame"},
		{"MethodCreateCharacter", MethodCreateCharacter, "createCharacter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.method) != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, string(tt.method), tt.expected)
			}
		})
	}
}

// TestRPCMethod_EquipmentMethods tests equipment management RPC method constants
func TestRPCMethod_EquipmentMethods(t *testing.T) {
	tests := []struct {
		name     string
		method   RPCMethod
		expected string
	}{
		{"MethodEquipItem", MethodEquipItem, "equipItem"},
		{"MethodUnequipItem", MethodUnequipItem, "unequipItem"},
		{"MethodGetEquipment", MethodGetEquipment, "getEquipment"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.method) != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, string(tt.method), tt.expected)
			}
		})
	}
}

// TestRPCMethod_QuestMethods tests quest management RPC method constants
func TestRPCMethod_QuestMethods(t *testing.T) {
	tests := []struct {
		name     string
		method   RPCMethod
		expected string
	}{
		{"MethodStartQuest", MethodStartQuest, "startQuest"},
		{"MethodCompleteQuest", MethodCompleteQuest, "completeQuest"},
		{"MethodUpdateObjective", MethodUpdateObjective, "updateObjective"},
		{"MethodFailQuest", MethodFailQuest, "failQuest"},
		{"MethodGetQuest", MethodGetQuest, "getQuest"},
		{"MethodGetActiveQuests", MethodGetActiveQuests, "getActiveQuests"},
		{"MethodGetCompletedQuests", MethodGetCompletedQuests, "getCompletedQuests"},
		{"MethodGetQuestLog", MethodGetQuestLog, "getQuestLog"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.method) != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, string(tt.method), tt.expected)
			}
		})
	}
}

// TestRPCMethod_SpellMethods tests spell management RPC method constants
func TestRPCMethod_SpellMethods(t *testing.T) {
	tests := []struct {
		name     string
		method   RPCMethod
		expected string
	}{
		{"MethodGetSpell", MethodGetSpell, "getSpell"},
		{"MethodGetSpellsByLevel", MethodGetSpellsByLevel, "getSpellsByLevel"},
		{"MethodGetSpellsBySchool", MethodGetSpellsBySchool, "getSpellsBySchool"},
		{"MethodGetAllSpells", MethodGetAllSpells, "getAllSpells"},
		{"MethodSearchSpells", MethodSearchSpells, "searchSpells"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.method) != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, string(tt.method), tt.expected)
			}
		})
	}
}

// TestRPCMethod_SpatialMethods tests spatial query RPC method constants
func TestRPCMethod_SpatialMethods(t *testing.T) {
	tests := []struct {
		name     string
		method   RPCMethod
		expected string
	}{
		{"MethodGetObjectsInRange", MethodGetObjectsInRange, "getObjectsInRange"},
		{"MethodGetObjectsInRadius", MethodGetObjectsInRadius, "getObjectsInRadius"},
		{"MethodGetNearestObjects", MethodGetNearestObjects, "getNearestObjects"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.method) != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, string(tt.method), tt.expected)
			}
		})
	}
}

// TestRPCMethod_StringConversion tests RPCMethod type conversion behavior
func TestRPCMethod_StringConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected RPCMethod
	}{
		{"string to RPCMethod", "testMethod", RPCMethod("testMethod")},
		{"empty string", "", RPCMethod("")},
		{"special characters", "test-method_123", RPCMethod("test-method_123")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := RPCMethod(tt.input)
			if method != tt.expected {
				t.Errorf("RPCMethod(%q) = %q, want %q", tt.input, method, tt.expected)
			}
			// Test reverse conversion
			if string(method) != tt.input {
				t.Errorf("string(RPCMethod(%q)) = %q, want %q", tt.input, string(method), tt.input)
			}
		})
	}
}

// TestEventType_CombatEvents tests combat event type constants
func TestEventType_CombatEvents(t *testing.T) {
	tests := []struct {
		name      string
		eventType game.EventType
		expected  game.EventType
	}{
		{"EventCombatStart", EventCombatStart, game.EventType(100)},
		{"EventCombatEnd", EventCombatEnd, game.EventType(101)},
		{"EventTurnStart", EventTurnStart, game.EventType(102)},
		{"EventTurnEnd", EventTurnEnd, game.EventType(103)},
		{"EventMovement", EventMovement, game.EventType(104)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.eventType != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.eventType, tt.expected)
			}
		})
	}
}

// TestEventType_Sequence tests that combat events follow the expected iota sequence
func TestEventType_Sequence(t *testing.T) {
	// Test that events follow proper iota sequence starting from 100
	if EventCombatStart != 100 {
		t.Errorf("EventCombatStart = %v, want 100", EventCombatStart)
	}
	if EventCombatEnd != EventCombatStart+1 {
		t.Errorf("EventCombatEnd = %v, want %v", EventCombatEnd, EventCombatStart+1)
	}
	if EventTurnStart != EventCombatEnd+1 {
		t.Errorf("EventTurnStart = %v, want %v", EventTurnStart, EventCombatEnd+1)
	}
	if EventTurnEnd != EventTurnStart+1 {
		t.Errorf("EventTurnEnd = %v, want %v", EventTurnEnd, EventTurnStart+1)
	}
	if EventMovement != EventTurnEnd+1 {
		t.Errorf("EventMovement = %v, want %v", EventMovement, EventTurnEnd+1)
	}
}

// TestConstants_TypeCompatibility tests compatibility with game package types
func TestConstants_TypeCompatibility(t *testing.T) {
	// Test that server event types are compatible with game.EventType
	var gameEvent game.EventType
	gameEvent = EventCombatStart
	if gameEvent != 100 {
		t.Errorf("EventCombatStart assignment to game.EventType failed: got %v, want 100", gameEvent)
	}
}

// TestRPCMethod_AllMethodsCovered tests that all defined methods are tested
func TestRPCMethod_AllMethodsCovered(t *testing.T) {
	// This test ensures we haven't missed any method constants
	allMethods := []RPCMethod{
		// Basic methods
		MethodMove, MethodAttack, MethodCastSpell, MethodUseItem, MethodApplyEffect,
		MethodStartCombat, MethodEndTurn, MethodGetGameState, MethodJoinGame,
		MethodLeaveGame, MethodCreateCharacter,
		// Equipment methods
		MethodEquipItem, MethodUnequipItem, MethodGetEquipment,
		// Quest methods
		MethodStartQuest, MethodCompleteQuest, MethodUpdateObjective, MethodFailQuest,
		MethodGetQuest, MethodGetActiveQuests, MethodGetCompletedQuests, MethodGetQuestLog,
		// Spell methods
		MethodGetSpell, MethodGetSpellsByLevel, MethodGetSpellsBySchool,
		MethodGetAllSpells, MethodSearchSpells,
		// Spatial methods
		MethodGetObjectsInRange, MethodGetObjectsInRadius, MethodGetNearestObjects,
	}

	// Test that all methods can be converted to strings
	for i, method := range allMethods {
		methodStr := string(method)
		if methodStr == "" {
			t.Errorf("Method at index %d has empty string value", i)
		}
		// Test that string conversion is consistent
		converted := RPCMethod(methodStr)
		if converted != method {
			t.Errorf("Round-trip conversion failed for method %v", method)
		}
	}
}

// TestSessionConfiguration_TimeUnits tests that time constants use correct units
func TestSessionConfiguration_TimeUnits(t *testing.T) {
	// Test that cleanup interval is reasonable (not too short or too long)
	if sessionCleanupInterval < time.Minute {
		t.Errorf("sessionCleanupInterval too short: %v, should be at least 1 minute", sessionCleanupInterval)
	}
	if sessionCleanupInterval > time.Hour {
		t.Errorf("sessionCleanupInterval too long: %v, should be less than 1 hour", sessionCleanupInterval)
	}

	// Test that session timeout is longer than cleanup interval
	if sessionTimeout <= sessionCleanupInterval {
		t.Errorf("sessionTimeout (%v) should be longer than sessionCleanupInterval (%v)", sessionTimeout, sessionCleanupInterval)
	}

	// Test that message timeout is reasonable for network operations
	if MessageSendTimeout < 10*time.Millisecond {
		t.Errorf("MessageSendTimeout too short: %v", MessageSendTimeout)
	}
	if MessageSendTimeout > 5*time.Second {
		t.Errorf("MessageSendTimeout too long: %v", MessageSendTimeout)
	}
}

// TestMessageChanBufferSize_Range tests buffer size is within reasonable limits
func TestMessageChanBufferSize_Range(t *testing.T) {
	// Buffer should be large enough to prevent blocking but not excessive
	if MessageChanBufferSize < 10 {
		t.Errorf("MessageChanBufferSize too small: %d", MessageChanBufferSize)
	}
	if MessageChanBufferSize > 10000 {
		t.Errorf("MessageChanBufferSize too large: %d", MessageChanBufferSize)
	}
}

// TestRPCMethod_StringValues_NoDuplicates tests that no method strings are duplicated
func TestRPCMethod_StringValues_NoDuplicates(t *testing.T) {
	allMethods := map[string]RPCMethod{
		string(MethodMove):               MethodMove,
		string(MethodAttack):             MethodAttack,
		string(MethodCastSpell):          MethodCastSpell,
		string(MethodUseItem):            MethodUseItem,
		string(MethodApplyEffect):        MethodApplyEffect,
		string(MethodStartCombat):        MethodStartCombat,
		string(MethodEndTurn):            MethodEndTurn,
		string(MethodGetGameState):       MethodGetGameState,
		string(MethodJoinGame):           MethodJoinGame,
		string(MethodLeaveGame):          MethodLeaveGame,
		string(MethodCreateCharacter):    MethodCreateCharacter,
		string(MethodEquipItem):          MethodEquipItem,
		string(MethodUnequipItem):        MethodUnequipItem,
		string(MethodGetEquipment):       MethodGetEquipment,
		string(MethodStartQuest):         MethodStartQuest,
		string(MethodCompleteQuest):      MethodCompleteQuest,
		string(MethodUpdateObjective):    MethodUpdateObjective,
		string(MethodFailQuest):          MethodFailQuest,
		string(MethodGetQuest):           MethodGetQuest,
		string(MethodGetActiveQuests):    MethodGetActiveQuests,
		string(MethodGetCompletedQuests): MethodGetCompletedQuests,
		string(MethodGetQuestLog):        MethodGetQuestLog,
		string(MethodGetSpell):           MethodGetSpell,
		string(MethodGetSpellsByLevel):   MethodGetSpellsByLevel,
		string(MethodGetSpellsBySchool):  MethodGetSpellsBySchool,
		string(MethodGetAllSpells):       MethodGetAllSpells,
		string(MethodSearchSpells):       MethodSearchSpells,
		string(MethodGetObjectsInRange):  MethodGetObjectsInRange,
		string(MethodGetObjectsInRadius): MethodGetObjectsInRadius,
		string(MethodGetNearestObjects):  MethodGetNearestObjects,
	}

	// Count expected methods (should match the constants defined)
	expectedCount := 30 // Update this if new methods are added
	if len(allMethods) != expectedCount {
		t.Errorf("Expected %d unique method strings, got %d", expectedCount, len(allMethods))
	}

	// Verify each method maps back to itself
	for methodStr, expectedMethod := range allMethods {
		if string(expectedMethod) != methodStr {
			t.Errorf("Method string mismatch: %s != %s", string(expectedMethod), methodStr)
		}
	}
}
