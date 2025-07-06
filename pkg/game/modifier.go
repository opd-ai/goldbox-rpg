package game

// Modifier represents a stat modification that can be applied to game entities.
// It defines how a specific attribute should be modified, including the target stat,
// the value to apply, and the mathematical operation to perform.
//
// Fields:
//   - Stat: String identifier of the stat to modify (e.g. "strength", "health")
//   - Value: Numeric value to apply in the modification
//   - Operation: Type of mathematical operation (add, multiply, set)
//
// Related types:
//   - ModOpType: Enumeration of supported modification operations
//   - Effect: Contains a list of Modifier objects
//
// Example usage:
//
//	mod := Modifier{
//	    Stat:      "strength",
//	    Value:     5,
//	    Operation: ModAdd,
//	}
//
// Moved from: effects.go
type Modifier struct {
	Stat      string    `yaml:"mod_stat"`
	Value     float64   `yaml:"mod_value"`
	Operation ModOpType `yaml:"mod_operation"`
}

// NewModifier creates a new Modifier with the specified parameters.
// Moved from: effects.go
func NewModifier(stat string, value float64, operation ModOpType) *Modifier {
	return &Modifier{
		Stat:      stat,
		Value:     value,
		Operation: operation,
	}
}

// Apply applies the modifier to a given base value and returns the result.
// The operation type determines how the modification is performed.
// Moved from: effects.go
func (m *Modifier) Apply(baseValue float64) float64 {
	switch m.Operation {
	case ModAdd:
		return baseValue + m.Value
	case ModMultiply:
		return baseValue * m.Value
	case ModSet:
		return m.Value
	default:
		return baseValue
	}
}
