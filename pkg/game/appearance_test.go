package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestAppearance_YAMLRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input Appearance
	}{
		{"zero value", Appearance{}},
		{"full fields", Appearance{
			SkinTone:            7,
			HairStyle:           "long braids",
			HairColor:           "auburn",
			BodyType:            BodyMuscular,
			GenderExpression:    "feminine",
			Pronouns:            "she/her",
			RomanticOrientation: "bisexual",
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
	tests := []struct {
		tone int
		want string
	}{
		{0, "medium"},
		{1, "light"},
		{3, "light"},
		{4, "medium"},
		{7, "medium"},
		{8, "dark"},
		{10, "dark"},
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
	tests := []struct {
		expr string
		want string
	}{
		{"masculine", "a"},
		{"male", "a"},
		{"feminine", "b"},
		{"female", "b"},
		{"non-binary", "nb"},
		{"", "nb"},
		{"genderqueer", "nb"},
	}
	for _, tt := range tests {
		a := Appearance{GenderExpression: tt.expr}
		assert.Equal(t, tt.want, a.PortraitTag())
	}
}

func TestDefaultAppearance(t *testing.T) {
	d := DefaultAppearance()
	assert.Equal(t, 5, d.SkinTone)
	assert.Equal(t, BodyAverage, d.BodyType)
	assert.Empty(t, d.HairStyle)
	assert.Empty(t, d.Pronouns)
}

func TestSkinTonePalette_Coverage(t *testing.T) {
	for i := 1; i <= 10; i++ {
		name := SkinToneName(i)
		assert.NotEqual(t, "unknown", name)
	}
}

func TestResolvePortraitFile(t *testing.T) {
	tests := []struct {
		class, race string
		appearance  Appearance
		want        string
	}{
		{
			"fighter", "human",
			Appearance{GenderExpression: "masculine", SkinTone: 2},
			"portrait_fighter_human_a_light.png",
		},
		{
			"mage", "elf",
			Appearance{GenderExpression: "feminine", SkinTone: 6},
			"portrait_mage_elf_b_medium.png",
		},
		{
			"cleric", "dwarf",
			Appearance{GenderExpression: "non-binary", SkinTone: 9},
			"portrait_cleric_dwarf_nb_dark.png",
		},
		{
			"thief", "halfling",
			Appearance{},
			"portrait_thief_halfling_nb_medium.png",
		},
	}
	for _, tt := range tests {
		got := ResolvePortraitFile(tt.class, tt.race, tt.appearance)
		assert.Equal(t, tt.want, got)
	}
}
