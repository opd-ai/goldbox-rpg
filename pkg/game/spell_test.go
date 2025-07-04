package game

import (
	"reflect"
	"testing"
)

// TestSpellSchool_Constants tests that all spell school constants have expected values
func TestSpellSchool_Constants(t *testing.T) {
	tests := []struct {
		name     string
		school   SpellSchool
		expected int
	}{
		{"SchoolAbjuration should be 0", SchoolAbjuration, 0},
		{"SchoolConjuration should be 1", SchoolConjuration, 1},
		{"SchoolDivination should be 2", SchoolDivination, 2},
		{"SchoolEnchantment should be 3", SchoolEnchantment, 3},
		{"SchoolEvocation should be 4", SchoolEvocation, 4},
		{"SchoolIllusion should be 5", SchoolIllusion, 5},
		{"SchoolNecromancy should be 6", SchoolNecromancy, 6},
		{"SchoolTransmutation should be 7", SchoolTransmutation, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.school) != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, int(tt.school), tt.expected)
			}
		})
	}
}

// TestSpellComponent_Constants tests that all spell component constants have expected values
func TestSpellComponent_Constants(t *testing.T) {
	tests := []struct {
		name      string
		component SpellComponent
		expected  int
	}{
		{"ComponentVerbal should be 0", ComponentVerbal, 0},
		{"ComponentSomatic should be 1", ComponentSomatic, 1},
		{"ComponentMaterial should be 2", ComponentMaterial, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.component) != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, int(tt.component), tt.expected)
			}
		})
	}
}

// TestSpell_StructFields tests Spell struct field assignment and initialization
func TestSpell_StructFields(t *testing.T) {
	components := []SpellComponent{ComponentVerbal, ComponentSomatic}
	spell := Spell{
		ID:          "fireball",
		Name:        "Fireball",
		Level:       3,
		School:      SchoolEvocation,
		Range:       150,
		Duration:    0,
		Components:  components,
		Description: "A bright streak flashes from your pointing finger to a point you choose.",
	}

	// Test field assignment
	if spell.ID != "fireball" {
		t.Errorf("ID = %q, want %q", spell.ID, "fireball")
	}
	if spell.Name != "Fireball" {
		t.Errorf("Name = %q, want %q", spell.Name, "Fireball")
	}
	if spell.Level != 3 {
		t.Errorf("Level = %d, want %d", spell.Level, 3)
	}
	if spell.School != SchoolEvocation {
		t.Errorf("School = %v, want %v", spell.School, SchoolEvocation)
	}
	if spell.Range != 150 {
		t.Errorf("Range = %d, want %d", spell.Range, 150)
	}
	if spell.Duration != 0 {
		t.Errorf("Duration = %d, want %d", spell.Duration, 0)
	}
	if !reflect.DeepEqual(spell.Components, components) {
		t.Errorf("Components = %v, want %v", spell.Components, components)
	}
	if spell.Description != "A bright streak flashes from your pointing finger to a point you choose." {
		t.Errorf("Description = %q, want %q", spell.Description, "A bright streak flashes from your pointing finger to a point you choose.")
	}
}

// TestSpell_EmptyValues tests Spell struct with default/empty values
func TestSpell_EmptyValues(t *testing.T) {
	spell := Spell{}

	// Test zero values
	if spell.ID != "" {
		t.Errorf("Default ID = %q, want empty string", spell.ID)
	}
	if spell.Name != "" {
		t.Errorf("Default Name = %q, want empty string", spell.Name)
	}
	if spell.Level != 0 {
		t.Errorf("Default Level = %d, want 0", spell.Level)
	}
	if spell.School != SchoolAbjuration { // SchoolAbjuration is 0, the zero value
		t.Errorf("Default School = %v, want %v", spell.School, SchoolAbjuration)
	}
	if spell.Range != 0 {
		t.Errorf("Default Range = %d, want 0", spell.Range)
	}
	if spell.Duration != 0 {
		t.Errorf("Default Duration = %d, want 0", spell.Duration)
	}
	if spell.Components != nil {
		t.Errorf("Default Components = %v, want nil", spell.Components)
	}
	if spell.Description != "" {
		t.Errorf("Default Description = %q, want empty string", spell.Description)
	}
}

// TestSpell_AllSpellSchools tests creating spells with all available spell schools
func TestSpell_AllSpellSchools(t *testing.T) {
	schools := []struct {
		school SpellSchool
		name   string
	}{
		{SchoolAbjuration, "Abjuration"},
		{SchoolConjuration, "Conjuration"},
		{SchoolDivination, "Divination"},
		{SchoolEnchantment, "Enchantment"},
		{SchoolEvocation, "Evocation"},
		{SchoolIllusion, "Illusion"},
		{SchoolNecromancy, "Necromancy"},
		{SchoolTransmutation, "Transmutation"},
	}

	for _, tt := range schools {
		t.Run(tt.name, func(t *testing.T) {
			spell := Spell{
				ID:     "test_spell",
				Name:   "Test Spell",
				School: tt.school,
			}

			if spell.School != tt.school {
				t.Errorf("Spell school = %v, want %v", spell.School, tt.school)
			}
		})
	}
}

// TestSpell_AllSpellComponents tests creating spells with all available spell components
func TestSpell_AllSpellComponents(t *testing.T) {
	tests := []struct {
		name       string
		components []SpellComponent
	}{
		{
			name:       "Verbal only",
			components: []SpellComponent{ComponentVerbal},
		},
		{
			name:       "Somatic only",
			components: []SpellComponent{ComponentSomatic},
		},
		{
			name:       "Material only",
			components: []SpellComponent{ComponentMaterial},
		},
		{
			name:       "Verbal and Somatic",
			components: []SpellComponent{ComponentVerbal, ComponentSomatic},
		},
		{
			name:       "Verbal and Material",
			components: []SpellComponent{ComponentVerbal, ComponentMaterial},
		},
		{
			name:       "Somatic and Material",
			components: []SpellComponent{ComponentSomatic, ComponentMaterial},
		},
		{
			name:       "All components",
			components: []SpellComponent{ComponentVerbal, ComponentSomatic, ComponentMaterial},
		},
		{
			name:       "No components",
			components: []SpellComponent{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spell := Spell{
				ID:         "test_spell",
				Name:       "Test Spell",
				Components: tt.components,
			}

			if !reflect.DeepEqual(spell.Components, tt.components) {
				t.Errorf("Spell components = %v, want %v", spell.Components, tt.components)
			}
		})
	}
}

// TestSpell_LevelValidation tests spell creation with various level values
func TestSpell_LevelValidation(t *testing.T) {
	tests := []struct {
		name  string
		level int
		valid bool
	}{
		{"Cantrip (level 0)", 0, true},
		{"First level", 1, true},
		{"Third level", 3, true},
		{"Ninth level", 9, true},
		{"High level spell", 15, true},
		{"Negative level", -1, false}, // Not typical but testing edge case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spell := Spell{
				ID:    "test_spell",
				Name:  "Test Spell",
				Level: tt.level,
			}

			// Verify the level was set correctly
			if spell.Level != tt.level {
				t.Errorf("Spell level = %d, want %d", spell.Level, tt.level)
			}

			// For negative levels, we can test the behavior
			if !tt.valid && spell.Level >= 0 {
				t.Errorf("Expected negative level to be preserved, got %d", spell.Level)
			}
		})
	}
}

// TestSpell_RangeValidation tests spell creation with various range values
func TestSpell_RangeValidation(t *testing.T) {
	tests := []struct {
		name   string
		range_ int
		desc   string
	}{
		{"Touch spell", 0, "Touch range spell"},
		{"Short range", 30, "Short range spell"},
		{"Medium range", 120, "Medium range spell"},
		{"Long range", 400, "Long range spell"},
		{"Sight range", 1000, "Sight range spell"},
		{"Unlimited range", -1, "Unlimited range spell"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spell := Spell{
				ID:    "test_spell",
				Name:  "Test Spell",
				Range: tt.range_,
			}

			if spell.Range != tt.range_ {
				t.Errorf("Spell range = %d, want %d", spell.Range, tt.range_)
			}
		})
	}
}

// TestSpell_DurationValidation tests spell creation with various duration values
func TestSpell_DurationValidation(t *testing.T) {
	tests := []struct {
		name     string
		duration int
		desc     string
	}{
		{"Instantaneous", 0, "Instantaneous effect"},
		{"One round", 1, "Lasts one round"},
		{"One minute", 10, "Lasts ten rounds (one minute)"},
		{"Ten minutes", 100, "Lasts one hundred rounds (ten minutes)"},
		{"One hour", 600, "Lasts six hundred rounds (one hour)"},
		{"Permanent", -1, "Permanent duration"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spell := Spell{
				ID:       "test_spell",
				Name:     "Test Spell",
				Duration: tt.duration,
			}

			if spell.Duration != tt.duration {
				t.Errorf("Spell duration = %d, want %d", spell.Duration, tt.duration)
			}
		})
	}
}

// TestSpell_ComplexSpellCreation tests creating realistic, complex spell configurations
func TestSpell_ComplexSpellCreation(t *testing.T) {
	tests := []struct {
		name  string
		spell Spell
	}{
		{
			name: "Fireball",
			spell: Spell{
				ID:          "fireball",
				Name:        "Fireball",
				Level:       3,
				School:      SchoolEvocation,
				Range:       150,
				Duration:    0,
				Components:  []SpellComponent{ComponentVerbal, ComponentSomatic, ComponentMaterial},
				Description: "A bright streak flashes from your pointing finger to a point you choose within range and then blossoms with a low roar into an explosion of flame.",
			},
		},
		{
			name: "Shield",
			spell: Spell{
				ID:          "shield",
				Name:        "Shield",
				Level:       1,
				School:      SchoolAbjuration,
				Range:       0,
				Duration:    10,
				Components:  []SpellComponent{ComponentVerbal, ComponentSomatic},
				Description: "An invisible barrier of magical force appears and protects you.",
			},
		},
		{
			name: "Detect Magic",
			spell: Spell{
				ID:          "detect_magic",
				Name:        "Detect Magic",
				Level:       1,
				School:      SchoolDivination,
				Range:       0,
				Duration:    600,
				Components:  []SpellComponent{ComponentVerbal, ComponentSomatic},
				Description: "For the duration, you sense the presence of magic within 30 feet of you.",
			},
		},
		{
			name: "Charm Person",
			spell: Spell{
				ID:          "charm_person",
				Name:        "Charm Person",
				Level:       1,
				School:      SchoolEnchantment,
				Range:       30,
				Duration:    3600,
				Components:  []SpellComponent{ComponentVerbal, ComponentSomatic},
				Description: "You attempt to charm a humanoid you can see within range.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spell := tt.spell

			// Test all fields are set correctly
			if spell.ID != tt.spell.ID {
				t.Errorf("ID = %q, want %q", spell.ID, tt.spell.ID)
			}
			if spell.Name != tt.spell.Name {
				t.Errorf("Name = %q, want %q", spell.Name, tt.spell.Name)
			}
			if spell.Level != tt.spell.Level {
				t.Errorf("Level = %d, want %d", spell.Level, tt.spell.Level)
			}
			if spell.School != tt.spell.School {
				t.Errorf("School = %v, want %v", spell.School, tt.spell.School)
			}
			if spell.Range != tt.spell.Range {
				t.Errorf("Range = %d, want %d", spell.Range, tt.spell.Range)
			}
			if spell.Duration != tt.spell.Duration {
				t.Errorf("Duration = %d, want %d", spell.Duration, tt.spell.Duration)
			}
			if !reflect.DeepEqual(spell.Components, tt.spell.Components) {
				t.Errorf("Components = %v, want %v", spell.Components, tt.spell.Components)
			}
			if spell.Description != tt.spell.Description {
				t.Errorf("Description = %q, want %q", spell.Description, tt.spell.Description)
			}
		})
	}
}

// TestSpellSchool_TypeConversion tests converting between SpellSchool and int
func TestSpellSchool_TypeConversion(t *testing.T) {
	tests := []struct {
		school   SpellSchool
		intValue int
	}{
		{SchoolAbjuration, 0},
		{SchoolConjuration, 1},
		{SchoolDivination, 2},
		{SchoolEnchantment, 3},
		{SchoolEvocation, 4},
		{SchoolIllusion, 5},
		{SchoolNecromancy, 6},
		{SchoolTransmutation, 7},
	}

	for _, tt := range tests {
		t.Run("School to int conversion", func(t *testing.T) {
			if int(tt.school) != tt.intValue {
				t.Errorf("int(%v) = %d, want %d", tt.school, int(tt.school), tt.intValue)
			}
		})

		t.Run("Int to school conversion", func(t *testing.T) {
			converted := SpellSchool(tt.intValue)
			if converted != tt.school {
				t.Errorf("SpellSchool(%d) = %v, want %v", tt.intValue, converted, tt.school)
			}
		})
	}
}

// TestSpellComponent_TypeConversion tests converting between SpellComponent and int
func TestSpellComponent_TypeConversion(t *testing.T) {
	tests := []struct {
		component SpellComponent
		intValue  int
	}{
		{ComponentVerbal, 0},
		{ComponentSomatic, 1},
		{ComponentMaterial, 2},
	}

	for _, tt := range tests {
		t.Run("Component to int conversion", func(t *testing.T) {
			if int(tt.component) != tt.intValue {
				t.Errorf("int(%v) = %d, want %d", tt.component, int(tt.component), tt.intValue)
			}
		})

		t.Run("Int to component conversion", func(t *testing.T) {
			converted := SpellComponent(tt.intValue)
			if converted != tt.component {
				t.Errorf("SpellComponent(%d) = %v, want %v", tt.intValue, converted, tt.component)
			}
		})
	}
}
