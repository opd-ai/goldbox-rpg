# Inclusive Character Customization — Implementation Plan

## Summary

This plan adds cosmetic appearance options (skin tone, hair, body type, gender expression) and biographical identity fields (pronouns, romantic orientation) to the GoldBox RPG Engine's character creation system. Changes span the `Character` data model, the JSON-RPC creation endpoint, input validation, the asset generation pipeline, and the Ebitengine/WASM frontend. All new fields are purely cosmetic/biographical — they never affect attributes, combat, or class eligibility. Existing saved characters, RPC callers, and tests continue to work unmodified because every new field defaults to its zero value.

---

## Phase 1 — Character Data Model

### Files to modify

| File | Action |
|------|--------|
| `pkg/game/character.go` | Add `Appearance` struct and field to `Character`; update `Clone()` |
| `pkg/game/character_creation.go` | Add appearance fields to `CharacterCreationConfig`; wire into `buildBaseCharacter()` |
| `pkg/game/appearance.go` *(new)* | Define `Appearance` struct, enums, palette map, defaults, and helpers |

### 1a. New file `pkg/game/appearance.go`

```go
package game

// BodyType represents a character's body type (cosmetic only).
type BodyType int

const (
	BodyDefault  BodyType = iota // zero-value; treated as "average" at render time
	BodySlim
	BodyAverage
	BodyMuscular
	BodyStocky
	BodyLarge
)

// SkinTone palette — numeric scale 1-10.
// Mapped to named entries for UI display; decoupled from fantasy race.
var SkinTonePalette = map[int]string{
	1:  "porcelain",
	2:  "ivory",
	3:  "beige",
	4:  "sand",
	5:  "tan",
	6:  "bronze",
	7:  "umber",
	8:  "chestnut",
	9:  "espresso",
	10: "obsidian",
}

// Appearance holds cosmetic and biographical character properties.
// None of these fields affect attributes, combat, or class eligibility.
// All fields are optional; zero values are valid defaults.
type Appearance struct {
	SkinTone            int      `yaml:"skin_tone"            json:"skin_tone,omitempty"`             // 1-10; 0 = unset (engine picks mid-range default)
	HairStyle           string   `yaml:"hair_style"           json:"hair_style,omitempty"`            // freeform e.g. "long braids", "shaved"
	HairColor           string   `yaml:"hair_color"           json:"hair_color,omitempty"`            // freeform e.g. "auburn", "silver"
	BodyType            BodyType `yaml:"body_type"            json:"body_type,omitempty"`             // enum; 0 = default
	GenderExpression    string   `yaml:"gender_expression"    json:"gender_expression,omitempty"`     // freeform; drives portrait selection + pronoun logic
	Pronouns            string   `yaml:"pronouns"             json:"pronouns,omitempty"`              // e.g. "he/him", "she/her", "they/them", custom
	RomanticOrientation string   `yaml:"romantic_orientation" json:"romantic_orientation,omitempty"`  // optional; stored, never mechanically referenced
}

// DefaultAppearance returns an Appearance with sensible mid-range defaults.
func DefaultAppearance() Appearance {
	return Appearance{
		SkinTone: 5,      // "tan" — visual midpoint
		BodyType: BodyAverage,
	}
}

// SkinToneName returns the named palette entry for the tone value,
// or "unknown" if out of range.
func SkinToneName(tone int) string {
	if name, ok := SkinTonePalette[tone]; ok {
		return name
	}
	return "unknown"
}

// SkinToneGroup returns "light" (1-3), "medium" (4-7), or "dark" (8-10).
// Used for portrait asset lookup.
func SkinToneGroup(tone int) string {
	switch {
	case tone <= 0:
		return "medium"
	case tone <= 3:
		return "light"
	case tone <= 7:
		return "medium"
	default:
		return "dark"
	}
}

// PortraitTag returns the gender-expression tag for asset lookup.
// Maps common expressions to "a" / "b" / "nb" portrait sets;
// unknown values default to "nb".
func (a Appearance) PortraitTag() string {
	switch a.GenderExpression {
	case "masculine", "male":
		return "a"
	case "feminine", "female":
		return "b"
	default:
		return "nb"
	}
}
```

### 1b. Add `Appearance` field to `Character` struct (`character.go`)

```go
// Inside Character struct — add after the `Description` field:
Appearance  Appearance `yaml:"char_appearance" json:"appearance,omitempty"` // Cosmetic & biographical (no gameplay impact)
```

Update `Clone()` to copy the new field (simple value copy — no pointers):

```go
clone := &Character{
    // ... existing fields ...
    Appearance: c.Appearance,
    // ...
}
```

Thread safety: `Appearance` is a plain value struct copied under the existing `mu.RLock()` in `Clone()`. Any future setter (e.g., `SetAppearance()`) must acquire `mu.Lock()`.

### 1c. Extend `CharacterCreationConfig` (`character_creation.go`)

```go
type CharacterCreationConfig struct {
    // ... existing fields ...
    Appearance Appearance `yaml:"creation_appearance" json:"appearance,omitempty"` // Cosmetic overrides
}
```

Update `buildBaseCharacter()`:

```go
func (cc *CharacterCreator) buildBaseCharacter(config CharacterCreationConfig, attributes map[string]int) *Character {
    char := &Character{
        // ... existing fields ...
        Appearance: config.Appearance,
    }
    // If caller did not set any appearance, apply defaults
    if char.Appearance == (Appearance{}) {
        char.Appearance = DefaultAppearance()
    }
    return char
}
```

### Estimated effort

| Item | LOC |
|------|-----|
| `appearance.go` (new) | ~100 |
| `character.go` changes | ~10 |
| `character_creation.go` changes | ~15 |
| **Subtotal** | **~125** |

---

## Phase 2 — RPC Layer

### Files to modify

| File | Action |
|------|--------|
| `pkg/server/handlers.go` | Extend `createCharacterRequest`; update `buildCharacterConfig()` |
| `pkg/validation/validation.go` | Add optional appearance validators |
| `pkg/validation/validation_helpers.go` | Add `validateOptionalIntRange`, `validateOptionalEnum` if missing |

### 2a. Extend `createCharacterRequest` (`handlers.go` ~L1342)

```go
type createCharacterRequest struct {
    Name              string         `json:"name"`
    Class             string         `json:"class"`
    AttributeMethod   string         `json:"attribute_method"`
    CustomAttributes  map[string]int `json:"custom_attributes,omitempty"`
    StartingEquipment bool           `json:"starting_equipment"`
    StartingGold      int            `json:"starting_gold"`
    // New cosmetic / identity fields — all optional
    SkinTone            int    `json:"skin_tone,omitempty"`
    HairStyle           string `json:"hair_style,omitempty"`
    HairColor           string `json:"hair_color,omitempty"`
    BodyType            int    `json:"body_type,omitempty"`
    GenderExpression    string `json:"gender_expression,omitempty"`
    Pronouns            string `json:"pronouns,omitempty"`
    RomanticOrientation string `json:"romantic_orientation,omitempty"`
}
```

### 2b. Update `buildCharacterConfig()` (`handlers.go` ~L1365)

Map the new request fields into `game.CharacterCreationConfig.Appearance`:

```go
config.Appearance = game.Appearance{
    SkinTone:            req.SkinTone,
    HairStyle:           req.HairStyle,
    HairColor:           req.HairColor,
    BodyType:            game.BodyType(req.BodyType),
    GenderExpression:    req.GenderExpression,
    Pronouns:            req.Pronouns,
    RomanticOrientation: req.RomanticOrientation,
}
```

### 2c. Add validators (`validation.go` ~L275)

Extend `validateCreateCharacter()` with **optional** field checks:

```go
// After existing name/class validation:

if skinTone, ok := paramMap["skin_tone"]; ok {
    v, err := toFloat64(skinTone)
    if err != nil || v < 1 || v > 10 {
        return fmt.Errorf("createCharacter: skin_tone must be 1-10")
    }
}

if bodyType, ok := paramMap["body_type"]; ok {
    v, err := toFloat64(bodyType)
    if err != nil || v < 0 || v > 5 {
        return fmt.Errorf("createCharacter: body_type must be 0-5")
    }
}

// String fields: pronouns, gender_expression, hair_style, hair_color, romantic_orientation
// Accept any non-empty UTF-8 string ≤100 chars (re-use validatePlayerName logic with raised limit)
for _, field := range []string{"pronouns", "gender_expression", "hair_style", "hair_color", "romantic_orientation"} {
    if val, ok := paramMap[field]; ok {
        s, ok := val.(string)
        if !ok || len(s) > 100 {
            return fmt.Errorf("createCharacter: %s must be a string ≤100 characters", field)
        }
    }
}
```

All new fields are **accepted but never required** — omitting them produces zero values that `DefaultAppearance()` fills in downstream.

### Estimated effort

| Item | LOC |
|------|-----|
| `handlers.go` changes | ~25 |
| `validation.go` changes | ~35 |
| **Subtotal** | **~60** |

---

## Phase 3 — Asset Pipeline

### Files to modify / create

| File | Action |
|------|--------|
| `game-assets.yaml` | Restructure portrait entries; add tone/expression/body-type variants |
| `assets/characters.yaml` | Mirror restructuring |
| `assets/appearance-fragments.yaml` *(new)* | Prompt-fragment library for skin tones and body types |
| `scripts/generate-all.sh` | Accept `--skin-tones` and handle expanded matrix |
| `Makefile` | Add `assets-portraits-expanded` target |

### 3a. Prompt-Fragment Library — `assets/appearance-fragments.yaml` *(new)*

```yaml
# Prompt fragments composed into portrait generation prompts.
# Keyed by field and variant; referenced by game-assets.yaml via `${fragment}` syntax.

skin_tone_groups:
  light:
    prompt_modifier: "light skin, fair complexion"
    seed_offset: 0
  medium:
    prompt_modifier: "medium skin, warm complexion"
    seed_offset: 1000
  dark:
    prompt_modifier: "dark skin, deep complexion"
    seed_offset: 2000

body_types:
  slim:
    prompt_modifier: "slim build"
    seed_offset: 0
  average:
    prompt_modifier: "average build"
    seed_offset: 100
  muscular:
    prompt_modifier: "muscular build, athletic"
    seed_offset: 200
  stocky:
    prompt_modifier: "stocky build, broad shouldered"
    seed_offset: 300
  large:
    prompt_modifier: "large build, heavy set"
    seed_offset: 400

expression_tags:
  a:
    prompt_modifier: "masculine presenting"
    label: "Expression A"
  b:
    prompt_modifier: "feminine presenting"
    label: "Expression B"
  nb:
    prompt_modifier: "androgynous, gender-neutral"
    label: "Expression C"
```

### 3b. Revised portrait naming convention

**Current**: `portrait_{class}_{race}_{male|female}.png`
**New**: `portrait_{class}_{race}_{expression}_{tone_group}.png`

Where:
- `{expression}` ∈ `{a, b, nb}` — replaces binary `male`/`female`
- `{tone_group}` ∈ `{light, medium, dark}`

Example: `portrait_fighter_human_a_medium.png`

**Legacy filenames are preserved as symlinks or aliases** so existing callers continue to resolve.

### 3c. Portrait entry template in `game-assets.yaml`

Replace each hand-written portrait entry with a templated expansion. Example for one class/race combination:

```yaml
# Before (1 entry):
- id: fighter_human_male
  prompt: "male human fighter, strong jawline, short brown hair, plate armor..."
  filename: "portrait_fighter_human_male.png"

# After (9 entries generated from template):
# fighter × human × {a,b,nb} × {light,medium,dark}
- id: fighter_human_a_light
  prompt: "human fighter, strong jawline, short brown hair, plate armor, determined expression, ${expression.a.prompt_modifier}, ${skin_tone.light.prompt_modifier}"
  filename: "portrait_fighter_human_a_light.png"
  seed_offset: 0       # base + expression(a=0) + tone(light=0)

- id: fighter_human_a_medium
  prompt: "human fighter, strong jawline, short brown hair, plate armor, determined expression, ${expression.a.prompt_modifier}, ${skin_tone.medium.prompt_modifier}"
  filename: "portrait_fighter_human_a_medium.png"
  seed_offset: 1000    # base + expression(a=0) + tone(medium=1000)

# ... 7 more variants per class/race combination
```

### 3d. Body-type variants for character sprites

Body-type variants apply to **character sprites** (which show the full or upper body), not to close-up portraits. Add a separate sprite subgroup:

```yaml
- name: Character Sprites (Body Variants)
  output_dir: characters/sprites
  seed_offset: 5000
  assets:
    # Template: sprite_{class}_{body_type}.png
    # 6 classes × 5 body types = 30 sprites
    - id: sprite_fighter_slim
      prompt: "fighter character sprite, slim build, plate armor, pixel art..."
      filename: "sprite_fighter_slim.png"
    # ...
```

### 3e. Update `scripts/generate-all.sh`

Add tone/expression matrix expansion:

```bash
# New flag
SKIN_TONE_GROUPS=("light" "medium" "dark")
EXPRESSION_TAGS=("a" "b" "nb")

# In the portrait generation loop, expand the matrix:
for class in fighter mage cleric thief ranger paladin; do
  for race in human elf dwarf halfling; do
    for expr in "${EXPRESSION_TAGS[@]}"; do
      for tone in "${SKIN_TONE_GROUPS[@]}"; do
        generate_portrait "${class}" "${race}" "${expr}" "${tone}"
      done
    done
  done
done
```

### 3f. Makefile additions

```makefile
# Generate expanded portrait matrix only
assets-portraits-expanded:
	@echo "Generating expanded portrait matrix (skin tones + expressions)..."
	./scripts/generate-all.sh --category portraits --seed $(SEED) --model $(MODEL)

# Backward-compat aliases: generate legacy symlinks
assets-legacy-symlinks:
	@echo "Creating legacy portrait symlinks..."
	./scripts/create-legacy-symlinks.sh
```

### Estimated effort

| Item | LOC / assets |
|------|--------------|
| `appearance-fragments.yaml` (new) | ~60 lines YAML |
| `game-assets.yaml` restructuring | ~800 lines YAML (mostly templated) |
| `assets/characters.yaml` update | ~200 lines YAML |
| `generate-all.sh` changes | ~40 lines Bash |
| `create-legacy-symlinks.sh` (new) | ~30 lines Bash |
| `Makefile` changes | ~10 lines |
| **Subtotal** | **~1 140 lines config/script** |

---

## Phase 4 — Frontend / WASM UI

### Files to modify / create

| File | Action |
|------|--------|
| `pkg/wasmui/types.go` | Add `ModeCharacterCreation` to `UIMode` enum; add `Appearance` mirror type |
| `pkg/wasmui/game.go` | Route `ModeCharacterCreation` in `Update()`/`Draw()` |
| `pkg/wasmui/character_creation_screen.go` *(new)* | Character creation screen with appearance fields |
| `pkg/wasmui/ui_widgets.go` *(new, optional)* | Reusable picker/dropdown/text-input widgets |

### 4a. New `UIMode` value

```go
const (
    ModeNormal UIMode = iota
    ModeCombat
    ModeInventory
    ModeSpellcasting
    ModeAdventureSelect
    ModeCharacterCreation // NEW
)
```

### 4b. Character creation screen layout

The screen presents fields top-to-bottom, grouped naturally alongside existing creation fields. No special section headers or explanatory labels for the new fields — they sit inline as standard options.

```
┌─────────────────────────────────────────┐
│           Create Your Character          │
├─────────────────────────────────────────┤
│  Name:       [________________]          │
│  Class:      [Fighter ▾]                 │
│                                          │
│  Skin Tone:  [ ◉◉◉◉◉◉◉◉◉◉ ] (1-10)    │
│              "tan"                        │
│  Hair Style: [________________]          │
│  Hair Color: [________________]          │
│  Body Type:  [Average ▾]                 │
│                                          │
│  Expression: [________________]          │
│  Pronouns:   [they/them ▾]  or custom    │
│  Orientation:[________________] optional  │
│                                          │
│  Attributes: [Roll ▾]                    │
│   STR 14  DEX 12  CON 13                │
│   INT 10  WIS 11  CHA  9                │
│                                          │
│         [ Create Character ]             │
└─────────────────────────────────────────┘
```

**Skin Tone** uses a horizontal picker — 10 cells rendered with palette colors using `drawRect()`. The selected cell is highlighted; the palette name appears below.

**Pronouns** defaults to a dropdown (`he/him`, `she/her`, `they/them`) but accepts free-text entry for custom pronouns.

**Body Type** is a simple dropdown: Slim / Average / Muscular / Stocky / Large.

**Gender Expression**, **Hair Style**, **Hair Color**, **Romantic Orientation** are free-text fields.

All labels use the same `ebitenutil.DebugPrintAt()` style as existing UI text. New widgets (text input, dropdown) need to be implemented since none exist:

- **TextInput widget**: Captures keyboard events, renders cursor, handles backspace/delete. Approx 80-100 LOC.
- **Dropdown widget**: Renders current value; on click, shows option list; picks on click. Approx 60-80 LOC.
- **PalettePicker widget**: Renders N colored cells horizontally; tracks selection. Approx 40-60 LOC.

### 4c. Portrait resolution at runtime

When displaying the character portrait, the UI resolves the filename:

```go
func resolvePortraitFile(class, race string, appearance Appearance) string {
    expr := appearance.PortraitTag()       // "a", "b", or "nb"
    tone := SkinToneGroup(appearance.SkinTone) // "light", "medium", "dark"
    file := fmt.Sprintf("portrait_%s_%s_%s_%s.png", class, race, expr, tone)
    // TODO: fallback to legacy filename if file not found
    return file
}
```

### Estimated effort

| Item | LOC |
|------|-----|
| `types.go` changes | ~15 |
| `game.go` routing | ~10 |
| `character_creation_screen.go` (new) | ~300 |
| `ui_widgets.go` (new) | ~200 |
| **Subtotal** | **~525** |

---

## Phase 5 — Tests

### Files to create / modify

| File | Action |
|------|--------|
| `pkg/game/appearance_test.go` *(new)* | Unit tests for `Appearance` struct |
| `pkg/game/character_creation_test.go` | Add test cases for appearance in creation flow |
| `pkg/validation/validation_test.go` | Add cases for appearance field validation |
| `pkg/server/handlers_test.go` | Add RPC round-trip test with appearance fields |

### 5a. `appearance_test.go` — YAML round-trip & helpers

```go
func TestAppearance_YAMLRoundTrip(t *testing.T) {
    tests := []struct {
        name string
        input Appearance
    }{
        {"zero value", Appearance{}},
        {"full fields", Appearance{
            SkinTone: 7, HairStyle: "long braids", HairColor: "auburn",
            BodyType: BodyMuscular, GenderExpression: "feminine",
            Pronouns: "she/her", RomanticOrientation: "bisexual",
        }},
        {"minimal", Appearance{SkinTone: 1}},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            data, err := yaml.Marshal(tt.input)
            require.NoError(t, err)
            var got Appearance
            require.NoError(t, yaml.Unmarshal(data, &got))
            assert.Equal(t, tt.input, got)
        })
    }
}

func TestSkinToneGroup(t *testing.T) {
    tests := []struct{ tone int; want string }{
        {0, "medium"}, {1, "light"}, {3, "light"},
        {4, "medium"}, {7, "medium"}, {8, "dark"}, {10, "dark"},
    }
    for _, tt := range tests {
        assert.Equal(t, tt.want, SkinToneGroup(tt.tone))
    }
}

func TestSkinToneName(t *testing.T) {
    assert.Equal(t, "porcelain", SkinToneName(1))
    assert.Equal(t, "obsidian", SkinToneName(10))
    assert.Equal(t, "unknown", SkinToneName(0))
    assert.Equal(t, "unknown", SkinToneName(11))
}

func TestAppearance_PortraitTag(t *testing.T) {
    tests := []struct{ expr, want string }{
        {"masculine", "a"}, {"male", "a"},
        {"feminine", "b"}, {"female", "b"},
        {"non-binary", "nb"}, {"", "nb"}, {"genderqueer", "nb"},
    }
    for _, tt := range tests {
        a := Appearance{GenderExpression: tt.expr}
        assert.Equal(t, tt.want, a.PortraitTag())
    }
}
```

### 5b. Backward-compatibility creation tests

Add to existing table in `character_creation_test.go`:

```go
{
    name: "creation without appearance fields (backward compat)",
    config: CharacterCreationConfig{
        Name: "OldSchool", Class: ClassFighter,
        AttributeMethod: "standard", StartingEquipment: true,
    },
    expectSuccess: true,
    validate: func(t *testing.T, result CharacterCreationResult) {
        // Appearance should have been filled with defaults
        assert.Equal(t, 5, result.Character.Appearance.SkinTone)
        assert.Equal(t, BodyAverage, result.Character.Appearance.BodyType)
    },
},
{
    name: "creation with full appearance",
    config: CharacterCreationConfig{
        Name: "NewStyle", Class: ClassMage,
        AttributeMethod: "standard", StartingEquipment: true,
        Appearance: Appearance{
            SkinTone: 9, HairStyle: "mohawk", HairColor: "teal",
            BodyType: BodyLarge, GenderExpression: "non-binary",
            Pronouns: "they/them",
        },
    },
    expectSuccess: true,
    validate: func(t *testing.T, result CharacterCreationResult) {
        assert.Equal(t, 9, result.Character.Appearance.SkinTone)
        assert.Equal(t, "mohawk", result.Character.Appearance.HairStyle)
        assert.Equal(t, "they/them", result.Character.Appearance.Pronouns)
    },
},
```

### 5c. Validation tests

Add to validation test table:

```go
{name: "skin_tone valid",    params: map[string]interface{}{"name":"X","class":"fighter","skin_tone":5.0},  wantErr: false},
{name: "skin_tone too high", params: map[string]interface{}{"name":"X","class":"fighter","skin_tone":11.0}, wantErr: true},
{name: "skin_tone too low",  params: map[string]interface{}{"name":"X","class":"fighter","skin_tone":0.0},  wantErr: true},
{name: "body_type valid",    params: map[string]interface{}{"name":"X","class":"fighter","body_type":3.0},  wantErr: false},
{name: "body_type invalid",  params: map[string]interface{}{"name":"X","class":"fighter","body_type":9.0},  wantErr: true},
{name: "pronouns too long",  params: map[string]interface{}{"name":"X","class":"fighter","pronouns":strings.Repeat("a",101)}, wantErr: true},
{name: "pronouns valid",     params: map[string]interface{}{"name":"X","class":"fighter","pronouns":"xe/xem"}, wantErr: false},
```

### 5d. Asset filename resolution test

```go
func TestResolvePortraitFilename(t *testing.T) {
    tests := []struct {
        class, race string
        appearance  Appearance
        want        string
    }{
        {"fighter", "human", Appearance{GenderExpression: "masculine", SkinTone: 2}, "portrait_fighter_human_a_light.png"},
        {"mage", "elf", Appearance{GenderExpression: "feminine", SkinTone: 6}, "portrait_mage_elf_b_medium.png"},
        {"cleric", "dwarf", Appearance{GenderExpression: "non-binary", SkinTone: 9}, "portrait_cleric_dwarf_nb_dark.png"},
        {"thief", "halfling", Appearance{}, "portrait_thief_halfling_nb_medium.png"}, // defaults
    }
    for _, tt := range tests {
        got := resolvePortraitFile(tt.class, tt.race, tt.appearance)
        assert.Equal(t, tt.want, got)
    }
}
```

### Estimated effort

| Item | LOC |
|------|-----|
| `appearance_test.go` (new) | ~120 |
| `character_creation_test.go` additions | ~50 |
| `validation_test.go` additions | ~30 |
| `handlers_test.go` additions | ~40 |
| **Subtotal** | **~240** |

---

## Asset Pipeline Delta

### Current portrait matrix

| Axis | Values | Count |
|------|--------|-------|
| Class | fighter, mage, cleric, thief, ranger, paladin | 6 |
| Race | human, elf, dwarf, halfling | 4 |
| Gender | male, female | 2 |
| **Total portraits** | | **48** |

### New portrait matrix

| Axis | Values | Count |
|------|--------|-------|
| Class | fighter, mage, cleric, thief, ranger, paladin | 6 |
| Race | human, elf, dwarf, halfling | 4 |
| Expression | a, b, nb | 3 |
| Skin tone group | light, medium, dark | 3 |
| **Total portraits** | | **216** |

### Additional body-type sprites

| Axis | Values | Count |
|------|--------|-------|
| Class | 6 classes | 6 |
| Body type | slim, average, muscular, stocky, large | 5 |
| **Total sprites** | | **30** |

### Overall asset count change

| Category | Current | New | Delta |
|----------|---------|-----|-------|
| Character portraits | 48 | 216 | +168 |
| Character sprites (body) | 0 | 30 | +30 |
| Monsters | 35 | 35 | 0 |
| Items | 60 | 60 | 0 |
| Terrain | 80 | 80 | 0 |
| Effects | 50 | 50 | 0 |
| UI | 100 | 100 | 0 |
| **Total** | **~248** | **~446** | **+198** |

(Counts are based on `game-assets.yaml` entries counted via `grep -c 'filename:'`; the README/docs reference "521 defined" which likely includes `assets/characters.yaml` and other config. The delta is additive regardless.)

### Generation time estimate

Current full generation: 4-6 hours for ~248 assets.
New full generation: ~7-10 hours for ~446 assets (portrait generation is the bottleneck).
Priority generation (`assets-priority`): unchanged — generates only one variant per class for quick testing.

---

## Migration Notes

### Saved characters

- Existing `Character` YAML files lack the `char_appearance` key.
- Go's `yaml.v3` deserializer produces a zero-value `Appearance{}` for missing keys.
- `buildBaseCharacter()` detects zero-value `Appearance` and applies `DefaultAppearance()`.
- **No migration script needed.** Existing saves load and play without modification. The first re-save writes the defaulted appearance.

### RPC callers

- All new JSON fields use `omitempty`. Existing clients omitting them get the same behavior as today.
- No breaking change to request or response schemas.

### Asset filenames

- **Legacy filenames are retained.** The existing 48 portraits remain in place; they map to the `{expression=a|b, tone=medium}` variants.
- A one-time `create-legacy-symlinks.sh` creates symlinks:
  - `portrait_fighter_human_male.png` → `portrait_fighter_human_a_medium.png`
  - `portrait_fighter_human_female.png` → `portrait_fighter_human_b_medium.png`
- Any code currently loading `portrait_{class}_{race}_{male|female}.png` continues to work.

### Database / persistence

- The `pkg/persistence/` package serializes `Character` via YAML. The new `Appearance` field serializes normally; old records without it deserialize to zero value.
- No schema migration required.

---

## Open Questions

1. **Race field in Go code**: The asset pipeline uses race (`human`, `elf`, `dwarf`, `halfling`) but no `CharacterRace` type exists in Go. Should this plan also add a `Race` field to `Character`, or should portrait lookup use `AdditionalData["race"]`?

2. **Prompt composition engine**: The fragment library (`appearance-fragments.yaml`) uses `${variable}` substitution. Should `generate-all.sh` handle this with `envsubst`/`sed`, or should a Go tool (e.g., `cmd/asset-gen/`) do the template expansion?

3. **Non-binary portrait art direction**: The `nb` expression tag uses "androgynous, gender-neutral" as the prompt modifier. Should there be more than one `nb` variant (e.g., multiple androgynous presentations), or is one per class/race/tone sufficient for the initial pass?

4. **Hair style/color in portraits**: Should hair attributes influence portrait generation prompts at all? This would further multiply the portrait matrix. Alternatively, hair could be a post-processing overlay or deferred to a later phase.

5. **Custom pronouns storage format**: Should `Pronouns` be a single string (`"they/them"`) or a structured type (`{subject: "they", object: "them", possessive: "their"}`)? A structured type enables correct pronoun insertion in game text but adds complexity.

6. **Romantic orientation visibility**: This field is stored but never displayed in-game. Should it appear on the character sheet UI, or remain purely data (e.g., for future narrative systems)?

7. **Full 10-tone portraits vs. 3-group approach**: The plan uses 3 skin-tone groups (light/medium/dark) for the asset matrix to keep generation tractable. Should a future phase generate all 10 tonal variants, or is runtime color-shifting (hue/brightness post-processing) a better approach for finer granularity?

---

## Implementation Order & Dependencies

```
Phase 1 ──► Phase 2 ──► Phase 5a/5b/5c (unit tests)
                │
                ▼
Phase 3 (asset pipeline — can run in parallel with Phase 2 tests)
                │
                ▼
Phase 4 (frontend — depends on Phase 1 types + Phase 3 filenames)
                │
                ▼
Phase 5d (integration / asset-resolution tests)
```

**Total estimated effort**: ~2,090 lines across Go, YAML, Bash, and tests.
