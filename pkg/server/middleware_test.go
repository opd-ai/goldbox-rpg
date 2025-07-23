package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TestRequestIDMiddleware tests the RequestIDMiddleware functionality
func TestRequestIDMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		existingHeader string
		expectNewID    bool
	}{
		{
			name:           "generates new ID when header is missing",
			existingHeader: "",
			expectNewID:    true,
		},
		{
			name:           "uses existing ID when header is present",
			existingHeader: "test-request-id-123",
			expectNewID:    false,
		},
		{
			name:           "generates new ID when header is empty",
			existingHeader: "",
			expectNewID:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that captures the request context
			var capturedRequestID string
			var capturedLogger *logrus.Entry
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequestID = GetRequestID(r.Context())
				if logger, ok := r.Context().Value("logger").(*logrus.Entry); ok {
					capturedLogger = logger
				}
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with RequestIDMiddleware
			middleware := RequestIDMiddleware(testHandler)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.existingHeader != "" {
				req.Header.Set("X-Request-ID", tt.existingHeader)
			}
			w := httptest.NewRecorder()

			// Execute request
			middleware.ServeHTTP(w, req)

			// Verify response header is set
			responseID := w.Header().Get("X-Request-ID")
			if responseID == "" {
				t.Error("X-Request-ID header not set in response")
			}

			// Verify context contains request ID
			if capturedRequestID == "" {
				t.Error("Request ID not found in context")
			}

			if tt.expectNewID {
				// Should generate a valid UUID
				if _, err := uuid.Parse(capturedRequestID); err != nil {
					t.Errorf("Generated request ID is not a valid UUID: %s", capturedRequestID)
				}
				// Response header should match context value
				if responseID != capturedRequestID {
					t.Errorf("Response header (%s) doesn't match context value (%s)", responseID, capturedRequestID)
				}
			} else {
				// Should use existing header value
				if capturedRequestID != tt.existingHeader {
					t.Errorf("Expected request ID %s, got %s", tt.existingHeader, capturedRequestID)
				}
				if responseID != tt.existingHeader {
					t.Errorf("Expected response header %s, got %s", tt.existingHeader, responseID)
				}
			}

			// Verify logger is available in context
			if capturedLogger == nil {
				t.Error("Logger not found in context")
			} else {
				// Check that logger has the request_id field
				entry := capturedLogger.WithField("test", "value")
				if entry.Data["request_id"] != capturedRequestID {
					t.Errorf("Logger doesn't contain correct request_id field. Expected %s, got %v",
						capturedRequestID, entry.Data["request_id"])
				}
			}
		})
	}
}

// TestLoggingMiddleware tests the LoggingMiddleware functionality
func TestLoggingMiddleware(t *testing.T) {
	// Set up a buffer to capture log output
	var logBuffer strings.Builder
	originalOutput := logrus.StandardLogger().Out
	logrus.SetOutput(&logBuffer)
	defer logrus.SetOutput(originalOutput)
	logrus.SetLevel(logrus.DebugLevel)

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware chain: RequestID -> Logging -> Handler
	chain := RequestIDMiddleware(LoggingMiddleware(testHandler))

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Execute request
	chain.ServeHTTP(w, req)

	// Verify log output contains expected fields
	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, "processing request") {
		t.Error("Debug log message 'processing request' not found")
	}
	if !strings.Contains(logOutput, "request completed") {
		t.Error("Info log message 'request completed' not found")
	}
	if !strings.Contains(logOutput, "request_id") {
		t.Error("request_id field not found in log output")
	}
	if !strings.Contains(logOutput, "status_code") {
		t.Error("status_code field not found in log output")
	}
}

// TestRecoveryMiddleware tests the RecoveryMiddleware panic handling
func TestRecoveryMiddleware(t *testing.T) {
	// Set up a buffer to capture log output
	var logBuffer strings.Builder
	originalOutput := logrus.StandardLogger().Out
	logrus.SetOutput(&logBuffer)
	defer logrus.SetOutput(originalOutput)
	logrus.SetLevel(logrus.ErrorLevel)

	// Create test handler that panics
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Create middleware chain: RequestID -> Recovery -> Handler
	chain := RequestIDMiddleware(RecoveryMiddleware(testHandler))

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Execute request (should not panic)
	chain.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	// Verify log output contains panic information
	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, "recovered from panic") {
		t.Error("Panic recovery log message not found")
	}
	if !strings.Contains(logOutput, "test panic") {
		t.Error("Panic message not found in log output")
	}
	if !strings.Contains(logOutput, "request_id") {
		t.Error("request_id field not found in panic log")
	}
}

// TestCORSMiddleware tests the CORS middleware functionality
func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		expectAllowed  bool
		isOptions      bool
	}{
		{
			name:           "wildcard allows all origins",
			allowedOrigins: []string{"*"},
			requestOrigin:  "https://example.com",
			expectAllowed:  true,
			isOptions:      false,
		},
		{
			name:           "specific origin allowed",
			allowedOrigins: []string{"https://example.com", "https://test.com"},
			requestOrigin:  "https://example.com",
			expectAllowed:  true,
			isOptions:      false,
		},
		{
			name:           "origin not in allowed list",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "https://malicious.com",
			expectAllowed:  false,
			isOptions:      false,
		},
		{
			name:           "OPTIONS preflight request",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "https://example.com",
			expectAllowed:  true,
			isOptions:      true,
		},
		{
			name:           "empty origin list denies all",
			allowedOrigins: []string{},
			requestOrigin:  "https://example.com",
			expectAllowed:  false,
			isOptions:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			handlerCalled := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			// Create CORS middleware
			middleware := CORSMiddleware(tt.allowedOrigins)(testHandler)

			// Create request
			method := "GET"
			if tt.isOptions {
				method = "OPTIONS"
			}
			req := httptest.NewRequest(method, "/test", nil)
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}
			w := httptest.NewRecorder()

			// Execute request
			middleware.ServeHTTP(w, req)

			// Check CORS headers
			if tt.expectAllowed {
				corsOrigin := w.Header().Get("Access-Control-Allow-Origin")
				if corsOrigin != tt.requestOrigin {
					t.Errorf("Expected Access-Control-Allow-Origin %s, got %s", tt.requestOrigin, corsOrigin)
				}
			} else {
				corsOrigin := w.Header().Get("Access-Control-Allow-Origin")
				if corsOrigin != "" {
					t.Errorf("Expected no Access-Control-Allow-Origin header, got %s", corsOrigin)
				}
			}

			// Check other CORS headers are always set
			if w.Header().Get("Access-Control-Allow-Methods") == "" {
				t.Error("Access-Control-Allow-Methods header not set")
			}
			if w.Header().Get("Access-Control-Allow-Headers") == "" {
				t.Error("Access-Control-Allow-Headers header not set")
			}

			// For OPTIONS requests, should return 200 and not call handler
			if tt.isOptions {
				if w.Code != http.StatusOK {
					t.Errorf("Expected status %d for OPTIONS request, got %d", http.StatusOK, w.Code)
				}
				if handlerCalled {
					t.Error("Handler should not be called for OPTIONS preflight request")
				}
			} else {
				// For non-OPTIONS requests, handler should be called
				if !handlerCalled {
					t.Error("Handler should be called for non-OPTIONS request")
				}
			}
		})
	}
}

// TestGetRequestID tests the GetRequestID helper function
func TestGetRequestID(t *testing.T) {
	tests := []struct {
		name       string
		setupCtx   func() context.Context
		expectedID string
	}{
		{
			name: "returns request ID from context",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), requestIDKey, "test-request-id")
			},
			expectedID: "test-request-id",
		},
		{
			name: "returns empty string when not in context",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectedID: "",
		},
		{
			name: "returns empty string when wrong type in context",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), requestIDKey, 123)
			},
			expectedID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result := GetRequestID(ctx)
			if result != tt.expectedID {
				t.Errorf("Expected %s, got %s", tt.expectedID, result)
			}
		})
	}
}

// TestGetSessionID tests the GetSessionID helper function
func TestGetSessionID(t *testing.T) {
	tests := []struct {
		name       string
		setupCtx   func() context.Context
		expectedID string
	}{
		{
			name: "returns session ID from context",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), sessionKey, "test-session-id")
			},
			expectedID: "test-session-id",
		},
		{
			name: "returns empty string when not in context",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectedID: "",
		},
		{
			name: "returns empty string when wrong type in context",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), sessionKey, 456)
			},
			expectedID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			result := GetSessionID(ctx)
			if result != tt.expectedID {
				t.Errorf("Expected %s, got %s", tt.expectedID, result)
			}
		})
	}
}

// TestGetClientIPMiddleware tests the getClientIP helper function
func TestGetClientIPMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		setupReq   func() *http.Request
		expectedIP string
	}{
		{
			name: "extracts IP from X-Forwarded-For header",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
				req.RemoteAddr = "127.0.0.1:8080"
				return req
			},
			expectedIP: "192.168.1.1",
		},
		{
			name: "extracts IP from X-Real-IP header",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Real-IP", "192.168.1.2")
				req.RemoteAddr = "127.0.0.1:8080"
				return req
			},
			expectedIP: "192.168.1.2",
		},
		{
			name: "falls back to RemoteAddr",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "192.168.1.3:9090"
				return req
			},
			expectedIP: "192.168.1.3",
		},
		{
			name: "handles malformed RemoteAddr",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "invalid-address"
				return req
			},
			expectedIP: "invalid-address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			result := getClientIP(req)
			if result != tt.expectedIP {
				t.Errorf("Expected %s, got %s", tt.expectedIP, result)
			}
		})
	}
}

// TestMiddlewareChain tests the complete middleware chain integration
func TestMiddlewareChain(t *testing.T) {
	// Set up logging capture
	var logBuffer strings.Builder
	originalOutput := logrus.StandardLogger().Out
	logrus.SetOutput(&logBuffer)
	defer logrus.SetOutput(originalOutput)
	logrus.SetLevel(logrus.DebugLevel)

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify context contains both request ID and logger
		requestID := GetRequestID(r.Context())
		if requestID == "" {
			t.Error("Request ID not found in handler context")
		}

		if logger, ok := r.Context().Value("logger").(*logrus.Entry); ok {
			logger.Info("handler executed")
		} else {
			t.Error("Logger not found in handler context")
		}

		w.WriteHeader(http.StatusOK)
	})

	// Create complete middleware chain
	allowedOrigins := []string{"https://example.com"}
	chain := RequestIDMiddleware(
		LoggingMiddleware(
			RecoveryMiddleware(
				CORSMiddleware(allowedOrigins)(testHandler))))

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	// Execute request
	chain.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify headers
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID header not set")
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Error("CORS headers not set correctly")
	}

	// Verify logs contain request correlation
	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, "processing request") {
		t.Error("Request processing log not found")
	}
	if !strings.Contains(logOutput, "request completed") {
		t.Error("Request completion log not found")
	}
	if !strings.Contains(logOutput, "handler executed") {
		t.Error("Handler execution log not found")
	}
}
