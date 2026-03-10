package wasmui

import (
	"image/color"
	"testing"
)

func TestMessageTypeColor(t *testing.T) {
	tests := []struct {
		name     string
		msgType  MessageType
		expected color.RGBA
	}{
		{
			name:     "info message",
			msgType:  MessageInfo,
			expected: color.RGBA{R: 220, G: 220, B: 220, A: 255},
		},
		{
			name:     "warning message",
			msgType:  MessageWarning,
			expected: color.RGBA{R: 255, G: 200, B: 0, A: 255},
		},
		{
			name:     "error message",
			msgType:  MessageError,
			expected: color.RGBA{R: 255, G: 100, B: 100, A: 255},
		},
		{
			name:     "combat message",
			msgType:  MessageCombat,
			expected: color.RGBA{R: 200, G: 150, B: 255, A: 255},
		},
		{
			name:     "system message",
			msgType:  MessageSystem,
			expected: color.RGBA{R: 150, G: 200, B: 255, A: 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.msgType.Color()
			if got != tt.expected {
				t.Errorf("MessageType.Color() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestUIMode(t *testing.T) {
	// Verify UIMode constants have expected values
	if ModeNormal != 0 {
		t.Errorf("ModeNormal = %d, want 0", ModeNormal)
	}
	if ModeCombat != 1 {
		t.Errorf("ModeCombat = %d, want 1", ModeCombat)
	}
	if ModeInventory != 2 {
		t.Errorf("ModeInventory = %d, want 2", ModeInventory)
	}
	if ModeSpellcasting != 3 {
		t.Errorf("ModeSpellcasting = %d, want 3", ModeSpellcasting)
	}
}

func TestPlayerAttributes(t *testing.T) {
	attrs := PlayerAttributes{
		Strength:     18,
		Dexterity:    16,
		Constitution: 14,
		Intelligence: 12,
		Wisdom:       10,
		Charisma:     8,
	}

	if attrs.Strength != 18 {
		t.Errorf("Strength = %d, want 18", attrs.Strength)
	}
	if attrs.Dexterity != 16 {
		t.Errorf("Dexterity = %d, want 16", attrs.Dexterity)
	}
	if attrs.Constitution != 14 {
		t.Errorf("Constitution = %d, want 14", attrs.Constitution)
	}
	if attrs.Intelligence != 12 {
		t.Errorf("Intelligence = %d, want 12", attrs.Intelligence)
	}
	if attrs.Wisdom != 10 {
		t.Errorf("Wisdom = %d, want 10", attrs.Wisdom)
	}
	if attrs.Charisma != 8 {
		t.Errorf("Charisma = %d, want 8", attrs.Charisma)
	}
}

func TestPlayerState(t *testing.T) {
	player := PlayerState{
		ID:   "player-001",
		Name: "TestHero",
		Position: Position{
			X: 10,
			Y: 20,
		},
		HP:         50,
		MaxHP:      100,
		Level:      5,
		Experience: 5000,
		Class:      "Fighter",
		Attributes: PlayerAttributes{
			Strength:     16,
			Dexterity:    14,
			Constitution: 15,
			Intelligence: 10,
			Wisdom:       12,
			Charisma:     11,
		},
	}

	if player.ID != "player-001" {
		t.Errorf("ID = %q, want %q", player.ID, "player-001")
	}
	if player.Name != "TestHero" {
		t.Errorf("Name = %q, want %q", player.Name, "TestHero")
	}
	if player.Position.X != 10 || player.Position.Y != 20 {
		t.Errorf("Position = (%d, %d), want (10, 20)", player.Position.X, player.Position.Y)
	}
	if player.HP != 50 {
		t.Errorf("HP = %d, want 50", player.HP)
	}
	if player.MaxHP != 100 {
		t.Errorf("MaxHP = %d, want 100", player.MaxHP)
	}
}

func TestCombatState(t *testing.T) {
	combat := CombatState{
		Active:      true,
		CurrentTurn: "player-001",
		Initiative: []InitiativeEntry{
			{ID: "player-001", Name: "Hero", Initiative: 18, IsPlayer: true},
			{ID: "monster-001", Name: "Goblin", Initiative: 12, IsPlayer: false},
		},
		Round:    3,
		InCombat: true,
	}

	if !combat.Active {
		t.Error("Active should be true")
	}
	if !combat.InCombat {
		t.Error("InCombat should be true")
	}
	if combat.Round != 3 {
		t.Errorf("Round = %d, want 3", combat.Round)
	}
	if combat.CurrentTurn != "player-001" {
		t.Errorf("CurrentTurn = %q, want %q", combat.CurrentTurn, "player-001")
	}
	if len(combat.Initiative) != 2 {
		t.Errorf("Initiative length = %d, want 2", len(combat.Initiative))
	}
}

func TestLogMessage(t *testing.T) {
	msg := LogMessage{
		Text:      "You hit the goblin for 8 damage!",
		Type:      MessageCombat,
		Timestamp: 1234567890,
	}

	if msg.Text != "You hit the goblin for 8 damage!" {
		t.Errorf("Text = %q, want %q", msg.Text, "You hit the goblin for 8 damage!")
	}
	if msg.Type != MessageCombat {
		t.Errorf("Type = %v, want %v", msg.Type, MessageCombat)
	}
	if msg.Timestamp != 1234567890 {
		t.Errorf("Timestamp = %d, want 1234567890", msg.Timestamp)
	}
}

func TestNewRPCClient(t *testing.T) {
	// Test stub for native builds
	client := NewRPCClient()
	if client == nil {
		t.Error("NewRPCClient() returned nil")
	}
}

func TestNewGame(t *testing.T) {
	// Test stub for native builds - should return error
	game, err := NewGame()
	if err == nil {
		t.Error("NewGame() should return error on native builds")
	}
	if game != nil {
		t.Error("NewGame() should return nil game on native builds")
	}
}
