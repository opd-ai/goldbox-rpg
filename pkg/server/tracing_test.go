package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewTraceContext(t *testing.T) {
	ctx := context.Background()
	correlationID := "trace-123"
	sessionID := "session-456"

	ctx = WithCorrelationID(ctx, correlationID)
	ctx = context.WithValue(ctx, sessionKey, sessionID)

	trace := NewTraceContext(ctx)

	assert.Equal(t, correlationID, trace.CorrelationID)
	assert.Equal(t, sessionID, trace.SessionID)
	assert.False(t, trace.StartTime.IsZero())
}

func TestGetLogger(t *testing.T) {
	t.Run("returns logger from context", func(t *testing.T) {
		ctx := context.Background()
		expectedLogger := logrus.WithField("test", "value")
		ctx = WithLogger(ctx, expectedLogger)

		logger := GetLogger(ctx)
		assert.NotNil(t, logger)
		// Verify it's the same logger by checking the data field
		assert.Equal(t, expectedLogger.Data["test"], logger.Data["test"])
	})

	t.Run("returns fallback logger when not in context", func(t *testing.T) {
		ctx := context.Background()
		logger := GetLogger(ctx)
		assert.NotNil(t, logger, "should return fallback logger")
	})
}

func TestWithLogger(t *testing.T) {
	ctx := context.Background()
	logger := logrus.WithField("custom", "logger")

	ctx = WithLogger(ctx, logger)
	retrieved := GetLogger(ctx)

	assert.Equal(t, logger.Data["custom"], retrieved.Data["custom"])
}

func TestLogWithTrace(t *testing.T) {
	// Note: This test verifies the function doesn't panic and processes fields correctly
	// Actual log output verification would require capturing logrus output

	ctx := context.Background()
	correlationID := "log-trace-123"
	sessionID := "session-789"

	ctx = WithCorrelationID(ctx, correlationID)
	ctx = context.WithValue(ctx, sessionKey, sessionID)

	// Create a logger for the context
	logger := logrus.WithField("test", "logger")
	ctx = WithLogger(ctx, logger)

	// This should not panic and should add correlation_id and session_id
	fields := logrus.Fields{
		"operation": "test_operation",
		"status":    "success",
	}

	// Should not panic
	assert.NotPanics(t, func() {
		LogWithTrace(ctx, logrus.InfoLevel, "test message", fields)
	})
}

func TestLogWithTrace_FieldPrecedence(t *testing.T) {
	// Test that existing correlation_id in fields is not overwritten
	ctx := context.Background()
	ctx = WithCorrelationID(ctx, "context-correlation-id")

	logger := logrus.WithField("base", "field")
	ctx = WithLogger(ctx, logger)

	fields := logrus.Fields{
		"correlation_id": "explicit-correlation-id",
		"custom":         "value",
	}

	// Should not panic - the explicit correlation_id should be preserved
	assert.NotPanics(t, func() {
		LogWithTrace(ctx, logrus.DebugLevel, "test with explicit correlation_id", fields)
	})

	// The fields map should still have the explicit value
	assert.Equal(t, "explicit-correlation-id", fields["correlation_id"])
}

func TestTraceOperation_Success(t *testing.T) {
	ctx := context.Background()
	correlationID := "operation-success-123"
	ctx = WithCorrelationID(ctx, correlationID)

	logger := logrus.WithField("correlation_id", correlationID)
	ctx = WithLogger(ctx, logger)

	operationCalled := false
	operation := func() error {
		operationCalled = true
		time.Sleep(10 * time.Millisecond) // Simulate some work
		return nil
	}

	err := TraceOperation(ctx, "test_operation", operation)

	assert.NoError(t, err)
	assert.True(t, operationCalled, "operation should have been called")
}

func TestTraceOperation_Failure(t *testing.T) {
	ctx := context.Background()
	correlationID := "operation-failure-123"
	ctx = WithCorrelationID(ctx, correlationID)

	logger := logrus.WithField("correlation_id", correlationID)
	ctx = WithLogger(ctx, logger)

	expectedError := errors.New("operation failed")
	operation := func() error {
		return expectedError
	}

	err := TraceOperation(ctx, "test_operation", operation)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestAddTraceFields(t *testing.T) {
	t.Run("adds correlation ID and session ID", func(t *testing.T) {
		ctx := context.Background()
		correlationID := "trace-field-123"
		sessionID := "session-456"

		ctx = WithCorrelationID(ctx, correlationID)
		ctx = context.WithValue(ctx, sessionKey, sessionID)

		logger := logrus.NewEntry(logrus.StandardLogger())
		enhancedLogger := AddTraceFields(ctx, logger)

		assert.NotNil(t, enhancedLogger)
		assert.Equal(t, correlationID, enhancedLogger.Data["correlation_id"])
		assert.Equal(t, sessionID, enhancedLogger.Data["session_id"])
	})

	t.Run("returns original logger when no trace fields", func(t *testing.T) {
		ctx := context.Background()
		logger := logrus.NewEntry(logrus.StandardLogger())

		enhancedLogger := AddTraceFields(ctx, logger)

		assert.NotNil(t, enhancedLogger)
		// Should be the same logger instance since no fields were added
		assert.Equal(t, logger, enhancedLogger)
	})

	t.Run("adds only correlation ID when session ID is missing", func(t *testing.T) {
		ctx := context.Background()
		correlationID := "trace-only-correlation"
		ctx = WithCorrelationID(ctx, correlationID)

		logger := logrus.NewEntry(logrus.StandardLogger())
		enhancedLogger := AddTraceFields(ctx, logger)

		assert.NotNil(t, enhancedLogger)
		assert.Equal(t, correlationID, enhancedLogger.Data["correlation_id"])
		assert.Nil(t, enhancedLogger.Data["session_id"])
	})
}

func TestMeasureDuration(t *testing.T) {
	ctx := context.Background()
	correlationID := "duration-test-123"
	ctx = WithCorrelationID(ctx, correlationID)

	logger := logrus.WithField("correlation_id", correlationID)
	ctx = WithLogger(ctx, logger)

	done := MeasureDuration(ctx, "test_operation")

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// Should not panic when called
	assert.NotPanics(t, func() {
		done()
	})
}

func TestMeasureDuration_MultipleCalls(t *testing.T) {
	ctx := context.Background()
	logger := logrus.NewEntry(logrus.StandardLogger())
	ctx = WithLogger(ctx, logger)

	// Create multiple duration measurers
	done1 := MeasureDuration(ctx, "operation_1")
	done2 := MeasureDuration(ctx, "operation_2")

	time.Sleep(5 * time.Millisecond)
	done1()

	time.Sleep(5 * time.Millisecond)
	done2()

	// Both should complete without panic
}

func TestTraceContext_Fields(t *testing.T) {
	ctx := context.Background()
	correlationID := "context-fields-123"
	sessionID := "session-789"

	ctx = WithCorrelationID(ctx, correlationID)
	ctx = context.WithValue(ctx, sessionKey, sessionID)

	trace := NewTraceContext(ctx)
	trace.Method = "POST"
	trace.Path = "/rpc"

	assert.Equal(t, correlationID, trace.CorrelationID)
	assert.Equal(t, sessionID, trace.SessionID)
	assert.Equal(t, "POST", trace.Method)
	assert.Equal(t, "/rpc", trace.Path)
	assert.False(t, trace.StartTime.IsZero())
}

func TestTracingIntegration(t *testing.T) {
	// Integration test: Verify tracing works end-to-end with correlation IDs
	ctx := context.Background()
	correlationID := "integration-trace-123"
	sessionID := "session-integration-456"

	// Set up context
	ctx = WithCorrelationID(ctx, correlationID)
	ctx = context.WithValue(ctx, sessionKey, sessionID)

	logger := logrus.WithFields(logrus.Fields{
		"correlation_id": correlationID,
		"session_id":     sessionID,
	})
	ctx = WithLogger(ctx, logger)

	// Create trace context
	trace := NewTraceContext(ctx)
	assert.Equal(t, correlationID, trace.CorrelationID)
	assert.Equal(t, sessionID, trace.SessionID)

	// Use TraceOperation
	operationRan := false
	err := TraceOperation(ctx, "integration_test", func() error {
		operationRan = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, operationRan)

	// Verify AddTraceFields
	enhancedLogger := AddTraceFields(ctx, logrus.NewEntry(logrus.StandardLogger()))
	assert.Equal(t, correlationID, enhancedLogger.Data["correlation_id"])
	assert.Equal(t, sessionID, enhancedLogger.Data["session_id"])
}
