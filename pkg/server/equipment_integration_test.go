package server

import (
	"encoding/json"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"
)

// TestEquipmentManagementIntegration tests the complete equipment management workflow via RPC
func TestEquipmentManagementIntegration(t *testing.T) {
	// Create test server
	server, err := NewRPCServer("../web")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create a test character with starting inventory
	character := &game.Character{
		ID:        "test-char-1",
		Name:      "Test Character",
		Class:     game.ClassFighter,
		Strength:  15,
		Equipment: make(map[game.EquipmentSlot]game.Item),
		Inventory: []game.Item{
			{
				ID:         "sword001",
				Name:       "Iron Sword",
				Type:       "weapon",
				Weight:     5,
				Properties: []string{"sharp"},
			},
			{
				ID:         "helmet001",
				Name:       "Iron Helmet",
				Type:       "helmet",
				Weight:     3,
				Properties: []string{"protective"},
			},
		},
	}

	// Create player from character
	player := &game.Player{
		Character: *character.Clone(),
		Level:     1,
	}

	// Create test session
	session := &PlayerSession{
		SessionID:   "test-session-123",
		Player:      player,
		LastActive:  time.Now(),
		CreatedAt:   time.Now(),
		Connected:   false,
		MessageChan: make(chan []byte, 100),
	}

	// Add session to server
	server.mu.Lock()
	server.sessions["test-session-123"] = session
	server.mu.Unlock()

	t.Run("equip weapon successfully", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": "test-session-123",
			"item_id":    "sword001",
			"slot":       "weapon_main",
		}

		paramBytes, _ := json.Marshal(params)
		result, err := server.handleEquipItem(paramBytes)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		response := result.(map[string]interface{})
		if !response["success"].(bool) {
			t.Errorf("Expected success=true, got: %v", response["success"])
		}

		// Verify item was equipped
		if equipped, exists := player.GetEquippedItem(game.SlotWeaponMain); !exists {
			t.Errorf("Item was not equipped")
		} else if equipped.ID != "sword001" {
			t.Errorf("Wrong item equipped: got %s, want sword001", equipped.ID)
		}

		// Verify item was removed from inventory
		if player.HasItem("sword001") {
			t.Errorf("Item was not removed from inventory")
		}
	})

	t.Run("get equipment information", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": "test-session-123",
		}

		paramBytes, _ := json.Marshal(params)
		result, err := server.handleGetEquipment(paramBytes)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		response := result.(map[string]interface{})
		if !response["success"].(bool) {
			t.Errorf("Expected success=true, got: %v", response["success"])
		}

		equipment := response["equipment"].(map[string]game.Item)
		if len(equipment) != 1 {
			t.Errorf("Expected 1 equipped item, got %d", len(equipment))
		}

		if weapon, exists := equipment["weapon_main"]; !exists {
			t.Errorf("Expected weapon_main to be equipped")
		} else if weapon.ID != "sword001" {
			t.Errorf("Wrong weapon equipped: got %s, want sword001", weapon.ID)
		}

		totalWeight := response["total_weight"].(int)
		if totalWeight != 5 { // Weight of the sword
			t.Errorf("Expected total weight 5, got %d", totalWeight)
		}
	})

	t.Run("unequip weapon successfully", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": "test-session-123",
			"slot":       "weapon_main",
		}

		paramBytes, _ := json.Marshal(params)
		result, err := server.handleUnequipItem(paramBytes)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		response := result.(map[string]interface{})
		if !response["success"].(bool) {
			t.Errorf("Expected success=true, got: %v", response["success"])
		}

		// Verify item was unequipped
		if _, exists := player.GetEquippedItem(game.SlotWeaponMain); exists {
			t.Errorf("Item was not unequipped")
		}

		// Verify item was returned to inventory
		if !player.HasItem("sword001") {
			t.Errorf("Item was not returned to inventory")
		}
	})

	t.Run("equip invalid item type", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": "test-session-123",
			"item_id":    "helmet001",
			"slot":       "weapon_main", // Wrong slot for helmet
		}

		paramBytes, _ := json.Marshal(params)
		_, err := server.handleEquipItem(paramBytes)
		if err == nil {
			t.Errorf("Expected error for invalid item type, got nil")
		}
	})

	t.Run("equip non-existent item", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": "test-session-123",
			"item_id":    "nonexistent",
			"slot":       "weapon_main",
		}

		paramBytes, _ := json.Marshal(params)
		_, err := server.handleEquipItem(paramBytes)
		if err == nil {
			t.Errorf("Expected error for non-existent item, got nil")
		}
	})

	t.Run("invalid session", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": "invalid-session",
			"item_id":    "sword001",
			"slot":       "weapon_main",
		}

		paramBytes, _ := json.Marshal(params)
		_, err := server.handleEquipItem(paramBytes)

		if err == nil {
			t.Errorf("Expected error for invalid session")
		}
	})
}

// TestSlotParsing tests the equipment slot parsing functions
func TestSlotParsing(t *testing.T) {
	tests := []struct {
		name     string
		slotName string
		expected game.EquipmentSlot
		hasError bool
	}{
		{"head slot", "head", game.SlotHead, false},
		{"weapon main", "weapon_main", game.SlotWeaponMain, false},
		{"main hand alternative", "main_hand", game.SlotWeaponMain, false},
		{"off hand", "off_hand", game.SlotWeaponOff, false},
		{"invalid slot", "invalid", game.SlotHead, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slot, err := parseEquipmentSlot(tt.slotName)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for slot %s", tt.slotName)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if slot != tt.expected {
					t.Errorf("parseEquipmentSlot(%s) = %v, want %v", tt.slotName, slot, tt.expected)
				}
			}
		})
	}
}

// TestSlotStringConversion tests the slot-to-string conversion
func TestSlotStringConversion(t *testing.T) {
	tests := []struct {
		slot     game.EquipmentSlot
		expected string
	}{
		{game.SlotHead, "head"},
		{game.SlotWeaponMain, "weapon_main"},
		{game.SlotWeaponOff, "weapon_off"},
		{game.SlotChest, "chest"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := equipmentSlotToString(tt.slot)
			if result != tt.expected {
				t.Errorf("equipmentSlotToString(%v) = %s, want %s", tt.slot, result, tt.expected)
			}
		})
	}
}

// TestGetCharacterResistances tests the character resistances RPC method
func TestGetCharacterResistances(t *testing.T) {
	// Create test server
	server, err := NewRPCServer("../web")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create a test character with equipment that provides resistances
	character := &game.Character{
		ID:        "resist-test-char",
		Name:      "Resistance Test Character",
		Class:     game.ClassFighter,
		Equipment: make(map[game.EquipmentSlot]game.Item),
		Inventory: []game.Item{
			{
				ID:         "fire-cloak",
				Name:       "Cloak of Fire Resistance",
				Type:       "armor",
				Weight:     2,
				Properties: []string{"fire_resistance+0.25"},
			},
		},
	}
	// Initialize effect manager via character initialization
	baseStats := game.NewDefaultStats()
	character.EffectManager = game.NewEffectManager(baseStats)

	// Create player from character
	player := &game.Player{
		Character: *character.Clone(),
		Level:     1,
	}
	// Make sure player also has effect manager
	player.Character.EffectManager = game.NewEffectManager(baseStats)

	// Create test session
	session := &PlayerSession{
		SessionID:   "resist-test-session",
		Player:      player,
		LastActive:  time.Now(),
		CreatedAt:   time.Now(),
		Connected:   false,
		MessageChan: make(chan []byte, 100),
	}
	server.mu.Lock()
	server.sessions["resist-test-session"] = session
	server.mu.Unlock()

	t.Run("get resistances without equipment", func(t *testing.T) {
		params := map[string]interface{}{
			"session_id": "resist-test-session",
		}

		paramBytes, _ := json.Marshal(params)
		result, err := server.handleGetCharacterResistances(paramBytes)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
			return
		}

		response := result.(map[string]interface{})
		if !response["success"].(bool) {
			t.Errorf("Expected success=true")
			return
		}

		resistances := response["resistances"].(map[string]int)
		// Without equipment, all resistances should be 0
		for resistType, value := range resistances {
			if value != 0 {
				t.Errorf("Expected %s resistance = 0 without equipment, got %d", resistType, value)
			}
		}
	})

	t.Run("get resistances with fire resistance equipment", func(t *testing.T) {
		// Equip the fire resistance cloak
		fireCloak := game.Item{
			ID:         "fire-cloak",
			Name:       "Cloak of Fire Resistance",
			Type:       "armor",
			Weight:     2,
			Properties: []string{"fire_resistance+0.25"},
		}
		session.Player.Equipment[game.SlotChest] = fireCloak
		session.Player.ApplyEquipmentResistances()

		params := map[string]interface{}{
			"session_id": "resist-test-session",
		}

		paramBytes, _ := json.Marshal(params)
		result, err := server.handleGetCharacterResistances(paramBytes)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
			return
		}

		response := result.(map[string]interface{})
		resistances := response["resistances"].(map[string]int)

		// Fire resistance should be 25% (0.25 * 100)
		if fireResist, exists := resistances["fire"]; !exists || fireResist != 25 {
			t.Errorf("Expected fire resistance = 25, got %d", fireResist)
		}
	})
}
