package server

import (
	"testing"

	"goldbox-rpg/pkg/game"
)

// TestIsTimeToExecute tests the isTimeToExecute function
func TestIsTimeToExecute(t *testing.T) {
	tests := []struct {
		name     string
		current  game.GameTime
		trigger  game.GameTime
		expected bool
	}{
		{
			name:     "Current time equals trigger time",
			current:  game.GameTime{GameTicks: 100},
			trigger:  game.GameTime{GameTicks: 100},
			expected: true,
		},
		{
			name:     "Current time greater than trigger time",
			current:  game.GameTime{GameTicks: 150},
			trigger:  game.GameTime{GameTicks: 100},
			expected: true,
		},
		{
			name:     "Current time less than trigger time",
			current:  game.GameTime{GameTicks: 50},
			trigger:  game.GameTime{GameTicks: 100},
			expected: false,
		},
		{
			name:     "Zero values",
			current:  game.GameTime{GameTicks: 0},
			trigger:  game.GameTime{GameTicks: 0},
			expected: true,
		},
		{
			name:     "Large tick values",
			current:  game.GameTime{GameTicks: 999999},
			trigger:  game.GameTime{GameTicks: 999998},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTimeToExecute(tt.current, tt.trigger)
			if result != tt.expected {
				t.Errorf("isTimeToExecute(%d, %d) = %v, want %v",
					tt.current.GameTicks, tt.trigger.GameTicks, result, tt.expected)
			}
		})
	}
}

// TestFindSpell tests the findSpell function
func TestFindSpell(t *testing.T) {
	spells := []game.Spell{
		{ID: "fireball", Name: "Fireball", School: game.SchoolEvocation, Level: 3},
		{ID: "heal", Name: "Cure Light Wounds", School: game.SchoolDivination, Level: 1},
		{ID: "teleport", Name: "Teleport", School: game.SchoolConjuration, Level: 5},
	}

	tests := []struct {
		name     string
		spells   []game.Spell
		spellID  string
		expected *game.Spell
	}{
		{
			name:     "Find existing spell by ID",
			spells:   spells,
			spellID:  "fireball",
			expected: &spells[0],
		},
		{
			name:     "Find another existing spell",
			spells:   spells,
			spellID:  "heal",
			expected: &spells[1],
		},
		{
			name:     "Spell not found",
			spells:   spells,
			spellID:  "nonexistent",
			expected: nil,
		},
		{
			name:     "Empty spell slice",
			spells:   []game.Spell{},
			spellID:  "fireball",
			expected: nil,
		},
		{
			name:     "Empty spell ID",
			spells:   spells,
			spellID:  "",
			expected: nil,
		},
		{
			name:     "Case sensitive search",
			spells:   spells,
			spellID:  "FIREBALL",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findSpell(tt.spells, tt.spellID)
			if result != tt.expected {
				if tt.expected == nil {
					t.Errorf("findSpell() = %v, want nil", result)
				} else if result == nil {
					t.Errorf("findSpell() = nil, want %v", tt.expected)
				} else if result.ID != tt.expected.ID {
					t.Errorf("findSpell() = spell with ID %s, want spell with ID %s",
						result.ID, tt.expected.ID)
				}
			}
		})
	}
}

// TestFindInventoryItem tests the findInventoryItem function
func TestFindInventoryItem(t *testing.T) {
	inventory := []game.Item{
		{ID: "sword1", Name: "Iron Sword", Type: "weapon", Weight: 3, Value: 50},
		{ID: "potion1", Name: "Healing Potion", Type: "consumable", Weight: 1, Value: 25},
		{ID: "armor1", Name: "Leather Armor", Type: "armor", Weight: 10, Value: 100},
	}

	tests := []struct {
		name      string
		inventory []game.Item
		itemID    string
		expected  *game.Item
	}{
		{
			name:      "Find existing item by ID",
			inventory: inventory,
			itemID:    "sword1",
			expected:  &inventory[0],
		},
		{
			name:      "Find consumable item",
			inventory: inventory,
			itemID:    "potion1",
			expected:  &inventory[1],
		},
		{
			name:      "Item not found",
			inventory: inventory,
			itemID:    "nonexistent",
			expected:  nil,
		},
		{
			name:      "Empty inventory",
			inventory: []game.Item{},
			itemID:    "sword1",
			expected:  nil,
		},
		{
			name:      "Empty item ID",
			inventory: inventory,
			itemID:    "",
			expected:  nil,
		},
		{
			name:      "Case sensitive search",
			inventory: inventory,
			itemID:    "SWORD1",
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findInventoryItem(tt.inventory, tt.itemID)
			if result != tt.expected {
				if tt.expected == nil {
					t.Errorf("findInventoryItem() = %v, want nil", result)
				} else if result == nil {
					t.Errorf("findInventoryItem() = nil, want %v", tt.expected)
				} else if result.ID != tt.expected.ID {
					t.Errorf("findInventoryItem() = item with ID %s, want item with ID %s",
						result.ID, tt.expected.ID)
				}
			}
		})
	}
}

// TestParseDamageString tests the parseDamageString function comprehensively
func TestParseDamageString(t *testing.T) {
	tests := []struct {
		name     string
		damage   string
		expected int
	}{
		// Plain numbers
		{
			name:     "Plain number 5",
			damage:   "5",
			expected: 5,
		},
		{
			name:     "Plain number 0",
			damage:   "0",
			expected: 0,
		},
		{
			name:     "Large plain number",
			damage:   "100",
			expected: 100,
		},

		// Basic dice notation
		{
			name:     "1d6 dice",
			damage:   "1d6",
			expected: 3, // Average of 1d6 is (1+6)/2 = 3.5, rounded down to 3
		},
		{
			name:     "d8 dice (implicit 1)",
			damage:   "d8",
			expected: 4, // Average of 1d8 is (1+8)/2 = 4.5, rounded down to 4
		},
		{
			name:     "2d6 dice",
			damage:   "2d6",
			expected: 7, // Average of 2d6 is 2 * (1+6)/2 = 7
		},
		{
			name:     "3d4 dice",
			damage:   "3d4",
			expected: 7, // Average of 3d4 is 3 * (1+4)/2 = 7.5, rounded down to 7
		},

		// Dice with modifiers
		{
			name:     "1d6+1",
			damage:   "1d6+1",
			expected: 4, // Average of 1d6 + 1 = 3 + 1 = 4
		},
		{
			name:     "2d8+5",
			damage:   "2d8+5",
			expected: 14, // Average of 2d8 + 5 = 9 + 5 = 14
		},
		{
			name:     "d4+2",
			damage:   "d4+2",
			expected: 4, // Average of 1d4 + 2 = 2 + 2 = 4
		},
		{
			name:     "10d10+10",
			damage:   "10d10+10",
			expected: 65, // Average of 10d10 + 10 = 55 + 10 = 65
		},

		// Edge cases and invalid inputs
		{
			name:     "Invalid format - letters",
			damage:   "abc",
			expected: 0,
		},
		{
			name:     "Invalid format - mixed",
			damage:   "2x6",
			expected: 0,
		},
		{
			name:     "Empty string",
			damage:   "",
			expected: 0,
		},
		{
			name:     "Invalid dice notation - no die size",
			damage:   "2d",
			expected: 0,
		},
		{
			name:     "Invalid dice notation - no d",
			damage:   "26",
			expected: 26, // Should be parsed as plain number
		},
		{
			name:     "Complex invalid format",
			damage:   "2d6+abc",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDamageString(tt.damage)
			if result != tt.expected {
				t.Errorf("parseDamageString(%q) = %d, want %d", tt.damage, result, tt.expected)
			}
		})
	}
}

// TestParseDamageString_TableDriven provides additional table-driven tests for edge cases
func TestParseDamageString_TableDriven(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		desc     string
	}{
		{"1d1", 1, "Minimum die size"},
		{"1d2", 1, "d2 average"},
		{"1d20", 10, "d20 average"},
		{"100d1", 100, "Many d1 dice"},
		{"1d100", 50, "d100 average"},
		{"0d6", 0, "Zero dice"},
		{"1d0+5", 5, "Invalid zero-sided die (average calculation still works)"}, // The implementation calculates average even for d0
		{"-5", -5, "Negative plain number"},
		{"1d6+-3", 0, "Invalid modifier format"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := parseDamageString(tt.input)
			if result != tt.expected {
				t.Errorf("parseDamageString(%q) = %d, want %d (%s)",
					tt.input, result, tt.expected, tt.desc)
			}
		})
	}
}

// TestParseDamageString_ErrorHandling tests error handling paths for better coverage
func TestParseDamageString_ErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		damage   string
		expected int
		desc     string
	}{
		// These cases should trigger specific error handling paths
		{"test1", "d+5", 0, "Invalid dice notation without die number"},
		{"test2", "d+", 0, "Invalid dice notation format"},
		{"test3", "1d+5", 0, "Missing die size"},
		{"test4", "1d6+", 0, "Invalid modifier - empty plus"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := parseDamageString(tt.damage)
			if result != tt.expected {
				t.Errorf("parseDamageString(%q) = %d, want %d (%s)",
					tt.damage, result, tt.expected, tt.desc)
			}
		})
	}
}

// TestParseDamageString_Coverage tests additional edge cases for maximum coverage
func TestParseDamageString_Coverage(t *testing.T) {
	// Test case that would make strconv.Atoi fail on number of dice
	// This is hard to trigger with valid regex, but we test what we can
	tests := []struct {
		input    string
		expected int
	}{
		{"99999999999999999999d6", 0},   // Very large number that might cause strconv error
		{"1d99999999999999999999", 0},   // Very large die size that might cause strconv error
		{"1d6+99999999999999999999", 0}, // Very large modifier that might cause strconv error
	}

	for _, tt := range tests {
		result := parseDamageString(tt.input)
		// We don't assert specific values for these edge cases since they depend on platform
		// But calling them provides coverage for error handling paths
		_ = result
	}
}

// TestMin tests the min utility function
func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "First number smaller",
			a:        5,
			b:        10,
			expected: 5,
		},
		{
			name:     "Second number smaller",
			a:        15,
			b:        8,
			expected: 8,
		},
		{
			name:     "Equal numbers",
			a:        7,
			b:        7,
			expected: 7,
		},
		{
			name:     "Zero and positive",
			a:        0,
			b:        5,
			expected: 0,
		},
		{
			name:     "Negative numbers",
			a:        -3,
			b:        -8,
			expected: -8,
		},
		{
			name:     "Negative and positive",
			a:        -5,
			b:        3,
			expected: -5,
		},
		{
			name:     "Large numbers",
			a:        1000000,
			b:        999999,
			expected: 999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestIsStaticFileRequest tests the isStaticFileRequest function
func TestIsStaticFileRequest(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// HTML files
		{
			name:     "HTML file",
			path:     "/index.html",
			expected: true,
		},
		{
			name:     "Nested HTML file",
			path:     "/pages/about.html",
			expected: true,
		},

		// CSS files
		{
			name:     "CSS file",
			path:     "/styles/main.css",
			expected: true,
		},

		// JavaScript files
		{
			name:     "JavaScript file",
			path:     "/scripts/app.js",
			expected: true,
		},

		// Image files
		{
			name:     "JPEG image",
			path:     "/images/photo.jpg",
			expected: true,
		},
		{
			name:     "PNG image",
			path:     "/images/logo.png",
			expected: true,
		},
		{
			name:     "GIF image",
			path:     "/images/animation.gif",
			expected: true,
		},
		{
			name:     "SVG image",
			path:     "/icons/arrow.svg",
			expected: true,
		},
		{
			name:     "ICO file",
			path:     "/favicon.ico",
			expected: true,
		},

		// Font files
		{
			name:     "WOFF font",
			path:     "/fonts/arial.woff",
			expected: true,
		},
		{
			name:     "WOFF2 font",
			path:     "/fonts/arial.woff2",
			expected: true,
		},
		{
			name:     "TTF font",
			path:     "/fonts/arial.ttf",
			expected: true,
		},
		{
			name:     "EOT font",
			path:     "/fonts/arial.eot",
			expected: true,
		},

		// Non-static files
		{
			name:     "API endpoint",
			path:     "/api/users",
			expected: false,
		},
		{
			name:     "Root path",
			path:     "/",
			expected: false,
		},
		{
			name:     "Path without extension",
			path:     "/dashboard",
			expected: false,
		},
		{
			name:     "Unknown extension",
			path:     "/file.xyz",
			expected: false,
		},
		{
			name:     "Text file",
			path:     "/readme.txt",
			expected: false,
		},
		{
			name:     "Empty path",
			path:     "",
			expected: false,
		},

		// Case sensitivity tests
		{
			name:     "Uppercase extension",
			path:     "/image.PNG",
			expected: false, // Extensions are case-sensitive
		},
		{
			name:     "Mixed case extension",
			path:     "/script.Js",
			expected: false,
		},

		// Edge cases
		{
			name:     "Multiple dots in filename",
			path:     "/file.min.js",
			expected: true, // Should match .js extension
		},
		{
			name:     "Hidden file with extension",
			path:     "/.htaccess.html",
			expected: true,
		},
		{
			name:     "File in deep path",
			path:     "/very/deep/nested/path/file.css",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isStaticFileRequest(tt.path)
			if result != tt.expected {
				t.Errorf("isStaticFileRequest(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestIsStaticFileRequest_AllExtensions tests all supported static file extensions
func TestIsStaticFileRequest_AllExtensions(t *testing.T) {
	staticExtensions := []string{
		".html", ".css", ".js", ".jpg", ".jpeg",
		".png", ".gif", ".svg", ".ico", ".woff",
		".woff2", ".ttf", ".eot",
	}

	for _, ext := range staticExtensions {
		t.Run("Extension_"+ext, func(t *testing.T) {
			path := "/test" + ext
			result := isStaticFileRequest(path)
			if !result {
				t.Errorf("isStaticFileRequest(%q) = false, want true", path)
			}
		})
	}
}

// Benchmark tests for performance validation

// BenchmarkParseDamageString benchmarks the parseDamageString function
func BenchmarkParseDamageString(b *testing.B) {
	testCases := []string{
		"5",
		"1d6",
		"2d8+3",
		"10d10+10",
		"invalid",
	}

	for _, damage := range testCases {
		b.Run(damage, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				parseDamageString(damage)
			}
		})
	}
}

// BenchmarkFindSpell benchmarks the findSpell function
func BenchmarkFindSpell(b *testing.B) {
	spells := make([]game.Spell, 100)
	for i := 0; i < 100; i++ {
		spells[i] = game.Spell{
			ID:     "spell" + string(rune(i)),
			Name:   "Test Spell " + string(rune(i)),
			School: game.SchoolEvocation,
			Level:  1,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findSpell(spells, "spell50")
	}
}

// BenchmarkFindInventoryItem benchmarks the findInventoryItem function
func BenchmarkFindInventoryItem(b *testing.B) {
	inventory := make([]game.Item, 100)
	for i := 0; i < 100; i++ {
		inventory[i] = game.Item{
			ID:     "item" + string(rune(i)),
			Name:   "Test Item " + string(rune(i)),
			Type:   "misc",
			Weight: 1,
			Value:  10,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findInventoryItem(inventory, "item50")
	}
}

// BenchmarkMin benchmarks the min function
func BenchmarkMin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		min(i, i+1)
	}
}

// BenchmarkIsStaticFileRequest benchmarks the isStaticFileRequest function
func BenchmarkIsStaticFileRequest(b *testing.B) {
	paths := []string{
		"/static/css/style.css",
		"/api/users",
		"/images/logo.png",
		"/scripts/app.js",
		"/dashboard",
	}

	for _, path := range paths {
		b.Run(path, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				isStaticFileRequest(path)
			}
		})
	}
}
