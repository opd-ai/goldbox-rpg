// Command menu definitions shared between WASM and native builds.
// Rendering helpers are in command_menu.go (js/wasm only).

package wasmui

// CommandDef defines a single command in the command menu.
type CommandDef struct {
	Key         string       // Keyboard shortcut (e.g., "M", "A", "Space")
	Label       string       // Display label (e.g., "Move", "Attack")
	Description string       // Tooltip/status description
	Action      CombatAction // Combat action if applicable (CombatActionNone for non-combat)
	Available   bool         // Whether the command can currently be used
	APCost      int          // AP cost for combat actions (0 = no cost/free)
}

// explorationCommands returns the command set for exploration mode.
func explorationCommands() []CommandDef {
	return []CommandDef{
		{Key: "W/Up", Label: "Forward", Description: "Move forward", Action: CombatActionNone, Available: true},
		{Key: "Q", Label: "Left", Description: "Turn left", Action: CombatActionNone, Available: true},
		{Key: "E", Label: "Right", Description: "Turn right", Action: CombatActionNone, Available: true},
		{Key: "I", Label: "Inventory", Description: "Open inventory", Action: CombatActionNone, Available: true},
		{Key: "C", Label: "Cast", Description: "Open spellbook", Action: CombatActionNone, Available: true},
		{Key: "J", Label: "Journal", Description: "Open quest log", Action: CombatActionNone, Available: true},
		{Key: "G", Label: "Guild", Description: "Open guild panel", Action: CombatActionNone, Available: true},
		{Key: "M", Label: "Map", Description: "Toggle minimap", Action: CombatActionNone, Available: true},
	}
}

// combatCommands returns the command set for combat mode.
func combatCommands(currentAP int) []CommandDef {
	return []CommandDef{
		{Key: "M", Label: "Move", Description: "Move to adjacent tile", Action: CombatActionMove, Available: currentAP >= 1, APCost: 1},
		{Key: "A", Label: "Attack", Description: "Attack target", Action: CombatActionAttack, Available: currentAP >= 1, APCost: 1},
		{Key: "C", Label: "Cast", Description: "Cast a spell", Action: CombatActionCast, Available: true, APCost: 0}, // Varies by spell
		{Key: "U", Label: "Use", Description: "Use an item", Action: CombatActionItem, Available: currentAP >= 1, APCost: 1},
		{Key: "D", Label: "Defend", Description: "Defensive stance (+2 AC)", Action: CombatActionDefend, Available: true, APCost: 0},
		{Key: "F", Label: "Flee", Description: "Attempt to flee combat", Action: CombatActionFlee, Available: true, APCost: 0},
		{Key: "Space", Label: "End", Description: "End your turn", Action: CombatActionNone, Available: true, APCost: 0},
	}
}
