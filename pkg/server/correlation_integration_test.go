package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TestCorrelationIDIntegration tests the complete correlation ID system
func TestCorrelationIDIntegration(t *testing.T) {
	// Set up log capture to verify correlation IDs in logs
	var logBuffer strings.Builder
	originalOutput := logrus.StandardLogger().Out
	logrus.SetOutput(&logBuffer)
	defer logrus.SetOutput(originalOutput)
	logrus.SetLevel(logrus.DebugLevel)

	// Create a test server
	srv, err := NewRPCServer("./test_web")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	tests := []struct {
		name              string
		existingRequestID string
		expectSameID      bool
		method            string
		jsonRPCPayload    string
	}{
		{
			name:              "auto-generated correlation ID",
			existingRequestID: "",
			expectSameID:      false,
			method:            "POST",
			jsonRPCPayload:    `{"jsonrpc":"2.0","method":"getGameState","params":{},"id":1}`,
		},
		{
			name:              "preserve existing correlation ID",
			existingRequestID: "user-provided-id-123",
			expectSameID:      true,
			method:            "POST",
			jsonRPCPayload:    `{"jsonrpc":"2.0","method":"getGameState","params":{},"id":2}`,
		},
		{
			name:              "correlation ID in GET requests",
			existingRequestID: "",
			expectSameID:      false,
			method:            "GET",
			jsonRPCPayload:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear log buffer for this test
			logBuffer.Reset()

			// Create request
			var req *http.Request
			if tt.method == "POST" {
				req = httptest.NewRequest(tt.method, "/", strings.NewReader(tt.jsonRPCPayload))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, "/", nil)
			}

			if tt.existingRequestID != "" {
				req.Header.Set("X-Request-ID", tt.existingRequestID)
			}

			w := httptest.NewRecorder()

			// Apply the full middleware chain that the server uses
			handler := RequestIDMiddleware(
				LoggingMiddleware(
					srv.withRecovery(
						srv.withTimeout(srv.config.RequestTimeout)(srv))))

			// Execute request
			handler.ServeHTTP(w, req)

			// Verify X-Request-ID header is set in response
			responseID := w.Header().Get("X-Request-ID")
			if responseID == "" {
				t.Error("X-Request-ID header not set in response")
			}

			if tt.expectSameID {
				if responseID != tt.existingRequestID {
					t.Errorf("Expected request ID %s, got %s", tt.existingRequestID, responseID)
				}
			} else {
				// Should be a valid UUID
				if _, err := uuid.Parse(responseID); err != nil {
					t.Errorf("Response ID is not a valid UUID: %s", responseID)
				}
			}

			// Verify logs contain the correlation ID
			logOutput := logBuffer.String()
			if !strings.Contains(logOutput, responseID) {
				t.Errorf("Log output does not contain request ID %s. Log: %s", responseID, logOutput)
			}

			// For POST requests, verify they're processed correctly
			if tt.method == "POST" {
				// Should get a JSON-RPC response
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse JSON response: %v", err)
				}

				// Verify response structure
				if response["jsonrpc"] != "2.0" {
					t.Error("Invalid JSON-RPC response format")
				}
			}
		})
	}
}

// TestCorrelationIDPropagation tests that correlation IDs are properly propagated through the request lifecycle
func TestCorrelationIDPropagation(t *testing.T) {
	// Create a test server
	srv, err := NewRPCServer("./test_web")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	// Capture the request ID within a handler
	var capturedRequestID string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRequestID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	// Apply middleware chain
	handler := RequestIDMiddleware(testHandler)

	// Test with custom request ID
	customID := "test-correlation-123"
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", customID)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify propagation
	if capturedRequestID != customID {
		t.Errorf("Request ID not properly propagated. Expected %s, got %s", customID, capturedRequestID)
	}

	// Verify response header
	responseID := w.Header().Get("X-Request-ID")
	if responseID != customID {
		t.Errorf("Response header doesn't match. Expected %s, got %s", customID, responseID)
	}
}

// TestCorrelationIDWithWebSocket tests correlation ID handling for WebSocket upgrades
func TestCorrelationIDWithWebSocket(t *testing.T) {
	// Set up log capture
	var logBuffer strings.Builder
	originalOutput := logrus.StandardLogger().Out
	logrus.SetOutput(&logBuffer)
	defer logrus.SetOutput(originalOutput)
	logrus.SetLevel(logrus.DebugLevel)

	// Create a test server
	srv, err := NewRPCServer("./test_web")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	// Create a request that looks like a WebSocket upgrade
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "test-key")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("X-Request-ID", "websocket-test-123")

	w := httptest.NewRecorder()

	// Apply middleware chain
	handler := RequestIDMiddleware(
		LoggingMiddleware(
			srv.withRecovery(
				srv.withTimeout(srv.config.RequestTimeout)(srv))))

	handler.ServeHTTP(w, req)

	// Verify request ID is in response headers
	responseID := w.Header().Get("X-Request-ID")
	if responseID != "websocket-test-123" {
		t.Errorf("WebSocket request ID not preserved. Expected websocket-test-123, got %s", responseID)
	}

	// Verify logs contain the correlation ID
	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, "websocket-test-123") {
		t.Error("WebSocket request ID not found in logs")
	}
}

// TestCorrelationIDWithError tests correlation ID handling during error conditions
func TestCorrelationIDWithError(t *testing.T) {
	// Set up log capture
	var logBuffer strings.Builder
	originalOutput := logrus.StandardLogger().Out
	logrus.SetOutput(&logBuffer)
	defer logrus.SetOutput(originalOutput)
	logrus.SetLevel(logrus.ErrorLevel)

	// Create handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic for correlation ID")
	})

	// Apply middleware chain
	handler := RequestIDMiddleware(
		LoggingMiddleware(
			RecoveryMiddleware(panicHandler)))

	// Create request with custom ID
	req := httptest.NewRequest("GET", "/panic", nil)
	req.Header.Set("X-Request-ID", "panic-test-456")
	w := httptest.NewRecorder()

	// Execute request (should not panic)
	handler.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	// Verify request ID is preserved in response
	responseID := w.Header().Get("X-Request-ID")
	if responseID != "panic-test-456" {
		t.Errorf("Request ID not preserved during panic. Expected panic-test-456, got %s", responseID)
	}

	// Verify panic log contains correlation ID
	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, "panic-test-456") {
		t.Error("Panic log does not contain correlation ID")
	}
	if !strings.Contains(logOutput, "recovered from panic") {
		t.Error("Panic recovery log message not found")
	}
}

// TestRequestIDUtilityFunctions tests the helper functions work correctly
func TestRequestIDUtilityFunctions(t *testing.T) {
	tests := []struct {
		name     string
		headerID string
	}{
		{
			name:     "with custom header ID",
			headerID: "custom-test-789",
		},
		{
			name:     "with auto-generated ID",
			headerID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test handler that uses utility functions
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestID := GetRequestID(r.Context())
				if requestID == "" {
					t.Error("GetRequestID returned empty string")
				}

				// Verify it matches the expected ID or is a valid UUID
				if tt.headerID != "" {
					if requestID != tt.headerID {
						t.Errorf("GetRequestID returned %s, expected %s", requestID, tt.headerID)
					}
				} else {
					if _, err := uuid.Parse(requestID); err != nil {
						t.Errorf("GetRequestID returned invalid UUID: %s", requestID)
					}
				}

				w.WriteHeader(http.StatusOK)
			})

			// Apply middleware
			handler := RequestIDMiddleware(testHandler)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.headerID != "" {
				req.Header.Set("X-Request-ID", tt.headerID)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		})
	}
}
