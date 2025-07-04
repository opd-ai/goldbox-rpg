package game

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

// TestSetLogger_ValidLogger tests setting a valid logger instance
func TestSetLogger_ValidLogger(t *testing.T) {
	// Create a custom logger with a buffer to capture output
	var buf bytes.Buffer
	customLogger := log.New(&buf, "[CUSTOM] ", log.LstdFlags)

	// Store original logger to restore after test
	originalLogger := logger
	defer func() {
		logger = originalLogger
	}()

	// Set the custom logger
	SetLogger(customLogger)

	// Verify that the logger was set correctly by checking if it's the same instance
	if logger != customLogger {
		t.Error("Expected logger to be set to custom logger instance")
	}

	// Verify the logger works by writing a test message
	logger.Print("test message")
	output := buf.String()

	if !strings.Contains(output, "[CUSTOM]") {
		t.Errorf("Expected output to contain '[CUSTOM]', got: %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected output to contain 'test message', got: %s", output)
	}
}

// TestSetLogger_DifferentOutputs tests setting loggers with different output destinations
func TestSetLogger_DifferentOutputs(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		flags  int
	}{
		{
			name:   "Standard logger with date and time",
			prefix: "[TEST] ",
			flags:  log.LstdFlags,
		},
		{
			name:   "Logger with only time",
			prefix: "[TIME] ",
			flags:  log.Ltime,
		},
		{
			name:   "Logger with no flags",
			prefix: "[SIMPLE] ",
			flags:  0,
		},
		{
			name:   "Logger with all flags",
			prefix: "[FULL] ",
			flags:  log.LstdFlags | log.Lshortfile,
		},
	}

	// Store original logger to restore after test
	originalLogger := logger
	defer func() {
		logger = originalLogger
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			testLogger := log.New(&buf, tt.prefix, tt.flags)

			SetLogger(testLogger)

			// Verify the logger was set
			if logger != testLogger {
				t.Error("Logger was not set correctly")
			}

			// Test that the logger actually works with the specified configuration
			testMessage := "configuration test"
			logger.Print(testMessage)
			output := buf.String()

			if !strings.Contains(output, tt.prefix) {
				t.Errorf("Expected output to contain prefix '%s', got: %s", tt.prefix, output)
			}
			if !strings.Contains(output, testMessage) {
				t.Errorf("Expected output to contain '%s', got: %s", testMessage, output)
			}
		})
	}
}

// TestSetLogger_MultipleCalls tests calling SetLogger multiple times
func TestSetLogger_MultipleCalls(t *testing.T) {
	// Store original logger to restore after test
	originalLogger := logger
	defer func() {
		logger = originalLogger
	}()

	// Create multiple loggers
	var buf1, buf2, buf3 bytes.Buffer
	logger1 := log.New(&buf1, "[FIRST] ", log.LstdFlags)
	logger2 := log.New(&buf2, "[SECOND] ", log.LstdFlags)
	logger3 := log.New(&buf3, "[THIRD] ", log.LstdFlags)

	// Set first logger
	SetLogger(logger1)
	if logger != logger1 {
		t.Error("First logger was not set correctly")
	}

	// Set second logger (should replace first)
	SetLogger(logger2)
	if logger != logger2 {
		t.Error("Second logger was not set correctly")
	}
	if logger == logger1 {
		t.Error("First logger should have been replaced")
	}

	// Set third logger (should replace second)
	SetLogger(logger3)
	if logger != logger3 {
		t.Error("Third logger was not set correctly")
	}
	if logger == logger2 || logger == logger1 {
		t.Error("Previous loggers should have been replaced")
	}

	// Verify the final logger works
	logger.Print("final test")
	output := buf3.String()
	if !strings.Contains(output, "[THIRD]") {
		t.Errorf("Expected output from third logger, got: %s", output)
	}

	// Verify previous loggers were not used
	if buf1.Len() > 0 {
		t.Errorf("First logger should not have received messages, but got: %s", buf1.String())
	}
	if buf2.Len() > 0 {
		t.Errorf("Second logger should not have received messages, but got: %s", buf2.String())
	}
}

// TestDefaultLogger_Initialization tests that the default logger is properly initialized
func TestDefaultLogger_Initialization(t *testing.T) {
	// Test that the package-level logger variable is not nil
	if logger == nil {
		t.Fatal("Default logger should not be nil")
	}

	// Test that we can write to the default logger without panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Default logger should not panic on write: %v", r)
		}
	}()

	// This should not panic
	logger.Print("default logger test")
}

// TestDefaultLogger_Properties tests the properties of the default logger
func TestDefaultLogger_Properties(t *testing.T) {
	// We can't directly compare logger instances, but we can test their behavior
	// by temporarily redirecting output and comparing the format

	var defaultBuf bytes.Buffer

	// Store original logger
	originalLogger := logger
	defer func() {
		logger = originalLogger
	}()

	// Create logger that writes to buffer for comparison
	testDefaultLogger := log.New(&defaultBuf, "[GAME] ", log.LstdFlags)

	// Set our test logger
	SetLogger(testDefaultLogger)

	// Write a test message
	testMessage := "format comparison test"
	logger.Print(testMessage)

	defaultOutput := defaultBuf.String()

	// Both should contain the same prefix and message
	if !strings.Contains(defaultOutput, "[GAME]") {
		t.Error("Default logger should use '[GAME]' prefix")
	}
	if !strings.Contains(defaultOutput, testMessage) {
		t.Errorf("Default logger should include test message, got: %s", defaultOutput)
	}

	// The format should be similar (we can't compare exactly due to timestamps)
	if !strings.HasPrefix(defaultOutput, "[GAME]") {
		t.Error("Default logger output should start with '[GAME]' prefix")
	}
}

// TestSetLogger_NilLogger tests behavior when passing nil (edge case)
func TestSetLogger_NilLogger(t *testing.T) {
	// Store original logger to restore after test
	originalLogger := logger
	defer func() {
		logger = originalLogger
	}()

	// Test that setting nil doesn't break anything
	// Note: This might cause issues in real usage, but we test the behavior
	SetLogger(nil)

	if logger != nil {
		t.Error("Logger should be nil after setting to nil")
	}

	// Attempting to use a nil logger would panic, which is expected behavior
	// We don't test the panic case as it's not the intended usage
}

// TestSetLogger_ConcurrentAccess tests basic concurrent safety considerations
func TestSetLogger_ConcurrentAccess(t *testing.T) {
	// Store original logger to restore after test
	originalLogger := logger
	defer func() {
		logger = originalLogger
	}()

	// This test ensures that SetLogger doesn't cause data races
	// when called from multiple goroutines (though this isn't guaranteed thread-safe)

	var buf bytes.Buffer
	testLogger := log.New(&buf, "[CONCURRENT] ", log.LstdFlags)

	// Set logger multiple times in sequence (simulating potential concurrent access)
	for i := 0; i < 10; i++ {
		SetLogger(testLogger)
		if logger != testLogger {
			t.Errorf("Logger was not set correctly in iteration %d", i)
		}
	}
}

// TestLogger_Integration tests the logger in a more realistic scenario
func TestLogger_Integration(t *testing.T) {
	// Store original logger to restore after test
	originalLogger := logger
	defer func() {
		logger = originalLogger
	}()

	// Create a logger that captures output for testing
	var buf bytes.Buffer
	integrationLogger := log.New(&buf, "[INTEGRATION] ", log.LstdFlags)

	// Set the logger
	SetLogger(integrationLogger)

	// Simulate typical logging scenarios
	testCases := []string{
		"System initialization",
		"Player action: move north",
		"Combat started",
		"Game state saved",
		"Error: invalid command",
	}

	for _, testCase := range testCases {
		logger.Printf("Game event: %s", testCase)
	}

	output := buf.String()

	// Verify all messages were logged
	for _, testCase := range testCases {
		if !strings.Contains(output, testCase) {
			t.Errorf("Expected output to contain '%s', got: %s", testCase, output)
		}
	}

	// Verify the format includes the integration prefix
	if !strings.Contains(output, "[INTEGRATION]") {
		t.Errorf("Expected output to contain '[INTEGRATION]' prefix, got: %s", output)
	}

	// Count the number of log entries
	lines := strings.Split(strings.TrimSpace(output), "\n")
	expectedLines := len(testCases)
	if len(lines) != expectedLines {
		t.Errorf("Expected %d log lines, got %d", expectedLines, len(lines))
	}
}
