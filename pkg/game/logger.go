package game

import (
	"log"
	"os"
)

// logger is the global game logger instance that writes log messages to standard output.
// It prefixes all messages with "[GAME]" and includes standard logging flags (date and time).
// The logger is configured to write to os.Stdout and uses the standard Go log package.
// Related: pkg/game package logging functions
var logger = log.New(os.Stdout, "[GAME] ", log.LstdFlags)

// SetLogger allows changing the default logger
// SetLogger sets the package-level logger instance used for logging throughout the game.
// It allows injection of a custom logger for different logging implementations.
//
// Parameters:
//   - l: pointer to a log.Logger instance to be used for logging. Must not be nil.
//
// Related:
//   - logger (package-level variable)
func SetLogger(l *log.Logger) {
	logger = l
}
