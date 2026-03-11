package game

import (
"testing"
)

// BenchmarkCharacter_Clone benchmarks deep cloning of a character
func BenchmarkCharacter_Clone(b *testing.B) {
creator := NewCharacterCreator()
config := CharacterCreationConfig{
Name:              "BenchChar",
Class:             ClassFighter,
AttributeMethod:   "standard",
StartingEquipment: true,
}
result := creator.CreateCharacter(config)
char := result.Character

b.ResetTimer()
for i := 0; i < b.N; i++ {
_ = char.Clone()
}
}

// BenchmarkCharacter_GetPosition benchmarks position retrieval
func BenchmarkCharacter_GetPosition(b *testing.B) {
creator := NewCharacterCreator()
config := CharacterCreationConfig{
Name:            "BenchChar",
Class:           ClassFighter,
AttributeMethod: "standard",
}
result := creator.CreateCharacter(config)
char := result.Character

b.ResetTimer()
for i := 0; i < b.N; i++ {
_ = char.GetPosition()
}
}

// BenchmarkCharacter_ToJSON benchmarks character serialization
func BenchmarkCharacter_ToJSON(b *testing.B) {
creator := NewCharacterCreator()
config := CharacterCreationConfig{
Name:              "BenchChar",
Class:             ClassFighter,
AttributeMethod:   "standard",
StartingEquipment: true,
}
result := creator.CreateCharacter(config)
char := result.Character

b.ResetTimer()
for i := 0; i < b.N; i++ {
_, _ = char.ToJSON()
}
}
