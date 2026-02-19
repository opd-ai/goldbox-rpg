// Package server tests for spell handler coverage
package server

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleGetSpell tests the handleGetSpell handler
func TestHandleGetSpell(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		checkResult func(t *testing.T, result interface{})
	}{
		{
			name: "valid spell_id returns spell",
			params: map[string]interface{}{
				"spell_id": "magic_missile",
			},
			expectError: false,
			checkResult: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, resultMap["success"])
				assert.NotNil(t, resultMap["spell"])
			},
		},
		{
			name: "empty spell_id returns error",
			params: map[string]interface{}{
				"spell_id": "",
			},
			expectError: true,
		},
		{
			name: "nonexistent spell_id returns error",
			params: map[string]interface{}{
				"spell_id": "nonexistent_spell",
			},
			expectError: true,
		},
		{
			name:        "invalid params returns error",
			params:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServerForHandlers(t)
			var paramBytes []byte
			if tt.params != nil {
				paramBytes, _ = json.Marshal(tt.params)
			} else {
				paramBytes = []byte("invalid json{")
			}

			result, err := server.handleGetSpell(paramBytes)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}

// TestHandleGetSpellsByLevel tests the handleGetSpellsByLevel handler
func TestHandleGetSpellsByLevel(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		checkResult func(t *testing.T, result interface{})
	}{
		{
			name: "valid level returns spells",
			params: map[string]interface{}{
				"level": 0,
			},
			expectError: false,
			checkResult: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, resultMap["success"])
				assert.NotNil(t, resultMap["spells"])
				assert.NotNil(t, resultMap["count"])
			},
		},
		{
			name: "negative level returns error",
			params: map[string]interface{}{
				"level": -1,
			},
			expectError: true,
		},
		{
			name:        "invalid params returns error",
			params:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServerForHandlers(t)
			var paramBytes []byte
			if tt.params != nil {
				paramBytes, _ = json.Marshal(tt.params)
			} else {
				paramBytes = []byte("invalid json{")
			}

			result, err := server.handleGetSpellsByLevel(paramBytes)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}

// TestHandleGetSpellsBySchool tests the handleGetSpellsBySchool handler
func TestHandleGetSpellsBySchool(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		checkResult func(t *testing.T, result interface{})
	}{
		{
			name: "valid school returns spells",
			params: map[string]interface{}{
				"school": "Evocation",
			},
			expectError: false,
			checkResult: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, resultMap["success"])
				assert.NotNil(t, resultMap["spells"])
			},
		},
		{
			name: "empty school returns error",
			params: map[string]interface{}{
				"school": "",
			},
			expectError: true,
		},
		{
			name:        "invalid params returns error",
			params:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServerForHandlers(t)
			var paramBytes []byte
			if tt.params != nil {
				paramBytes, _ = json.Marshal(tt.params)
			} else {
				paramBytes = []byte("invalid json{")
			}

			result, err := server.handleGetSpellsBySchool(paramBytes)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}

// TestHandleGetAllSpells tests the handleGetAllSpells handler
func TestHandleGetAllSpells(t *testing.T) {
	server := createTestServerForHandlers(t)

	// handleGetAllSpells doesn't require any params, but accepts empty params
	params := []byte("{}")

	result, err := server.handleGetAllSpells(params)

	assert.NoError(t, err)
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, resultMap["success"])
	assert.NotNil(t, resultMap["spells"])
	assert.NotNil(t, resultMap["count"])
	assert.NotNil(t, resultMap["by_level"])
}

// TestHandleSearchSpells tests the handleSearchSpells handler
func TestHandleSearchSpells(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		checkResult func(t *testing.T, result interface{})
	}{
		{
			name: "valid query returns spells",
			params: map[string]interface{}{
				"query": "fire",
			},
			expectError: false,
			checkResult: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, resultMap["success"])
				assert.NotNil(t, resultMap["spells"])
			},
		},
		{
			name: "empty query returns error",
			params: map[string]interface{}{
				"query": "",
			},
			expectError: true,
		},
		{
			name:        "invalid params returns error",
			params:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServerForHandlers(t)
			var paramBytes []byte
			if tt.params != nil {
				paramBytes, _ = json.Marshal(tt.params)
			} else {
				paramBytes = []byte("invalid json{")
			}

			result, err := server.handleSearchSpells(paramBytes)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkResult != nil {
					tt.checkResult(t, result)
				}
			}
		})
	}
}
