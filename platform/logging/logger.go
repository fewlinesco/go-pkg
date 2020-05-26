package logging

import (
	"log"
	"os"
)

// NewDefaultLogger creates a new logger with a default configuration
func NewDefaultLogger() *log.Logger {
	return log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
}
