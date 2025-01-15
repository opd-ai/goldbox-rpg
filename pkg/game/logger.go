package game

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "[GAME] ", log.LstdFlags)

// SetLogger allows changing the default logger
func SetLogger(l *log.Logger) {
	logger = l
}
