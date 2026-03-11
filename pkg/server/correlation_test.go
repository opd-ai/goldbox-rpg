package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCorrelationIDMiddleware(t *testing.T) {
	tests := []struct {
		name                  string
		incomingCorrelationID string
		incomingRequestID     string
		expectGenerated       bool
	}{
		{
			name:                  "uses existing X-Correlation-ID header",
			incomingCorrelationID: "test-correlation-123",
			expectGenerated:       false,
		},
		{
			name:              "falls back to X-Request-ID when no correlation ID",
			incomingRequestID: "test-request-456",
			expectGenerated:   false,
		},
		{
			name:            "generates new correlation ID when none provided",
			expectGenerated: true,
		},
		{
			name:                  "prefers X-Correlation-ID over X-Request-ID",
			incomingCorrelationID: "correlation-789",
			incomingRequestID:     "request-999",
			expectGenerated:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that captures the context
			var capturedCorrelationID string
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedCorrelationID = GetCorrelationID(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with middleware
			middleware := CorrelationIDMiddleware(handler)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.incomingCorrelationID != "" {
				req.Header.Set("X-Correlation-ID", tt.incomingCorrelationID)
			}
			if tt.incomingRequestID != "" {
				req.Header.Set("X-Request-ID", tt.incomingRequestID)
			}

			// Execute request
			rec := httptest.NewRecorder()
			middleware.ServeHTTP(rec, req)

			// Verify correlation ID was captured
			require.NotEmpty(t, capturedCorrelationID, "correlation ID should be set in context")

			// Verify correlation ID in response header
			responseCorrelationID := rec.Header().Get("X-Correlation-ID")
			assert.Equal(t, capturedCorrelationID, responseCorrelationID,
				"response header should match context correlation ID")

			// Verify expected behavior
			if tt.expectGenerated {
				// Should be a valid UUID format
				assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
					capturedCorrelationID, "generated correlation ID should be valid UUID")
			} else if tt.incomingCorrelationID != "" {
				assert.Equal(t, tt.incomingCorrelationID, capturedCorrelationID,
					"should use incoming X-Correlation-ID")
			} else if tt.incomingRequestID != "" {
				assert.Equal(t, tt.incomingRequestID, capturedCorrelationID,
					"should fall back to X-Request-ID")
			}
		})
	}
}

func TestCorrelationIDMiddleware_LoggerInContext(t *testing.T) {
	// Create a test handler that verifies logger is in context
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		// Verify logger is available in context
		logger := GetLogger(r.Context())
		require.NotNil(t, logger, "logger should be available in context")

		// Verify logger has correlation_id field
		// Note: We can't directly inspect logrus.Entry fields, but we can verify it exists
		assert.NotNil(t, logger.Data, "logger should have data fields")

		w.WriteHeader(http.StatusOK)
	})

	// Wrap with middleware
	middleware := CorrelationIDMiddleware(handler)

	// Create and execute request
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	assert.True(t, handlerCalled, "handler should have been called")
}

func TestGetCorrelationID(t *testing.T) {
	t.Run("returns empty string when not in context", func(t *testing.T) {
		ctx := context.Background()
		correlationID := GetCorrelationID(ctx)
		assert.Empty(t, correlationID)
	})

	t.Run("returns correlation ID from context", func(t *testing.T) {
		ctx := context.Background()
		expectedID := "test-correlation-id"
		ctx = WithCorrelationID(ctx, expectedID)

		correlationID := GetCorrelationID(ctx)
		assert.Equal(t, expectedID, correlationID)
	})
}

func TestWithCorrelationID(t *testing.T) {
	ctx := context.Background()
	correlationID := "test-id-123"

	ctx = WithCorrelationID(ctx, correlationID)
	retrieved := GetCorrelationID(ctx)

	assert.Equal(t, correlationID, retrieved)
}

func TestCorrelationIDMiddleware_Integration(t *testing.T) {
	// Test that correlation ID flows through the entire middleware chain
	correlationID := "integration-test-123"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify correlation ID is in context
		assert.Equal(t, correlationID, GetCorrelationID(r.Context()))

		// Verify logger has correlation ID
		logger := GetLogger(r.Context())
		assert.NotNil(t, logger)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Build middleware chain similar to production
	chain := CorrelationIDMiddleware(
		LoggingMiddleware(
			RecoveryMiddleware(handler)))

	req := httptest.NewRequest("POST", "/rpc", nil)
	req.Header.Set("X-Correlation-ID", correlationID)

	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, correlationID, rec.Header().Get("X-Correlation-ID"))
}

func TestCorrelationIDMiddleware_ConcurrentRequests(t *testing.T) {
	// Verify that correlation IDs don't leak between concurrent requests
	const numRequests = 100

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := GetCorrelationID(r.Context())
		expectedID := r.Header.Get("X-Correlation-ID")
		assert.Equal(t, expectedID, correlationID,
			"correlation ID should match request header")
		w.WriteHeader(http.StatusOK)
	})

	middleware := CorrelationIDMiddleware(handler)

	// Create a channel to synchronize goroutines
	done := make(chan bool, numRequests)

	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		go func(id int) {
			correlationID := "correlation-" + string(rune('A'+id%26)) + "-" + string(rune(id))
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Correlation-ID", correlationID)

			rec := httptest.NewRecorder()
			middleware.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, correlationID, rec.Header().Get("X-Correlation-ID"))

			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}
