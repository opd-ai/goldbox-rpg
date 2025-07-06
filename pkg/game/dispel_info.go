package game

// DispelInfo contains metadata about how a game effect can be dispelled or removed.
//
// Fields:
//   - Priority: Determines the order in which effects are dispelled (higher priority = dispelled first)
//   - Types: List of dispel types that can remove this effect (e.g. magic, poison, curse)
//   - Removable: Whether the effect can be removed at all
//
// Related types:
//   - DispelPriority: Priority level constants (0-100)
//   - DispelType: Type of dispel (magic, curse, poison, etc)
//   - Effect: Contains DispelInfo as a field
//
// Example usage:
//
//	info := DispelInfo{
//	    Priority: DispelPriorityNormal,
//	    Types: []DispelType{DispelMagic},
//	    Removable: true,
//	}
//
// Moved from: effects.go
type DispelInfo struct {
	Priority  DispelPriority `yaml:"dispel_priority"`
	Types     []DispelType   `yaml:"dispel_types"`
	Removable bool           `yaml:"dispel_removable"`
}

// NewDispelInfo creates a new DispelInfo with the specified parameters.
// Moved from: effects.go
func NewDispelInfo(priority DispelPriority, types []DispelType, removable bool) *DispelInfo {
	return &DispelInfo{
		Priority:  priority,
		Types:     types,
		Removable: removable,
	}
}

// CanBeDispelledBy checks if this effect can be removed by the given dispel type.
// Moved from: effects.go
func (di *DispelInfo) CanBeDispelledBy(dispelType DispelType) bool {
	if !di.Removable {
		return false
	}

	for _, t := range di.Types {
		if t == dispelType || t == DispelAll {
			return true
		}
	}

	return dispelType == DispelAll
}
