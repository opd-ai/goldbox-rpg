package server

import (
	"encoding/json"
	"fmt"
	"strings"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// handleEquipItem equips an item to a specified equipment slot for a player.
//
// Parameters (JSON):
//   - session_id: string - Player session identifier
//   - item_id: string - ID of the item to equip from inventory
//   - slot: string - Name of the equipment slot to equip to
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if equipping was successful
//   - message: string describing the result
//   - equipped_item: object containing details of the equipped item
//   - previous_item: object containing details of any previously equipped item (optional)
//
// Errors:
//   - "invalid session" if session is not found or inactive
//   - "invalid slot" if slot name is not recognized
//   - "item not found" if item ID is not in inventory
//   - "class restriction" if item cannot be equipped by character class
func (s *RPCServer) handleEquipItem(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleEquipItem",
	}).Debug("entering handleEquipItem")

	var req struct {
		SessionID string `json:"session_id"`
		ItemID    string `json:"item_id"`
		Slot      string `json:"slot"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleEquipItem",
			"error":    err.Error(),
		}).Error("failed to unmarshal equip item parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid equip item parameters", err.Error())
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return nil, err
	}

	player := session.Player

	// Parse slot name to EquipmentSlot
	slot, err := parseEquipmentSlot(req.Slot)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleEquipItem",
			"slot":     req.Slot,
		}).Error("invalid equipment slot")
		return nil, fmt.Errorf("invalid equipment slot: %s", req.Slot)
	}

	// Check if there's a previously equipped item
	var previousItem *game.Item
	if prevEquipped, exists := player.GetEquippedItem(slot); exists {
		previousItem = prevEquipped
	}

	// Equip the item
	if err := player.EquipItem(req.ItemID, slot); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleEquipItem",
			"itemID":   req.ItemID,
			"slot":     req.Slot,
			"error":    err.Error(),
		}).Error("failed to equip item")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid method parameters", "item not found: "+req.ItemID)
	}

	// Get the newly equipped item
	equippedItem, _ := player.GetEquippedItem(slot)

	logrus.WithFields(logrus.Fields{
		"function":     "handleEquipItem",
		"sessionID":    req.SessionID,
		"itemID":       req.ItemID,
		"slot":         req.Slot,
		"equippedItem": equippedItem.Name,
	}).Info("item equipped successfully")

	response := map[string]interface{}{
		"success":       true,
		"message":       fmt.Sprintf("Successfully equipped %s", equippedItem.Name),
		"equipped_item": equippedItem,
	}

	if previousItem != nil {
		response["previous_item"] = previousItem
	}

	return response, nil
}

// handleUnequipItem removes an equipped item and returns it to the player's inventory.
//
// Parameters (JSON):
//   - session_id: string - Player session identifier
//   - slot: string - Name of the equipment slot to unequip
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if unequipping was successful
//   - message: string describing the result
//   - unequipped_item: object containing details of the unequipped item
//
// Errors:
//   - "invalid session" if session is not found or inactive
//   - "invalid slot" if slot name is not recognized
//   - "no item equipped" if the specified slot is empty
func (s *RPCServer) handleUnequipItem(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleUnequipItem",
	}).Debug("entering handleUnequipItem")

	var req struct {
		SessionID string `json:"session_id"`
		Slot      string `json:"slot"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleUnequipItem",
			"error":    err.Error(),
		}).Error("failed to unmarshal unequip item parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid unequip item parameters", err.Error())
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return nil, err
	}

	player := session.Player

	// Parse slot name to EquipmentSlot
	slot, err := parseEquipmentSlot(req.Slot)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleUnequipItem",
			"slot":     req.Slot,
		}).Error("invalid equipment slot")
		return nil, fmt.Errorf("invalid equipment slot: %s", req.Slot)
	}

	// Unequip the item
	unequippedItem, err := player.UnequipItem(slot)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleUnequipItem",
			"slot":     req.Slot,
			"error":    err.Error(),
		}).Error("failed to unequip item")
		return map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}, nil
	}

	logrus.WithFields(logrus.Fields{
		"function":       "handleUnequipItem",
		"sessionID":      req.SessionID,
		"slot":           req.Slot,
		"unequippedItem": unequippedItem.Name,
	}).Info("item unequipped successfully")

	return map[string]interface{}{
		"success":         true,
		"message":         fmt.Sprintf("Successfully unequipped %s", unequippedItem.Name),
		"unequipped_item": unequippedItem,
	}, nil
}

// handleGetEquipment returns all currently equipped items for a player.
//
// Parameters (JSON):
//   - session_id: string - Player session identifier
//
// Returns:
//   - interface{}: Map containing:
//   - success: bool indicating if retrieval was successful
//   - equipment: map of slot names to equipped item objects
//   - total_weight: int total weight of all equipped items
//   - equipment_bonuses: map of stat bonuses from equipment
//
// Errors:
//   - "invalid session" if session is not found or inactive
func (s *RPCServer) handleGetEquipment(params json.RawMessage) (interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"function": "handleGetEquipment",
	}).Debug("entering handleGetEquipment")

	var req struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"function": "handleGetEquipment",
			"error":    err.Error(),
		}).Error("failed to unmarshal get equipment parameters")
		return nil, NewJSONRPCError(JSONRPCInvalidParams, "Invalid get equipment parameters", err.Error())
	}

	// Get player session
	session, err := s.getPlayerSession(req.SessionID)
	if err != nil {
		return nil, err
	}

	player := session.Player

	// Get all equipped items
	equippedItems := player.GetAllEquippedItems()

	// Convert equipment slots to string keys for JSON response
	equipment := make(map[string]game.Item)
	totalWeight := 0
	for slot, item := range equippedItems {
		slotName := equipmentSlotToString(slot)
		equipment[slotName] = item
		totalWeight += item.Weight
	}

	// Get equipment bonuses
	bonuses := player.CalculateEquipmentBonuses()

	logrus.WithFields(logrus.Fields{
		"function":    "handleGetEquipment",
		"sessionID":   req.SessionID,
		"numItems":    len(equipment),
		"totalWeight": totalWeight,
	}).Info("equipment retrieved successfully")

	return map[string]interface{}{
		"success":           true,
		"equipment":         equipment,
		"total_weight":      totalWeight,
		"equipment_bonuses": bonuses,
	}, nil
}

// parseEquipmentSlot converts a string slot name to an EquipmentSlot enum value
func parseEquipmentSlot(slotName string) (game.EquipmentSlot, error) {
	slotMap := map[string]game.EquipmentSlot{
		"head":        game.SlotHead,
		"neck":        game.SlotNeck,
		"chest":       game.SlotChest,
		"hands":       game.SlotHands,
		"rings":       game.SlotRings,
		"legs":        game.SlotLegs,
		"feet":        game.SlotFeet,
		"weapon_main": game.SlotWeaponMain,
		"weapon_off":  game.SlotWeaponOff,
		"main_hand":   game.SlotWeaponMain, // Alternative naming
		"off_hand":    game.SlotWeaponOff,  // Alternative naming
	}

	if slot, exists := slotMap[strings.ToLower(slotName)]; exists {
		return slot, nil
	}

	return game.SlotHead, fmt.Errorf("unknown equipment slot: %s", slotName)
}

// equipmentSlotToString converts an EquipmentSlot enum value to a string
func equipmentSlotToString(slot game.EquipmentSlot) string {
	slotNames := map[game.EquipmentSlot]string{
		game.SlotHead:       "head",
		game.SlotNeck:       "neck",
		game.SlotChest:      "chest",
		game.SlotHands:      "hands",
		game.SlotRings:      "rings",
		game.SlotLegs:       "legs",
		game.SlotFeet:       "feet",
		game.SlotWeaponMain: "weapon_main",
		game.SlotWeaponOff:  "weapon_off",
	}

	if name, exists := slotNames[slot]; exists {
		return name
	}

	return "unknown"
}
