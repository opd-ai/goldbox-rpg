package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleSessionRequest_GetSessionID(t *testing.T) {
	tests := []struct {
		name     string
		request  simpleSessionRequest
		expected string
	}{
		{
			name:     "valid session ID",
			request:  simpleSessionRequest{SessionID: "test-session-123"},
			expected: "test-session-123",
		},
		{
			name:     "empty session ID",
			request:  simpleSessionRequest{SessionID: ""},
			expected: "",
		},
		{
			name:     "UUID format session ID",
			request:  simpleSessionRequest{SessionID: "550e8400-e29b-41d4-a716-446655440000"},
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.GetSessionID()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGuildSuccessMsg(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected map[string]interface{}
	}{
		{
			name:    "standard success message",
			message: "Guild created successfully",
			expected: map[string]interface{}{
				"success": true,
				"message": "Guild created successfully",
			},
		},
		{
			name:    "empty message",
			message: "",
			expected: map[string]interface{}{
				"success": true,
				"message": "",
			},
		},
		{
			name:    "long message",
			message: "This is a very long success message that explains the result of a complex guild operation in detail",
			expected: map[string]interface{}{
				"success": true,
				"message": "This is a very long success message that explains the result of a complex guild operation in detail",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := guildSuccessMsg(tt.message)
			assert.Equal(t, tt.expected["success"], result["success"])
			assert.Equal(t, tt.expected["message"], result["message"])
		})
	}
}
