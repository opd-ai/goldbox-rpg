// Package game provides helper functions for faction diplomacy operations.
package game

// withRelationLock retrieves a relation with write locks on both the manager
// and relation mutexes. The callback receives the relation for modification.
// Locks are held for the duration of the callback and released after.
// Returns ErrRelationNotFound if the relation does not exist.
func (dm *DiplomacyManager) withRelationLock(faction1ID, faction2ID string, fn func(*FactionRelation) error) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	key := relationKey(faction1ID, faction2ID)
	rel, exists := dm.relations[key]
	if !exists {
		return ErrRelationNotFound
	}

	rel.mu.Lock()
	defer rel.mu.Unlock()

	return fn(rel)
}

// withRelationRLock retrieves a relation with read locks on both the manager
// and relation mutexes. The callback receives the relation for reading.
// Locks are held for the duration of the callback and released after.
// Returns ErrRelationNotFound if the relation does not exist.
func (dm *DiplomacyManager) withRelationRLock(faction1ID, faction2ID string, fn func(*FactionRelation) error) error {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	key := relationKey(faction1ID, faction2ID)
	rel, exists := dm.relations[key]
	if !exists {
		return ErrRelationNotFound
	}

	rel.mu.RLock()
	defer rel.mu.RUnlock()

	return fn(rel)
}

// readRelationState retrieves a relation's state using read locks.
// This is a convenience wrapper for simple state checks.
// Returns the zero value of T if the relation does not exist.
func readRelationState[T any](dm *DiplomacyManager, faction1ID, faction2ID string, fn func(*FactionRelation) T) (T, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	key := relationKey(faction1ID, faction2ID)
	rel, exists := dm.relations[key]
	if !exists {
		var zero T
		return zero, false
	}

	rel.mu.RLock()
	defer rel.mu.RUnlock()

	return fn(rel), true
}
