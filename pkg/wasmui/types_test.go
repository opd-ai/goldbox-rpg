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
	// Verify all 6 UIMode constants have expected values
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
	if ModeAdventureSelect != 4 {
		t.Errorf("ModeAdventureSelect = %d, want 4", ModeAdventureSelect)
	}
	if ModeCharacterCreation != 5 {
		t.Errorf("ModeCharacterCreation = %d, want 5", ModeCharacterCreation)
	}
}

func TestScreenState(t *testing.T) {
	if ScreenSplash != 0 {
		t.Errorf("ScreenSplash = %d, want 0", ScreenSplash)
	}
	if ScreenMainMenu != 1 {
		t.Errorf("ScreenMainMenu = %d, want 1", ScreenMainMenu)
	}
	if ScreenExploration != 2 {
		t.Errorf("ScreenExploration = %d, want 2", ScreenExploration)
	}
	if ScreenVictory != 3 {
		t.Errorf("ScreenVictory = %d, want 3", ScreenVictory)
	}
	if ScreenDefeat != 4 {
		t.Errorf("ScreenDefeat = %d, want 4", ScreenDefeat)
	}
}

func TestOverlayState(t *testing.T) {
	o := OverlayState{}
	if o.ShowQuestLog || o.ShowGuildPanel || o.ShowSettings {
		t.Error("OverlayState zero value should have all flags false")
	}
	o.ShowQuestLog = true
	if !o.ShowQuestLog {
		t.Error("ShowQuestLog should be true")
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
		AP:         2,
		MaxAP:      2,
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
	if player.AP != 2 {
		t.Errorf("AP = %d, want 2", player.AP)
	}
	if player.MaxAP != 2 {
		t.Errorf("MaxAP = %d, want 2", player.MaxAP)
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

func TestItemData(t *testing.T) {
	item := ItemData{
		ID:     "item-001",
		Name:   "Sword +1",
		Type:   "weapon",
		Slot:   "main_hand",
		Damage: "1d8+1",
		Weight: 3,
		Value:  250,
	}
	if item.Name != "Sword +1" {
		t.Errorf("Name = %q, want %q", item.Name, "Sword +1")
	}
	if item.Damage != "1d8+1" {
		t.Errorf("Damage = %q, want %q", item.Damage, "1d8+1")
	}
}

func TestSpellData(t *testing.T) {
	spell := SpellData{
		ID:         "spell-001",
		Name:       "Fireball",
		Level:      3,
		School:     4,
		DamageDice: "8d6",
		AreaEffect: true,
	}
	if spell.Name != "Fireball" {
		t.Errorf("Name = %q, want %q", spell.Name, "Fireball")
	}
	if spell.Level != 3 {
		t.Errorf("Level = %d, want 3", spell.Level)
	}
	if !spell.AreaEffect {
		t.Error("AreaEffect should be true")
	}
}

func TestQuestData(t *testing.T) {
	quest := QuestData{
		ID:          "quest-001",
		Title:       "Kill Goblins",
		Description: "Clear the goblin camp",
		Status:      "active",
		Objectives: []QuestObjective{
			{Description: "Kill 3 goblins", Progress: 2, Required: 3, Completed: false},
		},
		Rewards: []QuestReward{
			{Type: "gold", Value: 500},
			{Type: "exp", Value: 200},
		},
	}
	if quest.Title != "Kill Goblins" {
		t.Errorf("Title = %q, want %q", quest.Title, "Kill Goblins")
	}
	if len(quest.Objectives) != 1 {
		t.Fatalf("Objectives len = %d, want 1", len(quest.Objectives))
	}
	if quest.Objectives[0].Progress != 2 {
		t.Errorf("Objectives[0].Progress = %d, want 2", quest.Objectives[0].Progress)
	}
}

func TestAttributeModifier(t *testing.T) {
	tests := []struct {
		score    int
		expected int
	}{
		{8, -1},
		{10, 0},
		{12, 1},
		{14, 2},
		{16, 3},
		{18, 4},
		{20, 5},
	}
	for _, tt := range tests {
		got := AttributeModifier(tt.score)
		if got != tt.expected {
			t.Errorf("AttributeModifier(%d) = %d, want %d", tt.score, got, tt.expected)
		}
	}
}

func TestPointBuyCost(t *testing.T) {
	tests := []struct {
		score    int
		expected int
	}{
		{8, 0},
		{9, 1},
		{10, 2},
		{11, 3},
		{12, 4},
		{13, 5},
		{14, 7},
		{15, 9},
		{7, -1},
		{16, -1},
	}
	for _, tt := range tests {
		got := PointBuyCost(tt.score)
		if got != tt.expected {
			t.Errorf("PointBuyCost(%d) = %d, want %d", tt.score, got, tt.expected)
		}
	}
}

func TestSpellSchoolName(t *testing.T) {
	tests := []struct {
		school   int
		expected string
	}{
		{0, "Abjuration"},
		{4, "Evocation"},
		{7, "Transmutation"},
		{-1, "Unknown"},
		{8, "Unknown"},
	}
	for _, tt := range tests {
		got := SpellSchoolName(tt.school)
		if got != tt.expected {
			t.Errorf("SpellSchoolName(%d) = %q, want %q", tt.school, got, tt.expected)
		}
	}
}

func TestEffectIcon(t *testing.T) {
	tests := []struct {
		effectType string
		expected   string
	}{
		{"burning", "Fire"},
		{"poison", "Pois"},
		{"stun", "Stun"},
		{"unknown_effect", "Eff"},
	}
	for _, tt := range tests {
		got := EffectIcon(tt.effectType)
		if got != tt.expected {
			t.Errorf("EffectIcon(%q) = %q, want %q", tt.effectType, got, tt.expected)
		}
	}
}

func TestClassInfoList(t *testing.T) {
	if len(ClassInfoList) != 6 {
		t.Fatalf("ClassInfoList length = %d, want 6", len(ClassInfoList))
	}

	expectedNames := []string{"Fighter", "Mage", "Cleric", "Thief", "Ranger", "Paladin"}
	for i, expected := range expectedNames {
		if ClassInfoList[i].Name != expected {
			t.Errorf("ClassInfoList[%d].Name = %q, want %q", i, ClassInfoList[i].Name, expected)
		}
	}
}

func TestCharCreationStep(t *testing.T) {
	if CharStepName != 0 {
		t.Errorf("CharStepName = %d, want 0", CharStepName)
	}
	if CharStepClass != 1 {
		t.Errorf("CharStepClass = %d, want 1", CharStepClass)
	}
	if CharStepAttributes != 2 {
		t.Errorf("CharStepAttributes = %d, want 2", CharStepAttributes)
	}
	if CharStepReview != 3 {
		t.Errorf("CharStepReview = %d, want 3", CharStepReview)
	}
}

func TestAttributeMethodString(t *testing.T) {
	tests := []struct {
		method   AttributeMethod
		expected string
	}{
		{AttrMethodRoll, "Roll"},
		{AttrMethodStandard, "Standard"},
		{AttrMethodPointBuy, "Point Buy"},
		{AttrMethodCustom, "Custom"},
		{AttributeMethod(99), "Unknown"},
	}
	for _, tt := range tests {
		got := tt.method.String()
		if got != tt.expected {
			t.Errorf("AttributeMethod(%d).String() = %q, want %q", tt.method, got, tt.expected)
		}
	}
}

func TestCombatActionString(t *testing.T) {
	tests := []struct {
		action   CombatAction
		expected string
	}{
		{CombatActionNone, "None"},
		{CombatActionMove, "Move"},
		{CombatActionAttack, "Attack"},
		{CombatActionCast, "Cast"},
		{CombatActionItem, "Item"},
		{CombatActionDefend, "Defend"},
		{CombatActionFlee, "Flee"},
		{CombatAction(99), "Unknown"},
	}
	for _, tt := range tests {
		got := tt.action.String()
		if got != tt.expected {
			t.Errorf("CombatAction(%d).String() = %q, want %q", tt.action, got, tt.expected)
		}
	}
}

func TestCharCreationStateGetSetAttr(t *testing.T) {
	cc := &CharCreationState{}
	cc.SetStandardArray()

	// Verify standard array loaded correctly
	expected := StandardArray
	for i := 0; i < 6; i++ {
		got := cc.GetAttr(i)
		if got != expected[i] {
			t.Errorf("GetAttr(%d) after SetStandardArray = %d, want %d", i, got, expected[i])
		}
	}

	// Test SetAttr
	cc.SetAttr(0, 18) // STR
	cc.SetAttr(5, 3)  // CHA
	if cc.Attributes.Strength != 18 {
		t.Errorf("SetAttr(0, 18): Strength = %d, want 18", cc.Attributes.Strength)
	}
	if cc.Attributes.Charisma != 3 {
		t.Errorf("SetAttr(5, 3): Charisma = %d, want 3", cc.Attributes.Charisma)
	}

	// Test out-of-range index returns 0
	if cc.GetAttr(6) != 0 {
		t.Errorf("GetAttr(6) = %d, want 0", cc.GetAttr(6))
	}
	if cc.GetAttr(-1) != 0 {
		t.Errorf("GetAttr(-1) = %d, want 0", cc.GetAttr(-1))
	}
}

func TestCharCreationStateResetAttributes(t *testing.T) {
	cc := &CharCreationState{}
	cc.SetStandardArray()

	cc.ResetAttributes(8)
	for i := 0; i < 6; i++ {
		if cc.GetAttr(i) != 8 {
			t.Errorf("ResetAttributes(8): GetAttr(%d) = %d, want 8", i, cc.GetAttr(i))
		}
	}
}

func TestClassInfoHasDescription(t *testing.T) {
	for _, cls := range ClassInfoList {
		if cls.Description == "" {
			t.Errorf("ClassInfo %q has empty Description", cls.Name)
		}
	}
}

func TestItemDataEquippedField(t *testing.T) {
	item := ItemData{
		ID:       "sword-1",
		Name:     "Iron Sword",
		Equipped: true,
	}
	if !item.Equipped {
		t.Error("ItemData.Equipped should be true")
	}
}

func TestVictoryDefeatData(t *testing.T) {
	v := VictoryData{
		AdventureTitle:  "Test Quest",
		QuestsComplete:  3,
		QuestsTotal:     5,
		EnemiesDefeated: 10,
		GoldEarned:      500,
	}
	if v.AdventureTitle != "Test Quest" {
		t.Errorf("VictoryData.AdventureTitle = %q, want %q", v.AdventureTitle, "Test Quest")
	}

	d := DefeatData{
		CauseOfDeath: "HP reached 0",
		LastLocation: "Town Square",
	}
	if d.CauseOfDeath != "HP reached 0" {
		t.Errorf("DefeatData.CauseOfDeath = %q, want %q", d.CauseOfDeath, "HP reached 0")
	}
}
