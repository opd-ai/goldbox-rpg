package server

import (
	"encoding/json"
	"testing"
	"time"

	"goldbox-rpg/pkg/game"

	"github.com/sirupsen/logrus"
)

// TestPCGHandlers tests the newly implemented PCG handler methods
func TestPCGHandlers(t *testing.T) {
	// Create a test server
	server, err := NewRPCServer(":8080")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Stop()

	// Create a test session
	sessionID := "test_session_pcg"
	testCharacter := game.Character{
		ID:       "test_player_pcg",
		Name:     "PCG Test Player",
		Class:    game.ClassFighter,
		Position: game.Position{X: 5, Y: 5},
	}
	testPlayer := &game.Player{
		Character: *testCharacter.Clone(),
		Level:     1,
	}

	testSession := &PlayerSession{
		SessionID:   sessionID,
		Player:      testPlayer,
		LastActive:  time.Now(),
		CreatedAt:   time.Now(),
		Connected:   true,
		MessageChan: make(chan []byte, MessageChanBufferSize),
	}

	server.mu.Lock()
	server.sessions[sessionID] = testSession
	server.mu.Unlock()

	t.Run("TestHandleGetPCGStats", func(t *testing.T) {
		// Test getting PCG statistics
		params := map[string]interface{}{
			"session_id": sessionID,
		}

		paramsJSON, err := json.Marshal(params)
		if err != nil {
			t.Fatalf("Failed to marshal params: %v", err)
		}

		result, err := server.handleGetPCGStats(paramsJSON)
		if err != nil {
			t.Fatalf("handleGetPCGStats failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected result to be a map, got %T", result)
		}

		if !resultMap["success"].(bool) {
			t.Errorf("Expected success to be true")
		}

		if resultMap["stats"] == nil {
			t.Errorf("Expected stats to be present")
		}

		logrus.Info("PCG stats test passed successfully")
	})

	t.Run("TestHandleGenerateContent", func(t *testing.T) {
		// Test generating quest content
		params := map[string]interface{}{
			"session_id":   sessionID,
			"content_type": "quests",
			"location_id":  "test_location",
			"difficulty":   5,
			"constraints":  map[string]interface{}{},
		}

		paramsJSON, err := json.Marshal(params)
		if err != nil {
			t.Fatalf("Failed to marshal params: %v", err)
		}

		result, err := server.handleGenerateContent(paramsJSON)
		if err != nil {
			t.Fatalf("handleGenerateContent failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected result to be a map, got %T", result)
		}

		if !resultMap["success"].(bool) {
			t.Errorf("Expected success to be true")
		}

		if resultMap["content"] == nil {
			t.Errorf("Expected content to be present")
		}

		if resultMap["content_type"].(string) != "quests" {
			t.Errorf("Expected content_type to be 'quests', got %s", resultMap["content_type"])
		}

		logrus.Info("Content generation test passed successfully")
	})

	t.Run("TestHandleGenerateItems", func(t *testing.T) {
		// Test generating items
		params := map[string]interface{}{
			"session_id":   sessionID,
			"location_id":  "test_location",
			"count":        3,
			"min_rarity":   "common",
			"max_rarity":   "rare",
			"player_level": 5,
		}

		paramsJSON, err := json.Marshal(params)
		if err != nil {
			t.Fatalf("Failed to marshal params: %v", err)
		}

		result, err := server.handleGenerateItems(paramsJSON)
		if err != nil {
			t.Fatalf("handleGenerateItems failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected result to be a map, got %T", result)
		}

		if !resultMap["success"].(bool) {
			t.Errorf("Expected success to be true")
		}

		if resultMap["items"] == nil {
			t.Errorf("Expected items to be present")
		}

		count, ok := resultMap["count"].(int)
		if !ok || count <= 0 {
			t.Errorf("Expected count to be a positive integer, got %v", resultMap["count"])
		}

		logrus.Info("Item generation test passed successfully")
	})

	t.Run("TestHandleValidateContent", func(t *testing.T) {
		// Test content validation with a sample quest
		testQuest := &game.Quest{
			ID:          "test_quest",
			Title:       "Test Quest",
			Description: "A simple test quest",
			Status:      game.QuestNotStarted,
			Objectives:  []game.QuestObjective{},
			Rewards:     []game.QuestReward{},
		}

		params := map[string]interface{}{
			"session_id":   sessionID,
			"content_type": "quests",
			"content":      testQuest,
			"strict":       false,
		}

		paramsJSON, err := json.Marshal(params)
		if err != nil {
			t.Fatalf("Failed to marshal params: %v", err)
		}

		result, err := server.handleValidateContent(paramsJSON)
		if err != nil {
			t.Fatalf("handleValidateContent failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected result to be a map, got %T", result)
		}

		if !resultMap["success"].(bool) {
			t.Errorf("Expected success to be true")
		}

		// Note: validation may or may not pass depending on the specific content,
		// but the handler should not error out
		logrus.Info("Content validation test passed successfully")
	})
}

// TestPCGMethodConstants verifies that all PCG method constants are properly defined
func TestPCGMethodConstants(t *testing.T) {
	expectedMethods := []RPCMethod{
		MethodGenerateContent,
		MethodRegenerateTerrain,
		MethodGenerateItems,
		MethodGenerateLevel,
		MethodGenerateQuest,
		MethodGetPCGStats,
		MethodValidateContent,
	}

	for _, method := range expectedMethods {
		if string(method) == "" {
			t.Errorf("PCG method constant is empty: %v", method)
		}
	}

	logrus.Info("PCG method constants test passed successfully")
}
