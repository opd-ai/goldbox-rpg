package persistence

import "time"

// makeTestSessions returns a slice of test session data for use in tests.
// This consolidates the duplicated session data creation across test files.
func makeTestSessions() []*SessionData {
	return []*SessionData{
		{
			SessionID:  "session-1",
			PlayerID:   "player-1",
			PlayerName: "Player1",
			LastActive: time.Now(),
			CreatedAt:  time.Now(),
			Connected:  true,
		},
		{
			SessionID:  "session-2",
			PlayerID:   "player-2",
			PlayerName: "Player2",
			LastActive: time.Now(),
			CreatedAt:  time.Now(),
			Connected:  false,
		},
		{
			SessionID:  "session-3",
			PlayerID:   "player-3",
			PlayerName: "Player3",
			LastActive: time.Now(),
			CreatedAt:  time.Now(),
			Connected:  true,
		},
	}
}

// expectedSessionIDs returns the expected session IDs map for testing.
func expectedSessionIDs() map[string]bool {
	return map[string]bool{
		"session-1": true,
		"session-2": true,
		"session-3": true,
	}
}
